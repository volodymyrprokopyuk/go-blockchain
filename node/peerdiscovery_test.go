package node_test

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/node"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	"google.golang.org/grpc"
)

func TestPeerDiscovery(t *testing.T) {
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()
  wg := new(sync.WaitGroup)
  // Create peer discovery for the bootstrap node
  peerDiscCfg := node.PeerDiscoveryCfg{NodeAddr: bootAddr, Bootstrap: true}
  bootPeerDisc := node.NewPeerDiscovery(ctx, wg, peerDiscCfg)
  // Start the gRPC server on the bootstrap node
  grpcStartSvr(t, bootAddr, func(grpcSrv *grpc.Server) {
    node := rpc.NewNodeSrv(bootPeerDisc, nil)
    rpc.RegisterNodeServer(grpcSrv, node)
  })
  // Create peer discovery for the new node
  peerDiscCfg = node.PeerDiscoveryCfg{NodeAddr: nodeAddr, SeedAddr: bootAddr}
  nodePeerDisc := node.NewPeerDiscovery(ctx, wg, peerDiscCfg)
  // Start peer discover on the new node
  wg.Add(1)
  go nodePeerDisc.DiscoverPeers(100 * time.Millisecond)
  // Wait for the peer discovery to discover peers
  time.Sleep(150 * time.Millisecond)
  if !slices.Contains(bootPeerDisc.Peers(), nodeAddr) {
    t.Errorf("node address %v is not in bootstrap known peers", nodeAddr)
  }
  if !slices.Contains(nodePeerDisc.Peers(), bootAddr) {
    t.Errorf("bootstrap address %v is not in node known peers", bootAddr)
  }
}
