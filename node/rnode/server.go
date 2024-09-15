package rnode

import (
	context "context"
	"encoding/json"
	"fmt"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

type Peerer interface {
  Bootstrap() bool
  AddPeers(peers ...string);
  Peers() []string;
}

type NodeSrv struct {
  UnimplementedNodeServer
  blockStoreDir string
  peerer Peerer
}

func NewNodeSrv(blockStoreDir string, peerer Peerer) *NodeSrv {
  return &NodeSrv{blockStoreDir: blockStoreDir, peerer: peerer}
}

func (s *NodeSrv) GenesisSync(
  _ context.Context, req *GenesisSyncReq,
) (*GenesisSyncRes, error) {
  sgen, err := chain.ReadGenesis(s.blockStoreDir)
  if err != nil {
    fmt.Println("oh")
    return nil, err
  }
  valid, err := chain.VerifyGen(sgen)
  if err != nil {
    return nil, err
  }
  if !valid {
    return nil, fmt.Errorf("invalid genesis signature")
  }
  jsgen, err := json.Marshal(sgen)
  if err != nil {
    return nil, err
  }
  res := &GenesisSyncRes{Genesis: jsgen}
  return res, nil
}

func (s *NodeSrv) PeerDiscover(
  _ context.Context, req *PeerDiscoverReq,
) (*PeerDiscoverRes, error) {
  if s.peerer.Bootstrap() {
    s.peerer.AddPeers(req.Peer)
  }
  res := &PeerDiscoverRes{Peers: s.peerer.Peers()}
  return res, nil
}
