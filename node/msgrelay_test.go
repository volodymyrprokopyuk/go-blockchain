package node_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/node"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestTxRelay(t *testing.T) {
  defer os.RemoveAll(bootKeyStoreDir)
  defer os.RemoveAll(bootBlockStoreDir)
  defer os.RemoveAll(keyStoreDir)
  defer os.RemoveAll(blockStoreDir)
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()
  wg := new(sync.WaitGroup)
  // Create peer discovery for the bootstrap node
  peerDiscCfg := node.PeerDiscoveryCfg{NodeAddr: bootAddr, Bootstrap: true}
  bootPeerDisc := node.NewPeerDiscovery(ctx, wg, peerDiscCfg)
  // Create state sync for the bootstrap node
  nodeCfg := node.NodeCfg{
    NodeAddr: bootAddr, Bootstrap: true,
    KeyStoreDir: bootKeyStoreDir, BlockStoreDir: bootBlockStoreDir,
    Chain: chainName, AuthPass: authPass,
    OwnerPass: ownerPass, Balance: ownerBalance,
  }
  bootStateSync := node.NewStateSync(ctx, nodeCfg, bootPeerDisc)
  // Initialize the state on the bootstrap node by creating the genesis
  bootState, err := bootStateSync.SyncState()
  if err != nil {
    t.Fatal(err)
  }
  // Create transaction relay for the bootstrap node
  bootTxRelay := node.NewMsgRelay(
    ctx, wg, 10, node.GRPCTxRelay, false, bootPeerDisc,
  )
  // Start the transaction relay for the bootstrap node
  wg.Add(1)
  go bootTxRelay.RelayMsgs(100 * time.Millisecond)
  // Start the gRPC server on the bootstrap node
  grpcStartSvr(t, bootAddr, func(grpcSrv *grpc.Server) {
    node := rpc.NewNodeSrv(bootPeerDisc, nil)
    rpc.RegisterNodeServer(grpcSrv, node)
    tx := rpc.NewTxSrv(
      bootKeyStoreDir, bootBlockStoreDir, bootState.Pending, bootTxRelay,
    )
    rpc.RegisterTxServer(grpcSrv, tx)
    blk := rpc.NewBlockSrv(bootBlockStoreDir, nil, bootState, nil)
    rpc.RegisterBlockServer(grpcSrv, blk)
  })
  // Create peer discovery for the new node
  peerDiscCfg = node.PeerDiscoveryCfg{NodeAddr: nodeAddr, SeedAddr: bootAddr}
  nodePeerDisc := node.NewPeerDiscovery(ctx, wg, peerDiscCfg)
  // Start peer discover on the new node
  wg.Add(1)
  go nodePeerDisc.DiscoverPeers(100 * time.Millisecond)
  // Wait for the peer discovery to discover peers
  time.Sleep(150 * time.Millisecond)
  // Create state sync for the new node
  nodeCfg = node.NodeCfg{
    NodeAddr: nodeAddr, SeedAddr: bootAddr,
    KeyStoreDir: keyStoreDir, BlockStoreDir: blockStoreDir,
  }
  nodeStateSync := node.NewStateSync(ctx, nodeCfg, nodePeerDisc)
  // Synchronize the state on the new node by fetching the genesis and confirmed
  // blocks from the bootstrap node
  nodeState, err := nodeStateSync.SyncState()
  if err != nil {
    t.Fatal(err)
  }
  // Start the gRPC server on the new node
  grpcStartSvr(t, nodeAddr, func(grpcSrv *grpc.Server) {
    tx := rpc.NewTxSrv(keyStoreDir, blockStoreDir, nodeState.Pending, nil)
    rpc.RegisterTxServer(grpcSrv, tx)
  })
  // Wait for the gRPC server of the new node to start
  time.Sleep(100 * time.Millisecond)
  gen, err := chain.ReadGenesis(bootBlockStoreDir)
  if err != nil {
    t.Fatal(err)
  }
  // Get the initial owner account and its balance from the genesis
  ownerAcc, ownerBal := genesisAccount(gen)
  // Re-create the initial owner account from the genesis
  path := filepath.Join(bootKeyStoreDir, string(ownerAcc))
  acc, err := chain.ReadAccount(path, []byte(ownerPass))
  if err != nil {
    t.Fatal(err)
  }
  // Create the gRPC client connected to the bootstrap node
  conn, err := grpc.NewClient(
    bootAddr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    t.Fatal(err)
  }
  defer conn.Close()
  // Create the gRPC transaction client
  cln := rpc.NewTxClient(conn)
  // Sign and send transactions to the bootstrap node
  for _, value := range []uint64{12, 34} {
    // Create a new transaction
    tx := chain.NewTx(
      acc.Address(), chain.Address("to"), value,
      bootState.Pending.Nonce(acc.Address()) + 1,
    )
    // Sign the new transaction
    stx, err := acc.SignTx(tx)
    if err != nil {
      t.Fatal(err)
    }
    // Encode the new transaction
    jtx, err := json.Marshal(stx)
    if err != nil {
      t.Fatal(err)
    }
    // Send the signed transaction to the bootstrap node
    req := &rpc.TxSendReq{Tx: jtx}
    _, err = cln.TxSend(ctx, req)
    if err != nil {
      t.Fatal(err)
    }
    // Wait for the sent transaction to be applied to the pending state
    time.Sleep(50 * time.Millisecond)
  }
  expBalance := ownerBal - 12 - 34
  nodeBalance, exist := nodeState.Pending.Balance(acc.Address())
  if !exist {
    t.Fatalf("balance does not exist on the new node")
  }
  if nodeBalance != expBalance {
    t.Errorf(
      "invalid node balance: expected %v, got %v", expBalance, nodeBalance,
    )
  }
  bootBalance, exist := bootState.Pending.Balance(acc.Address())
  if !exist {
    t.Fatalf("balance does not exist on the bootstrap node")
  }
  if bootBalance != expBalance {
    t.Errorf(
      "invalid bootstrap balance: expected %v, got %v", expBalance, bootBalance,
    )
  }
}
