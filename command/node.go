package command

import (
	"github.com/spf13/cobra"
	"github.com/volodymyrprokopyuk/go-blockchain/node"
)

func nodeCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use: "node",
    Short: "Manages the blockchain node",
  }
  cmd.AddCommand(nodeStartCmd())
  return cmd
}

func nodeStartCmd() *cobra.Command {
  return &cobra.Command{
    Use: "start",
    Short: "Starts the blockchain node",
    RunE: func(cmd *cobra.Command, _ []string) error {
      keyStore, _ := cmd.Flags().GetString("keystore")
      blockStore, _ := cmd.Flags().GetString("blockstore")
      addr, _ := cmd.Flags().GetString("node")
      nd := node.NewNode(keyStore, blockStore, addr)
      return nd.Start()
    },
  }
}
