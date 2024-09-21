package command

import (
	"context"
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/volodymyrprokopyuk/go-blockchain/node"
)

func nodeCmd(ctx context.Context) *cobra.Command {
  cmd := &cobra.Command{
    Use: "node",
    Short: "Manages the blockchain node",
  }
  cmd.AddCommand(nodeStartCmd(ctx))
  return cmd
}

func nodeStartCmd(_ context.Context) *cobra.Command {
  cmd := &cobra.Command{
    Use: "start",
    Short: "Starts the blockchain node",
    RunE: func(cmd *cobra.Command, _ []string) error {
      nodeAddr, _ := cmd.Flags().GetString("node")
      reAddr := regexp.MustCompile(`[-\.\w]+:\d+`)
      if !reAddr.MatchString(nodeAddr) {
        return fmt.Errorf("expected --node host:port, got %v", nodeAddr)
      }
      bootstrap, _ := cmd.Flags().GetBool("bootstrap")
      seedAddr, _ := cmd.Flags().GetString("seed")
      if !bootstrap && len(seedAddr) == 0 {
        return fmt.Errorf("either --bootstrap or --seed <node> must be provided")
      }
      if !bootstrap && !reAddr.MatchString(seedAddr) {
        return fmt.Errorf("expected --seed host:port, got %v", seedAddr)
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
      name, _ := cmd.Flags().GetString("chain")
      pass, _ := cmd.Flags().GetString("password")
      balance, _ := cmd.Flags().GetUint64("balance")
      cfg := node.NodeCfg{
        NodeAddr: nodeAddr, Bootstrap: bootstrap, SeedAddr: seedAddr,
        KeyStoreDir: keyStoreDir, BlockStoreDir: blockStoreDir,
        Chain: name, Password: pass, Balance: balance,
      }
      nd := node.NewNode(cfg)
      return nd.Start()
    },
  }
  cmd.Flags().Bool("bootstrap", false, "peer discovery bootstrap node")
  cmd.Flags().String("seed", "", "peer discovery seed address host:port")
  cmd.MarkFlagsMutuallyExclusive("bootstrap", "seed")
  cmd.MarkFlagsOneRequired("bootstrap", "seed")
  cmd.Flags().String("keystore", "", "key store directory")
  cmd.Flags().String("blockstore", "", "block store directory")
  cmd.Flags().String("chain", "blockchain", "name of the blockchain")
  cmd.Flags().String("password", "", "password for the genesis account")
  cmd.Flags().Uint64("balance", 0, "initial balance for the genesis account")
  cmd.MarkFlagsRequiredTogether("password", "balance")
  return cmd
}
