package trieIterators

import (
	"context"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

type accountListProcessor struct {
	*commonStakingProcessor
	publicKeyConverter core.PubkeyConverter
	marshalizer        marshal.Marshalizer
}

// NewAccountListProcessor creates a new instance of accountListProcessor
func NewAccountListProcessor(arg ArgTrieIteratorProcessor) (*accountListProcessor, error) {
	err := checkArguments(arg)
	if err != nil {
		return nil, err
	}

	return &accountListProcessor{
		commonStakingProcessor: &commonStakingProcessor{
			queryService: arg.QueryService,
			blockChain:   arg.BlockChain,
			accounts:     arg.Accounts,
		},
		publicKeyConverter: arg.PublicKeyConverter,
		marshalizer:        arg.Marshalizer,
	}, nil
}

// GetAccountsList creates a list of all accounts in the trie with their balances
func (acp *accountListProcessor) GetAccountsList() ([]*api.Account, error) {
	currentHeader := acp.blockChain.GetCurrentBlockHeader()
	if check.IfNil(currentHeader) {
		return nil, ErrNodeNotInitialized
	}

	err := acp.accounts.RecreateTrie(currentHeader.GetRootHash())
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	chLeaves, err := acp.accounts.GetAllLeaves(currentHeader.GetRootHash(), ctx)
	if err != nil {
		return nil, err
	}

	accList := make([]*api.Account, 0)
	for leaf := range chLeaves {
		userAccount, errUnmarshal := unmarshalUserAccount(leaf.Key(), leaf.Value(), acp.marshalizer)
		if errUnmarshal != nil {
			// log.Debug("cannot unmarshal genesis user account. it may be a code leaf", "error", errUnmarshal)
			continue
		}

		accList = append(accList, &api.Account{
			Address: acp.publicKeyConverter.Encode(userAccount.AddressBytes()),
			Balance: userAccount.GetBalance().String(),
			BalanceAsBigInt: big.NewInt(0).Set(userAccount.GetBalance()),
		})
	}

	return accList, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (acp *accountListProcessor) IsInterfaceNil() bool {
	return acp == nil
}

func unmarshalUserAccount(address []byte, userAccountsBytes []byte, marshalizer marshal.Marshalizer) (state.UserAccountHandler, error) {
	userAccount, err := state.NewUserAccount(address)
	if err != nil {
		return nil, err
	}
	err = marshalizer.Unmarshal(userAccount, userAccountsBytes)
	if err != nil {
		return nil, err
	}

	return userAccount, nil
}
