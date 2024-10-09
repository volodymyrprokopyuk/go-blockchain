package node_test

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/node"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func createEventStream(
  ctx context.Context, wg *sync.WaitGroup,
) *node.EventStream {
  evStream := node.NewEventStream(ctx, wg, 10)
  wg.Add(1)
  go evStream.StreamEvents()
  return evStream
}

func TestEventStream(t *testing.T) {
  defer os.RemoveAll(bootKeyStoreDir)
  defer os.RemoveAll(bootBlockStoreDir)
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()
  wg := new(sync.WaitGroup)
  // Create the peer discovery without starting for the bootstrap node
  bootPeerDisc := createPeerDiscovery(ctx, wg, true, false)
  // Initialize the state on the bootstrap node by creating the genesis
  bootState, err := createStateSync(ctx, bootPeerDisc, true)
  if err != nil {
    t.Fatal(err)
  }
  // Create and start the block relay for the bootstrap node
  bootBlkRelay := createBlockRelay(ctx, wg, bootPeerDisc)
  // Re-create the authority account from the genesis to sign blocks
  path := filepath.Join(bootKeyStoreDir, string(bootState.Authority()))
  auth, err := chain.ReadAccount(path, []byte(authPass))
  if err != nil {
    t.Fatal(err)
  }
  // Create and start the block proposer on the bootstrap node
  _ = createBlockProposer(ctx, wg, bootBlkRelay, auth, bootState)
  // Create and start the event stream on the bootstrap node
  evStream := createEventStream(ctx, wg)
  // Start the gRPC server on the bootstrap node
  grpcStartSvr(t, bootAddr, func(grpcSrv *grpc.Server) {
    node := rpc.NewNodeSrv(bootPeerDisc, evStream)
    rpc.RegisterNodeServer(grpcSrv, node)
    tx := rpc.NewTxSrv(
      bootKeyStoreDir, bootBlockStoreDir, bootState.Pending, nil,
    )
    rpc.RegisterTxServer(grpcSrv, tx)
    blk := rpc.NewBlockSrv(bootBlockStoreDir, evStream, bootState, bootBlkRelay)
    rpc.RegisterBlockServer(grpcSrv, blk)
  })
  // Wait for the gRPC server of the bootstrap node to start
  time.Sleep(100 * time.Millisecond)
  // Get the initial owner account and its balance from the genesis
  gen, err := chain.ReadGenesis(bootBlockStoreDir)
  if err != nil {
    t.Fatal(err)
  }
  ownerAcc, _ := genesisAccount(gen)
  // Re-create the initial owner account from the genesis
  path = filepath.Join(bootKeyStoreDir, string(ownerAcc))
  acc, err := chain.ReadAccount(path, []byte(ownerPass))
  if err != nil {
    t.Fatal(err)
  }
  // Sign and send several signed transactions to the bootstrap node
  sendTxs(t, ctx, acc, []uint64{12, 34}, bootState.Pending, bootAddr)
  // Set up a gRPC client connection with the bootstrap node
  conn, err := grpc.NewClient(
    bootAddr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    t.Fatal(err)
  }
  // Create the gRPC node client
  cln := rpc.NewNodeClient(conn)
  // Call the StreamSubscribe method to subscribe to the node event stream and
  // establish the gRPC server stream of domain events
  req := &rpc.StreamSubscribeReq{EventTypes: []uint64{0}}
  stream, err := cln.StreamSubscribe(ctx, req)
  if err != nil {
    t.Fatal(err)
  }
  // Define the expected events to receive after a successful block proposal
  expEvents := []chain.Event{
    {Type: chain.EvBlock, Action: "validated", Body: nil},
    {Type: chain.EvTx, Action: "validated", Body: nil},
    {Type: chain.EvTx, Action: "validated", Body: nil},
  }
  // Start consuming events from the gRPC server stream of domain events
  for i := range len(expEvents) {
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
    exp := expEvents[i]
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
