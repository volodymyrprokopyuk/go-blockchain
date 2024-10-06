package node_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/node"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	"google.golang.org/grpc"
)

const (
  bootAddr = "localhost:1122"
  nodeAddr = "localhost:1123"
  bootKeyStoreDir = ".keystoreboot"
  bootBlockStoreDir = ".blockstoreboot"
  keyStoreDir = ".keystorenode"
  blockStoreDir = ".blockstorenode"
  chainName = "testblockchaint"
  authPass = "password"
  ownerPass = "password"
  ownerBalance = 1000
)

func genesisAccount(gen chain.SigGenesis) (chain.Address, uint64) {
  for acc, bal := range gen.Balances {
    return acc, bal
  }
  return "", 0
}

func createStateSync(
  ctx context.Context, peerReader node.PeerReader, bootstrap bool,
) (*chain.State, error) {
  var nodeCfg node.NodeCfg
  if bootstrap {
    nodeCfg = node.NodeCfg{
      NodeAddr: bootAddr, Bootstrap: true,
      KeyStoreDir: bootKeyStoreDir, BlockStoreDir: bootBlockStoreDir,
      Chain: chainName, AuthPass: authPass,
      OwnerPass: ownerPass, Balance: ownerBalance,
    }
  } else {
    nodeCfg = node.NodeCfg{
      NodeAddr: nodeAddr, SeedAddr: bootAddr,
      KeyStoreDir: keyStoreDir, BlockStoreDir: blockStoreDir,
    }
  }
  stateSync := node.NewStateSync(ctx, nodeCfg, peerReader)
  return stateSync.SyncState()
}

func createBlocks(
  keyStoreDir, blockStoreDir string, gen chain.SigGenesis, state *chain.State,
) error {
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

func grpcStartSvr(
  t *testing.T, nodeAddr string, grpcRegisterSrv func (grpcSrv *grpc.Server),
) {
  lis, err := net.Listen("tcp", nodeAddr)
  if err != nil {
    t.Fatal(err)
  }
  fmt.Printf("<=> gRPC test %v\n", nodeAddr)
  grpcSrv := grpc.NewServer()
  grpcRegisterSrv(grpcSrv)
  go func() {
    err := grpcSrv.Serve(lis)
    if err != nil {
      fmt.Println(err)
    }
  }()
  t.Cleanup(func() {
    lis.Close()
    grpcSrv.GracefulStop()
  })
}

func TestStateSync(t *testing.T) {
  defer os.RemoveAll(bootKeyStoreDir)
  defer os.RemoveAll(bootBlockStoreDir)
  defer os.RemoveAll(keyStoreDir)
  defer os.RemoveAll(blockStoreDir)
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()
  wg := new(sync.WaitGroup)
  // Create the peer discovery without starting for the bootstrap node
  bootPeerDisc := createPeerDiscovery(ctx, wg, true, false)
  // Initialize the state on the bootstrap node by creating the genesis
  bootState, err := createStateSync(ctx, bootPeerDisc, true)
  if err != nil {
    t.Fatal(err)
  }
  // Get the initial owner account and its balance from the genesis
  gen, err := chain.ReadGenesis(bootBlockStoreDir)
  if err != nil {
    t.Fatal(err)
  }
  ownerAcc, ownerBal := genesisAccount(gen)
  // Verify that the initial owner balance from the confirmed state on the
  // bootstrap node is equal to the initial owner balance from the genesis
  gotBalance, exist := bootState.Balance(ownerAcc)
  if !exist {
    t.Fatalf("balance does not exist")
  }
  expBalance := ownerBal
  if gotBalance != expBalance {
    t.Errorf("invalid balance: expected %v, got %v", expBalance, gotBalance)
  }
  // Create several confirmed blocks on the bootstrap node
  err = createBlocks(bootKeyStoreDir, bootBlockStoreDir, gen, bootState)
  if err != nil {
    t.Fatal(err)
  }
  // Start the gRPC server on the bootstrap node
  grpcStartSvr(t, bootAddr, func(grpcSrv *grpc.Server) {
    blk := rpc.NewBlockSrv(bootBlockStoreDir, nil, bootState, nil)
    rpc.RegisterBlockServer(grpcSrv, blk)
  })
  // Wait for the gRPC server of the bootstrap node to start
  time.Sleep(100 * time.Millisecond)
  // Create the peer discovery without starting for the new node
  nodePeerDisc := createPeerDiscovery(ctx, wg, false, false)
  // Synchronize the state on the new node by fetching the genesis and confirmed
  // blocks from the bootstrap node
  nodeState, err := createStateSync(ctx, nodePeerDisc, false)
  if err != nil {
    t.Fatal(err)
  }
  // Verify that the last block number and the last block parent on the
  // confirmed sates of the the new node and the bootstrap node are equal
  gotLastBlock, expLastBlock := nodeState.LastBlock(), bootState.LastBlock()
  if gotLastBlock.Number != expLastBlock.Number {
    t.Errorf(
      "invalid block number: expected %v, got %v",
      expLastBlock.Number, gotLastBlock.Number,
    )
  }
  if gotLastBlock.Parent != expLastBlock.Parent {
    t.Errorf("invalid block parent")
  }
}
