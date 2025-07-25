package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *KeeperTestSuite) TestGetSetProposers() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	// GetAllProposers should return empty initially.
	proposers, err := keeper.GetAllProposers(ctx)
	require.NoError(err)
	require.Empty(proposers)

	// Create and set bonded validators.
	numValidators := 3
	validators := make([]string, numValidators)
	for i := 0; i < numValidators; i++ {
		valPubKey := PKs[i]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
		validator := testutil.NewValidator(s.T(), valAddr, valPubKey)
		validator.Status = types.Bonded
		validators[i] = validator.OperatorAddress
		err := keeper.SetValidator(ctx, validator)
		require.NoError(err)
	}

	// When no proposers are set, all validators are eligible to propose.
	for i := 0; i < numValidators; i++ {
		isProposer, err := keeper.GetIsProposer(ctx, validators[i])
		require.NoError(err)
		require.True(isProposer, "validator %d should default to true when no proposers are set", i)
	}

	// Set the first validator as a proposer.
	err = keeper.SetProposer(ctx, validators[0])
	require.NoError(err)

	// Verify first is proposer and rest are not.
	for i := 0; i < numValidators; i++ {
		isProposer, err := keeper.GetIsProposer(ctx, validators[i])
		require.NoError(err)
		if i == 0 {
			require.True(isProposer, "validator %d should be a proposer", i)
		} else {
			require.False(isProposer, "validator %d should not be a proposer", i)
		}
	}

	// Set all validators as proposers.
	for i := 1; i < numValidators; i++ {
		err = keeper.SetProposer(ctx, validators[i])
		require.NoError(err)
	}

	// Verify all validators are proposers.
	for i := 0; i < numValidators; i++ {
		isProposer, err := keeper.GetIsProposer(ctx, validators[i])
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
	for _, expected := range validators {
		require.True(proposerMap[expected], "expected proposer %s not found", expected)
	}

}

func (s *KeeperTestSuite) TestGetSetProposersErrors() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	// GetIsProposer and SetProposer should return error given invalid addresses
	_, err := keeper.GetIsProposer(ctx, "invalid-address")
	require.Error(err, "GetIsProposer with invalid address should return error")

	err = keeper.SetProposer(ctx, "invalid-address")
	require.Error(err, "SetProposer with invalid address should return error")

	// Create a validator
	valPubKey := PKs[0]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	validator := testutil.NewValidator(s.T(), valAddr, valPubKey)

	// SetProposer should return error for non-existent validators
	nonExistentPubKey := PKs[1]
	nonExistentValAddr := sdk.ValAddress(nonExistentPubKey.Address().Bytes())
	err = keeper.SetProposer(ctx, nonExistentValAddr.String())
	require.Error(err)
	require.Contains(err.Error(), "validator does not exist")

	// SetProposer should return error for unbonded validators
	validator.Status = types.Unbonded
	err = keeper.SetValidator(ctx, validator)
	require.NoError(err)
	err = keeper.SetProposer(ctx, validator.OperatorAddress)
	require.Error(err)
	require.Contains(err.Error(), "is not bonded")

	// SetProposer should return error for non-operator addresses of bonded validators
	// such as account and consensus addresses
	accAddr := sdk.AccAddress(valAddr)
	err = keeper.SetProposer(ctx, accAddr.String())
	require.Error(err)
	require.Contains(err.Error(), "does not match bech32 prefix: expected 'cosmosvaloper'")

	consAddr := sdk.ConsAddress(valAddr)
	err = keeper.SetProposer(ctx, consAddr.String())
	require.Error(err)
	require.Contains(err.Error(), "does not match bech32 prefix: expected 'cosmosvaloper'")
}
