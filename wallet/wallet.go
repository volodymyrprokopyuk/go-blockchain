package wallet

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/dustinxie/ecc"
)

type P256k1PrivateKey struct {
  Curve string `json:"curve"`
  X *big.Int `json:"x"`
  Y *big.Int `json:"y"`
  D *big.Int `json:"d"`
}

func NewP256k1PrivateKey(prv *ecdsa.PrivateKey) *P256k1PrivateKey {
  return &P256k1PrivateKey{Curve: "P-256k1", X: prv.X, Y: prv.Y, D: prv.D}
}

func (pk *P256k1PrivateKey) PublicKey() *ecdsa.PublicKey {
  return &ecdsa.PublicKey{Curve: ecc.P256k1(), X: pk.X, Y: pk.Y}
}

func (pk *P256k1PrivateKey) PrivateKey() *ecdsa.PrivateKey {
  return &ecdsa.PrivateKey{PublicKey: *pk.PublicKey(), D: pk.D}
}

func SignVerify() error {
  // generate
  prv, err := ecdsa.GenerateKey(ecc.P256k1(), rand.Reader)
  if err != nil {
    return err
  }

  // marshal, unmarshal
  jsnPrv, err := json.Marshal(NewP256k1PrivateKey(prv))
  if err != nil {
    return err
  }
  fmt.Println(string(jsnPrv))
  var pk P256k1PrivateKey
  err = json.Unmarshal(jsnPrv, &pk)
  if err != nil {
    return err
  }
  pub := pk.PublicKey()
  prv = pk.PrivateKey()

  // sign
  msg := "abc"
  hash := sha256.Sum256([]byte(msg))
  sig, err := ecc.SignBytes(prv, hash[:], ecc.LowerS | ecc.RecID)
  if err != nil {
    return err
  }

  // verify
  pub, err = ecc.RecoverPubkey("P-256k1", hash[:], sig)
  if err != nil {
    return err
  }
  if pub.Equal(prv.Public()) {
    fmt.Println("valid")
  } else {
    fmt.Println("not valid")
  }
  return nil
}
