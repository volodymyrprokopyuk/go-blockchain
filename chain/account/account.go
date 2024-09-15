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

const encKeyLen = uint32(32)

type p256k1PrivateKey struct {
  chain.P256k1PublicKey
  D *big.Int `json:"d"`
}

func newP256k1PrivateKey(prv *ecdsa.PrivateKey) p256k1PrivateKey {
  return p256k1PrivateKey{
    P256k1PublicKey: chain.NewP256k1PublicKey(&prv.PublicKey), D: prv.D,
  }
}

func (k *p256k1PrivateKey) publicKey() *ecdsa.PublicKey {
  return &ecdsa.PublicKey{Curve: ecc.P256k1(), X: k.X, Y: k.Y}
}

func (k *p256k1PrivateKey) privateKey() *ecdsa.PrivateKey {
  return &ecdsa.PrivateKey{PublicKey: *k.publicKey(), D: k.D}
}

type Account struct {
  prv *ecdsa.PrivateKey
  addr chain.Address // derived
}

func (a Account) Address() chain.Address {
  return a.addr
}

func NewAccount() (Account, error) {
  prv, err := ecdsa.GenerateKey(ecc.P256k1(), rand.Reader)
  if err != nil {
    return Account{}, err
  }
  addr := chain.NewAddress(&prv.PublicKey)
  return Account{prv: prv, addr: addr}, nil
}

func (a Account) Write(dir string, pass []byte) error {
  jprv, err := a.encodePrivateKey()
  if err != nil {
    return err
  }
  cprv, err := encryptWithPassword(jprv, pass)
  if err != nil {
    return err
  }
  err = os.MkdirAll(dir, 0700)
  if err != nil {
    return err
  }
  path := filepath.Join(dir, string(a.Address()))
  return os.WriteFile(path, cprv, 0600)
}

func Read(path string, pass []byte) (Account, error) {
  cprv, err := os.ReadFile(path)
  if err != nil {
    return Account{}, err
  }
  jprv, err := decryptWithPassword(cprv, pass)
  if err != nil {
    return Account{}, err
  }
  return decodePrivateKey(jprv)
}

func (a Account) SignGen(gen chain.Genesis) (chain.SigGenesis, error) {
  sig, err := ecc.SignBytes(a.prv, gen.Hash().Bytes(), ecc.LowerS | ecc.RecID)
  if err != nil {
    return chain.SigGenesis{}, err
  }
  sgen := chain.NewSigGenesis(gen, sig)
  return sgen, nil
}

func (a Account) SignTx(tx chain.Tx) (chain.SigTx, error) {
  sig, err := ecc.SignBytes(a.prv, tx.Hash().Bytes(), ecc.LowerS | ecc.RecID)
  if err != nil {
    return chain.SigTx{}, err
  }
  stx := chain.NewSigTx(tx, sig)
  return stx, nil
}

func (a Account) encodePrivateKey() ([]byte, error) {
  return json.Marshal(newP256k1PrivateKey(a.prv))
}

func decodePrivateKey(jprv []byte) (Account, error) {
  var pk p256k1PrivateKey
  err := json.Unmarshal(jprv, &pk)
  if err != nil {
    return Account{}, err
  }
  prv := pk.privateKey()
  addr := chain.NewAddress(&prv.PublicKey)
  return Account{prv: prv, addr: addr}, nil
}

func encryptWithPassword(msg, pass []byte) ([]byte, error) {
  salt := make([]byte, encKeyLen)
  _, err := rand.Read(salt)
  if err != nil {
    return nil, err
  }
  key := argon2.IDKey(pass, salt, 1, 256, 1, encKeyLen)
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

func decryptWithPassword(ciph, pass []byte) ([]byte, error) {
  salt, ciph := ciph[:encKeyLen], ciph[encKeyLen:]
  key := argon2.IDKey(pass, salt, 1, 256, 1, encKeyLen)
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
