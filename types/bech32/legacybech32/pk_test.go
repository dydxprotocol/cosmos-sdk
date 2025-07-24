//go:build test_ledger_mock
// +build test_ledger_mock

package legacybech32

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/ledger"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestBeach32ifPbKey(t *testing.T) {
	require := require.New(t)
	path := *hd.NewFundraiserParams(0, sdk.CoinType, 0)
	priv, err := ledger.NewPrivKeySecp256k1Unsafe(path)

	// Skip test if ledger support is not available
	if err != nil && strings.Contains(err.Error(), "support for ledger devices is not available") {
		t.Skip("Skipping test: ledger support not available")
	}

	require.Nil(err, "%s", err)
	require.NotNil(priv)

	pubKeyAddr, err := MarshalPubKey(AccPK, priv.PubKey())
	require.NoError(err)
	require.Equal("cosmospub1addwnpepqd87l8xhcnrrtzxnkql7k55ph8fr9jarf4hn6udwukfprlalu8lgw0urza0",
		pubKeyAddr, "Is your device using test mnemonic: %s ?", testdata.TestMnemonic)
}
