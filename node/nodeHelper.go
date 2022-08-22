package node

import (
	"errors"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/factory"
	nodeDisabled "github.com/ElrondNetwork/elrond-go/node/disabled"
	"github.com/ElrondNetwork/elrond-go/node/nodeDebugFactory"
	procFactory "github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/blackList"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-vm-common/builtInFunctions"
)

// prepareOpenTopics will set to the anti flood handler the topics for which
// the node can receive messages from others than validators
func prepareOpenTopics(
	antiflood factory.P2PAntifloodHandler,
	shardCoordinator sharding.Coordinator,
) {
	selfID := shardCoordinator.SelfId()
	selfShardHeartbeatV2Topic := common.HeartbeatV2Topic + core.CommunicationIdentifierBetweenShards(selfID, selfID)
	if selfID == core.MetachainShardId {
		antiflood.SetTopicsForAll(common.HeartbeatTopic, common.PeerAuthenticationTopic, selfShardHeartbeatV2Topic, common.ConnectionTopic)
		return
	}

	selfShardTxTopic := procFactory.TransactionTopic + core.CommunicationIdentifierBetweenShards(selfID, selfID)
	antiflood.SetTopicsForAll(common.HeartbeatTopic, common.PeerAuthenticationTopic, selfShardHeartbeatV2Topic, common.ConnectionTopic, selfShardTxTopic)
}

