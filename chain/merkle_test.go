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
  if r == "" {
    return l
  }
  return l + r
}

func typeHash(s string) chain.Hash {
  return chain.NewHash(s)
}

func pairHash(l, r chain.Hash) chain.Hash {
  var nilHash chain.Hash
  if r == nilHash {
    return l
  }
  return chain.NewHash(l.String() + r.String())
  // return chain.NewHash([]chain.Hash{l, r})
}

func formatStrMerkleTree(merkleTree []string) string {
  mt := slices.Clone(merkleTree)
  for i := range mt {
    if mt[i] == "" {
      mt[i] = "_"
    }
  }
  return fmt.Sprintf("%v", mt)
}

func formatHashMerkleTree(merkleTree []chain.Hash) string {
  mt := make([]string, len(merkleTree))
  for i := range merkleTree {
    mt[i] = fmt.Sprintf("%.4s", merkleTree[i])
  }
  return fmt.Sprintf("%v", mt)
}

func TestMerkleHashProveVerify(t *testing.T) {
  for i := range 9 {
    txs := strRange(i + 1)
    merkleTree, err := chain.MerkleHash(txs, typeHash, pairHash)
    if err != nil {
      t.Fatal(err)
    }
    fmt.Printf("Tree (%v) %v\n", len(txs), formatHashMerkleTree(merkleTree))
    merkleRoot := merkleTree[0]
    for _, tx := range txs {
      merkleProof, err := chain.MerkleProve(typeHash(tx), merkleTree)
      if err != nil {
        t.Fatal(err)
      }
      fmt.Printf("Proof %v %.4s ", tx, typeHash(tx))
      valid := chain.MerkleVerify(typeHash(tx), merkleProof, merkleRoot, pairHash)
      if valid {
        fmt.Println("valid")
      } else {
        fmt.Println("INVALID")
      }
      if !valid {
        t.Errorf("invalid Merkle proof: %v %.4s", tx, typeHash(tx))
      }
    }
  }
}

func TestStrMerkleHashProveVerify(t *testing.T) {
  for i := range 9 {
    txs := strRange(i + 1)
    merkleTree, err := chain.MerkleHash(txs, strTypeHash, strPairHash)
    if err != nil {
      t.Fatal(err)
    }
    fmt.Println("Tree", formatStrMerkleTree(merkleTree))
    merkleRoot := merkleTree[0]
    for _, tx := range txs {
      merkleProof, err := chain.MerkleProve(tx, merkleTree)
      if err != nil {
        t.Fatal(err)
      }
      fmt.Printf("Proof %v %v ", tx, merkleProof)
      valid := chain.MerkleVerify(tx, merkleProof, merkleRoot, strPairHash)
      if valid {
        fmt.Println("valid")
      } else {
        fmt.Println("INVALID")
      }
      if !valid {
        t.Errorf("invalid Merkle proof: %v %v", tx, merkleProof)
      }
    }
  }
}
