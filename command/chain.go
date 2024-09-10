package command

import "github.com/spf13/cobra"

func ChainCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use: "chain",
    Short: "Manages the blockchain",
    Long: `Node administration, account management, tx management, block querying`,
    Example: `chain node | store | account | tx <command> [opts...]`,
    Version: "0.1.0",
    SilenceUsage: true,
    SilenceErrors: true,
  }
  cmd.PersistentFlags().StringP(
    "keystore", "k", ".keystore", "key store directory",
  )
  cmd.PersistentFlags().StringP(
    "blockstore", "b", ".blockstore", "block store directory",
  )
  cmd.PersistentFlags().StringP(
    "node", "n", "localhost:1122", "node address host:port",
  )
  cmd.AddCommand(nodeCmd(), storeCmd(), accountCmd(), txCmd())
  return cmd
}