// CreateNode is the node factory
func CreateNode(
	config *config.Config,
	bootstrapComponents factory.BootstrapComponentsHandler,
	coreComponents factory.CoreComponentsHandler,
	cryptoComponents factory.CryptoComponentsHandler,
	dataComponents factory.DataComponentsHandler,
	networkComponents factory.NetworkComponentsHandler,
	processComponents factory.ProcessComponentsHandler,
	stateComponents factory.StateComponentsHandler,
	statusComponents factory.StatusComponentsHandler,
	heartbeatComponents factory.HeartbeatComponentsHandler,
	heartbeatV2Components factory.HeartbeatV2ComponentsHandler,
	consensusComponents factory.ConsensusComponentsHandler,
	epochConfig config.EpochConfig,
	bootstrapRoundIndex uint64,
	isInImportMode bool,
) (*Node, error) {
	defer func() {
		log.Debug("REMOVE_ME CreateNode defer")
	}()
	log.Debug("REMOVE_ME start prepareOpenTopics")
	prepareOpenTopics(networkComponents.InputAntiFloodHandler(), processComponents.ShardCoordinator())
	log.Debug("REMOVE_ME finish prepareOpenTopics")

	log.Debug("REMOVE_ME start peerDenialEvaluator")
	peerDenialEvaluator, err := blackList.NewPeerDenialEvaluator(
		networkComponents.PeerBlackListHandler(),
		networkComponents.PubKeyCacher(),
		processComponents.PeerShardMapper(),
	)
	if err != nil {
		log.Debug("REMOVE_ME finish with error peerDenialEvaluator", "error", err)
		return nil, err
	}
	log.Debug("REMOVE_ME finish peerDenialEvaluator")

	log.Debug("REMOVE_ME start SetPeerDenialEvaluator")
	err = networkComponents.NetworkMessenger().SetPeerDenialEvaluator(peerDenialEvaluator)
	if err != nil {
		log.Debug("REMOVE_ME finish with error SetPeerDenialEvaluator", "error", err)
		return nil, err
	}
	log.Debug("REMOVE_ME finish SetPeerDenialEvaluator")

	log.Debug("REMOVE_ME start genesisTime")
	genesisTime := time.Unix(coreComponents.GenesisNodesSetup().GetStartTime(), 0)
	log.Debug("REMOVE_ME finish genesisTime")

	log.Debug("REMOVE_ME start ConsensusGroupSize")
	consensusGroupSize, err := consensusComponents.ConsensusGroupSize()
	if err != nil {
		log.Debug("REMOVE_ME finish with error ConsensusGroupSize", "error", err)
		return nil, err
	}
	log.Debug("REMOVE_ME finish ConsensusGroupSize")

	log.Debug("REMOVE_ME start NewESDTDataStorage")
	esdtNftStorage, err := builtInFunctions.NewESDTDataStorage(builtInFunctions.ArgsNewESDTDataStorage{
		Accounts:                stateComponents.AccountsAdapterAPI(),
		GlobalSettingsHandler:   nodeDisabled.NewDisabledGlobalSettingHandler(),
		Marshalizer:             coreComponents.InternalMarshalizer(),
		SaveToSystemEnableEpoch: epochConfig.EnableEpochs.OptimizeNFTStoreEnableEpoch,
		EpochNotifier:           coreComponents.EpochNotifier(),
		ShardCoordinator:        processComponents.ShardCoordinator(),
	})
	if err != nil {
		log.Debug("REMOVE_ME finish with error NewESDTDataStorage", "error", err)
		return nil, err
	}
	log.Debug("REMOVE_ME finish NewESDTDataStorage")

	var nd *Node
	log.Debug("REMOVE_ME start NewNode")
	nd, err = NewNode(
		WithCoreComponents(coreComponents),
		WithCryptoComponents(cryptoComponents),
		WithBootstrapComponents(bootstrapComponents),
		WithStateComponents(stateComponents),
		WithDataComponents(dataComponents),
		WithStatusComponents(statusComponents),
		WithProcessComponents(processComponents),
		WithHeartbeatComponents(heartbeatComponents),
		WithHeartbeatV2Components(heartbeatV2Components),
		WithConsensusComponents(consensusComponents),
		WithNetworkComponents(networkComponents),
		WithInitialNodesPubKeys(coreComponents.GenesisNodesSetup().InitialNodesPubKeys()),
		WithRoundDuration(coreComponents.GenesisNodesSetup().GetRoundDuration()),
		WithConsensusGroupSize(consensusGroupSize),
		WithGenesisTime(genesisTime),
		WithConsensusType(config.Consensus.Type),
		WithBootstrapRoundIndex(bootstrapRoundIndex),
		WithPeerDenialEvaluator(peerDenialEvaluator),
		WithRequestedItemsHandler(processComponents.RequestedItemsHandler()),
		WithAddressSignatureSize(config.AddressPubkeyConverter.SignatureLength),
		WithValidatorSignatureSize(config.ValidatorPubkeyConverter.SignatureLength),
		WithPublicKeySize(config.ValidatorPubkeyConverter.Length),
		WithNodeStopChannel(coreComponents.ChanStopNodeProcess()),
		WithImportMode(isInImportMode),
		WithESDTNFTStorageHandler(esdtNftStorage),
	)
	if err != nil {
		log.Debug("REMOVE_ME finish with error NewNode", "error", err)
		return nil, errors.New("error creating node: " + err.Error())
	}
	log.Debug("REMOVE_ME finish NewNode")

	if processComponents.ShardCoordinator().SelfId() < processComponents.ShardCoordinator().NumberOfShards() {
		log.Debug("REMOVE_ME start CreateShardedStores")
		err = nd.CreateShardedStores()
		if err != nil {
			log.Debug("REMOVE_ME finish with error CreateShardedStores", "error", err)
			return nil, err
		}
		log.Debug("REMOVE_ME finish CreateShardedStores")
	}

	log.Debug("REMOVE_ME start CreateInterceptedDebugHandler")
	err = nodeDebugFactory.CreateInterceptedDebugHandler(
		nd,
		processComponents.InterceptorsContainer(),
		processComponents.ResolversFinder(),
		config.Debug.InterceptorResolver,
	)
	if err != nil {
		log.Debug("REMOVE_ME finish with error CreateInterceptedDebugHandler", "error", err)
		return nil, err
	}
	log.Debug("REMOVE_ME finish CreateInterceptedDebugHandler")

	return nd, nil
}
