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

type P256k1PublicKey struct { // for Address hash
  Curve string `json:"curve"`
  X *big.Int `json:"x"`
  Y *big.Int `json:"y"`
}

func NewP256k1PublicKey(pub *ecdsa.PublicKey) P256k1PublicKey {
  return P256k1PublicKey{Curve: "P-256k1", X: pub.X, Y: pub.Y}
}

type Address string

func NewAddress(pub *ecdsa.PublicKey) (Address, error) {
  jsnPub, err := json.Marshal(NewP256k1PublicKey(pub))
  if err != nil {
    return Address(""), err
  }
  hashPub := make([]byte, 64)
  sha3.ShakeSum256(hashPub, jsnPub)
  return Address(hex.EncodeToString(hashPub[:32])), nil
}

type Hash [32]byte

func (h Hash) String() string {
  return hex.EncodeToString(h[:])
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
  Value uint `json:"value"`
  Nonce uint `json:"nonce"`
  Time time.Time `json:"time"`
}

func (t Tx) String() string {
  return fmt.Sprintf("%.7s -> %.7s %10d %5d", t.From, t.To, t.Value, t.Nonce)
}

func (t Tx) Hash() (Hash, error) {
  jsnTx, err := json.Marshal(t)
  if err != nil {
    return Hash{}, err
  }
  hash := make([]byte, 64)
  sha3.ShakeSum256(hash, jsnTx)
  return Hash(hash[:32]), nil
}

type SigTx struct {
  Tx
  Sig []byte `json:"sig"`
}

func (t SigTx) Hash() (Hash, error) {
  jsnTx, err := json.Marshal(t)
  if err != nil {
    return Hash{}, err
  }
  hash := make([]byte, 64)
  sha3.ShakeSum256(hash, jsnTx)
  return Hash(hash[:32]), nil
}
