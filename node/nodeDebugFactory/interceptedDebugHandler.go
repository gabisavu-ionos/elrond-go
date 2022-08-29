package nodeDebugFactory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/debug/factory"
	"github.com/ElrondNetwork/elrond-go/process"
)

// InterceptorResolverDebugger is the contant string for the debugger
const InterceptorResolverDebugger = "interceptor resolver debugger"

var log = logger.GetOrCreate("node")

// CreateInterceptedDebugHandler creates and applies an interceptor-resolver debug handler
func CreateInterceptedDebugHandler(
	node NodeWrapper,
	interceptors process.InterceptorsContainer,
	resolvers dataRetriever.ResolversFinder,
	config config.InterceptorResolverDebugConfig,
) error {
	if check.IfNil(node) {
		return ErrNilNodeWrapper
	}
	if check.IfNil(interceptors) {
		return ErrNilInterceptorContainer
	}
	if check.IfNil(resolvers) {
		return ErrNilResolverContainer
	}

	log.Debug("REMOVE_ME start factory.NewInterceptorResolverDebuggerFactory")
	debugHandler, err := factory.NewInterceptorResolverDebuggerFactory(config)
	if err != nil {
		log.Debug("REMOVE_ME finish with error factory.NewInterceptorResolverDebuggerFactory", "error", err)
		return err
	}
	log.Debug("REMOVE_ME finish factory.NewInterceptorResolverDebuggerFactory")

	log.Debug("REMOVE_ME start iterating on interceptors")
	var errFound error
	interceptors.Iterate(func(key string, interceptor process.Interceptor) bool {
		err = interceptor.SetInterceptedDebugHandler(debugHandler)
		if err != nil {
			errFound = err
			return false
		}

		return true
	})
	log.Debug("REMOVE_ME finished iterating on interceptors", "error found", errFound)
	if errFound != nil {
		return fmt.Errorf("%w while setting up debugger on interceptors", errFound)
	}

	log.Debug("REMOVE_ME start iterating on resolvers")
	resolvers.Iterate(func(key string, resolver dataRetriever.Resolver) bool {
		err = resolver.SetResolverDebugHandler(debugHandler)
		if err != nil {
			errFound = err
			return false
		}

		return true
	})
	log.Debug("REMOVE_ME finished iterating on resolvers", "error found", errFound)
	if errFound != nil {
		return fmt.Errorf("%w while setting up debugger on resolvers", errFound)
	}

	log.Debug("REMOVE_ME adding query handler")
	err = node.AddQueryHandler(InterceptorResolverDebugger, debugHandler)
	log.Debug("REMOVE_ME added query handler", "error", err)

	return err
}
