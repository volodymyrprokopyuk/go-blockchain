package command

import (
	"fmt"
	"regexp"

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
  cmd := &cobra.Command{
    Use: "start",
    Short: "Starts the blockchain node",
    RunE: func(cmd *cobra.Command, _ []string) error {
      nodeAddr, _ := cmd.Flags().GetString("node")
      reAddr := regexp.MustCompile(`[-\.\w]+:\d+`)
      if !reAddr.MatchString(nodeAddr) {
        return fmt.Errorf("expected node host:port, got %v", nodeAddr)
      }
      rePort := regexp.MustCompile(`\d+$`)
      port := rePort.FindString(nodeAddr)
      keyStoreDir, _ := cmd.Flags().GetString("keystore")
      if len(keyStoreDir) == 0 {
        keyStoreDir = ".keystore" + port
      }
      blockStoreDir, _ := cmd.Flags().GetString("blockstore")
      if len(blockStoreDir) == 0 {
        blockStoreDir = ".blockstore" + port
      }
      bootstrap, _ := cmd.Flags().GetBool("bootstrap")
      seedAddr, _ := cmd.Flags().GetString("seed")
      if !bootstrap && len(seedAddr) == 0 {
        return fmt.Errorf("either --bootstrap or --seed <node> must be provided")
      }
      cfg := node.NodeCfg{
        KeyStoreDir: keyStoreDir, BlockStoreDir: blockStoreDir,
        NodeAddr: nodeAddr, Bootstrap: bootstrap, SeedAddr: seedAddr,
      }
      nd := node.NewNode(cfg)
      return nd.Start()
    },
  }
  cmd.Flags().String("keystore", "", "key store directory")
  cmd.Flags().String("blockstore", "", "block store directory")
  cmd.Flags().Bool("bootstrap", false, "peer discovery bootstrap node")
  cmd.Flags().String("seed", "", "peer discovery seed address host:port")
  cmd.MarkFlagsMutuallyExclusive("bootstrap", "seed")
  return cmd
}
