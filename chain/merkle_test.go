package chain_test

import (
	"fmt"
	"testing"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

func strRange(end int) []string {
  slc := make([]string, 0, end)
  for i := range end {
    slc = append(slc, fmt.Sprintf("%v", i + 1))
  }
  return slc
}

func TestMerkleHashProveVerify(t *testing.T) {
  for i := range 9 {
    txs := strRange(i + 1)
    merkleTree, err := chain.MerkleHash(txs)
    if err != nil {
      t.Fatal(err)
    }
    fmt.Println(merkleTree)
    merkleRoot := merkleTree[0]
    for _, tx := range txs {
      merkleProof, err := chain.MerkleProve(tx, merkleTree)
      if err != nil {
        t.Fatal(err)
      }
      fmt.Println(tx, merkleProof)
      valid := chain.MerkleVerify(tx, merkleProof, merkleRoot)
      _ = valid
    }
  }
}
