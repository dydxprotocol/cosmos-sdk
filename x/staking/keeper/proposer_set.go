package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/staking/types"
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
// - at least one proposer is bonded
func (k Keeper) checkProposerSetInvariants(ctx context.Context, proposers []string) error {
	if len(proposers) == 0 {
		return nil // Valid as x/staking will default to all validators being proposers.
	}

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
			return nil
		}
	}

	return fmt.Errorf("at least one proposer must be bonded")
}
