package chain_test

import (
	"fmt"
	"slices"
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

func printMerkleTree(merkleTree []string) {
  mt := slices.Clone(merkleTree)
  for i := range mt {
    if mt[i] == "" {
      mt[i] = "_"
    }
  }
  fmt.Println("Tree", mt)
}

func TestMerkleHashProveVerify(t *testing.T) {
  for i := range 9 {
    txs := strRange(i + 1)
    merkleTree, err := chain.MerkleHash(txs, chain.StrTypeHash, chain.StrPairHash)
    if err != nil {
      t.Fatal(err)
    }
    printMerkleTree(merkleTree)
    // fmt.Println("Tree", merkleTree)
    // merkleRoot := merkleTree[0]
    // for _, tx := range txs {
    //   merkleProof, err := chain.MerkleProveStr(tx, merkleTree)
    //   if err != nil {
    //     t.Fatal(err)
    //   }
    //   // fmt.Println("Proof", tx, merkleProof)
    //   valid := chain.MerkleVerifyStr(tx, merkleProof, merkleRoot)
    //   if !valid {
    //     t.Errorf("invalid Merkle proof: %v %v", tx, merkleProof)
    //   }
    // }
  }
}

func TestMerkleHashProveVerifyStr(t *testing.T) {
  for i := range 9 {
    txs := strRange(i + 1)
    merkleTree, err := chain.MerkleHashStr(txs)
    if err != nil {
      t.Fatal(err)
    }
    fmt.Println("Tree", merkleTree)
    merkleRoot := merkleTree[0]
    for _, tx := range txs {
      merkleProof, err := chain.MerkleProveStr(tx, merkleTree)
      if err != nil {
        t.Fatal(err)
      }
      // fmt.Println("Proof", tx, merkleProof)
      valid := chain.MerkleVerifyStr(tx, merkleProof, merkleRoot)
      if !valid {
        t.Errorf("invalid Merkle proof: %v %v", tx, merkleProof)
      }
    }
  }
}
