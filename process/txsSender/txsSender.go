package txsSender

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/accumulator"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/partitioning"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("txsSender")
var numSecondsBetweenPrints = 20

// SendTransactionsPipe is the pipe used for sending new transactions
const SendTransactionsPipe = "send transactions pipe"

type txsSender struct {
	marshaller       marshal.Marshalizer
	shardCoordinator storage.ShardCoordinator
	networkMessenger NetworkMessenger

	ctx                      context.Context
	cancelFunc               context.CancelFunc
	txSentCounter            uint32
	txAccumulator            core.Accumulator
	currentSendingGoRoutines int32
}

// ArgsTxsSenderWithAccumulator is a holder struct for all necessary arguments to create a NewTxsSenderWithAccumulator
type ArgsTxsSenderWithAccumulator struct {
	Marshaller        marshal.Marshalizer
	ShardCoordinator  storage.ShardCoordinator
	NetworkMessenger  NetworkMessenger
	AccumulatorConfig config.TxAccumulatorConfig
}

// NewTxsSenderWithAccumulator creates a new instance of TxsSenderHandler, which initializes internally a accumulator.NewTimeAccumulator
func NewTxsSenderWithAccumulator(args ArgsTxsSenderWithAccumulator) (process.TxsSenderHandler, error) {
	txAccumulator, err := accumulator.NewTimeAccumulator(
		time.Duration(args.AccumulatorConfig.MaxAllowedTimeInMilliseconds)*time.Millisecond,
		time.Duration(args.AccumulatorConfig.MaxDeviationTimeInMilliseconds)*time.Millisecond,
		log,
	)
	if err != nil {
		return nil, err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	ret := &txsSender{
		marshaller:               args.Marshaller,
		shardCoordinator:         args.ShardCoordinator,
		networkMessenger:         args.NetworkMessenger,
		ctx:                      ctx,
		cancelFunc:               cancelFunc,
		txAccumulator:            txAccumulator,
		txSentCounter:            0,
		currentSendingGoRoutines: 0,
	}
	go ret.sendFromTxAccumulator(ret.ctx)
	go ret.printTxSentCounter(ret.ctx)

	return ret, nil
}

// SendBulkTransactions sends the provided transactions as a bulk, optimizing transfer between nodes
func (ts *txsSender) SendBulkTransactions(txs []data.TransactionHandler) (uint64, error) {
	if len(txs) == 0 {
		return 0, process.ErrNoTxToProcess
	}

	ts.addTransactionsToSendPipe(txs)

	return uint64(len(txs)), nil
}

func (ts *txsSender) addTransactionsToSendPipe(txs []data.TransactionHandler) {
	if check.IfNil(ts.txAccumulator) {
		log.Error("node has a nil tx accumulator instance")
		return
	}

	for _, tx := range txs {
		ts.txAccumulator.AddData(tx)
	}
}

func (ts *txsSender) sendFromTxAccumulator(ctx context.Context) {
	outputChannel := ts.txAccumulator.OutputChannel()

	for {
		select {
		case objs := <-outputChannel:
			{
				if len(objs) == 0 {
					break
				}

				txs := make([]*transaction.Transaction, 0, len(objs))
				for _, obj := range objs {
					tx, ok := obj.(*transaction.Transaction)
					if !ok {
						continue
					}

					txs = append(txs, tx)
				}

				atomic.AddUint32(&ts.txSentCounter, uint32(len(txs)))

				ts.sendBulkTransactions(txs)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (ts *txsSender) sendBulkTransactions(txs []*transaction.Transaction) {
	transactionsByShards := make(map[uint32][][]byte)
	log.Trace("txsSender.sendBulkTransactions sending txs",
		"num", len(txs),
	)

	for _, tx := range txs {
		senderShardId := ts.shardCoordinator.ComputeId(tx.SndAddr)

		marshalledTx, err := ts.marshaller.Marshal(tx)
		if err != nil {
			log.Warn("txsSender.sendBulkTransactions",
				"marshaller error", err,
			)
			continue
		}

		transactionsByShards[senderShardId] = append(transactionsByShards[senderShardId], marshalledTx)
	}

	numOfSentTxs := uint64(0)
	for shardId, txsForShard := range transactionsByShards {
		err := ts.sendBulkTransactionsFromShard(txsForShard, shardId)
		if err != nil {
			log.Debug("sendBulkTransactionsFromShard", "error", err.Error())
		} else {
			numOfSentTxs += uint64(len(txsForShard))
		}
	}
}

func (ts *txsSender) sendBulkTransactionsFromShard(transactions [][]byte, senderShardId uint32) error {
	dataPacker, err := partitioning.NewSimpleDataPacker(ts.marshaller)
	if err != nil {
		return err
	}

	// the topic identifier is made of the current shard id and sender's shard id
	identifier := factory.TransactionTopic + ts.shardCoordinator.CommunicationIdentifier(senderShardId)

	packets, err := dataPacker.PackDataInChunks(transactions, common.MaxBulkTransactionSize)
	if err != nil {
		return err
	}

	atomic.AddInt32(&ts.currentSendingGoRoutines, int32(len(packets)))
	for _, buff := range packets {
		go func(bufferToSend []byte) {
			log.Trace("txsSender.sendBulkTransactionsFromShard",
				"topic", identifier,
				"size", len(bufferToSend),
			)
			err = ts.networkMessenger.BroadcastOnChannelBlocking(
				SendTransactionsPipe,
				identifier,
				bufferToSend,
			)
			if err != nil {
				log.Debug("txsSender.BroadcastOnChannelBlocking", "error", err.Error())
			}

			atomic.AddInt32(&ts.currentSendingGoRoutines, -1)
		}(buff)
	}

	return nil
}

// printTxSentCounter prints the peak transaction counter from a time frame of about 'numSecondsBetweenPrints' seconds
// if this peak value is 0 (no transaction was sent through the REST API interface), the print will not be done
// the peak counter resets after each print. There is also a total number of transactions sent to p2p
// TODO make this function testable. Refactor if necessary.
func (ts *txsSender) printTxSentCounter(ctx context.Context) {
	maxTxCounter := uint32(0)
	totalTxCounter := uint64(0)
	counterSeconds := 0

	for {
		select {
		case <-time.After(time.Second):
			txSent := atomic.SwapUint32(&ts.txSentCounter, 0)
			if txSent > maxTxCounter {
				maxTxCounter = txSent
			}
			totalTxCounter += uint64(txSent)

			counterSeconds++
			if counterSeconds > numSecondsBetweenPrints {
				counterSeconds = 0

				if maxTxCounter > 0 {
					log.Info("sent transactions on network",
						"max/sec", maxTxCounter,
						"total", totalTxCounter,
					)
				}
				maxTxCounter = 0
			}
		case <-ctx.Done():
			return
		}
	}
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ts *txsSender) IsInterfaceNil() bool {
	return ts == nil
}

// Close calls the cancel function of the background context and closes the network messenger
func (ts *txsSender) Close() error {
	ts.cancelFunc()
	return ts.networkMessenger.Close()
}
