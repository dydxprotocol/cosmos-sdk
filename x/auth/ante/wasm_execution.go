package ante

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	// The IBC denom of `uusdc` (atomic unit of Noble USDC)
	UusdcDenom                  = "ibc/8E27BA2D5493AF5636760E354E46004562C46AB7EC0CC4C1CA14E9E20E2545B5"
	UusdcBalanceForGasExemption = 2_000_000_000 // 2_000 Noble USDC
)

// Returns true if the transaction includes a single MsgExecuteContract message.
// False otherwise.
func IsSingleWasmExecTx(
	sdkTx sdk.Tx,
) bool {
	msgs := sdkTx.GetMsgs()
	if len(msgs) > 1 {
		return false
	}

	for _, msg := range msgs {
		if sdk.MsgTypeURL(msg) == "/cosmwasm.wasm.v1.MsgExecuteContract" {
			return true
		}
	}
	return false
}

// Returns true if USDC balance is sufficient for gas exemption.
func WasmExecExemptFromGas(
	bankKeeper types.BankKeeper,
	ctx sdk.Context,
	deductFeeFromAcc sdk.AccountI,
) bool {
	balance := bankKeeper.GetBalance(ctx, deductFeeFromAcc.GetAddress(), UusdcDenom)
	return balance.Amount.GTE(math.NewInt(UusdcBalanceForGasExemption))
}
