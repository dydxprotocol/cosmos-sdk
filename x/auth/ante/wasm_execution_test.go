package ante_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/golang/mock/gomock"

	"github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/stretchr/testify/require"
)

// MockMsgExecuteContract is a mock message for testing purposes.
type MockMsgExecuteContract struct{}

func (m MockMsgExecuteContract) Reset()         {}
func (m MockMsgExecuteContract) String() string { return "MockMsgExecuteContract" }
func (m MockMsgExecuteContract) ProtoMessage()  {}

func (m MockMsgExecuteContract) XXX_MessageName() string {
	return "cosmwasm.wasm.v1.MsgExecuteContract"
}

func TestIsSingleWasmExecTx(t *testing.T) {
	// Create a mock message with the desired TypeURL
	msg := &MockMsgExecuteContract{}

	suite := SetupTestSuite(t, true)
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	err := suite.txBuilder.SetMsgs(msg)
	require.NoError(t, err)

	// Create the transaction
	sdkTx := suite.txBuilder.GetTx()

	// Test the function
	result := ante.IsSingleWasmExecTx(sdkTx)
	require.True(t, result, "Expected IsSingleWasmExecTx to return true for a single MsgExecuteContract message")

	// Test with multiple messages
	err = suite.txBuilder.SetMsgs(msg, msg)
	require.NoError(t, err)
	sdkTx = suite.txBuilder.GetTx()
	result = ante.IsSingleWasmExecTx(sdkTx)
	require.False(t, result, "Expected IsSingleWasmExecTx to return false for multiple messages")
}

func TestWasmExecExemptFromGas(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup the test suite
	suite := SetupTestSuite(t, false)

	// Create a mock account
	acc := &types.BaseAccount{}
	acc.SetAddress(sdk.AccAddress([]byte("test_address")))

	// Set up the mock bank keeper to return a specific balance
	suite.bankKeeper.EXPECT().
		GetBalance(gomock.Any(), acc.GetAddress(), ante.UusdcDenom).
		Return(sdk.Coin{Denom: ante.UusdcDenom, Amount: sdkmath.NewInt(ante.UusdcBalanceForGasExemption)}).
		Times(1)

	// Test the function with sufficient balance
	result := ante.WasmExecExemptFromGas(suite.bankKeeper, suite.ctx, acc)
	require.True(t, result, "Expected WasmExecExemptFromGas to return true for sufficient balance")

	// Set up the mock bank keeper to return an insufficient balance
	suite.bankKeeper.EXPECT().
		GetBalance(gomock.Any(), acc.GetAddress(), ante.UusdcDenom).
		Return(sdk.Coin{Denom: ante.UusdcDenom, Amount: sdkmath.NewInt(ante.UusdcBalanceForGasExemption - 1)}).
		Times(1)

	// Test the function with insufficient balance
	result = ante.WasmExecExemptFromGas(suite.bankKeeper, suite.ctx, acc)
	require.False(t, result, "Expected WasmExecExemptFromGas to return false for insufficient balance")
}
