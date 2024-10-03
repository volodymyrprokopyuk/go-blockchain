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
  // Re-create the authority account
  path := filepath.Join(keyStoreDir, string(gen.Authority))
  auth, err := chain.ReadAccount(path, []byte(authPass))
  if err != nil {
    return err
  }
  // Re-create the initial owner account
  ownerAcc, _ := genesisAccount(gen)
  path = filepath.Join(keyStoreDir, string(ownerAcc))
  acc, err := chain.ReadAccount(path, []byte(ownerPass))
  if err != nil {
    return err
  }
  // Create and persist a new auxiliary account
  aux, err := chain.NewAccount()
  err = aux.Write(keyStoreDir, []byte(ownerPass))
  if err != nil {
    return err
  }
  // Define transactions for blocks
  blocks := [][]struct{
    from, to chain.Account
    value uint64
  }{
    {{acc, aux, 2}, {aux, acc, 1}},
    {{acc, aux, 4}, {aux, acc, 3}},
  }
  for _, txs := range blocks {
    for _, t := range txs {
      // Create a new transaction
      tx := chain.NewTx(
        t.from.Address(), t.to.Address(), t.value,
        state.Pending.Nonce(t.from.Address()) + 1,
      )
      // Sign the new transaction
      stx, err := t.from.SignTx(tx)
      if err != nil {
        return err
      }
      // Apply the signed transaction to the pending state
      err = state.Pending.ApplyTx(stx)
      if err != nil {
        return err
      }
    }
    // Create a new block on the cloned state
    clone := state.Clone()
    blk, err := clone.CreateBlock(auth)
    if err != nil {
      return err
    }
    // Validate the new block on the cloned state
    clone = state.Clone()
    err = clone.ApplyBlock(blk)
    if err != nil {
      return err
    }
    // Apply the cloned state to the confirmed state
    state.Apply(clone)
    // Persist the confirmed block to the local block store
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
  // Create and persist the genesis
  gen, err := createGenesis()
  if err != nil {
    t.Fatal(err)
  }
  // Create the blockchain state
  state := chain.NewState(gen)
  // Set up the gRPC server and client
  conn := grpcClientConn(t, func(grpcSrv *grpc.Server) {
    blk := rpc.NewBlockSrv(blockStoreDir, nil, state, nil)
    rpc.RegisterBlockServer(grpcSrv, blk)
  })
  ctx := context.Background()
  // Create the gRPC block client
  cln := rpc.NewBlockClient(conn)
  req := &rpc.GenesisSyncReq{}
  // Call the GenesysSync method
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
  // Verify the genesis
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
  // Create and persist the genesis
  gen, err := createGenesis()
  if err != nil {
    t.Fatal(err)
  }
  // Create the blockchain state
  state := chain.NewState(gen)
  lastBlock := state.LastBlock()
  // Create several confirmed blocks
  err = createBlocks(gen, state)
  if err != nil {
    t.Fatal(err)
  }
  // Set up the gRPC server and client
  conn := grpcClientConn(t, func(grpcSrv *grpc.Server) {
    blk := rpc.NewBlockSrv(blockStoreDir, nil, state, nil)
    rpc.RegisterBlockServer(grpcSrv, blk)
  })
  ctx := context.Background()
  // Create the gRPC block client
  cln := rpc.NewBlockClient(conn)
  // Call the BlockSync method
  req := &rpc.BlockSyncReq{Number: lastBlock.Number + 1}
  stream, err := cln.BlockSync(ctx, req)
  if err != nil {
    t.Fatal(err)
  }
  for {
    // Receive blocks from the server stream
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
    // Verify the signature of the receive block
    valid, err := chain.VerifyBlock(blk, state.Authority())
    if err != nil {
      t.Fatal(err)
    }
    if !valid {
      t.Fatalf("invalid block signature")
    }
    // Check the correct block number regarding the last confirmed block
    got, exp := blk.Number, lastBlock.Number + 1
    if got != exp {
      t.Fatalf("invalid block number: expected %v, got %v", exp, got)
    }
    lastBlock = blk
  }
}
