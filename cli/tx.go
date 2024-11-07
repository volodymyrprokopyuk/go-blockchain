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
  cmd.AddCommand(
    txSignCmd(ctx), txSendCmd(ctx), txSearchCmd(ctx),
    txProveCmd(ctx), txVerifyCmd(ctx),
  )
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
    Short: "Signs a new transaction with the owner account private key",
    RunE: func(cmd *cobra.Command, _ []string) error {
      addr, _ := cmd.Flags().GetString("node")
      from, _ := cmd.Flags().GetString("from")
      to, _ := cmd.Flags().GetString("to")
      value, _ := cmd.Flags().GetUint64("value")
      ownerPass, _ := cmd.Flags().GetString("ownerpass")
      jtx, err := grpcTxSign(ctx, addr, from, to, value, ownerPass)
      if err != nil {
        return err
      }
      fmt.Printf("%s\n", jtx)
      return nil
    },
  }
  cmd.Flags().String("from", "", "sender address")
  _ = cmd.MarkFlagRequired("from")
  cmd.Flags().String("to", "", "recipient address")
  _ = cmd.MarkFlagRequired("to")
  cmd.Flags().Uint64("value", 0, "transfer amount")
  _ = cmd.MarkFlagRequired("value")
  cmd.Flags().String("ownerpass", "", "owner account password")
  _ = cmd.MarkFlagRequired("ownerpass")
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
    Short: "Sends the signed encoded transaction",
    RunE: func(cmd *cobra.Command, _ []string) error {
      addr, _ := cmd.Flags().GetString("node")
      tx, _ := cmd.Flags().GetString("sigtx")
      hash, err := grpcTxSend(ctx, addr, tx)
      if err != nil {
        return err
      }
      fmt.Printf("tx %s\n", hash)
      return nil
    },
  }
  cmd.Flags().String("sigtx", "", "signed encoded transaction")
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
    Short:
    "Searches transactions by the transaction hash, from, to, and account address",
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
        fmt.Printf("tx  %s\n", tx.Hash())
        fmt.Printf("%v\n", tx)
      }
      if !found {
        fmt.Println("no transactions found")
      }
      return nil
    },
  }
  cmd.Flags().String("hash", "", "transaction hash prefix")
  cmd.Flags().String("from", "", "sender address")
  cmd.Flags().String("to", "", "recipient address")
  cmd.Flags().String("account", "", "involved account address")
  cmd.MarkFlagsOneRequired("hash", "from", "to", "account")
  return cmd
}

func grpcTxProve(
  ctx context.Context, addr, hash string,
) ([]byte, string, error) {
  conn, err := grpc.NewClient(
    addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return nil, "", err
  }
  defer conn.Close()
  cln := rpc.NewTxClient(conn)
  req := &rpc.TxProveReq{Hash: hash}
  res, err := cln.TxProve(ctx, req)
  if err != nil {
    return nil, "", err
  }
  return res.MerkleProof, res.MerkleRoot, nil
}

func txProveCmd(ctx context.Context) *cobra.Command {
  cmd := &cobra.Command{
    Use: "prove",
    Short:
    "Receives Merkle proof and Merkle root for transactions hash",
    RunE: func(cmd *cobra.Command, _ []string) error {
      addr, _ := cmd.Flags().GetString("node")
      hash, _ := cmd.Flags().GetString("hash")
      merkleProof, merkleRoot, err := grpcTxProve(ctx, addr, hash)
      if err != nil {
        return err
      }
      fmt.Printf("%s\n%s", merkleProof, merkleRoot)
      return nil
    },
  }
  cmd.Flags().String("hash", "", "transaction hash")
  cmd.MarkFlagsOneRequired("hash")
  return cmd
}

func grpcTxVerify(
  ctx context.Context, addr, hash, merkleProof, merkleRoot string,
) (bool, error) {
  conn, err := grpc.NewClient(
    addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return false, err
  }
  defer conn.Close()
  cln := rpc.NewTxClient(conn)
  req := &rpc.TxVerifyReq{
    Hash: hash, MerkleProof: []byte(merkleProof), MerkleRoot: merkleRoot,
  }
  res, err := cln.TxVerify(ctx, req)
  if err != nil {
    return false, err
  }
  return res.Valid, nil
}

func txVerifyCmd(ctx context.Context) *cobra.Command {
  cmd := &cobra.Command{
    Use: "verify",
    Short:
    "Verifies Merkle proof for transactions hash against Merkle root",
    RunE: func(cmd *cobra.Command, _ []string) error {
      addr, _ := cmd.Flags().GetString("node")
      hash, _ := cmd.Flags().GetString("hash")
      merkleProof, _ := cmd.Flags().GetString("mrkproof")
      merkleRoot, _ := cmd.Flags().GetString("mrkroot")
      valid, err := grpcTxVerify(ctx, addr, hash, merkleProof, merkleRoot)
      if err != nil {
        return err
      }
      strValid := "valid"
      if !valid {
        strValid = "INVALID"
      }
      fmt.Printf("tx %s %v\n", hash, strValid)
      return nil
    },
  }
  cmd.Flags().String("hash", "", "transaction hash")
  cmd.Flags().String("mrkproof", "", "Merkle proof")
  cmd.Flags().String("mrkroot", "", "Merkle root")
  cmd.MarkFlagsOneRequired("hash", "mrkproof", "mrkroot")
  return cmd
}
