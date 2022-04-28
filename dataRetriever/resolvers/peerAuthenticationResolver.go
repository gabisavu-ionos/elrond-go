package resolvers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const minNumOfPeerAuthentication = 5
const bytesInUint32 = 4

// ArgPeerAuthenticationResolver is the argument structure used to create a new peer authentication resolver instance
type ArgPeerAuthenticationResolver struct {
	ArgBaseResolver
	PeerAuthenticationPool               storage.Cacher
	NodesCoordinator                     dataRetriever.NodesCoordinator
	PeerShardMapper                      process.PeerShardMapper
	MaxNumOfPeerAuthenticationInResponse int
}

// peerAuthenticationResolver is a wrapper over Resolver that is specialized in resolving peer authentication requests
type peerAuthenticationResolver struct {
	*baseResolver
	messageProcessor
	peerAuthenticationPool               storage.Cacher
	nodesCoordinator                     dataRetriever.NodesCoordinator
	peerShardMapper                      process.PeerShardMapper
	maxNumOfPeerAuthenticationInResponse int
}

// NewPeerAuthenticationResolver creates a peer authentication resolver
func NewPeerAuthenticationResolver(arg ArgPeerAuthenticationResolver) (*peerAuthenticationResolver, error) {
	err := checkArgPeerAuthenticationResolver(arg)
	if err != nil {
		return nil, err
	}

	return &peerAuthenticationResolver{
		baseResolver: &baseResolver{
			TopicResolverSender: arg.SenderResolver,
		},
		messageProcessor: messageProcessor{
			marshalizer:      arg.Marshalizer,
			antifloodHandler: arg.AntifloodHandler,
			throttler:        arg.Throttler,
			topic:            arg.SenderResolver.RequestTopic(),
		},
		peerAuthenticationPool:               arg.PeerAuthenticationPool,
		nodesCoordinator:                     arg.NodesCoordinator,
		peerShardMapper:                      arg.PeerShardMapper,
		maxNumOfPeerAuthenticationInResponse: arg.MaxNumOfPeerAuthenticationInResponse,
	}, nil
}

func checkArgPeerAuthenticationResolver(arg ArgPeerAuthenticationResolver) error {
	err := checkArgBase(arg.ArgBaseResolver)
	if err != nil {
		return err
	}
	if check.IfNil(arg.PeerAuthenticationPool) {
		return dataRetriever.ErrNilPeerAuthenticationPool
	}
	if check.IfNil(arg.NodesCoordinator) {
		return dataRetriever.ErrNilNodesCoordinator
	}
	if check.IfNil(arg.PeerShardMapper) {
		return dataRetriever.ErrNilPeerShardMapper
	}
	if arg.MaxNumOfPeerAuthenticationInResponse < minNumOfPeerAuthentication {
		return dataRetriever.ErrInvalidNumOfPeerAuthentication
	}
	return nil
}

// RequestDataFromHash requests peer authentication data from other peers having input a public key hash
func (res *peerAuthenticationResolver) RequestDataFromHash(hash []byte, _ uint32) error {
	return res.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashType,
			Value: hash,
		},
		[][]byte{hash},
	)
}

// RequestDataFromChunk requests peer authentication data from other peers having input a chunk index
func (res *peerAuthenticationResolver) RequestDataFromChunk(chunkIndex uint32, epoch uint32) error {
	chunkBuffer := make([]byte, bytesInUint32)
	binary.BigEndian.PutUint32(chunkBuffer, chunkIndex)

	b := &batch.Batch{
		Data: make([][]byte, 1),
	}
	b.Data[0] = chunkBuffer

	dataBuff, err := res.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	return res.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:       dataRetriever.ChunkType,
			ChunkIndex: chunkIndex,
			Epoch:      epoch,
			Value:      dataBuff,
		},
		[][]byte{chunkBuffer},
	)
}

// RequestDataFromHashArray requests peer authentication data from other peers having input multiple public key hashes
func (res *peerAuthenticationResolver) RequestDataFromHashArray(hashes [][]byte, _ uint32) error {
	b := &batch.Batch{
		Data: hashes,
	}
	buffHashes, err := res.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	return res.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashArrayType,
			Value: buffHashes,
		},
		hashes,
	)
}

