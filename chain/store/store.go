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

func NewGenesis(name string, acc chain.Address, balance uint64) Genesis {
  bals := make(map[chain.Address]uint64, 1)
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

type Block struct {
  Number uint64 `json:"number"`
  Parent chain.Hash `json:"parent"`
  Time time.Time `json:"time"`
  Txs []chain.SigTx `json:"txs"`
}

func NewBlock(number uint64, parent chain.Hash, txs []chain.SigTx) Block {
  return Block{Number: number, Parent: parent, Time: time.Now(), Txs: txs}
}

func (b Block) String() string {
  var bld strings.Builder
  bld.WriteString(
    fmt.Sprintf("%2d: %.7s -> %.7s\n", b.Number, b.Hash(), b.Parent),
  )
  for _, tx := range b.Txs {
    bld.WriteString(fmt.Sprintf("  %v\n", tx))
  }
  return bld.String()
}

func (b Block) Hash() chain.Hash {
  if b.Number == 0 && (b.Parent == chain.Hash{}) && len(b.Txs) == 0 {
    return chain.Hash{}
  }
  jblk, _ := json.Marshal(b)
  hash := make([]byte, 64)
  sha3.ShakeSum256(hash, jblk)
  return chain.Hash(hash[:32])
}

type storeBlock struct {
  Hash chain.Hash `json:"hash"`
  Block Block `json:"block"`
}

func (b Block) Write(dir string) error {
  blk := storeBlock{Hash: b.Hash(), Block: b}
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
  close := func() {
    file.Close()
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
      jblk := sca.Bytes()
      if len(jblk) == 0 {
        continue
      }
      var sblk storeBlock
      err = json.Unmarshal(jblk, &sblk)
      if err != nil {
        more = yield(err, Block{})
        continue
      }
      more = yield(nil, sblk.Block)
    }
  }
  return blocks, close, nil
}
