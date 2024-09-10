package command

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rstore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func storeCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use: "store",
    Short: "Manages the blockchain store",
  }
  cmd.AddCommand(storeInitCmd())
  return cmd
}

func storeInit(addr, chain, pwd string, bal uint64) (string, error) {
  conn, err := grpc.NewClient(
    addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return "", err
  }
  defer conn.Close()
  cln := rstore.NewStoreClient(conn)
  req := &rstore.StoreInitReq{Chain: chain, Password: pwd, Balance: bal}
  res, err := cln.StoreInit(context.Background(), req)
  if err != nil {
    return "", err
  }
  return res.Address, nil
}

func storeInitCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use: "init",
    Short: "Creates an account and writes the genesis to the blockstore",
    RunE: func(cmd *cobra.Command, _ []string) error {
      addr, _ := cmd.Flags().GetString("node")
      chain, _ := cmd.Flags().GetString("chain")
      pwd, _ := cmd.Flags().GetString("password")
      bal, _ := cmd.Flags().GetUint64("balance")
      acc, err := storeInit(addr, chain, pwd, bal)
      if err != nil {
        return err
      }
      fmt.Println(acc)
      return nil
    },
  }
  cmd.Flags().StringP(
    "chain", "", "blockchain", "name of the blockchain",
  )
  cmd.Flags().StringP(
    "password", "p", "", "password to encrypt the account private key",
  )
  _ = cmd.MarkFlagRequired("password")
  cmd.Flags().Uint64P(
    "balance", "", 0, "initial balance for the genesis account",
  )
  _ = cmd.MarkFlagRequired("balance")
  return cmd
}
