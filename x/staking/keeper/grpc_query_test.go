package keeper_test

import (
	gocontext "context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *KeeperTestSuite) TestGRPCQueryValidator() {
	ctx, keeper, queryClient := s.ctx, s.stakingKeeper, s.queryClient
	require := s.Require()

	validator := testutil.NewValidator(s.T(), sdk.ValAddress(PKs[0].Address().Bytes()), PKs[0])
	require.NoError(keeper.SetValidator(ctx, validator))
	var req *types.QueryValidatorRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryValidatorRequest{}
			},
			false,
		},
		{
			"with valid and not existing address",
			func() {
				req = &types.QueryValidatorRequest{
					ValidatorAddr: "cosmosvaloper15jkng8hytwt22lllv6mw4k89qkqehtahd84ptu",
				}
			},
			false,
		},
		{
			"valid request",
			func() {
				req = &types.QueryValidatorRequest{ValidatorAddr: validator.OperatorAddress}
			},
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.Validator(gocontext.Background(), req)
			if tc.expPass {
				require.NoError(err)
				require.True(validator.Equal(&res.Validator))
			} else {
				require.Error(err)
				require.Nil(res)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCQueryProposers() {
	ctx, keeper, queryClient := s.ctx, s.stakingKeeper, s.queryClient
	require := s.Require()

	// Create test validators
	val1 := testutil.NewValidator(s.T(), sdk.ValAddress(PKs[0].Address().Bytes()), PKs[0])
	val2 := testutil.NewValidator(s.T(), sdk.ValAddress(PKs[1].Address().Bytes()), PKs[1])
	require.NoError(keeper.SetValidator(ctx, val1))
	require.NoError(keeper.SetValidator(ctx, val2))

	testCases := []struct {
		name     string
		setup    func()
		expected []string
	}{
		{
			name:     "no proposers set",
			setup:    func() {},
			expected: []string{},
		},
		{
			name: "single proposer",
			setup: func() {
				require.NoError(keeper.SetProposer(ctx, val1.OperatorAddress))
			},
			expected: []string{val1.OperatorAddress},
		},
		{
			name: "multiple proposers",
			setup: func() {
				require.NoError(keeper.SetProposer(ctx, val1.OperatorAddress))
				require.NoError(keeper.SetProposer(ctx, val2.OperatorAddress))
			},
			expected: []string{val1.OperatorAddress, val2.OperatorAddress},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setup()
			
			res, err := queryClient.Proposers(gocontext.Background(), &types.QueryProposersRequest{})
			require.NoError(err)
			require.NotNil(res)
			// Handle nil vs empty slice difference
			if len(tc.expected) == 0 && res.Proposers == nil {
				// Both are effectively empty
				return
			}
			require.ElementsMatch(tc.expected, res.Proposers)
		})
	}
}
