package keeper

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetIsProposer returns whether a validator (given its operator address) is a block proposer.
// Note: all validators are proposers if none is set.
func (k Keeper) GetIsProposer(ctx context.Context, operatorAddrStr string) (bool, error) {
	operatorAddr, err := k.validatorAddressCodec.StringToBytes(operatorAddrStr)
	if err != nil {
		return false, err
	}

	store := k.storeService.OpenKVStore(ctx)

	iterator, err := store.Iterator(types.ProposerKeyPrefix, storetypes.PrefixEndBytes(types.ProposerKeyPrefix))
	if err != nil {
		return false, err
	}
	defer iterator.Close()

	// Return true if not a single proposer is set
	if !iterator.Valid() {
		return true, nil
	}

	bz, err := store.Get(types.GetProposerKey(operatorAddr))
	return len(bz) > 0, err
}

// SetProposer sets a validator (given its operator address) as a block proposer.
func (k Keeper) SetProposer(ctx context.Context, operatorAddrStr string) error {
	operatorAddr, err := k.validatorAddressCodec.StringToBytes(operatorAddrStr)
	if err != nil {
		return err
	}

	// Check that given operator address is a bonded validator
	validator, err := k.GetValidator(ctx, operatorAddr)
	if err != nil {
		return err
	}
	if validator.Status != types.Bonded {
		return fmt.Errorf("validator %s is not bonded", operatorAddrStr)
	}

	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.GetProposerKey(operatorAddr), []byte{1})
}

// GetAllProposers returns all proposers by their operator addresses. Used for genesis export.
// Note: This returns only explicitly set proposers, i.e. when no proposers are set, empty array
// is returned, whereas GetIsProposer defaults to all validators being proposers,
func (k Keeper) GetAllProposers(ctx context.Context) (proposers []string, err error) {
	store := k.storeService.OpenKVStore(ctx)

	iterator, err := store.Iterator(types.ProposerKeyPrefix, storetypes.PrefixEndBytes(types.ProposerKeyPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		addrBytes := types.GetValOpAddrFromProposerKey(iterator.Key())
		proposer, err := k.validatorAddressCodec.BytesToString(addrBytes)
		if err != nil {
			return nil, err
		}
		proposers = append(proposers, proposer)
	}

	return proposers, nil
}
