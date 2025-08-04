package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/golang/mock/gomock"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
)

func (s *KeeperTestSuite) TestApplyAndReturnValidatorSetUpdates_ProposerSet() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	// Helper function to get a validator update by public key
	getUpdateByPK := func(updates []abci.ValidatorUpdate, pk cryptotypes.PubKey) (abci.ValidatorUpdate, bool) {
		for _, update := range updates {
			pubKeyBytes := update.PubKey.GetEd25519()
			if string(pk.Bytes()) == string(pubKeyBytes) {
				return update, true
			}
		}
		return abci.ValidatorUpdate{}, false
	}

	// Helper function to create validators
	setupValidators := func(testCtx sdk.Context, count int) []string {
		valOpAddrs := make([]string, count)

		s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		for i := 0; i < count; i++ {
			valPubKey := PKs[i]
			valAddr := sdk.ValAddress(valPubKey.Address())

			// Create validator with 1000 power
			validator := testutil.NewValidator(s.T(), valAddr, valPubKey)
			tokens := keeper.TokensFromConsensusPower(testCtx, 1000)
			validator, _ = validator.AddTokensFromDel(tokens)
			valOpAddrs[i] = validator.OperatorAddress

			// apply=true so that initial validator creations won't be included in validator updates
			stakingkeeper.TestingUpdateValidator(keeper, testCtx, validator, true)
		}

		return valOpAddrs
	}

	testCases := []struct {
		name   string
		setup  func(testCtx sdk.Context) error
		verify func(updates []abci.ValidatorUpdate)
	}{
		{
			name: "proposers not set, validator power change, CanPropose should default to true",
			setup: func(testCtx sdk.Context) error {
				setupValidators(testCtx, 1)

				// Trigger a power update for val0
				val0, err := keeper.GetValidator(testCtx, sdk.ValAddress(PKs[0].Address()))
				if err != nil {
					return err
				}
				powerReduction := keeper.PowerReduction(testCtx)
				val0, _ = val0.AddTokensFromDel(powerReduction)
				stakingkeeper.TestingUpdateValidator(keeper, testCtx, val0, false)

				return nil
			},
			verify: func(updates []abci.ValidatorUpdate) {
				require.Len(updates, 1)
				require.True(updates[0].CanPropose)
				require.Equal(int64(1001), updates[0].Power) // 1000 + 1
			},
		},
		{
			name: "proposers set to empty, validator power change, CanPropose should default to true",
			setup: func(testCtx sdk.Context) error {
				setupValidators(testCtx, 1)

				if err := keeper.SetProposers(testCtx, []string{}); err != nil {
					return err
				}

				// Trigger a power update for val0
				val0, err := keeper.GetValidator(testCtx, sdk.ValAddress(PKs[0].Address()))
				if err != nil {
					return err
				}
				powerReduction := keeper.PowerReduction(testCtx)
				val0, _ = val0.AddTokensFromDel(powerReduction)
				stakingkeeper.TestingUpdateValidator(keeper, testCtx, val0, false)

				return nil
			},
			verify: func(updates []abci.ValidatorUpdate) {
				require.Len(updates, 1)
				require.True(updates[0].CanPropose)
				require.Equal(int64(1001), updates[0].Power) // 1000 + 1
			},
		},
		{
			name: "proposers set to all validators, no power change, all validators get updates",
			setup: func(testCtx sdk.Context) error {
				// Create 5 validators for this test
				valOpAddrs := setupValidators(testCtx, 5)

				return keeper.SetProposers(testCtx, valOpAddrs)
			},
			verify: func(updates []abci.ValidatorUpdate) {
				require.Len(updates, 5)

				for _, update := range updates {
					require.True(update.CanPropose)
					require.Equal(int64(1000), update.Power)
				}
			},
		},
		{
			name: "proposers set to first five, validator power changes, all validators get updates",
			setup: func(testCtx sdk.Context) error {
				// Create 7 validators
				valOpAddrs := setupValidators(testCtx, 7)

				// Set first 5 as proposers
				if err := keeper.SetProposers(testCtx, valOpAddrs[:5]); err != nil {
					return err
				}

				// Change power of val4
				val4, err := keeper.GetValidator(testCtx, sdk.ValAddress(PKs[4].Address()))
				if err != nil {
					return err
				}
				powerReduction := keeper.PowerReduction(testCtx)
				val4, _ = val4.AddTokensFromDel(powerReduction.Mul(math.NewInt(7)))
				stakingkeeper.TestingUpdateValidator(keeper, testCtx, val4, false)

				// Change power of val5
				val5, err := keeper.GetValidator(testCtx, sdk.ValAddress(PKs[5].Address()))
				if err != nil {
					return err
				}
				val5, _ = val5.AddTokensFromDel(powerReduction)
				stakingkeeper.TestingUpdateValidator(keeper, testCtx, val5, false)

				return nil
			},
			verify: func(updates []abci.ValidatorUpdate) {
				require.Len(updates, 7)

				for i := 0; i < 7; i++ {
					valUpdate, exists := getUpdateByPK(updates, PKs[i])
					require.True(exists)

					// Check `CanPropose`
					if i < 5 {
						require.True(valUpdate.CanPropose)
					} else {
						require.False(valUpdate.CanPropose)
					}

					// Check `Power`
					if i == 4 {
						require.Equal(int64(1007), valUpdate.Power)
					} else if i == 5 {
						require.Equal(int64(1001), valUpdate.Power)
					} else {
						require.Equal(int64(1000), valUpdate.Power)
					}
				}
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			// Use a cached context to isolate each test case's state
			testCtx, _ := ctx.CacheContext()

			err := tc.setup(testCtx)
			require.NoError(err)

			// Get validator updates
			updates, err := keeper.ApplyAndReturnValidatorSetUpdates(testCtx)
			require.NoError(err)

			// Verify updates
			tc.verify(updates)

			// Verify SendFullProposerSetAbciUpdate is reset to false
			fullUpdate, err := keeper.GetSendFullProposerSetAbciUpdate(ctx)
			require.NoError(err)
			require.False(fullUpdate)
		})
	}
}
