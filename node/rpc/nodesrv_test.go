package rpc_test

import (
	context "context"
	"slices"
	sync "sync"
	"testing"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/node"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	grpc "google.golang.org/grpc"
)

const (
  bootAddr = "localhost:1122"
  nodeAddr = "localhost:1123"
)

func createPeerDiscovery(
  ctx context.Context, wg *sync.WaitGroup, bootstrap, start bool,
) *node.PeerDiscovery {
  var peerDiscCfg node.PeerDiscoveryCfg
  if bootstrap {
    peerDiscCfg = node.PeerDiscoveryCfg{NodeAddr: bootAddr, Bootstrap: true}
  } else {
    peerDiscCfg = node.PeerDiscoveryCfg{NodeAddr: nodeAddr, SeedAddr: bootAddr}
  }
  peerDisc := node.NewPeerDiscovery(ctx, wg, peerDiscCfg)
  if start {
    wg.Add(1)
    go peerDisc.DiscoverPeers(100 * time.Millisecond)
  }
  return peerDisc
}

func TestPeerDiscover(t *testing.T) {
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()
  wg := new(sync.WaitGroup)
  // Create the peer discovery without starting for the bootstrap node
  bootPeerDisc := createPeerDiscovery(ctx, wg, true, false)
  // Set up the gRPC server and client for the bootstrap node
  conn := grpcClientConn(t, func(grpcSrv *grpc.Server) {
    node := rpc.NewNodeSrv(bootPeerDisc, nil)
    rpc.RegisterNodeServer(grpcSrv, node)
  })
  // Create the gRPC node client
  cln := rpc.NewNodeClient(conn)
  // Call the PeerDiscover method to discover peers
  req := &rpc.PeerDiscoverReq{Peer: nodeAddr}
  res, err := cln.PeerDiscover(context.Background(), req)
  if err != nil {
    t.Fatal(err)
  }
  // Verify that the new node address is returned in the list of discovered
  // peers
  if !slices.Contains(res.Peers, nodeAddr) {
    t.Errorf("peer not found: expected %v, got %v", nodeAddr, res.Peers)
  }
}
