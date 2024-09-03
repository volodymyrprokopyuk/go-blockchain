package account

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"

	"github.com/dustinxie/ecc"
	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"golang.org/x/crypto/argon2"
)

const (
  encKeyLen = uint32(32)
)

type p256k1PrivateKey struct { // for Account encoding
  chain.P256k1PublicKey
  D *big.Int `json:"d"`
}

func newP256k1PrivateKey(prv *ecdsa.PrivateKey) p256k1PrivateKey {
  return p256k1PrivateKey{
    P256k1PublicKey: chain.NewP256k1PublicKey(&prv.PublicKey), D: prv.D,
  }
}

func (pk *p256k1PrivateKey) publicKey() *ecdsa.PublicKey {
  return &ecdsa.PublicKey{Curve: ecc.P256k1(), X: pk.X, Y: pk.Y}
}

func (pk *p256k1PrivateKey) privateKey() *ecdsa.PrivateKey {
  return &ecdsa.PrivateKey{PublicKey: *pk.publicKey(), D: pk.D}
}

type Account struct {
  privateKey *ecdsa.PrivateKey
  address chain.Address // derived
}

func (a *Account) Address() chain.Address {
  return a.address
}

func NewAccount() (*Account, error) {
  prv, err := ecdsa.GenerateKey(ecc.P256k1(), rand.Reader)
  if err != nil {
    return nil, err
  }
  addr, err := chain.NewAddress(&prv.PublicKey)
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

func Read(path string, pwd []byte) (*Account, error) {
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

func (a *Account) Sign(tx chain.Tx) (chain.SigTx, error) {
  hash, err := tx.Hash()
  if err != nil {
    return chain.SigTx{}, err
  }
  sig, err := ecc.SignBytes(a.privateKey, hash[:], ecc.LowerS | ecc.RecID)
  if err != nil {
    return chain.SigTx{}, err
  }
  return chain.SigTx{Tx: tx, Sig: sig}, nil
}

func Verify(stx chain.SigTx) (bool, error) {
  hash, err := stx.Tx.Hash()
  if err != nil {
    return false, err
  }
  pub, err := ecc.RecoverPubkey("P-256k1", hash[:], stx.Sig)
  if err != nil {
    return false, err
  }
  addr, err := chain.NewAddress(pub)
  if err != nil {
    return false, err
  }
  return addr == stx.From, nil
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
  addr, err := chain.NewAddress(&prv.PublicKey)
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
