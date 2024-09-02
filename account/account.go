package account

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"

	"github.com/dustinxie/ecc"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/sha3"
)


const (
  curveName = "P-256k1"
  encKeyLen = uint32(32)
)

type p256k1PublicKey struct { // for Address hash
  Curve string `json:"curve"`
  X *big.Int `json:"x"`
  Y *big.Int `json:"y"`
}

func newP256k1PublicKey(pub *ecdsa.PublicKey) *p256k1PublicKey {
  return &p256k1PublicKey{Curve: curveName, X: pub.X, Y: pub.Y}
}

type p256k1PrivateKey struct { // for Account encoding
  Curve string `json:"curve"`
  X *big.Int `json:"x"`
  Y *big.Int `json:"y"`
  D *big.Int `json:"d"`
}

func newP256k1PrivateKey(prv *ecdsa.PrivateKey) *p256k1PrivateKey {
  return &p256k1PrivateKey{Curve: curveName, X: prv.X, Y: prv.Y, D: prv.D}
}

func (pk *p256k1PrivateKey) publicKey() *ecdsa.PublicKey {
  return &ecdsa.PublicKey{Curve: ecc.P256k1(), X: pk.X, Y: pk.Y}
}

func (pk *p256k1PrivateKey) privateKey() *ecdsa.PrivateKey {
  return &ecdsa.PrivateKey{PublicKey: *pk.publicKey(), D: pk.D}
}

type Address string

func NewAddress(pub *ecdsa.PublicKey) (Address, error) {
  jsnPub, err := json.Marshal(newP256k1PublicKey(pub))
  if err != nil {
    return Address(""), err
  }
  hashPub := make([]byte, 64)
  sha3.ShakeSum256(hashPub, jsnPub)
  return Address(hex.EncodeToString(hashPub[:32])), nil
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

func (a *Account) Write(dir string, pwd []byte) error {
  jsnPrv, err := a.encodePrivateKey()
  if err != nil {
    return err
  }
  ciphPrv, err := encryptWithPassword(jsnPrv, pwd)
  if err != nil {
    return err
  }
  err = os.MkdirAll(dir, 0700)
  if err != nil {
    return err
  }
  path := filepath.Join(dir, string(a.Address()))
  return os.WriteFile(path, ciphPrv, 0600)
}

func ReadAccount(path string, pwd []byte) (*Account, error) {
  ciphPrv, err := os.ReadFile(path)
  if err != nil {
    return nil, err
  }
  jsnPrv, err := decryptWithPassword(ciphPrv, pwd)
  if err != nil {
    return nil, err
  }
  return decodePrivateKey(jsnPrv)
}

func (a *Account) Sign(msg []byte) ([]byte, error) {
  hash := make([]byte, 64)
  sha3.ShakeSum256(hash, msg)
  sig, err := ecc.SignBytes(a.privateKey, hash, ecc.LowerS | ecc.RecID)
  if err != nil {
    return nil, err
  }
  return sig, nil
}

func VerifySig(sig, msg []byte, addr Address) (bool, error) {
  hash := make([]byte, 64)
  sha3.ShakeSum256(hash, msg)
  pub, err := ecc.RecoverPubkey(curveName, hash, sig)
  if err != nil {
    return false, err
  }
  pubAddr, err := NewAddress(pub)
  if err != nil {
    return false, err
  }
  return addr == pubAddr, nil
}

func (a *Account) encodePrivateKey() ([]byte, error) {
  return json.Marshal(newP256k1PrivateKey(a.privateKey))
}

func decodePrivateKey(jsnPrv []byte) (*Account, error) {
  var pk p256k1PrivateKey
  err := json.Unmarshal(jsnPrv, &pk)
  if err != nil {
    return nil, err
  }
  prv := pk.privateKey()
  addr, err := NewAddress(&prv.PublicKey)
  if err != nil {
    return nil, err
  }
  return &Account{privateKey: prv, address: addr}, nil
}

func encryptWithPassword(msg, pwd []byte) ([]byte, error) {
  salt := make([]byte, encKeyLen)
  _, err := rand.Read(salt)
  if err != nil {
    return nil, err
  }
  key := argon2.IDKey(pwd, salt, 1, 256, 1, encKeyLen)
  blk, err := aes.NewCipher(key)
  if err != nil {
    return nil, err
  }
  gcm, err := cipher.NewGCM(blk)
  if err != nil {
    return nil, err
  }
  nonce := make([]byte, gcm.NonceSize())
  _, err = rand.Read(nonce)
  if err != nil {
    return nil, err
  }
  ciph := gcm.Seal(nonce, nonce, msg, nil)
  ciph = append(salt, ciph...)
  return ciph, nil
}

func decryptWithPassword(ciph, pwd []byte) ([]byte, error) {
  salt, ciph := ciph[:encKeyLen], ciph[encKeyLen:]
  key := argon2.IDKey(pwd, salt, 1, 256, 1, encKeyLen)
  blk, err := aes.NewCipher(key)
  if err != nil {
    return nil, err
  }
  gcm, err := cipher.NewGCM(blk)
  if err != nil {
    return nil, err
  }
  nonceLen := gcm.NonceSize()
  nonce, ciph := ciph[:nonceLen], ciph[nonceLen:]
  msg, err := gcm.Open(nil, nonce, ciph, nil)
  if err != nil {
    return nil, err
  }
  return msg, nil
}
