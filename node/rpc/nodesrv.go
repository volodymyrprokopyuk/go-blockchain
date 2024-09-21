package rpc

import (
	context "context"
)

type PeerDiscoverer interface {
  Bootstrap() bool
  AddPeers(peers ...string);
  Peers() []string;
}

type NodeSrv struct {
  UnimplementedNodeServer
  peerDisc PeerDiscoverer
}

func NewNodeSrv(peerDisc PeerDiscoverer) *NodeSrv {
  return &NodeSrv{peerDisc: peerDisc}
}

func (s *NodeSrv) PeerDiscover(
  _ context.Context, req *PeerDiscoverReq,
) (*PeerDiscoverRes, error) {
  if s.peerDisc.Bootstrap() {
    s.peerDisc.AddPeers(req.Peer)
  }
  res := &PeerDiscoverRes{Peers: s.peerDisc.Peers()}
  return res, nil
}
