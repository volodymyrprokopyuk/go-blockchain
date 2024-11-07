package chain_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

func strRange(end int) []string {
  slc := make([]string, end)
  for i := range end {
    slc[i] = fmt.Sprintf("%v", i + 1)
  }
  return slc
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
}

func formatMerkleTree(merkleTree []chain.Hash) string {
  mt := make([]string, len(merkleTree))
  for i := range merkleTree {
    mt[i] = fmt.Sprintf("%.4s", merkleTree[i])
  }
  return fmt.Sprintf("%v", mt)
}

func formatMerkleProof(merkleProof []chain.Proof[chain.Hash]) string {
  mp := make([]string, len(merkleProof))
  for i, proof := range merkleProof {
    var pos string
    if proof.Pos == chain.Left {
      pos = "L"
    } else {
      pos = "R"
    }
    mp[i] = fmt.Sprintf("%.4s-%v", proof.Hash, pos)
  }
  return fmt.Sprintf("%v", mp)
}

func TestMerkleHashProveVerify(t *testing.T) {
  for i := range 9 {
    txs := strRange(i + 1)
    merkleTree, err := chain.MerkleHash(txs, typeHash, pairHash)
    if err != nil {
      t.Fatal(err)
    }
    fmt.Printf("Tree (%v) %v\n", len(txs), formatMerkleTree(merkleTree))
    merkleRoot := merkleTree[0]
    for _, tx := range txs {
      txh := typeHash(tx)
      merkleProof, err := chain.MerkleProve(txh, merkleTree)
      if err != nil {
        t.Fatal(err)
      }
      fmt.Printf("Proof %v %.4s %v ", tx, txh, formatMerkleProof(merkleProof))
      valid := chain.MerkleVerify(txh, merkleProof, merkleRoot, pairHash)
      if valid {
        fmt.Println("valid")
      } else {
        fmt.Println("INVALID")
      }
      if !valid {
        t.Errorf(
          "invalid Merkle proof: %v %.4s %v",
          tx, txh,formatMerkleProof(merkleProof),
        )
      }
    }
  }
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

func formatMerkleTreeStr(merkleTree []string) string {
  mt := slices.Clone(merkleTree)
  for i := range mt {
    if mt[i] == "" {
      mt[i] = "_"
    }
  }
  return fmt.Sprintf("%v", mt)
}

func formatMerkleProofStr(merkleProof []chain.Proof[string]) string {
  mp := make([]string, len(merkleProof))
  for i, proof := range merkleProof {
    var pos string
    if proof.Pos == chain.Left {
      pos = "L"
    } else {
      pos = "R"
    }
    mp[i] = fmt.Sprintf("%v-%v", proof.Hash, pos)
  }
  return fmt.Sprintf("%v", mp)
}

func TestMerkleHashProveVerifyStr(t *testing.T) {
  for i := range 9 {
    txs := strRange(i + 1)
    merkleTree, err := chain.MerkleHash(txs, strTypeHash, strPairHash)
    if err != nil {
      t.Fatal(err)
    }
    fmt.Printf("Tree (%v) %v\n", len(txs), formatMerkleTreeStr(merkleTree))
    merkleRoot := merkleTree[0]
    for _, tx := range txs {
      txh := strTypeHash(tx)
      merkleProof, err := chain.MerkleProve(txh, merkleTree)
      if err != nil {
        t.Fatal(err)
      }
      fmt.Printf("Proof %v %v ", txh, formatMerkleProofStr(merkleProof))
      valid := chain.MerkleVerify(txh, merkleProof, merkleRoot, strPairHash)
      if valid {
        fmt.Println("valid")
      } else {
        fmt.Println("INVALID")
      }
      if !valid {
        t.Errorf(
          "invalid Merkle proof: %v %v", txh, formatMerkleProofStr(merkleProof),
        )
      }
    }
  }
}
