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

type proposer struct {
  ctx context.Context
  wg *sync.WaitGroup
  state *chain.State
}

func newProposer(ctx context.Context, wg *sync.WaitGroup) *proposer {
  return &proposer{ctx: ctx, wg: wg}
}

func randPeriod(maxPeriod time.Duration) time.Duration {
  minPeriod := maxPeriod / 2
  randSpan, _ := rand.Int(rand.Reader, big.NewInt(int64(maxPeriod)))
  return minPeriod + time.Duration(randSpan.Int64())
}

func (p *proposer) proposeBlocks(maxPeriod time.Duration) {
  defer p.wg.Done()
  timer := time.NewTimer(randPeriod(maxPeriod))
  for {
    select {
    case <- p.ctx.Done():
      timer.Stop()
      return
    case <- timer.C:
      timer.Reset(randPeriod(maxPeriod))
      clo := p.state.Clone()
      blk := clo.CreateBlock()
      if len(blk.Txs) == 0 {
        continue
      }
      clo = p.state.Clone()
      err := clo.ApplyBlock(blk)
      if err != nil {
        fmt.Println(err)
        continue
      }
      fmt.Printf("* Proposed block: %v\n", blk)
      // p.blkRelay.relayBlock(blk)
    }
  }
}
