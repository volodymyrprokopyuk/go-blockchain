package node

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/chain/account"
	"github.com/volodymyrprokopyuk/go-blockchain/chain/state"
	"github.com/volodymyrprokopyuk/go-blockchain/node/raccount"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rnode"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rtx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
  wg sync.WaitGroup
  chErr chan error
  // components
  state *state.State
  grpcSrv *grpc.Server
  peers map[string]struct{}
  peersMtx sync.RWMutex
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
  return nd
}

func (n *Node) Bootstrap() bool {
  return n.cfg.Bootstrap
}

func (n *Node) AddPeers(peers ...string) {
  n.peersMtx.Lock()
  defer n.peersMtx.Unlock()
  for _, peer := range peers {
    if peer != n.cfg.NodeAddr {
      n.peers[peer] = struct{}{}
    }
  }
}

func (n *Node) Peers() []string {
  peers := make([]string, 0, len(n.peers))
  n.peersMtx.RLock()
  defer n.peersMtx.RUnlock()
  for peer := range n.peers {
    peers = append(peers, peer)
  }
  return peers
}

func (n *Node) Start() error {
  defer n.ctxCancel()
  err := n.initState()
  if err != nil {
    return err
  }
  go n.servegRPC()
  go n.discoverPeers(1 * time.Minute)
  go n.mine(10 * time.Second)
  select {
  case err = <- n.chErr:
  case <- n.ctx.Done():
  }
  n.ctxCancel() // restore default signal handling
  n.grpcSrv.GracefulStop()
  n.wg.Wait() // for discoverPeers
  return err
}

func (n *Node) createGenesis() (chain.SigGenesis, error) {
  pass := []byte(n.cfg.Password)
  if len(pass) < 5 {
    return chain.SigGenesis{}, fmt.Errorf("password length is less than 5")
  }
  if n.cfg.Balance == 0 {
    return chain.SigGenesis{}, fmt.Errorf("balance must be positive")
  }
  acc, err := account.NewAccount()
  if err != nil {
    return chain.SigGenesis{}, err
  }
  err = acc.Write(n.cfg.KeyStoreDir, pass)
  n.cfg.Password = "erase"
  if err != nil {
    return chain.SigGenesis{}, err
  }
  gen := chain.NewGenesis(n.cfg.Chain, acc.Address(), n.cfg.Balance)
  sgen, err := acc.SignGen(gen)
  if err != nil {
    return chain.SigGenesis{}, err
  }
  err = sgen.Write(n.cfg.BlockStoreDir)
  if err != nil {
    return chain.SigGenesis{}, err
  }
  return sgen, nil
}

func (n *Node) grpcGenesisSync() ([]byte, error) {
  conn, err := grpc.NewClient(
    n.cfg.SeedAddr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return nil, err
  }
  defer conn.Close()
  cln := rnode.NewNodeClient(conn)
  req := &rnode.GenesisSyncReq{}
  res, err := cln.GenesisSync(n.ctx, req)
  if err != nil {
    return nil, err
  }
  return res.Genesis, nil
}

func (n *Node) syncGenesis() (chain.SigGenesis, error) {
  jsgen, err := n.grpcGenesisSync()
  if err != nil {
    return chain.SigGenesis{}, err
  }
  var sgen chain.SigGenesis
  err = json.Unmarshal(jsgen, &sgen)
  if err != nil {
    return chain.SigGenesis{}, err
  }
  valid, err := chain.VerifyGen(sgen)
  if err != nil {
    return chain.SigGenesis{}, err
  }
  if !valid {
    return chain.SigGenesis{}, fmt.Errorf("invalid genesis signature")
  }
  err = sgen.Write(n.cfg.BlockStoreDir)
  if err != nil {
    return chain.SigGenesis{}, err
  }
  return sgen, nil
}

func (n *Node) readBlocks() error {
  blocks, closeBlocks, err := chain.ReadBlocks(n.cfg.BlockStoreDir)
  if err != nil {
    if _, assert := err.(*os.PathError); !assert {
      return err
    }
    fmt.Println("warning: blocks not yet created")
    return nil
  }
  defer closeBlocks()
  for err, blk := range blocks {
    if err != nil {
      return err
    }
    clo := n.state.Clone()
    err = clo.ApplyBlock(blk)
    if err != nil {
      return err
    }
    n.state.Apply(clo)
    n.state.ResetPending()
  }
  return nil
}

func (n *Node) grpcBlockSync(peer string) (
  func(yield (func(err error, jblk []byte) bool)), func(), error,
) {
  conn, err := grpc.NewClient(
    peer, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return nil, nil, err
  }
  close := func() {
    conn.Close()
  }
  cln := rnode.NewNodeClient(conn)
  req := &rnode.BlockSyncReq{Number: n.state.LastBlock().Number + 1}
  stream, err := cln.BlockSync(n.ctx, req)
  if err != nil {
    return nil, nil, err
  }
  more := true
  blocks := func(yield func(err error, jblk []byte) bool) {
    for more {
      res, err := stream.Recv()
      if err == io.EOF {
        return
      }
      if err != nil {
        yield(err, nil)
        return
      }
      more = yield(nil, res.Block)
    }
  }
  return blocks, close, nil
}

func (n *Node) syncBlocks() error {
  for _, peer := range n.Peers() {
    blocks, closeBlocks, err := n.grpcBlockSync(peer)
    if err != nil {
      return err
    }
    defer closeBlocks()
    for err, jblk := range blocks {
      if err != nil {
        return err
      }
      blk, err := chain.UnmarshalBlockBytes(jblk)
      if err != nil {
        return err
      }
      clo := n.state.Clone()
      err = clo.ApplyBlock(blk)
      if err != nil {
        return err
      }
      n.state.Apply(clo)
      n.state.ResetPending()
      err = blk.Write(n.cfg.BlockStoreDir)
      if err != nil {
        return err
      }
    }
  }
  return nil
}

func (n *Node) initState() error {
  sgen, err := chain.ReadGenesis(n.cfg.BlockStoreDir)
  if err != nil {
    if n.Bootstrap() {
      sgen, err = n.createGenesis()
      if err != nil {
        return err
      }
    } else {
      sgen, err = n.syncGenesis()
      if err != nil {
        return err
      }
    }
  }
  valid, err := chain.VerifyGen(sgen)
  if err != nil {
    return err
  }
  if !valid {
    return fmt.Errorf("invalid genesis signature")
  }
  n.state = state.NewState(sgen)
  err = n.readBlocks()
  if err != nil {
    return err
  }
  err = n.syncBlocks()
  if err != nil {
    return err
  }
  fmt.Printf("* Sync state (SyncBlocks)\n%v\n", n.state)
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
  nd := rnode.NewNodeSrv(n.cfg.BlockStoreDir, n)
  rnode.RegisterNodeServer(n.grpcSrv, nd)
  acc := raccount.NewAccountSrv(n.cfg.KeyStoreDir)
  raccount.RegisterAccountServer(n.grpcSrv, acc)
  tx := rtx.NewTxSrv(n.cfg.KeyStoreDir, n.state.Pending)
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
  tick := time.NewTicker(5 * time.Second)
  defer tick.Stop()
  for {
    select {
    case <- n.ctx.Done():
      return
    case <- tick.C:
      tick.Reset(interval)
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
