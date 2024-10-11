package rpc_test

import (
	context "context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	grpc "google.golang.org/grpc"
)

func createTxs(acc chain.Account, values []uint64, pending *chain.State) error {
  for _, value := range values {
    tx := chain.NewTx(
      acc.Address(), chain.Address("to"), value,
      pending.Nonce(acc.Address()) + 1,
    )
    stx, err := acc.SignTx(tx)
    if err != nil {
      return err
    }
    err = pending.ApplyTx(stx)
    if err != nil {
      return err
    }
  }
  return nil
}

func createBlocks(gen chain.SigGenesis, state *chain.State) error {
  path := filepath.Join(keyStoreDir, string(gen.Authority))
  auth, err := chain.ReadAccount(path, []byte(authPass))
  if err != nil {
    return err
  }
  ownerAcc, _ := genesisAccount(gen)
  path = filepath.Join(keyStoreDir, string(ownerAcc))
  acc, err := chain.ReadAccount(path, []byte(ownerPass))
  if err != nil {
    return err
  }
  aux, err := chain.NewAccount()
  err = aux.Write(keyStoreDir, []byte(ownerPass))
  if err != nil {
    return err
  }
  blocks := [][]struct{
    from, to chain.Account
    value uint64
  }{
    {{acc, aux, 2}, {aux, acc, 1}},
    {{acc, aux, 4}, {aux, acc, 3}},
  }
  for _, txs := range blocks {
    for _, t := range txs {
      tx := chain.NewTx(
        t.from.Address(), t.to.Address(), t.value,
        state.Pending.Nonce(t.from.Address()) + 1,
      )
      stx, err := t.from.SignTx(tx)
      if err != nil {
        return err
      }
      err = state.Pending.ApplyTx(stx)
      if err != nil {
        return err
      }
    }
    clone := state.Clone()
    blk, err := clone.CreateBlock(auth)
    if err != nil {
      return err
    }
    clone = state.Clone()
    err = clone.ApplyBlock(blk)
    if err != nil {
      return err
    }
    state.Apply(clone)
    err = blk.Write(blockStoreDir)
    if err != nil {
      return err
    }
  }
  return nil
}

func TestGenesisSync(t *testing.T) {
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
  // Set up the gRPC server and client
  conn := grpcClientConn(t, func(grpcSrv *grpc.Server) {
    blk := rpc.NewBlockSrv(blockStoreDir, nil, state, nil)
    rpc.RegisterBlockServer(grpcSrv, blk)
  })
  // Create the gRPC block client
  cln := rpc.NewBlockClient(conn)
  // Call the GenesysSync method to fetch the genesis
  req := &rpc.GenesisSyncReq{}
  res, err := cln.GenesisSync(ctx, req)
  if err != nil {
    t.Fatal(err)
  }
  jgen := res.Genesis
  // Decode the received genesis
  err = json.Unmarshal(jgen, &gen)
  if err != nil {
    t.Fatal(err)
  }
  // Verify that the signature of the received genesis is valid
  valid, err := chain.VerifyGen(gen)
  if err != nil {
    t.Fatal(err)
  }
  if !valid {
    t.Errorf("invalid genesis signature")
  }
}

