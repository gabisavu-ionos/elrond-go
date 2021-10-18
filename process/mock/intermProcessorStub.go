package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

// IntermediateTransactionHandlerStub -
type IntermediateTransactionHandlerStub struct {
	AddIntermediateTransactionsCalled        func(txs []data.TransactionHandler) error
	GetProcessedResultsCalled                func() map[uint32][]*process.TxInfo
	GetNumOfCrossInterMbsAndTxsCalled        func() (int, int)
	CreateAllInterMiniBlocksCalled           func() []*block.MiniBlock
	VerifyInterMiniBlocksCalled              func(body *block.Body) error
	SaveCurrentIntermediateTxToStorageCalled func() error
	CreateBlockStartedCalled                 func()
	CreateMarshalizedDataCalled              func(txHashes [][]byte) ([][]byte, error)
	GetAllCurrentFinishedTxsCalled           func() map[string]data.TransactionHandler
	RemoveProcessedResultsCalled             func()
	InitProcessedResultsCalled               func()
	intermediateTransactions                 []data.TransactionHandler
}

// RemoveProcessedResults -
func (ith *IntermediateTransactionHandlerStub) RemoveProcessedResults() {
	if ith.RemoveProcessedResultsCalled != nil {
		ith.RemoveProcessedResultsCalled()
	}
}

// InitProcessedResults -
func (ith *IntermediateTransactionHandlerStub) InitProcessedResults() {
	if ith.InitProcessedResultsCalled != nil {
		ith.InitProcessedResultsCalled()
	}
}

// CreateMarshalizedData -
func (ith *IntermediateTransactionHandlerStub) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
	if ith.CreateMarshalizedDataCalled == nil {
		return nil, nil
	}
	return ith.CreateMarshalizedDataCalled(txHashes)
}

// AddIntermediateTransactions -
func (ith *IntermediateTransactionHandlerStub) AddIntermediateTransactions(txs []data.TransactionHandler) error {
	if ith.AddIntermediateTransactionsCalled == nil {
		ith.intermediateTransactions = append(ith.intermediateTransactions, txs...)
		return nil
	}
	return ith.AddIntermediateTransactionsCalled(txs)
}

// GetAllIntermediateTxsForTxHash -
func (ith *IntermediateTransactionHandlerStub) GetProcessedResults() map[uint32][]*process.TxInfo {
	if ith.GetProcessedResultsCalled != nil {
		return ith.GetProcessedResultsCalled()
	}
	return nil
}

// GetIntermediateTransactions -
func (ith *IntermediateTransactionHandlerStub) GetIntermediateTransactions() []data.TransactionHandler {
	return ith.intermediateTransactions
}

// GetNumOfCrossInterMbsAndTxs -
func (ith *IntermediateTransactionHandlerStub) GetNumOfCrossInterMbsAndTxs() (int, int) {
	if ith.GetNumOfCrossInterMbsAndTxsCalled == nil {
		return 0, 0
	}
	return ith.GetNumOfCrossInterMbsAndTxsCalled()
}

// CreateAllInterMiniBlocks -
func (ith *IntermediateTransactionHandlerStub) CreateAllInterMiniBlocks() []*block.MiniBlock {
	if ith.CreateAllInterMiniBlocksCalled == nil {
		return nil
	}
	return ith.CreateAllInterMiniBlocksCalled()
}

// VerifyInterMiniBlocks -
func (ith *IntermediateTransactionHandlerStub) VerifyInterMiniBlocks(body *block.Body) error {
	if ith.VerifyInterMiniBlocksCalled == nil {
		return nil
	}
	return ith.VerifyInterMiniBlocksCalled(body)
}

// SaveCurrentIntermediateTxToStorage -
func (ith *IntermediateTransactionHandlerStub) SaveCurrentIntermediateTxToStorage() error {
	if ith.SaveCurrentIntermediateTxToStorageCalled == nil {
		return nil
	}
	return ith.SaveCurrentIntermediateTxToStorageCalled()
}

// CreateBlockStarted -
func (ith *IntermediateTransactionHandlerStub) CreateBlockStarted() {
	if ith.CreateBlockStartedCalled != nil {
		ith.CreateBlockStartedCalled()
	}
}

// GetAllCurrentFinishedTxs -
func (ith *IntermediateTransactionHandlerStub) GetAllCurrentFinishedTxs() map[string]data.TransactionHandler {
	if ith.GetAllCurrentFinishedTxsCalled != nil {
		return ith.GetAllCurrentFinishedTxsCalled()
	}
	return nil
}

// GetCreatedInShardMiniBlock -
func (ith *IntermediateTransactionHandlerStub) GetCreatedInShardMiniBlock() *block.MiniBlock {
	return &block.MiniBlock{}
}

// IsInterfaceNil returns true if there is no value under the interface
func (ith *IntermediateTransactionHandlerStub) IsInterfaceNil() bool {
	return ith == nil
}
