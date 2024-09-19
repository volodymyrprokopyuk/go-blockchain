package node

import (
	"context"
	"sync"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

type blockRelay struct {
  ctx context.Context
  wg *sync.WaitGroup
  chBlk chan chain.Block
  dis *discovery
}

func newBlockRelay(
  ctx context.Context, wg *sync.WaitGroup, cap int, dis *discovery,
) *blockRelay {
  return &blockRelay{
    ctx: ctx, wg: wg, chBlk: make(chan chain.Block, cap), dis: dis,
  }
}

func (r *blockRelay) RelayBlock(blk chain.Block) {
  r.chBlk <- blk
}
