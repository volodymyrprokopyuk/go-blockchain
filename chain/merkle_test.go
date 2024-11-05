package chain_test

import (
	"testing"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

func TestMerkleHash(t *testing.T) {
  txs := []string{"1", "2", "3", "4", "5", "6", "7", "8"}
  // txs := []string{"1", "2", "3", "4", "5", "6", "7"}
  // txs := []string{"1", "2", "3", "4", "5", "6"}
  // txs := []string{"1", "2", "3", "4", "5"}
  // txs := []string{"1", "2", "3", "4"}
  // txs := []string{"1", "2", "3"}
  // txs := []string{"1", "2"}
  // txs := []string{"1"}
  // txs := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14"}
  merkleTree, err := chain.MerkleHash(txs)
  if err != nil {
    t.Fatal(err)
  }
  _ = merkleTree
  // merkleRoot := merkleTree[0]
  // for _, tx := range txs {
  //   merkleProof, err := chain.MerkleProve(tx, merkleTree)
  //   if err != nil {
  //     t.Fatal(err)
  //   }
  //   valid := chain.MerkleVerify(tx, merkleProof, merkleRoot)
  //   fmt.Println(valid)
  // }
}
