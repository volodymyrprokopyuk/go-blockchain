package node

import (
	"context"
	"fmt"
	"net"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/chain/state"
	"github.com/volodymyrprokopyuk/go-blockchain/node/raccount"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rblock"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rnode"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rtx"
	"google.golang.org/grpc"
)

type NodeCfg struct {
  NodeAddr string
  Bootstrap bool
  SeedAddr string
  KeyStoreDir string
  BlockStoreDir string
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
  state *state.State
  stateSync *stateSync
  grpcSrv *grpc.Server
  dis *discovery
  txRelay *txRelay
}

func NewNode(cfg NodeCfg) *Node {
  // configure
  nd := &Node{cfg: cfg}
  ctx, cancel := signal.NotifyContext(
    context.Background(), syscall.SIGINT, syscall.SIGINT, syscall.SIGKILL,
  )
  // context
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
  nd.txRelay = newTxRelay(nd.ctx, nd.wg, 100, nd.dis)
  return nd
}

func (n *Node) Start() error {
  defer n.ctxCancel()
  sta, err := n.stateSync.syncState()
  if err != nil {
    return err
  }
  n.state = sta
  go n.servegRPC()
  go n.dis.discoverPeers(5 * time.Second)
  go n.txRelay.relayTxs(5 * time.Second)
  // go n.mine(10 * time.Second)
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
  n.wg.Add(1)
  defer n.wg.Done()
  lis, err := net.Listen("tcp", n.cfg.NodeAddr)
  if err != nil {
    n.chErr <- err
    return
  }
  defer lis.Close()
  fmt.Printf("* gRPC address %v\n", n.cfg.NodeAddr)
  n.grpcSrv = grpc.NewServer()
  nd := rnode.NewNodeSrv(n.dis)
  rnode.RegisterNodeServer(n.grpcSrv, nd)
  acc := raccount.NewAccountSrv(n.cfg.KeyStoreDir)
  raccount.RegisterAccountServer(n.grpcSrv, acc)
  tx := rtx.NewTxSrv(n.cfg.KeyStoreDir, n.state.Pending, n.txRelay)
  rtx.RegisterTxServer(n.grpcSrv, tx)
  blk := rblock.NewBlockSrv(n.cfg.BlockStoreDir)
  rblock.RegisterBlockServer(n.grpcSrv, blk)
  err = n.grpcSrv.Serve(lis)
  if err != nil {
    n.chErr <- err
    return
  }
}

func (n *Node) mine(interval time.Duration) {
  n.wg.Add(1)
  defer n.wg.Done()
  tick := time.NewTicker(interval)
  defer tick.Stop()
  for {
    select {
    case <- n.ctx.Done():
      return
    case <- tick.C:
      // create block
      clo := n.state.Clone()
      blk := clo.CreateBlock()
      if len(blk.Txs) == 0 {
        continue
      }
      fmt.Printf("* Block\n%v\n", blk)
      // apply block
      clo = n.state.Clone()
      err := clo.ApplyBlock(blk)
      if err != nil {
        fmt.Println(err)
        continue
      }
      n.state.Apply(clo)
      fmt.Printf("* Block state (ApplyBlock)\n%v\n", n.state)
      // write block
      err = blk.Write(n.cfg.BlockStoreDir)
      if err != nil {
        fmt.Println(err)
        continue
      }
    }
  }
}
