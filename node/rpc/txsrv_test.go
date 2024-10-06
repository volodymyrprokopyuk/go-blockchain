package rpc_test

import (
	context "context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
    From: string(acc.Address()), To: "to", Value: 12, Password: ownerPass,
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

func TestTxSend(t *testing.T) {
  defer os.RemoveAll(keyStoreDir)
  defer os.RemoveAll(blockStoreDir)
  // Create and persist the genesis
  gen, err := createGenesis()
  if err != nil {
    t.Fatal(err)
  }
  // Create the blockchain state
  state := chain.NewState(gen)
  // Lookup the initial owner account address and balance
  ownerAcc, ownerBal := genesisAccount(gen)
  // Re-create the initial owner account
  path := filepath.Join(keyStoreDir, string(ownerAcc))
  acc, err := chain.ReadAccount(path, []byte(ownerPass))
  if err != nil {
    t.Fatal(err)
  }
  // Set up the gRPC server and client
  conn := grpcClientConn(t, func(grpcSrv *grpc.Server) {
    tx := rpc.NewTxSrv(keyStoreDir, blockStoreDir, state.Pending, nil)
    rpc.RegisterTxServer(grpcSrv, tx)
  })
  ctx := context.Background()
  // Create the gRPC transaction client
  cln := rpc.NewTxClient(conn)
  cases := []struct{
    name string
    value uint64
    err error
  }{
    {"valid tx", 12, nil},
    {"insufficient funds", 1000, fmt.Errorf("insufficient funds")},
  }
  for _, c := range cases {
    t.Run(c.name, func(t *testing.T) {
      // Create and sign a transaction
      tx := chain.NewTx(
        acc.Address(), chain.Address("to"), c.value,
        state.Pending.Nonce(acc.Address()) + 1,
      )
      stx, err := acc.SignTx(tx)
      if err != nil {
        t.Fatal(err)
      }
      // Call the TxSend method
      jtx, err := json.Marshal(stx)
      if err != nil {
        t.Fatal(err)
      }
      req := &rpc.TxSendReq{Tx: jtx}
      res, err := cln.TxSend(ctx, req)
      if c.err == nil && err != nil {
        t.Error(err)
      }
      if c.err != nil && err == nil {
        t.Errorf("expected TxSend error, got none")
      }
      if err != nil {
        got, exp := status.Code(err), codes.FailedPrecondition
        if got != exp {
          t.Errorf("wrong error: expected %v, got %v", exp, got)
        }
      }
      if err == nil {
        got, exp := res.Hash, stx.Hash().String()
        if got != exp {
          t.Errorf("invalid transaction hash")
        }
      }
    })
  }
  // Lookup the balance of the initial owner
  got, exist := state.Pending.Balance(acc.Address())
  exp := ownerBal - 12
  if !exist {
    t.Fatalf("balance does not exist")
  }
  if got != exp {
    t.Errorf("invalid balance: expected %v, got %v", exp, got)
  }
}

func TestTxReceive(t *testing.T) {
  defer os.RemoveAll(keyStoreDir)
  defer os.RemoveAll(blockStoreDir)
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()
  // Create and persist the genesis
  gen, err := createGenesis()
  if err != nil {
    t.Fatal(err)
  }
  // Create the state from the genesis
  state := chain.NewState(gen)
  pending := state.Pending
  // Look up the initial owner account and balance from the genesis
  ownerAcc, ownerBal := genesisAccount(gen)
  path := filepath.Join(keyStoreDir, string(ownerAcc))
  // Re-create the initial owner account from the genesis
  acc, err := chain.ReadAccount(path, []byte(ownerPass))
  if err != nil {
    t.Fatal(err)
  }
  // Set up the gRPC server and gRPC client connection
  conn := grpcClientConn(t, func(grpcSrv *grpc.Server) {
    tx := rpc.NewTxSrv(keyStoreDir, blockStoreDir, pending, nil)
    rpc.RegisterTxServer(grpcSrv, tx)
  })
  // Create the gRPC transaction client
  cln := rpc.NewTxClient(conn)
  // Establish the gRPC client stream
  stream, err := cln.TxReceive(ctx)
  defer stream.CloseAndRecv()
  for _, value := range []uint64{12, 1000} {
    // Create and sign a transaction
    tx := chain.NewTx(
      acc.Address(), chain.Address("to"), value,
      pending.Nonce(acc.Address()) + 1,
    )
    stx, err := acc.SignTx(tx)
    if err != nil {
      t.Fatal(err)
    }
    // Encode the signed transaction
    jtx, err := json.Marshal(stx)
    if err != nil {
      t.Fatal(err)
    }
    // Call the gRPC TxReceive method
    req := &rpc.TxReceiveReq{Tx: jtx}
    err = stream.Send(req)
    if err != nil {
      t.Fatal(err)
    }
    // Wait for the transaction to be received and processed
    time.Sleep(100 * time.Millisecond)
  }
  // Verify the correct pending balance after receiving the transactions
  got, exist := pending.Balance(acc.Address())
  if !exist {
    t.Errorf("balance does not exist %v", acc.Address())
  }
  exp := ownerBal - 12
  if got != exp {
    t.Errorf("invalid balance: expected %v, got %v", exp, got)
  }
}
