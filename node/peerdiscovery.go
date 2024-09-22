package node

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type peerDiscoveryCfg struct {
  nodeAddr string
  bootstrap bool
  seedAddr string
}

type peerDiscovery struct {
  cfg peerDiscoveryCfg
  ctx context.Context
  wg *sync.WaitGroup
  mtx sync.RWMutex
  peers map[string]struct{}
}

func newPeerDiscovery(
  ctx context.Context, wg *sync.WaitGroup, cfg peerDiscoveryCfg,
) *peerDiscovery {
  peerDisc := &peerDiscovery{
    ctx: ctx, wg: wg, cfg: cfg, peers: make(map[string]struct{}),
  }
  if !peerDisc.Bootstrap() {
    peerDisc.AddPeers(peerDisc.cfg.seedAddr)
  }
  return peerDisc
}

func (d *peerDiscovery) Bootstrap() bool {
  return d.cfg.bootstrap
}

func (d *peerDiscovery) AddPeers(peers ...string) {
  d.mtx.Lock()
  defer d.mtx.Unlock()
  for _, peer := range peers {
    if peer != d.cfg.nodeAddr {
      d.peers[peer] = struct{}{}
    }
  }
}

func (d *peerDiscovery) Peers() []string {
  d.mtx.RLock()
  defer d.mtx.RUnlock()
  peers := make([]string, 0, len(d.peers))
  for peer := range d.peers {
    peers = append(peers, peer)
  }
  return peers
}

func (d *peerDiscovery) SelfPeers() []string {
  return append(d.Peers(), d.cfg.nodeAddr)
}

func (d *peerDiscovery) grpcPeerDiscover(peer string) ([]string, error) {
  conn, err := grpc.NewClient(
    peer, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return nil, err
  }
  defer conn.Close()
  cln := rpc.NewNodeClient(conn)
  req := &rpc.PeerDiscoverReq{Peer: d.cfg.nodeAddr}
  res, err := cln.PeerDiscover(d.ctx, req)
  if err != nil {
    return nil, err
  }
  return res.Peers, nil
}

func (d *peerDiscovery) discoverPeers(period time.Duration) {
  defer d.wg.Done()
  tick := time.NewTicker(5 * time.Second) // initial early peer discovery
  defer tick.Stop()
  reset := false
  for {
    select {
    case <- d.ctx.Done():
      return
    case <- tick.C:
      if !reset {
        tick.Reset(period)
        reset = true
      }
      for _, peer := range d.Peers() {
        if peer != d.cfg.nodeAddr {
          peers, err := d.grpcPeerDiscover(peer)
          if err != nil {
            fmt.Println(err)
            continue
          }
          d.AddPeers(peers...)
        }
      }
      // fmt.Printf("* Peers discovered: %v => %v\n", d.cfg.nodeAddr, d.Peers())
    }
  }
}
