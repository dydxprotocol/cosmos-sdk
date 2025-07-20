package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
)

func (s *KeeperTestSuite) TestGetSetProposers() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	// GetAllProposers should return empty initially.
	proposers, err := keeper.GetAllProposers(ctx)
	require.NoError(err)
	require.Empty(proposers)

	// Create validators.
	numValidators := 3
	validators := make([]string, numValidators)
	for i := 0; i < numValidators; i++ {
		valPubKey := PKs[i]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
		validator := testutil.NewValidator(s.T(), valAddr, valPubKey)
		validators[i] = validator.OperatorAddress
	}

	// GetIsProposer should return false for non-proposers.
	isProposer, err := keeper.GetIsProposer(ctx, validators[0])
	require.NoError(err)
	require.False(isProposer)

	// GetIsProposer should return true for proposers.
	err = keeper.SetProposer(ctx, validators[0])
	require.NoError(err)

	isProposer, err = keeper.GetIsProposer(ctx, validators[0])
	require.NoError(err)
	require.True(isProposer)

	// Set all validators as proposers.
	for i := 1; i < numValidators; i++ {
		err = keeper.SetProposer(ctx, validators[i])
		require.NoError(err)
	}

	// Verify all validators are proposers.
	for i := 0; i < numValidators; i++ {
		isProposer, err = keeper.GetIsProposer(ctx, validators[i])
		require.NoError(err)
		require.True(isProposer, "validator %d should be a proposer", i)
	}

	proposers, err = keeper.GetAllProposers(ctx)
	require.NoError(err)
	require.Len(proposers, numValidators)

	proposerMap := make(map[string]bool)
	for _, p := range proposers {
		proposerMap[p] = true
	}
	fmt.Println("validators", validators)
	fmt.Println("proposers", proposers)
	for _, expected := range validators {
		require.True(proposerMap[expected], "expected proposer %s not found", expected)
	}

	// GetIsProposer and SetProposer should return error given invalid addresses.
	_, err = keeper.GetIsProposer(ctx, "invalid-address")
	require.Error(err, "GetIsProposer with invalid address should return error")

	err = keeper.SetProposer(ctx, "invalid-address")
	require.Error(err, "SetProposer with invalid address should return error")
}
