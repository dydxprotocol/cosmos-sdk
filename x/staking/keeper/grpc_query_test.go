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

	// Create test validators (5 bonded + 1 unbonding + 1 unbonded)
	validators := make([]types.Validator, 7)
	valOpAddrs := make([]string, 7)
	for i := 0; i < 7; i++ {
		validators[i] = testutil.NewValidator(s.T(), sdk.ValAddress(PKs[i].Address().Bytes()), PKs[i])
		if i < 5 {
			validators[i].Status = types.Bonded
		} else if i == 5 {
			validators[i].Status = types.Unbonding
		} else {
			validators[i].Status = types.Unbonded
		}
		require.NoError(keeper.SetValidator(ctx, validators[i]))
		valOpAddrs[i] = validators[i].OperatorAddress
	}

	testCases := []struct {
		name     string
		setup    func()
		expected []string
	}{
		{
			name:     "empty",
			setup:    func() {},
			expected: []string(nil),
		},
		{
			name: "5 bonded",
			setup: func() {
				require.NoError(keeper.SetProposers(ctx, valOpAddrs[:5]))
			},
			expected: valOpAddrs[:5],
		},
		{
			name: "5 bonded + 1 unbonding + 1 unbonded",
			setup: func() {
				require.NoError(keeper.SetProposers(ctx, valOpAddrs))
			},
			expected: valOpAddrs,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setup()

			res, err := queryClient.Proposers(gocontext.Background(), &types.QueryProposersRequest{})
			require.NoError(err)
			require.NotNil(res)
			require.Equal(tc.expected, res.Proposers)
		})
	}
}
