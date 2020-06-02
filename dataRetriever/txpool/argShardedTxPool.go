package txpool

import (
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// ArgShardedTxPool is the argument for ShardedTxPool's constructor
type ArgShardedTxPool struct {
	Config         storageUnit.CacheConfig
	MinGasPrice    uint64
	NumberOfShards uint32
	SelfShardID    uint32
}

// TODO: Upon further analysis and brainstorming, add some sensible minimum accepted values for the appropriate fields.
func (args *ArgShardedTxPool) verify() error {
	config := args.Config

	if config.SizeInBytes == 0 {
		return fmt.Errorf("%w: config.SizeInBytes is not valid", dataRetriever.ErrCacheConfigInvalidSizeInBytes)
	}
	if config.SizeInBytesPerSender == 0 {
		return fmt.Errorf("%w: config.SizeInBytesPerSender is not valid", dataRetriever.ErrCacheConfigInvalidSizeInBytes)
	}
	if config.Capacity == 0 {
		return fmt.Errorf("%w: config.Capacity is not valid", dataRetriever.ErrCacheConfigInvalidSize)
	}
	if config.SizePerSender == 0 {
		return fmt.Errorf("%w: config.SizePerSender is not valid", dataRetriever.ErrCacheConfigInvalidSize)
	}
	if config.Shards == 0 {
		return fmt.Errorf("%w: config.Shards (map chunks) is not valid", dataRetriever.ErrCacheConfigInvalidShards)
	}
	if args.MinGasPrice == 0 {
		return fmt.Errorf("%w: MinGasPrice is not valid", dataRetriever.ErrCacheConfigInvalidEconomics)
	}
	if args.NumberOfShards == 0 {
		return fmt.Errorf("%w: NumberOfShards is not valid", dataRetriever.ErrCacheConfigInvalidSharding)
	}

	return nil
}

// String returns a readable representation of the object
func (args *ArgShardedTxPool) String() string {
	bytes, _ := json.Marshal(args)
	return string(bytes)
}
