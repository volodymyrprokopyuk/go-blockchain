package cli

import (
	"context"

	"github.com/spf13/cobra"
)

func ChainCmd(ctx context.Context) *cobra.Command {
  cmd := &cobra.Command{
    Use: "bcn",
    Short: "Manages the blockchain",
    Long: `Node, account, transaction management, and block querying`,
    Example: `bcn node | account | tx | block <command> [opts...]`,
    Version: "0.1.0",
    SilenceUsage: true,
    SilenceErrors: true,
  }
  cmd.PersistentFlags().String("node", "", "node address host:port")
  cmd.MarkFlagRequired("node")
  cmd.AddCommand(nodeCmd(ctx), accountCmd(ctx), txCmd(ctx), blockCmd(ctx))
  return cmd
}
