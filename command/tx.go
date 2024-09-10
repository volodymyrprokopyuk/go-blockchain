package command

import (
	"context"
	"encoding/hex"
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

func txSign(addr, from, to string, value uint, pwd string) ([]byte, error) {
  conn, err := grpc.NewClient(
    addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return nil, err
  }
  defer conn.Close()
  cln := rtx.NewTxClient(conn)
  req := &rtx.TxSignReq{From: from, To: to, Value: uint64(value), Password: pwd}
  res, err := cln.TxSign(context.Background(), req)
  if err != nil {
    return nil, err
  }
  return res.SigTx, nil
}

func txSignCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use: "sign",
    Short: "Signs a transaction with the private key of the account",
    RunE: func(cmd *cobra.Command, _ []string) error {
      addr, _ := cmd.Flags().GetString("node")
      from, _ := cmd.Flags().GetString("from")
      to, _ := cmd.Flags().GetString("to")
      value, _ := cmd.Flags().GetUint("value")
      pwd, _ := cmd.Flags().GetString("password")
      jsnSTx, err := txSign(addr, from, to, value, pwd)
      if err != nil {
        return err
      }
      fmt.Printf("%s\n", jsnSTx)
      return nil
    },
  }
  cmd.Flags().StringP("from", "", "", "debtor address")
  cmd.MarkFlagRequired("from")
  cmd.Flags().StringP("to", "", "", "creditor address")
  cmd.MarkFlagRequired("to")
  cmd.Flags().UintP("value", "", 0, "transfer amount")
  cmd.MarkFlagRequired("value")
  cmd.Flags().StringP(
    "password", "p", "", "password to encrypt the account private key",
  )
  cmd.MarkFlagRequired("password")
  return cmd
}

func txSend(addr, sigtx string) ([]byte, error) {
  conn, err := grpc.NewClient(
    addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return nil, err
  }
  defer conn.Close()
  cln := rtx.NewTxClient(conn)
  req := &rtx.TxSendReq{SigTx: []byte(sigtx)}
  res, err := cln.TxSend(context.Background(), req)
  if err != nil {
    return nil, err
  }
  return res.Hash, nil
}

func txSendCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use: "send",
    Short: "send a signed transaction",
    RunE: func(cmd *cobra.Command, _ []string) error {
      addr, _ := cmd.Flags().GetString("node")
      stx, _ := cmd.Flags().GetString("sigtx")
      hash, err := txSend(addr, stx)
      if err != nil {
        return err
      }
      fmt.Printf("%s\n", hex.EncodeToString(hash))
      return nil
    },
  }
  cmd.Flags().StringP("sigtx", "", "", "signed transaction")
  cmd.MarkFlagRequired("sigtx")
  return cmd
}
