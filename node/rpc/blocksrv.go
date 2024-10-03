package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BlockApplier interface {
  ApplyBlockToState(blk chain.SigBlock) error
}

type BlockRelayer interface {
  RelayBlock(blk chain.SigBlock)
}

type BlockSrv struct {
  UnimplementedBlockServer
  blockStoreDir string
  eventPub chain.EventPublisher
  blkApplier BlockApplier
  blkRelayer BlockRelayer
}

func NewBlockSrv(
  blockStoreDir string, eventPub chain.EventPublisher,
  blkApplier BlockApplier, blkRelayer BlockRelayer,
) *BlockSrv {
  return &BlockSrv{
    blockStoreDir: blockStoreDir, eventPub: eventPub,
    blkApplier: blkApplier, blkRelayer: blkRelayer,
  }
}

func (s *BlockSrv) GenesisSync(
  _ context.Context, req *GenesisSyncReq,
) (*GenesisSyncRes, error) {
  jgen, err := chain.ReadGenesisBytes(s.blockStoreDir)
  if err != nil {
    return nil, status.Errorf(codes.NotFound, err.Error())
  }
  res := &GenesisSyncRes{Genesis: jgen}
  return res, nil
}

func (s *BlockSrv) BlockSync(
  req *BlockSyncReq, stream grpc.ServerStreamingServer[BlockSyncRes],
) error {
  blocks, closeBlocks, err := chain.ReadBlocksBytes(s.blockStoreDir)
  if err != nil {
    return status.Errorf(codes.NotFound, err.Error())
  }
  defer closeBlocks()
  num, i := int(req.Number), 1
  for err, jblk := range blocks {
    if err != nil {
      return status.Errorf(codes.Internal, err.Error())
    }
    if i >= num {
      res := &BlockSyncRes{Block: jblk}
      err = stream.Send(res)
      if err != nil {
        return status.Errorf(codes.Internal, err.Error())
      }
    }
    i++
  }
  return nil
}

func (s *BlockSrv) publishBlock(blk chain.SigBlock) {
  jblk, _ := json.Marshal(blk)
  event := chain.NewEvent(chain.EvBlock, "validated", jblk)
  s.eventPub.PublishEvent(event)
  for _, tx := range blk.Txs {
    jtx, _ := json.Marshal(tx)
    event := chain.NewEvent(chain.EvTx, "validated", jtx)
    s.eventPub.PublishEvent(event)
  }
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
    var blk chain.SigBlock
    err = json.Unmarshal(req.Block, &blk)
    if err != nil {
      fmt.Println(err)
      continue
    }
    fmt.Printf("* Block received\n%v", blk)
    err = s.blkApplier.ApplyBlockToState(blk)
    if err != nil {
      fmt.Print(err)
      continue
    }
    err = blk.Write(s.blockStoreDir)
    if err != nil {
      fmt.Println(err)
      continue
    }
    s.blkRelayer.RelayBlock(blk)
    s.publishBlock(blk)
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
      len(req.Hash) > 0 && prefix(blk.Hash().String(), req.Hash) ||
      len(req.Parent) > 0 && prefix(blk.Parent.String(), req.Parent) {
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
