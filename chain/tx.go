package chain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dustinxie/ecc"
	"golang.org/x/crypto/sha3"
)

type Tx struct {
  From Address `json:"from"`
  To Address `json:"to"`
  Value uint64 `json:"value"`
  Nonce uint64 `json:"nonce"`
  Time time.Time `json:"time"`
}

func NewTx(from, to Address, value, nonce uint64) Tx {
  return Tx{From: from, To: to, Value: value, Nonce: nonce, Time: time.Now()}
}

func (t Tx) Hash() Hash {
  jtx, _ := json.Marshal(t)
  hash := make([]byte, 64)
  sha3.ShakeSum256(hash, jtx)
  return Hash(hash[:32])
}

func (t Tx) String() string {
  return fmt.Sprintf(
    "%.7s: %.7s -> %.7s %10d %5d", t.Hash(), t.From, t.To, t.Value, t.Nonce,
  )
}

type SigTx struct {
  Tx
  Sig []byte `json:"sig"`
}

func NewSigTx(tx Tx, sig []byte) SigTx {
  return SigTx{Tx: tx, Sig: sig}
}

func (t SigTx) Hash() Hash {
  jtx, _ := json.Marshal(t)
  hash := make([]byte, 64)
  sha3.ShakeSum256(hash, jtx)
  return Hash(hash[:32])
}

func VerifyTx(tx SigTx) (bool, error) {
  pub, err := ecc.RecoverPubkey("P-256k1", tx.Tx.Hash().Bytes(), tx.Sig)
  if err != nil {
    return false, err
  }
  acc := NewAddress(pub)
  return acc == tx.From, nil
}
