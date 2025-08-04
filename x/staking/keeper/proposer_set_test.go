package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *KeeperTestSuite) TestGetSetProposers() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	// Proposer set should be empty by default
	proposers, err := keeper.GetProposers(ctx)
	require.NoError(err)
	require.Empty(proposers)

	// Create six validators: 5 bonded and 1 unbonding
	validators := make([]string, 6)
	for i := 0; i < 6; i++ {
		valPubKey := PKs[i]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
		validator := testutil.NewValidator(s.T(), valAddr, valPubKey)
		if i == 5 {
			validator.Status = types.Unbonding
		} else {
			validator.Status = types.Bonded
		}
		validators[i] = validator.OperatorAddress
		err := keeper.SetValidator(ctx, validator)
		require.NoError(err)
	}

	// Set first 5 validators as proposers and verify
	err = keeper.SetProposers(ctx, validators[:5])
	require.NoError(err)

	proposers, err = keeper.GetProposers(ctx)
	require.NoError(err)
	require.Equal(validators[:5], proposers)

	// Set last 5 validators as proposers
	// Should fail as there are insufficient bonded proposers
	err = keeper.SetProposers(ctx, validators[1:])
	require.Error(err)
	require.ErrorContains(err, "proposer set only has 4 bonded validators, less than the required 5")

	// Set all validators as proposers and verify
	err = keeper.SetProposers(ctx, validators)
	require.NoError(err)

	proposers, err = keeper.GetProposers(ctx)
	require.NoError(err)
	require.ElementsMatch(validators, proposers)

	// Set proposer set back to empty and verify
	err = keeper.SetProposers(ctx, []string{})
	require.NoError(err)

	proposers, err = keeper.GetProposers(ctx)
	require.NoError(err)
	require.Empty(proposers)
}

func (s *KeeperTestSuite) TestSetProposersErrors() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	// Create a validator
	valPubKey := PKs[0]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	validator := testutil.NewValidator(s.T(), valAddr, valPubKey)
	err := keeper.SetValidator(ctx, validator)
	require.NoError(err)

	// Should return error given invalid addresses
	err = keeper.SetProposers(ctx, []string{"invalid-address"})
	require.Error(err, "SetProposers with invalid address should return error")

	// Should return error for non-operator addresses (e.g. account and consensus addresses)
	accAddr := sdk.AccAddress(valAddr)
	err = keeper.SetProposers(ctx, []string{accAddr.String()})
	require.ErrorContains(err, "does not match bech32 prefix: expected 'cosmosvaloper'")

	consAddr := sdk.ConsAddress(valAddr)
	err = keeper.SetProposers(ctx, []string{consAddr.String()})
	require.ErrorContains(err, "does not match bech32 prefix: expected 'cosmosvaloper'")

	// Should return error for non-existent validators
	nonExistentPubKey := PKs[1]
	nonExistentValAddr := sdk.ValAddress(nonExistentPubKey.Address().Bytes())
	err = keeper.SetProposers(ctx, []string{nonExistentValAddr.String()})
	require.ErrorContains(err, "validator does not exist")
}

func (s *KeeperTestSuite) TestGetSetSendFullProposerSetAbciUpdate() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	// Should be false when not set yet.
	send, err := keeper.GetSendFullProposerSetAbciUpdate(ctx)
	require.NoError(err)
	require.False(send, "default value should be false")

	// Set to true and verify
	err = keeper.SetSendFullProposerSetAbciUpdate(ctx, true)
	require.NoError(err)

	send, err = keeper.GetSendFullProposerSetAbciUpdate(ctx)
	require.NoError(err)
	require.True(send)

	// Set to false and verify
	err = keeper.SetSendFullProposerSetAbciUpdate(ctx, false)
	require.NoError(err)

	send, err = keeper.GetSendFullProposerSetAbciUpdate(ctx)
	require.NoError(err)
	require.False(send)

	// Set to true again and verify.
	err = keeper.SetSendFullProposerSetAbciUpdate(ctx, true)
	require.NoError(err)

	send, err = keeper.GetSendFullProposerSetAbciUpdate(ctx)
	require.NoError(err)
	require.True(send)
}
