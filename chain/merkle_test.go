package chain_test

import (
	"fmt"
	"testing"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

func TestMerkleHash(t *testing.T) {
  txs := []int{1, 2, 3, 4, 5, 6, 7, 8}
  // txs := []int{1, 2, 3, 4, 5, 6, 7}
  // txs := []int{1, 2, 3, 4, 5, 6}
  // txs := []int{1, 2, 3, 4, 5}
  // txs := []int{1, 2, 3, 4}
  // txs := []int{1, 2, 3}
  // txs := []int{1, 2}
  // txs := []int{1}
  // txs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}
  merkleTree, err := chain.MerkleHash(txs)
  if err != nil {
    t.Fatal(err)
  }
  merkleRoot := merkleTree[0]
  // for _, tx := range []int{6, 11} {
  for _, tx := range []int{4} {
    merkleProof, err := chain.MerkleProve(tx, merkleTree)
    if err != nil {
      t.Fatal(err)
    }
    valid := chain.MerkleVerify(tx, merkleProof, merkleRoot)
    fmt.Println(valid)
  }
}
