package rpc_test

import (
	context "context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	grpc "google.golang.org/grpc"
)

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
  // Call the BlockSync method to get the server stream of confirmed blocks
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