// ProcessReceivedMessage represents the callback func from the p2p.Messenger that is called each time a new message is received
// (for the topic this validator was registered to, usually a request topic)
func (res *peerAuthenticationResolver) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	err := res.canProcessMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	res.throttler.StartProcessing()
	defer res.throttler.EndProcessing()

	rd, err := res.parseReceivedMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	switch rd.Type {
	case dataRetriever.ChunkType:
		return res.resolveChunkRequest(int(rd.ChunkIndex), rd.Epoch, message.Peer())
	case dataRetriever.HashArrayType:
		return res.resolveMultipleHashesRequest(rd.Value, message.Peer())
	default:
		err = dataRetriever.ErrRequestTypeNotImplemented
	}
	if err != nil {
		err = fmt.Errorf("%w for value %s", err, logger.DisplayByteSlice(rd.Value))
	}

	return err
}

// resolveChunkRequest sends the response for a chunk request
func (res *peerAuthenticationResolver) resolveChunkRequest(chunkIndex int, epoch uint32, pid core.PeerID) error {
	log.Debug("testing---resolveChunkRequest", "chunk", chunkIndex)
	sortedPKs, err := res.getSortedValidatorsKeys(epoch)
	if err != nil {
		return err
	}
	if len(sortedPKs) == 0 {
		return nil
	}

	maxChunks := res.getMaxChunks(sortedPKs)
	pksChunk, err := res.extractChunk(sortedPKs, chunkIndex, res.maxNumOfPeerAuthenticationInResponse, maxChunks)
	if err != nil {
		return err
	}

	dataSlice, err := res.fetchPeerAuthenticationSlicesForPublicKeys(pksChunk)
	if err != nil {
		return fmt.Errorf("resolveChunkRequest error %w from chunk %d", err, chunkIndex)
	}

	return res.sendData(dataSlice, nil, chunkIndex, maxChunks, pid)
}

// getSortedValidatorsKeys returns the sorted slice of validators keys from all shards
func (res *peerAuthenticationResolver) getSortedValidatorsKeys(epoch uint32) ([][]byte, error) {
	validatorsPKsMap, err := res.nodesCoordinator.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		return nil, err
	}

	validatorsPKs := make([][]byte, 0)
	for _, shardValidators := range validatorsPKsMap {
		validatorsPKs = append(validatorsPKs, shardValidators...)
	}

	sort.Slice(validatorsPKs, func(i, j int) bool {
		return bytes.Compare(validatorsPKs[i], validatorsPKs[j]) < 0
	})

	return validatorsPKs, nil
}

// extractChunk returns the chunk from dataBuff at the specified index
func (res *peerAuthenticationResolver) extractChunk(dataBuff [][]byte, chunkIndex int, chunkSize int, maxChunks int) ([][]byte, error) {
	chunkIndexOutOfBounds := chunkIndex < 0 || chunkIndex > maxChunks
	if chunkIndexOutOfBounds {
		return nil, dataRetriever.InvalidChunkIndex
	}

	startingIndex := chunkIndex * chunkSize
	endIndex := startingIndex + chunkSize
	if endIndex > len(dataBuff) {
		endIndex = len(dataBuff)
	}
	return dataBuff[startingIndex:endIndex], nil
}

// resolveMultipleHashesRequest sends the response for multiple hashes request
func (res *peerAuthenticationResolver) resolveMultipleHashesRequest(hashesBuff []byte, pid core.PeerID) error {
	log.Debug("testing---resolveMultipleHashesRequest", "num of hashes", len(hashesBuff), "max messages allowed", res.maxNumOfPeerAuthenticationInResponse)

	b := batch.Batch{}
	err := res.marshalizer.Unmarshal(&b, hashesBuff)
	if err != nil {
		return err
	}
	hashes := b.Data

	peerAuthsForHashes, err := res.fetchPeerAuthenticationSlicesForPublicKeys(hashes)
	if err != nil {
		return fmt.Errorf("resolveMultipleHashesRequest error %w from buff %s", err, hashesBuff)
	}

	return res.sendPeerAuthsForHashes(peerAuthsForHashes, hashesBuff, pid)
}

