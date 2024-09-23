package node

import (
	"context"
	"crypto/rand"
	"math/big"
	"sync"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

type blockProposer struct {
  ctx context.Context
  wg *sync.WaitGroup
  state *chain.State
  blkRelay *msgRelay[chain.Block, grpcMsgRelay[chain.Block]]
}

func newBlockProposer(
  ctx context.Context, wg *sync.WaitGroup,
  blkRelay *msgRelay[chain.Block, grpcMsgRelay[chain.Block]],
) *blockProposer {
  return &blockProposer{ctx: ctx, wg: wg, blkRelay: blkRelay}
}

func randPeriod(maxPeriod time.Duration) time.Duration {
  minPeriod := maxPeriod / 2
  randSpan, _ := rand.Int(rand.Reader, big.NewInt(int64(maxPeriod)))
  return minPeriod + time.Duration(randSpan.Int64())
}

func (p *blockProposer) proposeBlocks(maxPeriod time.Duration) {
  defer p.wg.Done()
  randPropose := time.NewTimer(randPeriod(maxPeriod))
  for {
    select {
    case <- p.ctx.Done():
      randPropose.Stop()
      return
    case <- randPropose.C:
      randPropose.Reset(randPeriod(maxPeriod))
      // clone := p.state.Clone()
      // blk := clone.CreateBlock()
      // if len(blk.Txs) == 0 {
      //   continue
      // }
      // clone = p.state.Clone()
      // err := clone.ApplyBlock(blk)
      // if err != nil {
      //   fmt.Println(err)
      //   continue
      // }
      // p.blkRelay.RelayBlock(blk)
      // fmt.Printf("* Block proposed: %v", blk)
    }
  }
}
