package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func accountCmd(ctx context.Context) *cobra.Command {
  cmd := &cobra.Command{
    Use: "account",
    Short: "Manages accounts on the blockchain",
  }
  cmd.AddCommand(accountCreateCmd(ctx), accountBalanceCmd(ctx))
  return cmd
}

func grpcAccountCreate(
  ctx context.Context, addr, ownerPass string,
) (string, error) {
  conn, err := grpc.NewClient(
    addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return "", err
  }
  defer conn.Close()
  cln := rpc.NewAccountClient(conn)
  req := &rpc.AccountCreateReq{Password: ownerPass}
  res, err := cln.AccountCreate(ctx, req)
  if err != nil {
    return "", err
  }
  return res.Address, nil
}

func accountCreateCmd(ctx context.Context) *cobra.Command {
  cmd := &cobra.Command{
    Use: "create",
    Short: "Creates an account protected with a password",
    RunE: func(cmd *cobra.Command, _ []string) error {
      addr, _ := cmd.Flags().GetString("node")
      ownerPass, _ := cmd.Flags().GetString("ownerpass")
      acc, err := grpcAccountCreate(ctx, addr, ownerPass)
      if err != nil {
        return err
      }
      fmt.Printf("acc %v\n", acc)
      return nil
    },
  }
  cmd.Flags().String("ownerpass", "", "password to encrypt the account private key")
  _ = cmd.MarkFlagRequired("ownerpass")
  return cmd
}

func grpcAccountBalance(ctx context.Context, addr, acc string) (uint64, error) {
  conn, err := grpc.NewClient(
    addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return 0, err
  }
  defer conn.Close()
  cln := rpc.NewAccountClient(conn)
  req := &rpc.AccountBalanceReq{Address: acc}
  res, err := cln.AccountBalance(ctx, req)
  if err != nil {
    return 0, err
  }
  return res.Balance, nil
}

func accountBalanceCmd(ctx context.Context) *cobra.Command {
  cmd := &cobra.Command{
    Use: "balance",
    Short: "Returns an account balance",
    RunE: func(cmd *cobra.Command, _ []string) error {
      addr, _ := cmd.Flags().GetString("node")
      acc, _ := cmd.Flags().GetString("account")
      balance, err := grpcAccountBalance(ctx, addr, acc)
      if err != nil {
        return err
      }
      fmt.Printf("acc %v: %v\n", acc, balance)
      return nil
    },
  }
  cmd.Flags().String("account", "", "account address")
  _ = cmd.MarkFlagRequired("account")
  return cmd
}
