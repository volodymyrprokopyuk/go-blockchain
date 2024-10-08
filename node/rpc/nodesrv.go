package rpc

import (
	context "context"
	"encoding/json"
	"fmt"
	"math/rand"
	"slices"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PeerDiscoverer interface {
  Bootstrap() bool
  AddPeers(peers ...string);
  Peers() []string;
}

type EventStreamer interface {
  AddSubscriber(sub string) chan chain.Event
  RemoveSubscriber(sub string)
}

type NodeSrv struct {
  UnimplementedNodeServer
  peerDisc PeerDiscoverer
  evStreamer EventStreamer
}

func NewNodeSrv(peerDisc PeerDiscoverer, evStreamer EventStreamer) *NodeSrv {
  return &NodeSrv{peerDisc: peerDisc, evStreamer: evStreamer}
}

func (s *NodeSrv) PeerDiscover(
  _ context.Context, req *PeerDiscoverReq,
) (*PeerDiscoverRes, error) {
  if s.peerDisc.Bootstrap() {
    s.peerDisc.AddPeers(req.Peer)
  }
  peers := s.peerDisc.Peers()
  res := &PeerDiscoverRes{Peers: peers}
  return res, nil
}

func (s *NodeSrv) StreamSubscribe(
  req *StreamSubscribeReq, stream grpc.ServerStreamingServer[StreamSubscribeRes],
) error {
  sub := fmt.Sprint(rand.Intn(999999))
  chStream := s.evStreamer.AddSubscriber(sub)
  defer s.evStreamer.RemoveSubscriber(sub)
  for {
    select {
    case <- stream.Context().Done():
      return nil
    case event, open := <- chStream:
      if !open {
        return nil
      }
      if slices.Contains(req.EventTypes, uint64(0)) ||
        slices.Contains(req.EventTypes, uint64(event.Type)) {
        jev, err := json.Marshal(event)
        if err != nil {
          fmt.Println(err)
          continue
        }
        res := &StreamSubscribeRes{Event: jev}
        err = stream.Send(res)
        if err != nil {
          return status.Errorf(codes.Internal, err.Error())
        }
      }
    }
  }
}
