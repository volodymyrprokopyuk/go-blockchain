package command

import "github.com/spf13/cobra"

func ChainCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use: "bcn",
    Short: "Manages the blockchain",
    Long: `Node, account, transaction management, and block querying`,
    Example: `bcn node | account | tx | block <command> [opts...]`,
    Version: "0.1.0",
    SilenceUsage: true,
    SilenceErrors: true,
  }
  cmd.PersistentFlags().String("node", "localhost:1122", "node address host:port")
  cmd.AddCommand(nodeCmd(), accountCmd(), txCmd(), blockCmd())
  return cmd
}
