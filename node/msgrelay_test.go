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

func createTxRelay(
  ctx context.Context, wg *sync.WaitGroup, peerReader node.PeerReader,
) *node.MsgRelay[chain.SigTx, node.GRPCMsgRelay[chain.SigTx]] {
  txRelay := node.NewMsgRelay(ctx, wg, 10, node.GRPCTxRelay, false, peerReader)
  wg.Add(1)
  go txRelay.RelayMsgs(100 * time.Millisecond)
  return txRelay
}

func sendTxs(
  t *testing.T, ctx context.Context, acc chain.Account, values []uint64,
  pending *chain.State, nodeAddr string,
) {
  conn, err := grpc.NewClient(
    nodeAddr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    t.Fatal(err)
  }
  defer conn.Close()
  cln := rpc.NewTxClient(conn)
  for _, value := range values {
    tx := chain.NewTx(
      acc.Address(), chain.Address("to"), value,
      pending.Nonce(acc.Address()) + 1,
    )
    stx, err := acc.SignTx(tx)
    if err != nil {
      t.Fatal(err)
    }
    jtx, err := json.Marshal(stx)
    if err != nil {
      t.Fatal(err)
    }
    req := &rpc.TxSendReq{Tx: jtx}
    _, err = cln.TxSend(ctx, req)
    if err != nil {
      t.Fatal(err)
    }
    time.Sleep(50 * time.Millisecond)
  }
}

func TestTxRelay(t *testing.T) {
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
  // Create and start the transaction relay for the bootstrap node
  bootTxRelay := createTxRelay(ctx, wg, bootPeerDisc)
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
  // Create and start the peer discovery for the new node
  nodePeerDisc := createPeerDiscovery(ctx, wg, false, true)
  // Wait for the peer discovery to discover peers
  time.Sleep(150 * time.Millisecond)
  // Synchronize the state on the new node by fetching the genesis and confirmed
  // blocks from the bootstrap node
  nodeState, err := createStateSync(ctx, nodePeerDisc, false)
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
  // Get the initial owner account and its balance from the genesis
  gen, err := chain.ReadGenesis(bootBlockStoreDir)
  if err != nil {
    t.Fatal(err)
  }
  ownerAcc, ownerBal := genesisAccount(gen)
  // Re-create the initial owner account from the genesis
  path := filepath.Join(bootKeyStoreDir, string(ownerAcc))
  acc, err := chain.ReadAccount(path, []byte(ownerPass))
  if err != nil {
    t.Fatal(err)
  }
  // Sign and send several signed transactions to the bootstrap node
  sendTxs(t, ctx, acc, []uint64{12, 34}, bootState.Pending, bootAddr)
  // Verify that the initial account balance on the pending states of the new
  // node and the bootstrap node are equal
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
