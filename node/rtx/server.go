package rtx

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/chain/account"
	"github.com/volodymyrprokopyuk/go-blockchain/chain/state"
)

type TxSrv struct {
  UnimplementedTxServer
  keyStoreDir string
  state *state.State
}

func NewTxSrv(keyStoreDir string, sta *state.State) *TxSrv {
  return &TxSrv{keyStoreDir: keyStoreDir, state: sta}
}

func (s *TxSrv) TxSign(_ context.Context, req *TxSignReq,) (*TxSignRes, error) {
  path := filepath.Join(s.keyStoreDir, req.From)
  acc, err := account.Read(path, []byte(req.Password))
  if err != nil {
    return nil, err
  }
  tx := chain.NewTx(
    chain.Address(req.From), chain.Address(req.To), req.Value,
    s.state.Pending.Nonce(chain.Address(req.From)) + 1,
  )
  stx, err := acc.Sign(tx)
  if err != nil {
    return nil, err
  }
  jstx, err := json.Marshal(stx)
  if err != nil {
    return nil, err
  }
  res := &TxSignRes{SigTx: jstx}
  return res, nil
}

func (s *TxSrv) TxSend(_ context.Context, req *TxSendReq) (*TxSendRes, error) {
  var stx chain.SigTx
  err := json.Unmarshal(req.SigTx, &stx)
  if err != nil {
    return nil, err
  }
  err = s.state.Pending.ApplyTx(stx)
  if err != nil {
    return nil, err
  }
  fmt.Printf("* Pending state (ApplyTx)\n%v\n", s.state)
  res := &TxSendRes{Hash: stx.Hash().String()}
  return res, nil
}
