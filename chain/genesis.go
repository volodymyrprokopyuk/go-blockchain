package chain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const genesisFile = "genesis.json"

type Genesis struct {
  Chain string `json:"chain"`
  Time time.Time `json:"time"`
  Balances map[Address]uint64 `json:"balances"`
}

func NewGenesis(name string, acc Address, balance uint64) Genesis {
  bals := make(map[Address]uint64, 1)
  bals[acc] = balance
  return Genesis{Chain: name, Time: time.Now(), Balances: bals}
}

func (g Genesis) Write(dir string) error {
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

func ReadGenesis(dir string) (Genesis, error) {
  path := filepath.Join(dir, genesisFile)
  jgen, err := os.ReadFile(path)
  if err != nil {
    return Genesis{}, err
  }
  var gen Genesis
  err = json.Unmarshal(jgen, &gen)
  return gen, err
}
