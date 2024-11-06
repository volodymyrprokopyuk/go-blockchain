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

func strTypeHash(s string) string {
  return s
}

func strPairHash(l, r string) string {
  if r != "" {
    return l + r
  } else {
    return l
  }
}

func typeHash(s string) chain.Hash {
  return chain.NewHash(s)
}

func pairHash(l, r chain.Hash) chain.Hash {
  return chain.NewHash([]chain.Hash{l, r})
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

func TestMerkleHash_ProveVerifyStr(t *testing.T) {
  for i := range 9 {
    txs := strRange(i + 1)
    merkleTree, err := chain.MerkleHash(txs, strTypeHash, strPairHash)
    if err != nil {
      t.Fatal(err)
    }
    printMerkleTree(merkleTree)
    merkleRoot := merkleTree[0]
    for _, tx := range txs {
      merkleProof, err := chain.MerkleProve(tx, merkleTree)
      if err != nil {
        t.Fatal(err)
      }
      fmt.Println("Proof", tx, merkleProof)
      valid := chain.MerkleVerify(tx, merkleProof, merkleRoot, strPairHash)
      if !valid {
        t.Errorf("invalid Merkle proof: %v %v", tx, merkleProof)
      }
    }
  }
}

func TestMerkleHashProveVerify(t *testing.T) {
  for i := range 9 {
    txs := strRange(i + 1)
    merkleTree, err := chain.MerkleHash(txs, typeHash, pairHash)
    if err != nil {
      t.Fatal(err)
    }
    fmt.Println("Tree", merkleTree)
    merkleRoot := merkleTree[0]
    for _, tx := range txs {
      merkleProof, err := chain.MerkleProve(typeHash(tx), merkleTree)
      if err != nil {
        t.Fatal(err)
      }
      fmt.Println("Proof", tx, merkleProof)
      valid := chain.MerkleVerify(typeHash(tx), merkleProof, merkleRoot, pairHash)
      if !valid {
        t.Fatalf("invalid Merkle proof: %v %v", tx, merkleProof)
      }
    }
  }
}
