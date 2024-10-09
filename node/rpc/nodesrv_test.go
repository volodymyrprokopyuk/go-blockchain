package rpc_test

import (
	context "context"
	"encoding/json"
	"io"
	"slices"
	sync "sync"
	"testing"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
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

func createEventStream(
  ctx context.Context, wg *sync.WaitGroup,
) *node.EventStream {
  evStream := node.NewEventStream(ctx, wg, 10)
  wg.Add(1)
  go evStream.StreamEvents()
  return evStream
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

func TestStreamSubscribe(t *testing.T) {
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()
  wg := new(sync.WaitGroup)
  // Create and start the event stream on the node
  evStream := createEventStream(ctx, wg)
  // Set up the gRPC server and client
  conn := grpcClientConn(t, func(grpcSrv *grpc.Server) {
    node := rpc.NewNodeSrv(nil, evStream)
    rpc.RegisterNodeServer(grpcSrv, node)
  })
  // Create the gRPC node client
  cln := rpc.NewNodeClient(conn)
  // Call the StreamSubscribe method to subscribe to the node event stream and
  // establish the gRPC server stream of domain events
  req := &rpc.StreamSubscribeReq{EventTypes: []uint64{0}}
  stream, err := cln.StreamSubscribe(ctx, req)
  if err != nil {
    t.Fatal(err)
  }
  // Start publishing domain events to the node event stream through the event
  // publisher interface
  events := []chain.Event{
    {Type: chain.EvTx, Action: "validated", Body: nil},
    {Type: chain.EvBlock, Action: "validated", Body: nil},
  }
  go func(eventPub chain.EventPublisher) {
    for _, event := range events {
      time.Sleep(50 * time.Millisecond)
      eventPub.PublishEvent(event)
    }
  }(evStream)
  // Start consuming events from the gRPC server stream of domain events
  for i := range len(events) {
    // Receive a domain event
    res, err := stream.Recv()
    if err == io.EOF {
      break
    }
    if err != nil {
      t.Fatal(err)
    }
    // Decode the received domain event
    var got chain.Event
    err = json.Unmarshal(res.Event, &got)
    if err != nil {
      t.Fatal(err)
    }
    // Verify that the type and the action of the domain event are correct
    exp := events[i]
    if got.Type != exp.Type {
      t.Errorf("invalid event type: expected %v, got %v", exp.Type, got.Type)
    }
    if got.Action != exp.Action {
      t.Errorf(
        "invalid event action: expected %v, got %v", exp.Action, got.Action,
      )
    }
  }
}

