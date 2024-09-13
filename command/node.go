package command

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/volodymyrprokopyuk/go-blockchain/node"
)

func nodeCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use: "node",
    Short: "Manages the blockchain node",
  }
  cmd.PersistentFlags().String("keystore", ".keystore", "key store directory")
  cmd.PersistentFlags().String("blockstore", ".blockstore", "block store directory")
  cmd.AddCommand(nodeStartCmd())
  return cmd
}

func nodeStartCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use: "start",
    Short: "Starts the blockchain node",
    RunE: func(cmd *cobra.Command, _ []string) error {
      keyStoreDir, _ := cmd.Flags().GetString("keystore")
      blockStoreDir, _ := cmd.Flags().GetString("blockstore")
      nodeAddr, _ := cmd.Flags().GetString("node")
      bootstrap, _ := cmd.Flags().GetBool("bootstrap")
      seedAddr, _ := cmd.Flags().GetString("seed")
      if !bootstrap && len(seedAddr) == 0 {
        return fmt.Errorf("either --bootstrap or --seed must be provided")
      }
      cfg := node.NodeCfg{
        KeyStoreDir: keyStoreDir, BlockStoreDir: blockStoreDir,
        NodeAddr: nodeAddr, Bootstrap: bootstrap, SeedAddr: seedAddr,
      }
      nd := node.NewNode(cfg)
      return nd.Start()
    },
  }
  cmd.Flags().Bool("bootstrap", false, "peer discovery bootstrap node")
  cmd.Flags().String("seed", "", "peer discovery seed address host:port")
  cmd.MarkFlagsMutuallyExclusive("bootstrap", "seed")
  return cmd
}
