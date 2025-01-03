package chain

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustinxie/ecc"
)

const blocksFile = "block.store"

func InitBlockStore(dir string) error {
  path := filepath.Join(dir, blocksFile)
  file, err := os.OpenFile(path, os.O_CREATE | os.O_RDONLY, 0600)
  if err != nil {
    return err
  }
  defer file.Close()
  return nil
}

type Block struct {
  Number uint64 `json:"number"`
  Parent Hash `json:"parent"`
  Txs []SigTx `json:"txs"`
  merkleTree []Hash
  MerkleRoot Hash `json:"merkleRoot"`
  Time time.Time `json:"time"`
}

func NewBlock(number uint64, parent Hash, txs []SigTx) (Block, error) {
  merkleTree, err := MerkleHash(txs, TxHash, TxPairHash)
  if err != nil {
    return Block{}, err
  }
  blk := Block{
    Number: number, Parent: parent, Txs: txs,
    merkleTree: merkleTree, MerkleRoot: merkleTree[0],
    Time: time.Now(),
  }
  return blk, nil
}

func (b Block) Hash() Hash {
  return NewHash(b)
}

type SigBlock struct {
  Block
  Sig []byte `json:"sig"`
}

func NewSigBlock(blk Block, sig []byte) SigBlock {
  return SigBlock{Block: blk, Sig: sig}
}

func (b SigBlock) Hash() Hash {
  return NewHash(b)
}

func (b SigBlock) String() string {
  var bld strings.Builder
  bld.WriteString(
    fmt.Sprintf(
      "blk %7d: %.7s -> %.7s   mrk %.7s\n",
      b.Number, b.Hash(), b.Parent, b.MerkleRoot,
    ),
  )
  for _, tx := range b.Txs {
    bld.WriteString(fmt.Sprintf("%v\n", tx))
  }
  return bld.String()
}

func (b SigBlock) Write(dir string) error {
  path := filepath.Join(dir, blocksFile)
  file, err := os.OpenFile(path, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0600)
  if err != nil {
    return err
  }
  defer file.Close()
  return json.NewEncoder(file).Encode(b)
}

func ReadBlocks(dir string) (
  func(yield func(err error, blk SigBlock) bool), func(), error,
) {
  path := filepath.Join(dir, blocksFile)
  file, err := os.Open(path)
  if err != nil {
    return nil, nil, err
  }
  close := func() {
    file.Close()
  }
  blocks := func(yield func(err error, blk SigBlock) bool) {
    sca := bufio.NewScanner(file)
    more := true
    for sca.Scan() && more {
      err := sca.Err()
      if err != nil {
        yield(err, SigBlock{})
        return
      }
      var blk SigBlock
      err = json.Unmarshal(sca.Bytes(), &blk)
      if err != nil {
        more = yield(err, SigBlock{})
        continue
      }
      more = yield(nil, blk)
    }
  }
  return blocks, close, nil
}

func ReadBlocksBytes(dir string) (
  func (yield func(err error, jblk []byte) bool), func(), error,
) {
  path := filepath.Join(dir, blocksFile)
  file, err := os.Open(path)
  if err != nil {
    return nil, nil, err
  }
  close := func() {
    file.Close()
  }
  blocks := func(yield func(err error, jblk []byte) bool) {
    sca := bufio.NewScanner(file)
    more := true
    for sca.Scan() && more {
      err := sca.Err()
      if err != nil {
        yield(err, nil)
        return
      }
      more = yield(nil, sca.Bytes())
    }
  }
  return blocks, close, nil
}

func VerifyBlock(blk SigBlock, authority Address) (bool, error) {
  hash := blk.Block.Hash().Bytes()
  pub, err := ecc.RecoverPubkey("P-256k1", hash, blk.Sig)
  if err != nil {
    return false, err
  }
  acc := NewAddress(pub)
  return acc == authority, nil
}
