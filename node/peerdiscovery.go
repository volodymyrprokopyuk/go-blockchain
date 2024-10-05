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

type PeerReader interface {
  Peers() []string
  SelfPeers() []string
}

type PeerDiscoveryCfg struct {
  NodeAddr string
  Bootstrap bool
  SeedAddr string
}

type PeerDiscovery struct {
  cfg PeerDiscoveryCfg
  ctx context.Context
  wg *sync.WaitGroup
  mtx sync.RWMutex
  peers map[string]struct{}
}

func NewPeerDiscovery(
  ctx context.Context, wg *sync.WaitGroup, cfg PeerDiscoveryCfg,
) *PeerDiscovery {
  peerDisc := &PeerDiscovery{
    ctx: ctx, wg: wg, cfg: cfg, peers: make(map[string]struct{}),
  }
  if !peerDisc.Bootstrap() {
    peerDisc.AddPeers(peerDisc.cfg.SeedAddr)
  }
  return peerDisc
}

func (d *PeerDiscovery) Bootstrap() bool {
  return d.cfg.Bootstrap
}

func (d *PeerDiscovery) AddPeers(peers ...string) {
  d.mtx.Lock()
  defer d.mtx.Unlock()
  for _, peer := range peers {
    if peer != d.cfg.NodeAddr {
      _, exist := d.peers[peer]
      if !exist {
        fmt.Printf("<=> Peer %v\n", peer)
      }
      d.peers[peer] = struct{}{}
    }
  }
}

func (d *PeerDiscovery) Peers() []string {
  d.mtx.RLock()
  defer d.mtx.RUnlock()
  peers := make([]string, 0, len(d.peers))
  for peer := range d.peers {
    peers = append(peers, peer)
  }
  return peers
}

func (d *PeerDiscovery) SelfPeers() []string {
  return append(d.Peers(), d.cfg.NodeAddr)
}

func (d *PeerDiscovery) grpcPeerDiscover(peer string) ([]string, error) {
  conn, err := grpc.NewClient(
    peer, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return nil, err
  }
  defer conn.Close()
  cln := rpc.NewNodeClient(conn)
  req := &rpc.PeerDiscoverReq{Peer: d.cfg.NodeAddr}
  res, err := cln.PeerDiscover(d.ctx, req)
  if err != nil {
    return nil, err
  }
  return res.Peers, nil
}

func (d *PeerDiscovery) DiscoverPeers(period time.Duration) {
  defer d.wg.Done()
  tick := time.NewTicker(period)
  defer tick.Stop()
  for {
    select {
    case <- d.ctx.Done():
      return
    case <- tick.C:
      for _, peer := range d.Peers() {
        if peer != d.cfg.NodeAddr {
          peers, err := d.grpcPeerDiscover(peer)
          if err != nil {
            fmt.Println(err)
            continue
          }
          d.AddPeers(peers...)
        }
      }
    }
  }
}
