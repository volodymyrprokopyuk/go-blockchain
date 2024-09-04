package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"golang.org/x/crypto/sha3"
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
  Parent chain.Hash `json:"parent"`
  Time time.Time `json:"time"`
  Txs []chain.SigTx `json:"txs"`
}

func (b Block) String() string {
  hash, _ := b.Hash()
  var bld strings.Builder
  bld.WriteString(fmt.Sprintf("%d %.5x -> %.5x\n", b.Number, hash, b.Parent))
  for _, tx := range b.Txs {
    bld.WriteString(fmt.Sprintf("  %s\n", tx))
  }
  return bld.String()
}

func (b Block) Hash() (chain.Hash, error) {
  jsnBlk, err := json.Marshal(b)
  if err != nil {
    return chain.Hash{}, err
  }
  hash := make([]byte, 64)
  sha3.ShakeSum256(hash, jsnBlk)
  return chain.Hash(hash[:32]), nil
}
