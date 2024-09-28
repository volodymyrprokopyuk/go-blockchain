package chain

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

const encKeyLen = uint32(32)

type p256k1PublicKey struct {
  Curve string `json:"curve"`
  X *big.Int `json:"x"`
  Y *big.Int `json:"y"`
}

func newP256k1PublicKey(pub *ecdsa.PublicKey) p256k1PublicKey {
  return p256k1PublicKey{Curve: "P-256k1", X: pub.X, Y: pub.Y}
}

type p256k1PrivateKey struct {
  p256k1PublicKey
  D *big.Int `json:"d"`
}

func newP256k1PrivateKey(prv *ecdsa.PrivateKey) p256k1PrivateKey {
  return p256k1PrivateKey{
    p256k1PublicKey: newP256k1PublicKey(&prv.PublicKey), D: prv.D,
  }
}

func (k *p256k1PrivateKey) publicKey() *ecdsa.PublicKey {
  return &ecdsa.PublicKey{Curve: ecc.P256k1(), X: k.X, Y: k.Y}
}

func (k *p256k1PrivateKey) privateKey() *ecdsa.PrivateKey {
  return &ecdsa.PrivateKey{PublicKey: *k.publicKey(), D: k.D}
}

type Address string

func NewAddress(pub *ecdsa.PublicKey) Address {
  jpub, _ := json.Marshal(newP256k1PublicKey(pub))
  hash := make([]byte, 64)
  sha3.ShakeSum256(hash, jpub)
  return Address(hex.EncodeToString(hash[:32]))
}

type Account struct {
  prv *ecdsa.PrivateKey
  addr Address // derived
}

func NewAccount() (Account, error) {
  prv, err := ecdsa.GenerateKey(ecc.P256k1(), rand.Reader)
  if err != nil {
    return Account{}, err
  }
  addr := NewAddress(&prv.PublicKey)
  return Account{prv: prv, addr: addr}, nil
}

func (a Account) Address() Address {
  return a.addr
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

func ReadAccount(path string, pass []byte) (Account, error) {
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

func (a Account) SignGen(gen Genesis) (SigGenesis, error) {
  hash := gen.Hash().Bytes()
  sig, err := ecc.SignBytes(a.prv, hash, ecc.LowerS | ecc.RecID)
  if err != nil {
    return SigGenesis{}, err
  }
  sgen := NewSigGenesis(gen, sig)
  return sgen, nil
}

func (a Account) SignTx(tx Tx) (SigTx, error) {
  hash := tx.Hash().Bytes()
  sig, err := ecc.SignBytes(a.prv, hash, ecc.LowerS | ecc.RecID)
  if err != nil {
    return SigTx{}, err
  }
  stx := NewSigTx(tx, sig)
  return stx, nil
}

func (a Account) SignBlock(blk Block) (SigBlock, error) {
  hash := blk.Hash().Bytes()
  sig, err := ecc.SignBytes(a.prv, hash, ecc.LowerS | ecc.RecID)
  if err != nil {
    return SigBlock{}, err
  }
  sblk := NewSigBlock(blk, sig)
  return sblk, nil
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
  addr := NewAddress(&prv.PublicKey)
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
