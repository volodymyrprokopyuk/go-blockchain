package chain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/dustinxie/ecc"
)

const genesisFile = "genesis.json"

type Genesis struct {
  Chain string `json:"chain"`
  Authority Address `json:"authority"`
  Balances map[Address]uint64 `json:"balances"`
  Time time.Time `json:"time"`
}

func NewGenesis(name string, authority, acc Address, balance uint64) Genesis {
  balances := make(map[Address]uint64, 1)
  balances[acc] = balance
  return Genesis{
    Chain: name, Authority: authority, Balances: balances, Time: time.Now(),
  }
}

func (g Genesis) Hash() Hash {
  return NewHash(g)
}

type SigGenesis struct {
  Genesis
  Sig []byte `json:"sig"`
}

func NewSigGenesis(gen Genesis, sig []byte) SigGenesis {
  return SigGenesis{Genesis: gen, Sig: sig}
}

func (g SigGenesis) Hash() Hash {
  return NewHash(g)
}

func (g SigGenesis) Write(dir string) error {
  jgen, err := json.Marshal(g)
  if err != nil {
    return err
  }
  err = os.MkdirAll(dir, 0700)
  if err != nil {
    return err
  }
  path := filepath.Join(dir, genesisFile)
  return os.WriteFile(path, jgen, 0600)
}

func VerifyGen(gen SigGenesis) (bool, error) {
  hash := gen.Genesis.Hash().Bytes()
  pub, err := ecc.RecoverPubkey("P-256k1", hash, gen.Sig)
  if err != nil {
    return false, err
  }
  acc := NewAddress(pub)
  return acc == Address(gen.Authority), nil
}

func ReadGenesis(dir string) (SigGenesis, error) {
  path := filepath.Join(dir, genesisFile)
  jgen, err := os.ReadFile(path)
  if err != nil {
    return SigGenesis{}, err
  }
  var gen SigGenesis
  err = json.Unmarshal(jgen, &gen)
  return gen, err
}

func ReadGenesisBytes(dir string) ([]byte, error) {
  path := filepath.Join(dir, genesisFile)
  return os.ReadFile(path)
}
