package node

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type txRelay struct {
  ctx context.Context
  wg *sync.WaitGroup
  chTx chan chain.SigTx
  dis *discovery
}

func newTxRelay(
  ctx context.Context, wg *sync.WaitGroup, cap int, dis *discovery,
) *txRelay {
  return &txRelay{
    ctx: ctx, wg: wg, chTx: make(chan chain.SigTx, cap), dis: dis,
  }
}

func (r *txRelay) RelayTx(tx chain.SigTx) {
  r.chTx <- tx
}

func (r *txRelay) grpcRelays() []chan chain.SigTx {
  peers := r.dis.Peers()
  chRelays := make([]chan chain.SigTx, len(peers))
  for i, peer := range peers {
    chRelay := make(chan chain.SigTx)
    chRelays[i] = chRelay
    go func () {
      r.wg.Add(1)
      defer r.wg.Done()
      conn, err := grpc.NewClient(
        peer, grpc.WithTransportCredentials(insecure.NewCredentials()),
      )
      if err != nil {
        fmt.Println(err)
        return
      }
      defer conn.Close()
      cln := rpc.NewTxClient(conn)
      stream, err := cln.TxReceive(r.ctx)
      if err != nil {
        fmt.Println(err)
        return
      }
      for tx := range chRelay {
        jtx, err := json.Marshal(tx)
        if err != nil {
          fmt.Println(err)
          continue
        }
        req := &rpc.TxReceiveReq{SigTx: jtx}
        err = stream.Send(req)
        if err != nil {
          fmt.Println(err)
          continue
        }
      }
      _, err = stream.CloseAndRecv()
      if err != nil {
        fmt.Println(err)
        return
      }
    }()
  }
  return chRelays
}

func (r *txRelay) relayTxs(interval time.Duration) {
  r.wg.Add(1)
  defer r.wg.Done()
  tick := time.NewTicker(interval)
  defer tick.Stop()
  for {
    select {
    case <- r.ctx.Done():
      return
    case <- tick.C:
      chRelays := r.grpcRelays()
      closeRelays := func() {
        for _, chRelay := range chRelays {
          close(chRelay)
        }
      }
      timer := time.NewTimer(interval - 1 * time.Second)
      relay: for {
        select {
        case <- r.ctx.Done():
          closeRelays()
          return
        case <- timer.C:
          closeRelays()
          break relay
        case tx := <- r.chTx:
          for _, chRelay := range chRelays {
            chRelay <- tx
          }
        }
      }
    }
  }
}
