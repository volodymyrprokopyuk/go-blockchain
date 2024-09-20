package command

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func accountCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use: "account",
    Short: "Manages accounts on the blockchain",
  }
  cmd.AddCommand(accountCreateCmd(), accountBalanceCmd())
  return cmd
}

func grpcAccountCreate(addr, pass string) (string, error) {
  conn, err := grpc.NewClient(
    addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return "", err
  }
  defer conn.Close()
  cln := rpc.NewAccountClient(conn)
  req := &rpc.AccountCreateReq{Password: pass}
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
      pass, _ := cmd.Flags().GetString("password")
      acc, err := grpcAccountCreate(addr, pass)
      if err != nil {
        return err
      }
      fmt.Printf("%v\n", acc)
      return nil
    },
  }
  cmd.Flags().String("password", "", "password to encrypt the account private key")
  _ = cmd.MarkFlagRequired("password")
  return cmd
}

func grpcAccountBalance(addr, acc string) (uint64, error) {
  conn, err := grpc.NewClient(
    addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return 0, err
  }
  defer conn.Close()
  cln := rpc.NewAccountClient(conn)
  req := &rpc.AccountBalanceReq{Address: acc}
  res, err := cln.AccountBalance(context.Background(), req)
  if err != nil {
    return 0, err
  }
  return res.Balance, nil
}

func accountBalanceCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use: "balance",
    Short: "Returns an account balance",
    RunE: func(cmd *cobra.Command, _ []string) error {
      addr, _ := cmd.Flags().GetString("node")
      acc, _ := cmd.Flags().GetString("account")
      balance, err := grpcAccountBalance(addr, acc)
      if err != nil {
        return err
      }
      fmt.Printf("%v: %v\n", acc, balance)
      return nil
    },
  }
  cmd.Flags().String("account", "", "account address")
  _ = cmd.MarkFlagRequired("account")
  return cmd
}
