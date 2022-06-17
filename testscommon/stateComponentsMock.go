package testscommon

import (
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/state"
)

// StateComponentsMock -
type StateComponentsMock struct {
	PeersAcc               state.AccountsAdapter
	Accounts               state.AccountsAdapter
	AccountsAPI            state.AccountsAdapter
	AccountsAPIFactory     factory.AccountsAdapterAPIFactory
	PeerAccountsAPIFactory factory.AccountsAdapterAPIFactory
	Tries                  common.TriesHolder
	StorageManagers        map[string]common.StorageManager
}

// Create -
func (scm *StateComponentsMock) Create() error {
	return nil
}

// Close -
func (scm *StateComponentsMock) Close() error {
	return nil
}

// CheckSubcomponents -
func (scm *StateComponentsMock) CheckSubcomponents() error {
	return nil
}

// PeerAccounts -
func (scm *StateComponentsMock) PeerAccounts() state.AccountsAdapter {
	return scm.PeersAcc
}

// AccountsAdapter -
func (scm *StateComponentsMock) AccountsAdapter() state.AccountsAdapter {
	return scm.Accounts
}

// AccountsAdapterAPI -
func (scm *StateComponentsMock) AccountsAdapterAPI() state.AccountsAdapter {
	return scm.AccountsAPI
}

// AccountsAdapterAPIFactory -
func (scm *StateComponentsMock) AccountsAdapterAPIFactory() factory.AccountsAdapterAPIFactory {
	return scm.AccountsAPIFactory
}

// PeerAccountsAdapterAPIFactory -
func (scm *StateComponentsMock) PeerAccountsAdapterAPIFactory() factory.AccountsAdapterAPIFactory {
	return scm.AccountsAPIFactory
}

// TriesContainer -
func (scm *StateComponentsMock) TriesContainer() common.TriesHolder {
	return scm.Tries
}

// TrieStorageManagers -
func (scm *StateComponentsMock) TrieStorageManagers() map[string]common.StorageManager {
	return scm.StorageManagers
}

// String -
func (scm *StateComponentsMock) String() string {
	return "StateComponentsMock"
}

// IsInterfaceNil -
func (scm *StateComponentsMock) IsInterfaceNil() bool {
	return scm == nil
}
