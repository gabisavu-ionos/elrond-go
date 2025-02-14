package state

import (
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// LastSnapshotStarted -
const LastSnapshotStarted = lastSnapshotStarted

// NewEmptyBaseAccount -
func NewEmptyBaseAccount(address []byte, tracker DataTrieTracker) *baseAccount {
	return &baseAccount{
		address:         address,
		dataTrieTracker: tracker,
	}
}

// LoadCode -
func (adb *AccountsDB) LoadCode(accountHandler baseAccountHandler) error {
	return adb.loadCode(accountHandler)
}

// LoadDataTrie -
func (adb *AccountsDB) LoadDataTrie(accountHandler baseAccountHandler) error {
	return adb.loadDataTrie(accountHandler)
}

// GetAccount -
func (adb *AccountsDB) GetAccount(address []byte) (vmcommon.AccountHandler, error) {
	return adb.getAccount(address)
}

// GetObsoleteHashes -
func (adb *AccountsDB) GetObsoleteHashes() map[string][][]byte {
	return adb.obsoleteDataTrieHashes
}

// WaitForCompletionIfRunningInImportDB -
func (adb *AccountsDB) WaitForCompletionIfRunningInImportDB(stats common.SnapshotStatisticsHandler) {
	adb.waitForCompletionIfRunningInImportDB(stats)
}

// GetCode -
func GetCode(account baseAccountHandler) []byte {
	return account.GetCodeHash()
}

// GetCodeEntry -
func GetCodeEntry(codeHash []byte, trie Updater, marshalizer marshal.Marshalizer) (*CodeEntry, error) {
	return getCodeEntry(codeHash, trie, marshalizer)
}

// RecreateTrieIfNecessary -
func (accountsDB *accountsDBApi) RecreateTrieIfNecessary() error {
	return accountsDB.recreateTrieIfNecessary()
}

// DoRecreateTrie -
func (accountsDB *accountsDBApi) DoRecreateTrie(targetRootHash []byte) error {
	return accountsDB.doRecreateTrie(targetRootHash)
}

// SetLastRootHash -
func (accountsDB *accountsDBApi) SetLastRootHash(rootHash []byte) {
	accountsDB.mutLastRootHash.Lock()
	accountsDB.lastRootHash = rootHash
	accountsDB.mutLastRootHash.Unlock()
}

// EmptyErrChanReturningHadContained -
func EmptyErrChanReturningHadContained(errChan chan error) bool {
	return emptyErrChanReturningHadContained(errChan)
}
