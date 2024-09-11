package chain

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"golang.org/x/crypto/sha3"
)

type P256k1PublicKey struct {
  Curve string `json:"curve"`
  X *big.Int `json:"x"`
  Y *big.Int `json:"y"`
}

func NewP256k1PublicKey(pub *ecdsa.PublicKey) P256k1PublicKey {
  return P256k1PublicKey{Curve: "P-256k1", X: pub.X, Y: pub.Y}
}

type Address string

func NewAddress(pub *ecdsa.PublicKey) Address {
  jpub, _ := json.Marshal(NewP256k1PublicKey(pub))
  hash := make([]byte, 64)
  sha3.ShakeSum256(hash, jpub)
  return Address(hex.EncodeToString(hash[:32]))
}

type Hash [32]byte

func NewHash(str string) Hash {
  var hash Hash
  hex.Decode(hash[:], []byte(str))
  return hash
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
  jstx, _ := json.Marshal(t)
  hash := make([]byte, 64)
  sha3.ShakeSum256(hash, jstx)
  return Hash(hash[:32])
}
