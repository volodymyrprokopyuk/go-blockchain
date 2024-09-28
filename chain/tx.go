package chain

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dustinxie/ecc"
	"golang.org/x/crypto/sha3"
)

type Hash [32]byte

func NewHash(val any) Hash {
  jval, _ := json.Marshal(val)
  hash := make([]byte, 64)
  sha3.ShakeSum256(hash, jval)
  return Hash(hash[:32])
}

func (h Hash) String() string {
  return hex.EncodeToString(h[:])
}

func (h Hash) Bytes() []byte {
  hash := [32]byte(h)
  return hash[:]
}

func (h Hash) MarshalText() ([]byte, error) {
  return []byte(hex.EncodeToString(h[:])), nil
}

func (h *Hash) UnmarshalText(hash []byte) error {
  _, err := hex.Decode(h[:], hash)
  return err
}

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
  hash := tx.Tx.Hash().Bytes()
  pub, err := ecc.RecoverPubkey("P-256k1", hash, tx.Sig)
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
