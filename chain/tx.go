package chain

import (
	"fmt"
	"time"

	"github.com/dustinxie/ecc"
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
  return NewHash(t)
}

type SigTx struct {
  Tx
  Sig []byte `json:"sig"`
}

func NewSigTx(tx Tx, sig []byte) SigTx {
  return SigTx{Tx: tx, Sig: sig}
}

func (t SigTx) Hash() Hash {
  return NewHash(t)
}

func (t SigTx) String() string {
  return fmt.Sprintf(
    "tx %.7s: %.7s -> %.7s %8d %8d", t.Hash(), t.From, t.To, t.Value, t.Nonce,
  )
}

func VerifyTx(tx SigTx) (bool, error) {
  pub, err := ecc.RecoverPubkey("P-256k1", tx.Tx.Hash().Bytes(), tx.Sig)
  if err != nil {
    return false, err
  }
  acc := NewAddress(pub)
  return acc == tx.From, nil
}

type SearchTx struct {
  SigTx
  BlockNumber uint64 `json:"blockNumber"`
  BlockHash Hash `json:"blockHash"`
}

func NewSearchTx(tx SigTx, blkNumber uint64, blkHash Hash) SearchTx {
  return SearchTx{SigTx: tx, BlockNumber: blkNumber, BlockHash: blkHash}
}

func (t SearchTx) String() string {
  return fmt.Sprintf(
    "%v    blk: %5d    %.7s", t.SigTx, t.BlockNumber, t.BlockHash,
  )
}
