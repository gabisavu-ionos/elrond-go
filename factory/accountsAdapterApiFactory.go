package factory

import (
	"fmt"

	chainData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/state"
)

type accountsAdapterAPIFactory struct {
	accountsDBArgs state.ArgsAccountsDB
	chainHandler   chainData.ChainHandler
}

func newAccountsAdapterAPIFactory(accountsDBArgs state.ArgsAccountsDB, chainHandler chainData.ChainHandler) *accountsAdapterAPIFactory {
	return &accountsAdapterAPIFactory{
		accountsDBArgs: accountsDBArgs,
		chainHandler:   chainHandler,
	}
}

// Create will create a new instance of AccountsAdapter that is going to be used in API operations
func (factory *accountsAdapterAPIFactory) Create() (state.AccountsAdapter, error) {
	accountsAdapterAPI, err := state.NewAccountsDB(factory.accountsDBArgs)
	if err != nil {
		return nil, fmt.Errorf("accounts adapter API: %w: %s", errors.ErrAccountsAdapterCreation, err.Error())
	}

	log.Debug("accountsAdapterAPIFactory: created an AccountsAdapter instance")

	return state.NewAccountsDBApi(accountsAdapterAPI, factory.chainHandler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (factory *accountsAdapterAPIFactory) IsInterfaceNil() bool {
	return factory == nil
}
