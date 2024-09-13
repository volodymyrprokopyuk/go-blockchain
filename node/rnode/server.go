package rnode

import context "context"

type Peerer interface {
  Bootstrap() bool
  AddPeers(peers ...string);
  Peers() []string;
}

type NodeSrv struct {
  UnimplementedNodeServer
  peerer Peerer
}

func NewNodeSrv(peerer Peerer) *NodeSrv {
  return &NodeSrv{peerer: peerer}
}

func (s *NodeSrv) PeedDiscover(
  _ context.Context, req *PeerDiscoverReq,
) (*PeerDiscoverRes, error) {
  if s.peerer.Bootstrap() {
    s.peerer.AddPeers(req.Peer)
  }
  res := &PeerDiscoverRes{Peers: s.peerer.Peers()}
  return res, nil
}
