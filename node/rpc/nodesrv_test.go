package rpc_test

import (
	context "context"
	sync "sync"
	"testing"

	"github.com/volodymyrprokopyuk/go-blockchain/node"
)

const (
  nodeAddr = "localhost:1122"
  node2Addr = "localhost:1123"
)

func TestPeerDiscover(t *testing.T) {
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()
  wg := new(sync.WaitGroup)
  // Create peer discovery for the bootstrap node
  peerDiscCfg := node.PeerDiscoveryCfg{NodeAddr: nodeAddr, Bootstrap: true}
  peerDisc := node.NewPeerDiscovery(ctx, wg, peerDiscCfg)
}
