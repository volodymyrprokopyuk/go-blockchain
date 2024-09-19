package rpc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"google.golang.org/grpc"
)

type BlockApplier interface {
  ApplyBlockToState(blk chain.Block) error
}

type BlockRelayer interface {
  RelayBlock(blk chain.Block)
}

type BlockSrv struct {
  UnimplementedBlockServer
  blockStoreDir string
  blkApplier BlockApplier
  blkRelayer BlockRelayer
}

func NewBlockSrv(
  blockStoreDir string, blkApplicer BlockApplier,
  blkRelayer BlockRelayer,
) *BlockSrv {
  return &BlockSrv{
    blockStoreDir: blockStoreDir, blkApplier: blkApplicer,
    blkRelayer: blkRelayer,
  }
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

func (s *BlockSrv) BlockReceive(
  _ context.Context, req *BlockReceiveReq,
) (*BlockReceiveRes, error) {
  var blk chain.Block
  err := json.Unmarshal(req.Block, &blk)
  if err != nil {
    return nil, err
  }
  fmt.Printf("* Block Receive\n%v\n", blk)
  err = s.blkApplier.ApplyBlockToState(blk)
  if err != nil {
    return nil, err
  }
  err = blk.Write(s.blockStoreDir)
  if err != nil {
    return nil, err
  }
  s.blkRelayer.RelayBlock(blk)
  res := &BlockReceiveRes{}
  return res, nil
}
