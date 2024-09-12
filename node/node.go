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

type Node struct {
  keyStoreDir string
  blockStoreDir string
  nodeAddr string
  ctx context.Context
  state *state.State
  grpcSrv *grpc.Server
  wg sync.WaitGroup
  chErr chan error
}

func NewNode(keyStoreDir string, blockStoreDir string, nodeAddr string) *Node {
  return &Node{
    keyStoreDir: keyStoreDir, blockStoreDir: blockStoreDir, nodeAddr: nodeAddr,
    chErr: make(chan error, 1),
  }
}

func (n *Node) Start() error {
  ctx, cancel := signal.NotifyContext(
    context.Background(), syscall.SIGINT, syscall.SIGINT, syscall.SIGKILL,
  )
  n.ctx = ctx
  defer cancel()
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
  cancel() // restore default signal handling
  n.grpcSrv.GracefulStop()
  n.wg.Wait()
  return err
}

func (n *Node) readState() error {
  gen, err := store.ReadGenesis(n.blockStoreDir)
  if err != nil {
    fmt.Println("warning: genesis not found: > bcn store init, then restart")
    return nil
  }
  n.state = state.NewState(gen)
  err = n.state.ReadBlocks(n.blockStoreDir)
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
  lis, err := net.Listen("tcp", n.nodeAddr)
  if err != nil {
    n.chErr <- err
    return
  }
  defer lis.Close()
  fmt.Printf("* gRPC listening on %v\n", n.nodeAddr)
  n.grpcSrv = grpc.NewServer()
  acc := raccount.NewAccountSrv(n.keyStoreDir)
  raccount.RegisterAccountServer(n.grpcSrv, acc)
  sto := rstore.NewStoreSrv(n.keyStoreDir, n.blockStoreDir)
  rstore.RegisterStoreServer(n.grpcSrv, sto)
  tx := rtx.NewTxSrv(n.keyStoreDir, n.state)
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
      err = blk.Write(n.blockStoreDir)
      if err != nil {
        fmt.Println(err)
        continue
      }
    }
  }
}
