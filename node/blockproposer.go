package node

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

type BlockRelayer interface {
  RelayBlock(blk chain.SigBlock)
}

type BlockProposer struct {
  ctx context.Context
  wg *sync.WaitGroup
  authority chain.Account
  state *chain.State
  blkRelayer BlockRelayer
}

func NewBlockProposer(
  ctx context.Context, wg *sync.WaitGroup, blkRelayer BlockRelayer,
) *BlockProposer {
  return &BlockProposer{ctx: ctx, wg: wg, blkRelayer: blkRelayer}
}

func randPeriod(maxPeriod time.Duration) time.Duration {
  minPeriod := maxPeriod / 2
  randSpan, _ := rand.Int(rand.Reader, big.NewInt(int64(maxPeriod)))
  return minPeriod + time.Duration(randSpan.Int64())
}

func (p *BlockProposer) ProposeBlocks(maxPeriod time.Duration) {
  defer p.wg.Done()
  randPropose := time.NewTimer(randPeriod(maxPeriod))
  for {
    select {
    case <- p.ctx.Done():
      randPropose.Stop()
      return
    case <- randPropose.C:
      randPropose.Reset(randPeriod(maxPeriod))
      clone := p.state.Clone()
      blk, err := clone.CreateBlock(p.authority)
      if err != nil {
        fmt.Println(err)
        continue
      }
      if len(blk.Txs) == 0 {
        continue
      }
      clone = p.state.Clone()
      err = clone.ApplyBlock(blk)
      if err != nil {
        fmt.Println(err)
        continue
      }
      if p.blkRelayer != nil {
        p.blkRelayer.RelayBlock(blk)
      }
      fmt.Printf("==> Block proposed: %v", blk)
    }
  }
}
