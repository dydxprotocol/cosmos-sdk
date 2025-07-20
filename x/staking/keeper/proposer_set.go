package keeper

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetIsProposer returns whether a validator (given its operator address) is a block proposer.
func (k Keeper) GetIsProposer(ctx context.Context, operatorAddrStr string) (bool, error) {
	operatorAddr, err := k.validatorAddressCodec.StringToBytes(operatorAddrStr)
	if err != nil {
		return false, err
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetProposerKey(operatorAddr))
	return len(bz) > 0, err
}

// SetProposer sets a validator (given its operator address) as a block proposer.
func (k Keeper) SetProposer(ctx context.Context, operatorAddrStr string) error {
	operatorAddr, err := k.validatorAddressCodec.StringToBytes(operatorAddrStr)
	if err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.GetProposerKey(operatorAddr), []byte{1})
}

// GetAllProposers returns all proposers by their operator addresses. Used for genesis export.
func (k Keeper) GetAllProposers(ctx context.Context) (proposers []string, err error) {
	store := k.storeService.OpenKVStore(ctx)

	iterator, err := store.Iterator(types.ProposerKeyPrefix, storetypes.PrefixEndBytes(types.ProposerKeyPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		// Skip prefix and length prefix to get actual address bytes
		addrBytesWithLen := key[len(types.ProposerKeyPrefix):]
		addrLen := int(addrBytesWithLen[0])
		addrBytes := addrBytesWithLen[1 : 1+addrLen]
		proposer, err := k.validatorAddressCodec.BytesToString(addrBytes)
		if err != nil {
			return nil, err
		}
		proposers = append(proposers, proposer)
	}

	return proposers, nil
}