func TestBlockSync(t *testing.T) {
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
  lastBlock := state.LastBlock()
  // Create several confirmed blocks on the state and on the local block store
  err = createBlocks(gen, state)
  if err != nil {
    t.Fatal(err)
  }
  // Set up the gRPC server and client
  conn := grpcClientConn(t, func(grpcSrv *grpc.Server) {
    blk := rpc.NewBlockSrv(blockStoreDir, nil, state, nil)
    rpc.RegisterBlockServer(grpcSrv, blk)
  })
  // Create the gRPC block client
  cln := rpc.NewBlockClient(conn)
  // Call the BlockSync method to get the gRPC server stream of confirmed blocks
  req := &rpc.BlockSyncReq{Number: lastBlock.Number + 1}
  stream, err := cln.BlockSync(ctx, req)
  if err != nil {
    t.Fatal(err)
  }
  // Start receiving confirmed blocks from the gRPC server stream
  for {
    // Receive a block from the server stream
    res, err := stream.Recv()
    if err == io.EOF {
      break
    }
    if err != nil {
      t.Fatal(err)
    }
    // Decode the received block
    jblk := res.Block
    var blk chain.SigBlock
    err = json.Unmarshal(jblk, &blk)
    if err != nil {
      t.Fatal(err)
    }
    // Verify that the signature of the received block is valid
    valid, err := chain.VerifyBlock(blk, state.Authority())
    if err != nil {
      t.Fatal(err)
    }
    if !valid {
      t.Fatalf("invalid block signature")
    }
    // Verify that the received block number and its parent hash equal to the
    // block number and the parent hash of the last confirmed block
    gotNumber, expNumber := blk.Number, lastBlock.Number + 1
    if gotNumber != expNumber {
      t.Fatalf(
        "invalid block number: expected %v, got %v", expNumber, gotNumber,
      )
    }
    gotParent, expParent := blk.Parent, lastBlock.Hash()
    if blk.Number == 1 {
      expParent = gen.Hash()
    }
    if gotParent != expParent {
      t.Fatalf("invalid parent hash: \n%v", blk)
    }
    lastBlock = blk
  }
}

func TestBlockReceive(t *testing.T) {
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
  // Get the initial owner account and its balance from the genesis
  ownerAcc, ownerBal := genesisAccount(gen)
  // Re-create the initial owner account from the genesis
  path := filepath.Join(keyStoreDir, string(ownerAcc))
  acc, err := chain.ReadAccount(path, []byte(ownerPass))
  if err != nil {
    t.Fatal(err)
  }
  // Re-create the authority account from the genesis to sign blocks
  path = filepath.Join(keyStoreDir, string(gen.Authority))
  auth, err := chain.ReadAccount(path, []byte(authPass))
  if err != nil {
    t.Fatal(err)
  }
  // Create several transactions on the pending state
  err = createTxs(acc, []uint64{12, 34}, state.Pending)
  if err != nil {
    t.Fatal(err)
  }
  // Create a new block on the cloned state
  clone := state.Clone()
  blk, err := clone.CreateBlock(auth)
  if err != nil {
    t.Fatal(err)
  }
  // Set up the gRPC server and gRPC client
  conn := grpcClientConn(t, func(grpcSrv *grpc.Server) {
    blk := rpc.NewBlockSrv(blockStoreDir, nil, state, nil)
    rpc.RegisterBlockServer(grpcSrv, blk)
  })
  // Create the gRPC block client
  cln := rpc.NewBlockClient(conn)
  // Call the BlockReceive method go get the gRPC client stream to relay
  // validated blocks
  stream, err := cln.BlockReceive(ctx)
  if err != nil {
    t.Fatal(err)
  }
  defer stream.CloseAndRecv()
  // Start relaying validated blocks to the gRPC client stream
  for _, blk := range []chain.SigBlock{blk} {
    // Encode the validated block
    jblk, err := json.Marshal(blk)
    if err != nil {
      t.Fatal(err)
    }
    // Send the encoded block over the gRPC client stream
    req := &rpc.BlockReceiveReq{Block: jblk}
    err = stream.Send(req)
    if err != nil {
      t.Fatal(err)
    }
    // Wait for the relayed block to be received and processed
    time.Sleep(50 * time.Millisecond)
  }
  // Verify that the balance of the initial owner account on the confirmed state
  // after receiving the relayed block is correct
  got, exist := state.Balance(acc.Address())
  if !exist {
    t.Errorf("balance does not exist %v", acc.Address())
  }
  exp := ownerBal - 12 - 34
  if got != exp {
    t.Errorf("invalid balance: expected %v, got %v", exp, got)
  }
}
