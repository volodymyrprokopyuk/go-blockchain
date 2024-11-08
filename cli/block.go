package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func blockCmd(ctx context.Context) *cobra.Command {
  cmd := &cobra.Command{
    Use: "block",
    Short: "Queries blocks on the blockchain",
  }
  cmd.AddCommand(blockSearchCmd(ctx))
  return cmd
}

func grpcBlockSearch(
  ctx context.Context, addr string, number uint64, hash, parent string,
) (func(yield func(err error, blk chain.SigBlock) bool), func(), error) {
  conn, err := grpc.NewClient(
    addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return nil, nil, err
  }
  close := func() {
    conn.Close()
  }
  cln := rpc.NewBlockClient(conn)
  req := &rpc.BlockSearchReq{Number: number, Hash: hash, Parent: parent}
  stream, err := cln.BlockSearch(ctx, req)
  if err != nil {
    return nil, nil, err
  }
  more := true
  blocks := func(yield func(err error, blk chain.SigBlock) bool) {
    for more {
      res, err := stream.Recv()
      if err == io.EOF {
        return
      }
      if err != nil {
        yield(err, chain.SigBlock{})
        return
      }
      var blk chain.SigBlock
      err = json.Unmarshal(res.Block, &blk)
      if err != nil {
        yield(err, chain.SigBlock{})
        return
      }
      more = yield(nil, blk)
    }
  }
  return blocks, close, nil
}

func blockSearchCmd(ctx context.Context) *cobra.Command {
  cmd := &cobra.Command{
    Use: "search",
    Short: "Searches blocks by the block number, block hash, and parent hash",
    RunE: func(cmd *cobra.Command, _ []string) error {
      addr, _ := cmd.Flags().GetString("node")
      number, _ := cmd.Flags().GetUint64("number")
      hash, _ := cmd.Flags().GetString("hash")
      parent, _ := cmd.Flags().GetString("parent")
      blocks, closeBlocks, err := grpcBlockSearch(ctx, addr, number, hash, parent)
      if err != nil {
        return err
      }
      defer closeBlocks()
      found := false
      for err, blk := range blocks {
        if err != nil {
          return err
        }
        found = true
        fmt.Printf("blk %s\n", blk.Hash())
        fmt.Printf("mrk %s\n", blk.MerkleRoot)
        fmt.Printf("%v", blk)
      }
      if !found {
        fmt.Println("no blocks found")
      }
      return nil
    },
  }
  cmd.Flags().Uint64("number", 0, "block number")
  cmd.Flags().String("hash", "", "block hash prefix")
  cmd.Flags().String("parent", "", "parent hash prefix")
  cmd.MarkFlagsMutuallyExclusive("number", "hash", "parent")
  cmd.MarkFlagsOneRequired("number", "hash", "parent")
  return cmd
}
