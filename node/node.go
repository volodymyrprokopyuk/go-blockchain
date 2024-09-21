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
  Password string
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
  // components
  state *chain.State
  stateSync *stateSync
  grpcSrv *grpc.Server
  dis *discovery
  txRelay *msgRelay[chain.SigTx, grpcMsgRelay[chain.SigTx]]
  prop *proposer
  blkRelay *msgRelay[chain.Block, grpcMsgRelay[chain.Block]]
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
  // components
  disCfg := discoveryCfg{
    bootstrap: nd.cfg.Bootstrap,
    nodeAddr: nd.cfg.NodeAddr,
    seedAddr: nd.cfg.SeedAddr,
  }
  nd.dis = newDiscovery(nd.ctx, nd.wg, disCfg)
  nd.stateSync = newStateSync(nd.ctx, nd.cfg, nd.dis)
  nd.txRelay = newMsgRelay(nd.ctx, nd.wg, 100, nd.dis, grpcTxRelay)
  nd.blkRelay = newMsgRelay(nd.ctx, nd.wg, 10, nd.dis, grpcBlockRelay)
  nd.prop = newProposer(nd.ctx, nd.wg, nd.blkRelay)
  return nd
}

func (n *Node) Start() error {
  defer n.ctxCancel()
  sta, err := n.stateSync.syncState()
  if err != nil {
    return err
  }
  n.state = sta
  n.prop.state = n.state
  n.wg.Add(1)
  go n.servegRPC()
  n.wg.Add(1)
  go n.dis.discoverPeers(30 * time.Second)
  n.wg.Add(1)
  go n.txRelay.relayMsgs(30 * time.Second)
  n.wg.Add(1)
  go n.prop.proposeBlocks(10 * time.Second)
  n.wg.Add(1)
  go n.blkRelay.relayMsgs(30 * time.Second)
  select {
  case err = <- n.chErr:
  case <- n.ctx.Done():
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
  fmt.Printf("* gRPC address %v\n", n.cfg.NodeAddr)
  n.grpcSrv = grpc.NewServer()
  nd := rpc.NewNodeSrv(n.dis)
  rpc.RegisterNodeServer(n.grpcSrv, nd)
  acc := rpc.NewAccountSrv(n.cfg.KeyStoreDir, n.state)
  rpc.RegisterAccountServer(n.grpcSrv, acc)
  tx := rpc.NewTxSrv(n.cfg.KeyStoreDir, n.state.Pending, n.txRelay)
  rpc.RegisterTxServer(n.grpcSrv, tx)
  blk := rpc.NewBlockSrv(n.cfg.BlockStoreDir, n.state, n.blkRelay)
  rpc.RegisterBlockServer(n.grpcSrv, blk)
  err = n.grpcSrv.Serve(lis)
  if err != nil {
    n.chErr <- err
    return
  }
}