// sendPeerAuthsForHashes sends multiple peer authentication messages for specific hashes
func (res *peerAuthenticationResolver) sendPeerAuthsForHashes(dataBuff [][]byte, hashesBuff []byte, pid core.PeerID) error {
	log.Debug("testing---sendPeerAuthsForHashes", "count data buff", len(dataBuff), "max num in response", res.maxNumOfPeerAuthenticationInResponse)

	if len(dataBuff) > res.maxNumOfPeerAuthenticationInResponse {
		return res.sendLargeDataBuff(dataBuff, hashesBuff, res.maxNumOfPeerAuthenticationInResponse, pid)
	}

	return res.sendData(dataBuff, hashesBuff, 0, 0, pid)
}

// sendLargeDataBuff splits dataBuff into chunks and sends a message for the first chunk
func (res *peerAuthenticationResolver) sendLargeDataBuff(dataBuff [][]byte, reference []byte, chunkSize int, pid core.PeerID) error {
	maxChunks := res.getMaxChunks(dataBuff)
	chunk, err := res.extractChunk(dataBuff, 0, chunkSize, maxChunks)
	if err != nil {
		return err
	}

	return res.sendData(chunk, reference, 0, maxChunks, pid)
}

// getMaxChunks returns the max num of chunks from a buffer
func (res *peerAuthenticationResolver) getMaxChunks(dataBuff [][]byte) int {
	maxChunks := len(dataBuff) / res.maxNumOfPeerAuthenticationInResponse
	if len(dataBuff)%res.maxNumOfPeerAuthenticationInResponse != 0 {
		maxChunks++
	}
	return maxChunks
}

// sendData sends a message to a peer
func (res *peerAuthenticationResolver) sendData(dataSlice [][]byte, reference []byte, chunkIndex int, maxChunks int, pid core.PeerID) error {
	b := batch.Batch{
		Data:       dataSlice,
		Reference:  reference,
		ChunkIndex: uint32(chunkIndex),
		MaxChunks:  uint32(maxChunks),
	}
	buffToSend, err := res.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	log.Debug("testing---sending peer auth", "count", len(dataSlice))
	for i := 0; i < len(dataSlice); i++ {
		log.Debug("testing---sendData", "peer auth slice", string(dataSlice[i]))
	}

	return res.Send(buffToSend, pid)
}

// fetchPeerAuthenticationSlicesForPublicKeys fetches all peer authentications for all pks
func (res *peerAuthenticationResolver) fetchPeerAuthenticationSlicesForPublicKeys(pks [][]byte) ([][]byte, error) {
	peerAuths := make([][]byte, 0)
	for _, pk := range pks {
		peerAuthForHash, _ := res.fetchPeerAuthenticationAsByteSlice(pk)
		if peerAuthForHash != nil {
			log.Debug("testing---fetchPeerAuthenticationSlicesForPublicKeys", "pk", string(pk), "peer auth bytes", string(peerAuthForHash))
			peerAuths = append(peerAuths, peerAuthForHash)
		}
	}

	if len(peerAuths) == 0 {
		return nil, dataRetriever.ErrPeerAuthNotFound
	}

	return peerAuths, nil
}

// fetchPeerAuthenticationAsByteSlice returns the value from authentication pool if exists
func (res *peerAuthenticationResolver) fetchPeerAuthenticationAsByteSlice(pk []byte) ([]byte, error) {
	pid, ok := res.peerShardMapper.GetLastKnownPeerID(pk)
	if !ok {
		return nil, dataRetriever.ErrPeerAuthNotFound
	}

	value, ok := res.peerAuthenticationPool.Peek(pid.Bytes())
	if ok {
		return res.marshalizer.Marshal(value)
	}

	return nil, dataRetriever.ErrPeerAuthNotFound
}

// IsInterfaceNil returns true if there is no value under the interface
func (res *peerAuthenticationResolver) IsInterfaceNil() bool {
	return res == nil
}
