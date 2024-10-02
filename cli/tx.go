package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func txCmd(ctx context.Context) *cobra.Command {
  cmd := &cobra.Command{
    Use: "tx",
    Short: "Manages transactions on the blockchain",
  }
  cmd.AddCommand(txSignCmd(ctx), txSendCmd(ctx), txSearchCmd(ctx))
  return cmd
}

func grpcTxSign(
  ctx context.Context, addr, from, to string, value uint64, ownerPass string,
) ([]byte, error) {
  conn, err := grpc.NewClient(
    addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return nil, err
  }
  defer conn.Close()
  cln := rpc.NewTxClient(conn)
  req := &rpc.TxSignReq{From: from, To: to, Value: value, Password: ownerPass}
  res, err := cln.TxSign(ctx, req)
  if err != nil {
    return nil, err
  }
  return res.Tx, nil
}

func txSignCmd(ctx context.Context) *cobra.Command {
  cmd := &cobra.Command{
    Use: "sign",
    Short: "Signs a transaction",
    RunE: func(cmd *cobra.Command, _ []string) error {
      addr, _ := cmd.Flags().GetString("node")
      ownerPass, _ := cmd.Flags().GetString("ownerpass")
      from, _ := cmd.Flags().GetString("from")
      to, _ := cmd.Flags().GetString("to")
      value, _ := cmd.Flags().GetUint64("value")
      jtx, err := grpcTxSign(ctx, addr, from, to, value, ownerPass)
      if err != nil {
        return err
      }
      fmt.Printf("%s\n", jtx)
      return nil
    },
  }
  cmd.Flags().String("ownerpass", "", "password of debtor account")
  _ = cmd.MarkFlagRequired("ownerpass")
  cmd.Flags().String("from", "", "debtor address")
  _ = cmd.MarkFlagRequired("from")
  cmd.Flags().String("to", "", "creditor address")
  _ = cmd.MarkFlagRequired("to")
  cmd.Flags().Uint64("value", 0, "transfer amount")
  _ = cmd.MarkFlagRequired("value")
  return cmd
}

func grpcTxSend(ctx context.Context, addr, tx string) (string, error) {
  conn, err := grpc.NewClient(
    addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return "", err
  }
  defer conn.Close()
  cln := rpc.NewTxClient(conn)
  req := &rpc.TxSendReq{Tx: []byte(tx)}
  res, err := cln.TxSend(ctx, req)
  if err != nil {
    return "", err
  }
  return res.Hash, nil
}

func txSendCmd(ctx context.Context) *cobra.Command {
  cmd := &cobra.Command{
    Use: "send",
    Short: "Sends a signed transaction",
    RunE: func(cmd *cobra.Command, _ []string) error {
      addr, _ := cmd.Flags().GetString("node")
      tx, _ := cmd.Flags().GetString("sigtx")
      hash, err := grpcTxSend(ctx, addr, tx)
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

func grpcTxSearch(
  ctx context.Context, addr, hash, from, to, account string,
) (func(yeild func(err error, tx chain.SearchTx) bool), func(), error) {
  conn, err := grpc.NewClient(
    addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return nil, nil, err
  }
  close := func() {
    conn.Close()
  }
  cln := rpc.NewTxClient(conn)
  req := &rpc.TxSearchReq{Hash: hash, From: from, To: to, Account: account}
  stream, err := cln.TxSearch(ctx, req)
  if err != nil {
    return nil, nil, err
  }
  more := true
  txs := func(yield func(err error, tx chain.SearchTx) bool) {
    for more {
      res, err := stream.Recv()
      if err == io.EOF {
        return
      }
      if err != nil {
        yield(err, chain.SearchTx{})
        return
      }
      var tx chain.SearchTx
      err = json.Unmarshal(res.Tx, &tx)
      if err != nil {
        yield(err, chain.SearchTx{})
        return
      }
      more = yield(nil, tx)
    }
  }
  return txs, close, nil
}

func txSearchCmd(ctx context.Context) *cobra.Command {
  cmd := &cobra.Command{
    Use: "search",
    Short: "Searches transactions by transaction hash, from, to, account address",
    RunE: func(cmd *cobra.Command, _ []string) error {
      addr, _ := cmd.Flags().GetString("node")
      hash, _ := cmd.Flags().GetString("hash")
      from, _ := cmd.Flags().GetString("from")
      to, _ := cmd.Flags().GetString("to")
      account, _ := cmd.Flags().GetString("account")
      txs, closeTxs, err := grpcTxSearch(ctx, addr, hash, from, to, account)
      if err != nil {
        return err
      }
      defer closeTxs()
      found := false
      for err, tx := range txs {
        if err != nil {
          return err
        }
        found = true
        fmt.Printf("%v\n", tx)
      }
      if !found {
        fmt.Println("no transactions found")
      }
      return nil
    },
  }
  cmd.Flags().String("hash", "", "transaction hash")
  cmd.Flags().String("from", "", "debtor address")
  cmd.Flags().String("to", "", "creditor address")
  cmd.Flags().String("account", "", "account address")
  cmd.MarkFlagsOneRequired("hash", "from", "to", "account")
  return cmd
}
