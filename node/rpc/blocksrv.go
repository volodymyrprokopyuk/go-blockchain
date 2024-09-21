package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

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
  stream grpc.ClientStreamingServer[BlockReceiveReq, BlockReceiveRes],
) error {
  for {
    req, err := stream.Recv()
    if err == io.EOF {
      res := &BlockReceiveRes{}
      return stream.SendAndClose(res)
    }
    if err != nil {
      return err
    }
    var blk chain.Block
    err = json.Unmarshal(req.Block, &blk)
    if err != nil {
      fmt.Println(err)
      continue
    }
    fmt.Printf("* Block Receive\n%v\n", blk)
    err = s.blkApplier.ApplyBlockToState(blk)
    if err != nil {
      fmt.Println(err)
      continue
    }
    err = blk.Write(s.blockStoreDir)
    if err != nil {
      fmt.Println(err)
      continue
    }
    s.blkRelayer.RelayBlock(blk)
  }
}

func (s *BlockSrv) BlockSearch(
  req *BlockSearchReq, stream grpc.ServerStreamingServer[BlockSearchRes],
) error {
  blocks, closeBlocks, err := chain.ReadBlocks(s.blockStoreDir)
  if err != nil {
    return err
  }
  defer closeBlocks()
  prefix := strings.HasPrefix
  for err, blk := range blocks {
    if err != nil {
      return err
    }
    if req.Number != 0 && blk.Number == req.Number ||
      len(req.BlockHash) > 0 && prefix(blk.Hash().String(), req.BlockHash) ||
      len(req.ParentHash) > 0 && prefix(blk.Parent.String(), req.ParentHash) {
      jblk, err := json.Marshal(blk)
      if err != nil {
        return err
      }
      res := &BlockSearchRes{Block: jblk}
      err = stream.Send(res)
      if err != nil {
        return err
      }
      break
    }
  }
  return nil
}
