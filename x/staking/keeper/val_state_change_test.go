package keeper_test

import (
	"crypto"
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

				// Use TestingUpdateValidator to update the validator
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
			name: "empty proposers, validator power change, CanPropose should default to true",
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

				// Use TestingUpdateValidator to update the validator
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
			name: "empty proposers, no updates without power changes",
			setup: func(testCtx sdk.Context) error {
				setupValidators(testCtx, 1)
				return keeper.SetProposers(testCtx, []string{})
			},
			verify: func(updates []abci.ValidatorUpdate) {
				require.Empty(updates)
			},
		},
		{
			name: "proposers not set, all validators should get CanPropose=true updates during a force update",
			setup: func(testCtx sdk.Context) error {
				setupValidators(testCtx, 3)

				// Force full proposer set update
				return keeper.SetSendFullProposerSetAbciUpdate(testCtx, true)
			},
			verify: func(updates []abci.ValidatorUpdate) {
				seen := make(map[crypto.PublicKey]bool)
				for _, update := range updates {
					require.True(update.CanPropose)
					require.Equal(int64(1000), update.Power)

					seen[update.PubKey] = true
				}

				require.Len(updates, 3)
				require.Len(seen, 3)
			},
		},
		{
			name: "proposers set, no updates without power changes",
			setup: func(testCtx sdk.Context) error {
				// Create 5 validators for this test
				valOpAddrs := setupValidators(testCtx, 5)

				proposers := make([]string, 5)
				for i := 0; i < 5; i++ {
					proposers[i] = valOpAddrs[i]
				}
				return keeper.SetProposers(testCtx, proposers)
			},
			verify: func(updates []abci.ValidatorUpdate) {
				require.Empty(updates)
			},
		},
		{
			name: "proposers set, updates with appropriate CanPropose when power changes",
			setup: func(testCtx sdk.Context) error {
				// Create 7 validators for this test (need more than 5 to have validators outside proposer set)
				valOpAddrs := setupValidators(testCtx, 7)

				// Set first 5 as proposers
				proposers := valOpAddrs[:5]
				if err := keeper.SetProposers(testCtx, proposers); err != nil {
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
				require.Len(updates, 2)

				// Check update for val4 (in proposer set)
				val4Update, exists := getUpdateByPK(updates, PKs[4])
				require.True(exists)
				require.True(val4Update.CanPropose)
				require.Equal(int64(1007), val4Update.Power)

				// Check update for val5 (not in proposer set)
				val5Update, exists := getUpdateByPK(updates, PKs[5])
				require.True(exists)
				require.False(val5Update.CanPropose)
				require.Equal(int64(1001), val5Update.Power)
			},
		},
		{
			name: "proposers set, no power change, force update",
			setup: func(testCtx sdk.Context) error {
				valOpAddrs := setupValidators(testCtx, 7)

				// Set first 5 validators as proposers
				proposers := make([]string, 5)
				for i := 0; i < 5; i++ {
					proposers[i] = valOpAddrs[i]
				}
				if err := keeper.SetProposers(testCtx, proposers); err != nil {
					return err
				}
				return keeper.SetSendFullProposerSetAbciUpdate(testCtx, true)
			},
			verify: func(updates []abci.ValidatorUpdate) {
				require.Len(updates, 7)

				// Check proposers
				for i := 0; i < 5; i++ {
					update, exists := getUpdateByPK(updates, PKs[i])
					require.True(exists)
					require.True(update.CanPropose)
					require.Equal(int64(1000), update.Power)
				}

				// Check non-proposers
				for i := 5; i < 7; i++ {
					update, exists := getUpdateByPK(updates, PKs[i])
					require.True(exists)
					require.False(update.CanPropose)
					require.Equal(int64(1000), update.Power)
				}
			},
		},
		{
			name: "proposers set, validator power change, force update",
			setup: func(testCtx sdk.Context) error {
				valOpAddrs := setupValidators(testCtx, 7)

				// Set first 5 validators as proposers
				proposers := make([]string, 5)
				for i := 0; i < 5; i++ {
					proposers[i] = valOpAddrs[i]
				}
				if err := keeper.SetProposers(testCtx, proposers); err != nil {
					return err
				}

				// Add 123 to power of val5 (not in proposer set)
				val5, err := keeper.GetValidator(testCtx, sdk.ValAddress(PKs[5].Address()))
				if err != nil {
					return err
				}
				powerReduction := keeper.PowerReduction(testCtx)
				val5, _ = val5.AddTokensFromDel(powerReduction.Mul(math.NewInt(123)))
				stakingkeeper.TestingUpdateValidator(keeper, testCtx, val5, false)

				return keeper.SetSendFullProposerSetAbciUpdate(testCtx, true)
			},
			verify: func(updates []abci.ValidatorUpdate) {
				require.Len(updates, 7)

				// Check validators 0-4 (in proposer set)
				for i := 0; i < 5; i++ {
					update, exists := getUpdateByPK(updates, PKs[i])
					require.True(exists)
					require.True(update.CanPropose)
					require.Equal(int64(1000), update.Power)
				}

				// Check val5 (not in proposer set, power change)
				val5Update, exists := getUpdateByPK(updates, PKs[5])
				require.True(exists)
				require.False(val5Update.CanPropose)
				require.Equal(int64(1123), val5Update.Power)

				// Check val6 (not in proposer set, no power change)
				val6Update, exists := getUpdateByPK(updates, PKs[6])
				require.True(exists)
				require.False(val6Update.CanPropose)
				require.Equal(int64(1000), val6Update.Power)
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
