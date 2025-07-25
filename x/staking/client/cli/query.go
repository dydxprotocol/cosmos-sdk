package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// CmdQueryProposers queries the set of validators eligible to propose blocks
func CmdQueryProposers() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposers",
		Short: "Query the set of validators eligible to propose blocks",
		Long:  "Query the set of validator operator addresses that are eligible to propose blocks.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Proposers(cmd.Context(), &types.QueryProposersRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
