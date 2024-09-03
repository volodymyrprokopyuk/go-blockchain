package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

const (
  genesisFile = "genesis.json"
)

type Genesis struct {
  Chain string `json:"chain"`
  Time time.Time `json:"time"`
  Balances map[chain.Address]uint `json:"balances"`
}

func NewGenesis(name string, addr chain.Address, amount uint) Genesis {
  bals := make(map[chain.Address]uint, 1)
  bals[addr] = amount
  return Genesis{Chain: name, Time: time.Now(), Balances: bals}
}

func (g Genesis) Write(dir string) error {
  jsnGen, err := json.Marshal(g)
  if err != nil {
    return err
  }
  err = os.MkdirAll(dir, 0700)
  if err != nil {
    return err
  }
  path := filepath.Join(dir, genesisFile)
  return os.WriteFile(path, jsnGen, 0600)
}

func ReadGenesis(dir string) (Genesis, error) {
  path := filepath.Join(dir, genesisFile)
  jsnGen, err := os.ReadFile(path)
  if err != nil {
    return Genesis{}, err
  }
  var gen Genesis
  err = json.Unmarshal(jsnGen, &gen)
  return gen, err
}

type Block struct {
  Number uint `json:"number"`
  Parent []byte `json:"parent"`
  Time time.Time `json:"time"`
  Txs []chain.Tx `json:"txs"`
}
