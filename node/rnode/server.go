package rnode

import (
	context "context"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"google.golang.org/grpc"
)

type Discoverer interface {
  Bootstrap() bool
  AddPeers(peers ...string);
  Peers() []string;
}

type NodeSrv struct {
  UnimplementedNodeServer
  blockStoreDir string
  dis Discoverer
}

func NewNodeSrv(blockStoreDir string, dis Discoverer) *NodeSrv {
  return &NodeSrv{blockStoreDir: blockStoreDir, dis: dis}
}

func (s *NodeSrv) GenesisSync(
  _ context.Context, req *GenesisSyncReq,
) (*GenesisSyncRes, error) {
  jsgen, err := chain.ReadGenesisBytes(s.blockStoreDir)
  if err != nil {
    return nil, err
  }
  res := &GenesisSyncRes{Genesis: jsgen}
  return res, nil
}

func (s *NodeSrv) BlockSync(
  req *BlockSyncReq, stream grpc.ServerStreamingServer[BlockSyncRes],
) error {
  blocks, closeBlocks, err := chain.ReadBlocksBytes(s.blockStoreDir)
  if err != nil {
    return err
  }
  defer closeBlocks()
  num, i := int(req.Number), 1
  for err, jblk := range blocks {
    if err != nil {
      return err
    }
    if i >= num {
      res := &BlockSyncRes{Block: jblk}
      err = stream.Send(res)
      if err != nil {
        return err
      }
    }
    i++
  }
  return nil
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
