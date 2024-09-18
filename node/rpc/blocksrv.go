package rpc

import (
	"context"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"google.golang.org/grpc"
)

type BlockSrv struct {
  UnimplementedBlockServer
  blockStoreDir string
}

func NewBlockSrv(blockStoreDir string) *BlockSrv {
  return &BlockSrv{blockStoreDir: blockStoreDir}
}

func (s *BlockSrv) GenesisSync(
  _ context.Context, req *GenesisSyncReq,
) (*GenesisSyncRes, error) {
  jgen, err := chain.ReadGenesisBytes(s.blockStoreDir)
  if err != nil {
    return nil, err
  }
  res := &GenesisSyncRes{Genesis: jgen}
  return res, nil
}

func (s *BlockSrv) BlockSync(
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
