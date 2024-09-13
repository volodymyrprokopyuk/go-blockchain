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
	"github.com/volodymyrprokopyuk/go-blockchain/node/rnode"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rstore"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rtx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type NodeCfg struct {
  KeyStoreDir string
  BlockStoreDir string
  NodeAddr string
  Bootstrap bool
  SeedAddr string
}

type Node struct {
  // configure
  cfg NodeCfg
  // context
  ctx context.Context
  ctxCancel func()
  wg sync.WaitGroup
  chErr chan error
  // components
  state *state.State
  grpcSrv *grpc.Server
  peers map[string]struct{}
  peersMtx *sync.RWMutex
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
  nd.chErr = make(chan error, 1)
  // components
  nd.peers = make(map[string]struct{})
  if !cfg.Bootstrap {
    nd.AddPeers(cfg.SeedAddr)
  }
  nd.peersMtx = new(sync.RWMutex)
  return nd
}

func (n *Node) Bootstrap() bool {
  return n.cfg.Bootstrap
}

func (n *Node) AddPeers(peers ...string) {
  // n.peersMtx.Lock()
  // defer n.peersMtx.Unlock()
  for _, peer := range peers {
    if peer != n.cfg.NodeAddr {
      n.peers[peer] = struct{}{}
    }
  }
}

func (n *Node) Peers() []string {
  // n.peersMtx.RLock()
  // defer n.peersMtx.RUnlock()
  peers := make([]string, 0, len(n.peers))
  for peer := range n.peers {
    peers = append(peers, peer)
  }
  return peers
}

func (n *Node) Start() error {
  defer n.ctxCancel()
  err := n.readState()
  if err != nil {
    return err
  }
  go n.servegRPC()
  go n.discoverPeers(5 * time.Second)
  // go n.mine(5 * time.Second)
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
  nd := rnode.NewNodeSrv(n)
  rnode.RegisterNodeServer(n.grpcSrv, nd)
  sto := rstore.NewStoreSrv(n.cfg.KeyStoreDir, n.cfg.BlockStoreDir)
  rstore.RegisterStoreServer(n.grpcSrv, sto)
  acc := raccount.NewAccountSrv(n.cfg.KeyStoreDir)
  raccount.RegisterAccountServer(n.grpcSrv, acc)
  tx := rtx.NewTxSrv(n.cfg.KeyStoreDir, n.state)
  rtx.RegisterTxServer(n.grpcSrv, tx)
  err = n.grpcSrv.Serve(lis)
  if err != nil {
    n.chErr <- err
  }
}

func (n *Node) grpcPeerDiscover(peer string) ([]string, error) {
  conn, err := grpc.NewClient(
    peer, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return nil, err
  }
  defer conn.Close()
  cln := rnode.NewNodeClient(conn)
  req := &rnode.PeerDiscoverReq{Peer: n.cfg.NodeAddr}
  res, err := cln.PeerDiscover(n.ctx, req)
  if err != nil {
    return nil, err
  }
  return res.Peers, nil
}

func (n *Node) discoverPeers(interval time.Duration) {
  n.wg.Add(1)
  defer n.wg.Done()
  tick := time.NewTicker(interval)
  defer tick.Stop()
  for {
    select {
    case <- n.ctx.Done():
      return
    case <- tick.C:
      for _, peer := range n.Peers() {
        peers, err := n.grpcPeerDiscover(peer)
        if err != nil {
          fmt.Println(err)
          continue
        }
        n.AddPeers(peers...)
      }
      fmt.Printf("%v peers: %v\n", n.cfg.NodeAddr, n.Peers())
    }
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
