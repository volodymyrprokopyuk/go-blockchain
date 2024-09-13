package node

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/chain/state"
	"github.com/volodymyrprokopyuk/go-blockchain/chain/store"
	"github.com/volodymyrprokopyuk/go-blockchain/node/raccount"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rstore"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rtx"
	"google.golang.org/grpc"
)

type NodeCfg struct {
  KeyStoreDir string
  BlockStoreDir string
  NodeAddr string
  Bootstrap bool
  SeedAddr string
}

type Node struct {
  cfg NodeCfg
  // context
  ctx context.Context
  ctxCancel func()
  wg sync.WaitGroup
  chErr chan error
  // components
  state *state.State
  grpcSrv *grpc.Server
}

func NewNode(cfg NodeCfg) *Node {
  nd := &Node{cfg: cfg}
  ctx, cancel := signal.NotifyContext(
    context.Background(), syscall.SIGINT, syscall.SIGINT, syscall.SIGKILL,
  )
  nd.ctx = ctx
  nd.ctxCancel = cancel
  nd.chErr = make(chan error, 1)
  return nd
}

func (n *Node) Start() error {
  defer n.ctxCancel()
  err := n.readState()
  if err != nil {
    return err
  }
  go n.servegRPC()
  go n.mine()
  select {
  case err = <- n.chErr:
  case <- n.ctx.Done():
  }
  n.ctxCancel() // restore default signal handling
  n.grpcSrv.GracefulStop()
  n.wg.Wait()
  return err
}

func (n *Node) readState() error {
  gen, err := store.ReadGenesis(n.cfg.BlockStoreDir)
  if err != nil {
    fmt.Println("warning: genesis not found: > bcn store init, then restart")
    return nil
  }
  n.state = state.NewState(gen)
  err = n.state.ReadBlocks(n.cfg.BlockStoreDir)
  if err != nil {
    if _, assert := err.(*os.PathError); !assert {
      return err
    }
    fmt.Println("warning: blocks not yet created")
  }
  fmt.Printf("* Read state (ReadBlocks)\n%v\n", n.state)
  return nil
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
  fmt.Printf("* gRPC listening on %v\n", n.cfg.NodeAddr)
  n.grpcSrv = grpc.NewServer()
  acc := raccount.NewAccountSrv(n.cfg.KeyStoreDir)
  raccount.RegisterAccountServer(n.grpcSrv, acc)
  sto := rstore.NewStoreSrv(n.cfg.KeyStoreDir, n.cfg.BlockStoreDir)
  rstore.RegisterStoreServer(n.grpcSrv, sto)
  tx := rtx.NewTxSrv(n.cfg.KeyStoreDir, n.state)
  rtx.RegisterTxServer(n.grpcSrv, tx)
  err = n.grpcSrv.Serve(lis)
  if err != nil {
    n.chErr <- err
  }
}

func (n *Node) mine() {
  n.wg.Add(1)
  defer n.wg.Done()
  tick := time.NewTicker(5 * time.Second)
  defer tick.Stop()
  for {
    select {
    case <- n.ctx.Done():
      return
    case <- tick.C:
      // create block
      clo := n.state.Clone()
      blk := clo.CreateBlock()
      fmt.Printf("* Block\n%v\n", blk)
      if len(blk.Txs) == 0 {
        fmt.Println("warning: empty block")
        continue
      }
      // apply block
      clo = n.state.Clone()
      err := clo.ApplyBlock(blk)
      if err != nil {
        fmt.Println(err)
        continue
      }
      n.state.Apply(clo)
      n.state.ResetPending()
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
