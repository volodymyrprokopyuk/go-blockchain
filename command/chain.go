package command

import "github.com/spf13/cobra"

func ChainCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use: "chain",
    Short: "Manages the blockchain",
    Long: `Node administration, account management, tx management, block querying`,
    Example: `chain node | account | tx | block <command> [opts...]`,
    Version: "0.1.0",
    SilenceUsage: true,
    SilenceErrors: true,
  }
  cmd.AddCommand(nodeCmd(), accountCmd())
  return cmd
}
