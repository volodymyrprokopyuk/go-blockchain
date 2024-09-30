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

func TestTxSign(t *testing.T) {
  defer os.RemoveAll(keyStoreDir)
  defer os.RemoveAll(blockStoreDir)
  // Create and persist the genesis
  gen, err := createGenesis()
  if err != nil {
    t.Fatal(err)
  }
  // Create the blockchain state
  state := chain.NewState(gen)
  // Create and persist a new account
  acc, err := createAccount()
  if err != nil {
    t.Fatal(err)
  }
  // Set up the gRPC server and client
  conn := grpcClientConn(t, func(grpcSrv *grpc.Server) {
    tx := rpc.NewTxSrv(keyStoreDir, blockStoreDir, state, nil)
    rpc.RegisterTxServer(grpcSrv, tx)
  })
  ctx := context.Background()
  // Create the gRPC transaction client
  cln := rpc.NewTxClient(conn)
  // Call the TxSign method
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
  // Verify the signature of the signed transaction
  valid, err := chain.VerifyTx(tx)
  if err != nil {
    t.Fatal(err)
  }
  if !valid {
    t.Errorf("invalid transaction signature")
  }
}
