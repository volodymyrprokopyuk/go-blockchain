package command

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rtx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func txCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use: "tx",
    Short: "Manages transactions on the blockchain",
  }
  cmd.AddCommand(txSignCmd(), txSendCmd())
  return cmd
}

func txSign(addr, from, to string, value uint64, pwd string) ([]byte, error) {
  conn, err := grpc.NewClient(
    addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return nil, err
  }
  defer conn.Close()
  cln := rtx.NewTxClient(conn)
  req := &rtx.TxSignReq{From: from, To: to, Value: value, Password: pwd}
  res, err := cln.TxSign(context.Background(), req)
  if err != nil {
    return nil, err
  }
  return res.SigTx, nil
}

func txSignCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use: "sign",
    Short: "Signs a transaction",
    RunE: func(cmd *cobra.Command, _ []string) error {
      addr, _ := cmd.Flags().GetString("node")
      from, _ := cmd.Flags().GetString("from")
      to, _ := cmd.Flags().GetString("to")
      value, _ := cmd.Flags().GetUint64("value")
      pwd, _ := cmd.Flags().GetString("password")
      jstx, err := txSign(addr, from, to, value, pwd)
      if err != nil {
        return err
      }
      fmt.Printf("%s\n", jstx)
      return nil
    },
  }
  cmd.Flags().String("from", "", "debtor address")
  _ = cmd.MarkFlagRequired("from")
  cmd.Flags().String("to", "", "creditor address")
  _ = cmd.MarkFlagRequired("to")
  cmd.Flags().Uint64("value", 0, "transfer amount")
  _ = cmd.MarkFlagRequired("value")
  cmd.Flags().String("password", "", "password to encrypt the account private key")
  _ = cmd.MarkFlagRequired("password")
  return cmd
}

func txSend(addr, sigtx string) (string, error) {
  conn, err := grpc.NewClient(
    addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return "", err
  }
  defer conn.Close()
  cln := rtx.NewTxClient(conn)
  req := &rtx.TxSendReq{SigTx: []byte(sigtx)}
  res, err := cln.TxSend(context.Background(), req)
  if err != nil {
    return "", err
  }
  return res.Hash, nil
}

func txSendCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use: "send",
    Short: "Sends a signed transaction",
    RunE: func(cmd *cobra.Command, _ []string) error {
      addr, _ := cmd.Flags().GetString("node")
      stx, _ := cmd.Flags().GetString("sigtx")
      hash, err := txSend(addr, stx)
      if err != nil {
        return err
      }
      fmt.Printf("%s\n", hash)
      return nil
    },
  }
  cmd.Flags().String("sigtx", "", "signed transaction")
  _ = cmd.MarkFlagRequired("sigtx")
  return cmd
}
