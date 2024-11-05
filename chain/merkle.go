package chain

import (
	"fmt"
	"math"
	"slices"
)

func hashPair(l, r string) string {
  if r != "_" {
    return l + r
  } else {
    return l
  }
}

func MerkleHash(txs []string) ([]string, error) {
  if len(txs) == 0 {
    return nil, fmt.Errorf("merkle hash: empty transaction list")
  }
  halfFloor := func(i int) int {
    return int(math.Floor(float64(i / 2)))
  }
  l := int(math.Pow(2, math.Ceil(math.Log2(float64(len(txs)))) + 1) - 1)
  merkleTree := make([]string, 0, l)
  for range l {
    merkleTree = append(merkleTree, "_")
  }
  chd := halfFloor(l)
  for i, j := 0, chd; i < len(txs); i, j = i + 1, j + 1 {
    merkleTree[j] = txs[i]
  }
  l, par := chd * 2, halfFloor(chd)
  for chd > 0 {
    for i, j := chd, par; i < l; i, j = i + 2, j + 1 {
      merkleTree[j] = hashPair(merkleTree[i], merkleTree[i + 1])
    }
    chd = halfFloor(chd)
    l, par = chd * 2, halfFloor(chd)
  }
  return merkleTree, nil
}

func MerkleProve(tx string, merkleTree []string) ([]string, error) {
  start := int(math.Floor(float64(len(merkleTree) / 2)))
  i := slices.Index(merkleTree[start:], tx)
  if i == -1 {
    return nil, fmt.Errorf("merkle prove: transaction %v not found", tx)
  }
  i = start + i
  merkleProof := make([]string, 0)
  if len(merkleTree) == 1 {
    merkleProof = append(merkleProof, merkleTree[0])
    return merkleProof, nil
  }
  if len(merkleTree) == 3 {
    merkleProof = append(merkleProof, merkleTree[1], merkleTree[2])
    return merkleProof, nil
  }
  if i % 2 == 0 {
    i--
    merkleProof = append(merkleProof, merkleTree[i])
    merkleProof = append(merkleProof, merkleTree[i + 1])
  } else {
    merkleProof = append(merkleProof, merkleTree[i])
    hash := merkleTree[i + 1]
    if hash != "_" {
      merkleProof = append(merkleProof, hash)
    }
    i++
  }
  for {
    if i % 2 == 0 {
      i = (i - 2) / 2
    } else {
      i = (i - 1) / 2
    }
    if i % 2 == 0 {
      i--
    } else {
      i++
    }
    hash := merkleTree[i]
    if hash != "_" {
      merkleProof = append(merkleProof, hash)
    }
    if i == 2 || i == 1 {
      break
    }
  }
  return merkleProof, nil
}

func MerkleVerify(tx string, merkleProof []string, merkleRoot string) bool {
  return true
}
