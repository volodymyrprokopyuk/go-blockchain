package node

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/node/rnode"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type discoveryCfg struct {
  bootstrap bool
  nodeAddr string
  seedAddr string
}

type discovery struct {
  cfg discoveryCfg
  ctx context.Context
  wg *sync.WaitGroup
  peers map[string]struct{}
  mtx sync.RWMutex
}

func newDiscovery(
  ctx context.Context, wg *sync.WaitGroup, cfg discoveryCfg,
) *discovery {
  dis := &discovery{
    ctx: ctx, wg: wg, cfg: cfg, peers: make(map[string]struct{}),
  }
  if !dis.Bootstrap() {
    dis.AddPeers(dis.cfg.seedAddr)
  }
  return dis
}

func (d *discovery) Bootstrap() bool {
  return d.cfg.bootstrap
}

func (d *discovery) AddPeers(peers ...string) {
  d.mtx.Lock()
  defer d.mtx.Unlock()
  for _, peer := range peers {
    if peer != d.cfg.nodeAddr {
      d.peers[peer] = struct{}{}
    }
  }
}

func (d *discovery) Peers() []string {
  d.mtx.RLock()
  defer d.mtx.RUnlock()
  peers := make([]string, 0, len(d.peers))
  for peer := range d.peers {
    peers = append(peers, peer)
  }
  return peers
}

func (d *discovery) grpcPeerDiscover(peer string) ([]string, error) {
  conn, err := grpc.NewClient(
    peer, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return nil, err
  }
  defer conn.Close()
  cln := rnode.NewNodeClient(conn)
  req := &rnode.PeerDiscoverReq{Peer: d.cfg.nodeAddr}
  res, err := cln.PeerDiscover(d.ctx, req)
  if err != nil {
    return nil, err
  }
  return res.Peers, nil
}

func (d *discovery) discoverPeers(interval time.Duration) {
  d.wg.Add(1)
  defer d.wg.Done()
  tick := time.NewTicker(5 * time.Second)
  defer tick.Stop()
  for {
    select {
    case <- d.ctx.Done():
      return
    case <- tick.C:
      tick.Reset(interval)
      for _, peer := range d.Peers() {
        peers, err := d.grpcPeerDiscover(peer)
        if err != nil {
          fmt.Println(err)
          continue
        }
        d.AddPeers(peers...)
      }
      fmt.Printf("* Discovery: %v => %v\n", d.cfg.nodeAddr, d.Peers())
    }
  }
}
