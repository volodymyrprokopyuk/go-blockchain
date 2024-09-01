package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/dustinxie/ecc"
	"golang.org/x/crypto/argon2"
)

// type p256k1PrivateKey struct {
//   Curve string `json:"curve"`
//   X *big.Int `json:"x"`
//   Y *big.Int `json:"y"`
//   D *big.Int `json:"d"`
// }

// func Newp256k1PrivateKey(prv *ecdsa.PrivateKey) *p256k1PrivateKey {
//   return &p256k1PrivateKey{Curve: "P-256k1", X: prv.X, Y: prv.Y, D: prv.D}
// }

// func (pk *p256k1PrivateKey) PublicKey() *ecdsa.PublicKey {
//   return &ecdsa.PublicKey{Curve: ecc.P256k1(), X: pk.X, Y: pk.Y}
// }

// func (pk *p256k1PrivateKey) PrivateKey() *ecdsa.PrivateKey {
//   return &ecdsa.PrivateKey{PublicKey: *pk.PublicKey(), D: pk.D}
// }

func SignVerify() error {
  // generate
  prv, err := ecdsa.GenerateKey(ecc.P256k1(), rand.Reader)
  if err != nil {
    return err
  }

  // marshal
  jsnPrv, err := json.Marshal(NewP256k1PrivateKey(prv))
  if err != nil {
    return err
  }
  fmt.Printf("%s\n", jsnPrv)

  // encrypt
  keyLen := uint32(32)
  pwd := "password"
  salt := make([]byte, keyLen)
  _, err = rand.Read(salt)
  if err != nil {
    return err
  }
  key := argon2.IDKey([]byte(pwd), salt, 1, 256, 1, keyLen)
  blk, err := aes.NewCipher(key)
  if err != nil {
    return err
  }
  gcm, err := cipher.NewGCM(blk)
  if err != nil {
    return err
  }
  nonceLen := gcm.NonceSize()
  nonce := make([]byte, nonceLen)
  _, err = rand.Read(nonce)
  if err != nil {
    return err
  }
  ciph := gcm.Seal(nonce, nonce, jsnPrv, nil)
  ciph = append(salt, ciph...)
  fmt.Printf("%x\n", ciph)

  // decrypt
  salt, ciph = ciph[:keyLen], ciph[keyLen:]
  key = argon2.IDKey([]byte(pwd), salt, 1, 256, 1, keyLen)
  blk, err = aes.NewCipher(key)
  if err != nil {
    return err
  }
  gcm, err = cipher.NewGCM(blk)
  if err != nil {
    return err
  }
  nonce, ciph = ciph[:nonceLen], ciph[nonceLen:]
  pln, err := gcm.Open(nil, nonce, ciph, nil)
  if err != nil {
    return err
  }
  fmt.Printf("%s\n", pln)

  // unmarshal
  var pk p256k1PrivateKey
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
