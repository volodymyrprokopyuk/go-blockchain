package rtx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/chain/account"
	"google.golang.org/grpc"
)

type TxApplier interface {
  Nonce(acc chain.Address) uint64
  ApplyTx(tx chain.SigTx) error
}

type TxRelayer interface {
  RelayTx(tx chain.SigTx)
}

type TxSrv struct {
  UnimplementedTxServer
  keyStoreDir string
  txApplier TxApplier
  txRelayer TxRelayer
}

func NewTxSrv(
  keyStoreDir string, txApplier TxApplier, txRelayer TxRelayer,
) *TxSrv {
  return &TxSrv{
    keyStoreDir: keyStoreDir, txApplier: txApplier, txRelayer: txRelayer,
  }
}

func (s *TxSrv) TxSign(_ context.Context, req *TxSignReq) (*TxSignRes, error) {
  path := filepath.Join(s.keyStoreDir, req.From)
  acc, err := account.Read(path, []byte(req.Password))
  if err != nil {
    return nil, err
  }
  tx := chain.NewTx(
    chain.Address(req.From), chain.Address(req.To), req.Value,
    s.txApplier.Nonce(chain.Address(req.From)) + 1,
  )
  stx, err := acc.SignTx(tx)
  if err != nil {
    return nil, err
  }
  jtx, err := json.Marshal(stx)
  if err != nil {
    return nil, err
  }
  res := &TxSignRes{SigTx: jtx}
  return res, nil
}

func (s *TxSrv) TxSend(_ context.Context, req *TxSendReq) (*TxSendRes, error) {
  var tx chain.SigTx
  err := json.Unmarshal(req.SigTx, &tx)
  if err != nil {
    return nil, err
  }
  err = s.txApplier.ApplyTx(tx)
  if err != nil {
    return nil, err
  }
  s.txRelayer.RelayTx(tx)
  res := &TxSendRes{Hash: tx.Hash().String()}
  return res, nil
}

func (s *TxSrv) TxReceive(
  stream grpc.ClientStreamingServer[TxReceiveReq, TxReceiveRes],
) error {
  for {
    req, err := stream.Recv()
    if err == io.EOF {
      res := &TxReceiveRes{}
      return stream.SendAndClose(res)
    }
    if err != nil {
      return err
    }
    var tx chain.SigTx
    err = json.Unmarshal(req.SigTx, &tx)
    err = s.txApplier.ApplyTx(tx)
    if err != nil {
      fmt.Println(err)
      continue
    }
    s.txRelayer.RelayTx(tx)
  }
}
