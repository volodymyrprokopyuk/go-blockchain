package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/account"
)

type Genesis struct {
  Blockchain string `json:"blockchain"`
  Time time.Time `json:"time"`
  Balances map[account.Address]uint `json:"balances"`
}

func NewGenesis(bcn string, addr account.Address, amount uint) *Genesis {
  bals := make(map[account.Address]uint, 1)
  bals[addr] = amount
  return &Genesis{Blockchain: bcn, Time: time.Now(), Balances: bals}
}

func (g *Genesis) Write(dir string) error {
  jsnGen, err := json.Marshal(g)
  if err != nil {
    return err
  }
  err = os.MkdirAll(dir, 0700)
  if err != nil {
    return err
  }
  path := filepath.Join(dir, "genesis.json")
  return os.WriteFile(path, jsnGen, 0600)
}

type Tx struct {
  From account.Address `json:"from"`
  To account.Address `json:"to"`
  Value uint `json:"value"`
  Nonce uint `json:"nonce"`
  Time time.Time `json:"time"`
  Sig []byte `json:"sig"`
}

type Block struct {
  Number uint `json:"number"`
  Hash []byte `json:"hash"`
  Parent []byte `json:"parent"`
  Time time.Time `json:"time"`
  Txs []Tx `json:"txs"`
}
