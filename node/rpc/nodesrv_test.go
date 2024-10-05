package rpc_test

import (
	context "context"
	"slices"
	sync "sync"
	"testing"

	"github.com/volodymyrprokopyuk/go-blockchain/node"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	grpc "google.golang.org/grpc"
)

const (
  bootAddr = "localhost:1122"
  nodeAddr = "localhost:1123"
)

func TestPeerDiscover(t *testing.T) {
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()
  wg := new(sync.WaitGroup)
  // Create peer discovery for the bootstrap node
  peerDiscCfg := node.PeerDiscoveryCfg{NodeAddr: bootAddr, Bootstrap: true}
  peerDisc := node.NewPeerDiscovery(ctx, wg, peerDiscCfg)
  // Set up the gRPC server and client
  conn := grpcClientConn(t, func(grpcSrv *grpc.Server) {
    node := rpc.NewNodeSrv(peerDisc, nil)
    rpc.RegisterNodeServer(grpcSrv, node)
  })
  // Create the gRPC node client
  cln := rpc.NewNodeClient(conn)
  // Call the PeerDiscover method
  req := &rpc.PeerDiscoverReq{Peer: nodeAddr}
  res, err := cln.PeerDiscover(context.Background(), req)
  if err != nil {
    t.Fatal(err)
  }
  peers := res.Peers
  if !slices.Contains(peers, nodeAddr) {
    t.Errorf("peer not found: expected %v, got %v", nodeAddr, peers)
  }
}
