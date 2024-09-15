package chain

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/sha3"
)

const blocksFile = "blocks.store"

type Block struct {
  Number uint64 `json:"number"`
  Parent Hash `json:"parent"`
  Time time.Time `json:"time"`
  Txs []SigTx `json:"txs"`
}

func NewBlock(number uint64, parent Hash, txs []SigTx) Block {
  return Block{Number: number, Parent: parent, Time: time.Now(), Txs: txs}
}

func (b Block) Hash() Hash {
  jblk, _ := json.Marshal(b)
  hash := make([]byte, 64)
  sha3.ShakeSum256(hash, jblk)
  return Hash(hash[:32])
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

type storeBlock struct {
  Hash Hash `json:"hash"`
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
  blocks := func(yield func(err error, blk Block) bool) {
    sca := bufio.NewScanner(file)
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
