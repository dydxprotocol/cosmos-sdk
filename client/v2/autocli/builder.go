package autocli

import (
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"cosmossdk.io/client/v2/autocli/flag"

	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

// Builder manages options for building CLI commands.
type Builder struct {
	// flag.Builder embeds the flag builder and its options.
	flag.Builder

	// GetClientConn specifies how CLI commands will resolve a grpc.ClientConnInterface
	// from a given context.
	GetClientConn func(*cobra.Command) (grpc.ClientConnInterface, error)

	// TxConfigOptions is required to support sign mode textual
	TxConfigOpts authtx.ConfigOptions

	// AddQueryConnFlags and AddTxConnFlags are functions that add flags to query and transaction commands
	AddQueryConnFlags func(*cobra.Command)
	AddTxConnFlags    func(*cobra.Command)
}

// ValidateAndComplete the builder fields.
// It returns an error if any of the required fields are missing.
func (b *Builder) ValidateAndComplete() error {
	return b.Builder.ValidateAndComplete()
}
