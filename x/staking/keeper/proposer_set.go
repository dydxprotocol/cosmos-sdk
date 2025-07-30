package keeper

import (
	"context"
	"encoding/json"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	// minimum number of bonded validators required in a proposer set
	MIN_BONDED_IN_PROPOSER_SET = 5
)

// GetProposers returns all proposers by their operator addresses.
// Note: Returns empty array when no proposers are set
func (k Keeper) GetProposers(ctx context.Context) ([]string, error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(types.ProposerSetKey)
	if err != nil {
		return nil, err
	}

	if bz == nil {
		return []string{}, nil
	}

	var proposers []string
	if err := json.Unmarshal(bz, &proposers); err != nil {
		return nil, err
	}

	return proposers, nil
}

// SetProposers sets proposers in state by storing their operator addresses.
// Returns error if proposer set invariants are violated.
func (k Keeper) SetProposers(ctx context.Context, proposers []string) error {
	if err := k.checkProposerSetInvariants(ctx, proposers); err != nil {
		return err
	}

	bz, err := json.Marshal(proposers)
	if err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.ProposerSetKey, bz)
}

// checkProposerSetInvariant validates invariants of a proposer set, which are:
// - all proposers are valid operator addresses
// - all proposers correspond to existing validators
// - at least MIN_BONDED_IN_PROPOSER_SET proposers are bonded
func (k Keeper) checkProposerSetInvariants(ctx context.Context, proposers []string) error {
	if len(proposers) == 0 {
		return nil // Valid as x/staking will default to all validators being proposers.
	}

	bonded := 0
	for _, proposerAddr := range proposers {
		operatorAddr, err := k.validatorAddressCodec.StringToBytes(proposerAddr)
		if err != nil {
			return err
		}

		validator, err := k.GetValidator(ctx, operatorAddr)
		if err != nil {
			return err
		}

		if validator.Status == types.Bonded {
			bonded++
		}
	}

	if bonded < MIN_BONDED_IN_PROPOSER_SET {
		return errorsmod.Wrapf(types.ErrInsufficientBondedValidators,
			"proposer set only has %d bonded validators, less than the required %d",
			bonded, MIN_BONDED_IN_PROPOSER_SET)
	}

	return nil
}
