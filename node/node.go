package node

import (
	"context"
	"fmt"
	"net"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	"google.golang.org/grpc"
)

type NodeCfg struct {
  // Addressing
  NodeAddr string
  Bootstrap bool
  SeedAddr string
  // Stores
  KeyStoreDir string
  BlockStoreDir string
  // Genesis
  Chain string
  AuthPass string
  OwnerPass string
  Balance uint64
}

type Node struct {
  // Configuration
  cfg NodeCfg
  // Graceful shutdown
  ctx context.Context
  ctxCancel func()
  wg *sync.WaitGroup
  chErr chan error
  // Event stream
  evStream *EventStream
  // Blockchain state
  state *chain.State
  stateSync *StateSync
  // gRPC server
  grpcSrv *grpc.Server
  // Peer discovery
  peerDisc *PeerDiscovery
  // Transaction relay
  txRelay *MsgRelay[chain.SigTx, GRPCMsgRelay[chain.SigTx]]
  // Block proposer
  blockProp *BlockProposer
  // Message relay
  blkRelay *MsgRelay[chain.SigBlock, GRPCMsgRelay[chain.SigBlock]]
}

func NewNode(cfg NodeCfg) *Node {
  ctx, cancel := signal.NotifyContext(
    context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL,
  )
  wg := new(sync.WaitGroup)
  evStream := NewEventStream(ctx, wg, 100)
  peerDiscCfg := PeerDiscoveryCfg{
    NodeAddr: cfg.NodeAddr, Bootstrap: cfg.Bootstrap, SeedAddr: cfg.SeedAddr,
  }
  peerDisc := NewPeerDiscovery(ctx, wg, peerDiscCfg)
  stateSync := NewStateSync(ctx, cfg, peerDisc)
  txRelay := NewMsgRelay(ctx, wg, 100, GRPCTxRelay, false, peerDisc)
  blkRelay := NewMsgRelay(ctx, wg, 10, GRPCBlockRelay, true, peerDisc)
  blockProp := NewBlockProposer(ctx, wg, blkRelay)
  return &Node{
    cfg: cfg, ctx: ctx, ctxCancel: cancel, wg: wg, chErr: make(chan error, 1),
    evStream: evStream, stateSync: stateSync, peerDisc: peerDisc,
    txRelay: txRelay, blockProp: blockProp, blkRelay: blkRelay,
  }
}

func (n *Node) Start() error {
  defer n.ctxCancel()
  n.wg.Add(1)
  go n.evStream.StreamEvents()
  state, err := n.stateSync.SyncState()
  if err != nil {
    return err
  }
  n.state = state
  n.wg.Add(1)
  go n.servegRPC()
  n.wg.Add(1)
  go n.peerDisc.DiscoverPeers(5 * time.Second)
  n.wg.Add(1)
  go n.txRelay.RelayMsgs(5 * time.Second)
  if n.cfg.Bootstrap {
    path := filepath.Join(n.cfg.KeyStoreDir, string(n.state.Authority()))
    auth, err := chain.ReadAccount(path, []byte(n.cfg.AuthPass))
    if err != nil {
      return err
    }
    n.blockProp.SetAuthority(auth)
    n.blockProp.SetState(n.state)
    n.wg.Add(1)
    go n.blockProp.ProposeBlocks(10 * time.Second)
  }
  n.wg.Add(1)
  go n.blkRelay.RelayMsgs(5 * time.Second)
  select {
  case <- n.ctx.Done():
  case err = <- n.chErr:
    fmt.Println(err)
  }
  n.ctxCancel() // restore default signal handling
  n.grpcSrv.GracefulStop()
  n.wg.Wait()
  return err
}

func (n *Node) servegRPC() {
  defer n.wg.Done()
  lis, err := net.Listen("tcp", n.cfg.NodeAddr)
  if err != nil {
    n.chErr <- err
    return
  }
  defer lis.Close()
  fmt.Printf("<=> gRPC %v\n", n.cfg.NodeAddr)
  n.grpcSrv = grpc.NewServer()
  node := rpc.NewNodeSrv(n.peerDisc, n.evStream)
  rpc.RegisterNodeServer(n.grpcSrv, node)
  acc := rpc.NewAccountSrv(n.cfg.KeyStoreDir, n.state)
  rpc.RegisterAccountServer(n.grpcSrv, acc)
  tx := rpc.NewTxSrv(
    n.cfg.KeyStoreDir, n.cfg.BlockStoreDir, n.state.Pending, n.txRelay,
  )
  rpc.RegisterTxServer(n.grpcSrv, tx)
  blk := rpc.NewBlockSrv(n.cfg.BlockStoreDir, n.evStream, n.state, n.blkRelay)
  rpc.RegisterBlockServer(n.grpcSrv, blk)
  err = n.grpcSrv.Serve(lis)
  if err != nil {
    n.chErr <- err
    return
  }
}
