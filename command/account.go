package command

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/volodymyrprokopyuk/go-blockchain/node/raccount"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func accountCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use: "account",
    Short: "Manages accounts on the blockchain",
  }
  cmd.AddCommand(accountCreateCmd())
  return cmd
}

func accountCreate(addr, pwd string) (string, error) {
  conn, err := grpc.NewClient(
    addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return "", err
  }
  defer conn.Close()
  cln := raccount.NewAccountClient(conn)
  req := &raccount.AccountCreateReq{Password: pwd}
  res, err := cln.AccountCreate(context.Background(), req)
  if err != nil {
    return "", err
  }
  return res.Address, nil
}

func accountCreateCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use: "create",
    Short: "Creates an account protected with a password",
    RunE: func(cmd *cobra.Command, _ []string) error {
      addr, _ := cmd.Flags().GetString("node")
      pwd, _ := cmd.Flags().GetString("password")
      acc, err := accountCreate(addr, pwd)
      if err != nil {
        return err
      }
      fmt.Println(acc)
      return nil
    },
  }
  cmd.Flags().StringP(
    "password", "p", "", "password to encrypt the account private key",
  )
  cmd.MarkFlagRequired("password")
  return cmd
}
