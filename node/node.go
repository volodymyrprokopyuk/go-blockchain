package node

import (
	"context"
	"fmt"
	"net"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	"google.golang.org/grpc"
)

type NodeCfg struct {
  // addressing
  NodeAddr string
  Bootstrap bool
  SeedAddr string
  // stores
  KeyStoreDir string
  BlockStoreDir string
  // genesis
  Chain string
  AuthPass string
  OwnerPass string
  Balance uint64
}

type Node struct {
  // configure
  cfg NodeCfg
  // context
  ctx context.Context
  ctxCancel func()
  wg *sync.WaitGroup
  chErr chan error
  // events
  evStream *eventStream
  // components
  state *chain.State
  stateSync *StateSync
  grpcSrv *grpc.Server
  peerDisc *PeerDiscovery
  txRelay *MsgRelay[chain.SigTx, GRPCMsgRelay[chain.SigTx]]
  blockProp *BlockProposer
  blkRelay *MsgRelay[chain.SigBlock, GRPCMsgRelay[chain.SigBlock]]
}

func NewNode(cfg NodeCfg) *Node {
  // configure
  nd := &Node{cfg: cfg}
  // context
  ctx, cancel := signal.NotifyContext(
    context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL,
  )
  nd.ctx = ctx
  nd.ctxCancel = cancel
  nd.wg = new(sync.WaitGroup)
  nd.chErr = make(chan error, 1)
  // events
  nd.evStream = newEventStream(nd.ctx, nd.wg, 100)
  // components
  peerDiscCfg := PeerDiscoveryCfg{
    NodeAddr: nd.cfg.NodeAddr,
    Bootstrap: nd.cfg.Bootstrap,
    SeedAddr: nd.cfg.SeedAddr,
  }
  nd.peerDisc = NewPeerDiscovery(nd.ctx, nd.wg, peerDiscCfg)
  nd.stateSync = NewStateSync(nd.ctx, nd.cfg, nd.peerDisc)
  nd.txRelay = NewMsgRelay(nd.ctx, nd.wg, 100, GRPCTxRelay, false, nd.peerDisc)
  nd.blkRelay = NewMsgRelay(nd.ctx, nd.wg, 10, GRPCBlockRelay, true, nd.peerDisc)
  nd.blockProp = NewBlockProposer(nd.ctx, nd.wg, nd.blkRelay)
  return nd
}

func (n *Node) Start() error {
  defer n.ctxCancel()
  // n.wg.Add(1)
  // go n.evStream.streamEvents()
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
  // if n.cfg.Bootstrap {
  //   path := filepath.Join(n.cfg.KeyStoreDir, string(n.state.Authority()))
  //   auth, err := chain.ReadAccount(path, []byte(n.cfg.AuthPass))
  //   if err != nil {
  //     return err
  //   }
  //   n.blockProp.authority = auth
  //   n.blockProp.state = n.state
  //   n.wg.Add(1)
  //   go n.blockProp.proposeBlocks(10 * time.Second)
  // }
  // n.wg.Add(1)
  // go n.blkRelay.relayMsgs()
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
