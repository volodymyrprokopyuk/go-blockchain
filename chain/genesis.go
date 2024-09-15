package chain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/dustinxie/ecc"
	"golang.org/x/crypto/sha3"
)

const genesisFile = "genesis.json"

type Genesis struct {
  Chain string `json:"chain"`
  Time time.Time `json:"time"`
  Balances map[Address]uint64 `json:"balances"`
}

func NewGenesis(name string, acc Address, balance uint64) Genesis {
  balances := make(map[Address]uint64, 1)
  balances[acc] = balance
  return Genesis{Chain: name, Time: time.Now(), Balances: balances}
}

func (g Genesis) Hash() Hash {
  jgen, _ := json.Marshal(g)
  hash := make([]byte, 64)
  sha3.ShakeSum256(hash, jgen)
  return Hash(hash[:32])
}

type SigGenesis struct {
  Genesis
  Sig []byte `json:"sig"`
}

func NewSigGenesis(gen Genesis, sig []byte) SigGenesis {
  return SigGenesis{Genesis: gen, Sig: sig}
}

func (g SigGenesis) Hash() Hash {
  jsgen, _ := json.Marshal(g)
  hash := make([]byte, 64)
  sha3.ShakeSum256(hash, jsgen)
  return Hash(hash[:32])
}

func VerifyGen(sgen SigGenesis) (bool, error) {
  pub, err := ecc.RecoverPubkey("P-256k1", sgen.Genesis.Hash().Bytes(), sgen.Sig)
  if err != nil {
    return false, err
  }
  acc := NewAddress(pub)
  _, exist := sgen.Balances[acc]
  return exist, nil
}

func (g SigGenesis) Write(dir string) error {
  jsgen, err := json.Marshal(g)
  if err != nil {
    return err
  }
  err = os.MkdirAll(dir, 0700)
  if err != nil {
    return err
  }
  path := filepath.Join(dir, genesisFile)
  return os.WriteFile(path, jsgen, 0600)
}

func ReadGenesis(dir string) (SigGenesis, error) {
  path := filepath.Join(dir, genesisFile)
  jsgen, err := os.ReadFile(path)
  if err != nil {
    return SigGenesis{}, err
  }
  var sgen SigGenesis
  err = json.Unmarshal(jsgen, &sgen)
  return sgen, err
}

func ReadGenesisBytes(dir string) ([]byte, error) {
  path := filepath.Join(dir, genesisFile)
  return os.ReadFile(path)
}
