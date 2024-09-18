package rpc

import (
	context "context"
)

type Discoverer interface {
  Bootstrap() bool
  AddPeers(peers ...string);
  Peers() []string;
}

type NodeSrv struct {
  UnimplementedNodeServer
  dis Discoverer
}

func NewNodeSrv(dis Discoverer) *NodeSrv {
  return &NodeSrv{dis: dis}
}

func (s *NodeSrv) PeerDiscover(
  _ context.Context, req *PeerDiscoverReq,
) (*PeerDiscoverRes, error) {
  if s.dis.Bootstrap() {
    s.dis.AddPeers(req.Peer)
  }
  res := &PeerDiscoverRes{Peers: s.dis.Peers()}
  return res, nil
}
