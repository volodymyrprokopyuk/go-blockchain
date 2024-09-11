package store

import (
	"bufio"
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
  blocksFile = "blocks.store"
)

type Genesis struct {
  Chain string `json:"chain"`
  Time time.Time `json:"time"`
  Balances map[chain.Address]uint64 `json:"balances"`
}

func NewGenesis(name string, addr chain.Address, balance uint64) Genesis {
  bals := make(map[chain.Address]uint64, 1)
  bals[addr] = balance
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
  bld.WriteString(fmt.Sprintf("%d %.7s -> %.7s\n", b.Number, hash, b.Parent))
  for _, tx := range b.Txs {
    bld.WriteString(fmt.Sprintf("  %s\n", tx))
  }
  return bld.String()
}

func (b Block) Hash() (chain.Hash, error) {
  if b.Number == 0 && (b.Parent == chain.Hash{}) && len(b.Txs) == 0 {
    return chain.Hash{}, nil
  }
  jsnBlk, err := json.Marshal(b)
  if err != nil {
    return chain.Hash{}, err
  }
  hash := make([]byte, 64)
  sha3.ShakeSum256(hash, jsnBlk)
  return chain.Hash(hash[:32]), nil
}

type storeBlock struct {
  Hash chain.Hash `json:"hash"`
  Block Block `json:"block"`
}

func (b Block) Write(dir string) error {
  hash, err := b.Hash()
  if err != nil {
    return err
  }
  blk := storeBlock{Hash: hash, Block: b}
  path := filepath.Join(dir, blocksFile)
  file, err := os.OpenFile(path, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0600)
  if err != nil {
    return err
  }
  defer file.Close()
  err = json.NewEncoder(file).Encode(blk)
  if err != nil {
    return err
  }
  _, err = file.WriteString("\n")
  return err
}

func ReadBlocks(dir string) (
  func(yield func(err error, blk Block) bool), func(), error,
) {
  path := filepath.Join(dir, blocksFile)
  file, err := os.Open(path)
  if err != nil {
    return nil, nil, err
  }
  sca := bufio.NewScanner(file)
  blocks := func(yield func(err error, blk Block) bool) {
    more := true
    for sca.Scan() && more {
      err := sca.Err()
      if err != nil {
        yield(err, Block{})
        return
      }
      jsnBlk := sca.Bytes()
      if len(jsnBlk) == 0 {
        continue
      }
      var stoBlk storeBlock
      err = json.Unmarshal(jsnBlk, &stoBlk)
      if err != nil {
        more = yield(err, Block{})
        continue
      }
      more = yield(nil, stoBlk.Block)
    }
  }
  close := func() {
    file.Close()
  }
  return blocks, close, nil
}
