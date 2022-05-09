package processor

import (
	"strconv"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
)

type shardProvider interface {
	ShardID() string
}

// ArgDirectConnectionInfoInterceptorProcessor is the argument for the interceptor processor used for direct connection info
type ArgDirectConnectionInfoInterceptorProcessor struct {
	PeerShardMapper process.PeerShardMapper
}

type DirectConnectionInfoInterceptorProcessor struct {
	peerShardMapper process.PeerShardMapper
}

// NewDirectConnectionInfoInterceptorProcessor creates an instance of DirectConnectionInfoInterceptorProcessor
func NewDirectConnectionInfoInterceptorProcessor(args ArgDirectConnectionInfoInterceptorProcessor) (*DirectConnectionInfoInterceptorProcessor, error) {
	if check.IfNil(args.PeerShardMapper) {
		return nil, process.ErrNilPeerShardMapper
	}

	return &DirectConnectionInfoInterceptorProcessor{
		peerShardMapper: args.PeerShardMapper,
	}, nil
}

// Validate checks if the intercepted data can be processed
// returns nil as proper validity checks are done at intercepted data level
func (processor *DirectConnectionInfoInterceptorProcessor) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save will save the intercepted validator info into peer shard mapper
func (processor *DirectConnectionInfoInterceptorProcessor) Save(data process.InterceptedData, fromConnectedPeer core.PeerID, _ string) error {
	shardDirectConnectionInfo, ok := data.(shardProvider)
	if !ok {
		log.Debug("testing---DirectConnectionInfoInterceptorProcessor save wrong type assert")
		return process.ErrWrongTypeAssertion
	}

	shardID, err := strconv.Atoi(shardDirectConnectionInfo.ShardID())
	if err != nil {
		log.Debug("testing---DirectConnectionInfoInterceptorProcessor save can't cast to string")
		return err
	}

	processor.peerShardMapper.PutPeerIdShardId(fromConnectedPeer, uint32(shardID))

	return nil
}

// RegisterHandler registers a callback function to be notified of incoming shard validator info, currently not implemented
func (processor *DirectConnectionInfoInterceptorProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
	log.Error("DirectConnectionInfoInterceptorProcessor.RegisterHandler", "error", "not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (processor *DirectConnectionInfoInterceptorProcessor) IsInterfaceNil() bool {
	return processor == nil
}
