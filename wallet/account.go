package wallet

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/big"

	"github.com/dustinxie/ecc"
)

type p256k1PublicKey struct {
  Curve string `json:"curve"`
  X *big.Int `json:"x"`
  Y *big.Int `json:"y"`
}

func NewP256k1PublicKey(pub *ecdsa.PublicKey) *p256k1PublicKey {
  return &p256k1PublicKey{Curve: "P-256k1", X: pub.X, Y: pub.Y}
}

type p256k1PrivateKey struct {
  Curve string `json:"curve"`
  X *big.Int `json:"x"`
  Y *big.Int `json:"y"`
  D *big.Int `json:"d"`
}

func NewP256k1PrivateKey(prv *ecdsa.PrivateKey) *p256k1PrivateKey {
  return &p256k1PrivateKey{Curve: "P-256k1", X: prv.X, Y: prv.Y, D: prv.D}
}

func (pk *p256k1PrivateKey) PublicKey() *ecdsa.PublicKey {
  return &ecdsa.PublicKey{Curve: ecc.P256k1(), X: pk.X, Y: pk.Y}
}

func (pk *p256k1PrivateKey) PrivateKey() *ecdsa.PrivateKey {
  return &ecdsa.PrivateKey{PublicKey: *pk.PublicKey(), D: pk.D}
}

type Address string

func NewAddress(pub *ecdsa.PublicKey) (Address, error) {
  jsnPub, err := json.Marshal(NewP256k1PublicKey(pub))
  if err != nil {
    return Address(""), err
  }
  hashPub := sha256.Sum256(jsnPub)
  return Address(hex.EncodeToString(hashPub[:])), nil
}

type Account struct {
  privateKey *ecdsa.PrivateKey
  address Address
}

func (a *Account) Address() Address {
  return a.address
}

func NewAccount() (*Account, error) {
  prv, err := ecdsa.GenerateKey(ecc.P256k1(), rand.Reader)
  if err != nil {
    return nil, err
  }
  addr, err := NewAddress(&prv.PublicKey)
  if err != nil {
    return nil, err
  }
  return &Account{privateKey: prv, address: addr}, nil
}

func (a *Account) Encode() ([]byte, error) {
  return json.Marshal(NewP256k1PrivateKey(a.privateKey))
}

func DecodeAccount(jsnPrv []byte) (*Account, error) {
  var pk p256k1PrivateKey
  err := json.Unmarshal(jsnPrv, &pk)
  if err != nil {
    return nil, err
  }
  prv := pk.PrivateKey()
  addr, err := NewAddress(&prv.PublicKey)
  if err != nil {
    return nil, err
  }
  return &Account{privateKey: prv, address: addr}, nil
}
