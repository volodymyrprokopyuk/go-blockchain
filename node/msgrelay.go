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

type grpcMsgRelay[Msg any] func(
  ctx context.Context, conn *grpc.ClientConn, chRelay chan Msg,
) error

var grpcTxRelay grpcMsgRelay[chain.SigTx] = func(
  ctx context.Context, conn *grpc.ClientConn, chRelay chan chain.SigTx,
) error {
  cln := rpc.NewTxClient(conn)
  stream, err := cln.TxReceive(ctx)
  if err != nil {
    return err
  }
  for {
    select {
    case <- ctx.Done():
      _, err = stream.CloseAndRecv()
      return err
    case tx, open := <- chRelay:
      if !open {
        _, err = stream.CloseAndRecv()
        return err
      }
      jtx, err := json.Marshal(tx)
      if err != nil {
        fmt.Println(err)
        continue
      }
      req := &rpc.TxReceiveReq{Tx: jtx}
      err = stream.Send(req)
      if err != nil {
        fmt.Println(err)
        continue
      }
    }
  }
}

var grpcBlockRelay grpcMsgRelay[chain.Block] = func(
  ctx context.Context, conn *grpc.ClientConn, chRelay chan chain.Block,
) error {
  cln := rpc.NewBlockClient(conn)
  stream, err := cln.BlockReceive(ctx)
  if err != nil {
    return err
  }
  for {
    select {
    case <- ctx.Done():
      _, err = stream.CloseAndRecv()
      return err
    case blk, open := <- chRelay:
      if !open {
        _, err = stream.CloseAndRecv()
        return err
      }
      jblk, err := json.Marshal(blk)
      if err != nil {
        fmt.Println(err)
        continue
      }
      req := &rpc.BlockReceiveReq{Block: jblk}
      err = stream.Send(req)
      if err != nil {
        fmt.Println(err)
        continue
      }
    }
  }
}

type msgRelay[Msg any, Relay grpcMsgRelay[Msg]] struct {
  ctx context.Context
  wg *sync.WaitGroup
  chMsg chan Msg
  selfRelay bool
  peerDisc *peerDiscovery
  grpcRelay Relay
}

func newMsgRelay[Msg any, Relay grpcMsgRelay[Msg]](
  ctx context.Context, wg *sync.WaitGroup, cap int, selfRelay bool,
  peerDisc *peerDiscovery, grpcRelay Relay,
) *msgRelay[Msg, Relay] {
  return &msgRelay[Msg, Relay]{
    ctx: ctx, wg: wg, chMsg: make(chan Msg, cap), selfRelay: selfRelay,
    peerDisc: peerDisc, grpcRelay: grpcRelay,
  }
}

func (r *msgRelay[Msg, Relay]) RelayTx(tx Msg) {
  r.chMsg <- tx
}

func (r *msgRelay[Msg, Relay]) RelayBlock(blk Msg) {
  r.chMsg <- blk
}

func (r *msgRelay[Msg, Relay]) grpcRelays() []chan Msg {
  var peers []string
  if r.selfRelay {
    peers = r.peerDisc.SelfPeers()
  } else {
    peers = r.peerDisc.Peers()
  }
  chRelays := make([]chan Msg, len(peers))
  for i, peer := range peers {
    chRelay := make(chan Msg)
    chRelays[i] = chRelay
    r.wg.Add(1)
    go func() {
      defer r.wg.Done()
      conn, err := grpc.NewClient(
        peer, grpc.WithTransportCredentials(insecure.NewCredentials()),
      )
      if err != nil {
        fmt.Println(err)
        return
      }
      defer conn.Close()
      err = r.grpcRelay(r.ctx, conn, chRelay)
      if err != nil {
        fmt.Println(err)
        return
      }
    }()
  }
  return chRelays
}

func (r *msgRelay[Msg, Relay]) relayMsgs(period time.Duration) {
  defer r.wg.Done()
  tick := time.NewTicker(period)
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
      stopRelay := time.NewTimer(period - 1 * time.Second)
      relay: for {
        select {
        case <- r.ctx.Done():
          stopRelay.Stop()
          closeRelays()
          return
        case <- stopRelay.C:
          closeRelays()
          break relay
        case msg := <- r.chMsg:
          for _, chRelay := range chRelays {
            chRelay <- msg
          }
        }
      }
    }
  }
}
