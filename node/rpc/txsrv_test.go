package rpc_test

import (
	context "context"
	"encoding/json"
	"os"
	"testing"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	grpc "google.golang.org/grpc"
)

func createAccount() (chain.Account, error) {
  acc, err := chain.NewAccount()
  if err != nil {
    return chain.Account{}, err
  }
  err = acc.Write(keyStoreDir, []byte(ownerPass))
  if err != nil {
    return chain.Account{}, err
  }
  return acc, nil
}

func TestTxSign(t *testing.T) {
  gen, err := createGenesis()
  if err != nil {
    t.Fatal(err)
  }
  state := chain.NewState(gen)
  acc, err := createAccount()
  if err != nil {
    t.Fatal(err)
  }
  defer os.RemoveAll(keyStoreDir)
  defer os.RemoveAll(blockStoreDir)
  conn := grpcClientConn(t, func(grpcSrv *grpc.Server) {
    tx := rpc.NewTxSrv(keyStoreDir, blockStoreDir, state, nil)
    rpc.RegisterTxServer(grpcSrv, tx)
  })
  ctx := context.Background()
  cln := rpc.NewTxClient(conn)
  req := &rpc.TxSignReq{
    From: string(acc.Address()), To: "recipient", Value: 12, Password: ownerPass,
  }
  res, err := cln.TxSign(ctx, req)
  if err != nil {
    t.Fatal(err)
  }
  jtx := res.Tx
  var tx chain.SigTx
  err = json.Unmarshal(jtx, &tx)
  if err != nil {
    t.Fatal(err)
  }
  valid, err := chain.VerifyTx(tx)
  if err != nil {
    t.Fatal(err)
  }
  if !valid {
    t.Errorf("invalid tx signature")
  }
}
