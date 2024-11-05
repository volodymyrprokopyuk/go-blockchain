package chain

import (
	"fmt"
	"math"
	"slices"
	"strconv"
)

func pairHash(l, r int) int {
  if r != 0 {
    hash, _ := strconv.Atoi(fmt.Sprintf("%d%d", l, r))
    return hash
  }
  return l
}

func MerkleHash(txs []int) ([]int, error) {
  if len(txs) == 0 {
    return nil, fmt.Errorf("merkle hash: empty transaction list")
  }
  halfFloor := func(i int) int {
    return int(math.Floor(float64(i / 2)))
  }
  l := int(math.Pow(2, math.Ceil(math.Log2(float64(len(txs)))) + 1) - 1)
  merkleTree := make([]int, l)
  chd := halfFloor(l)
  for i, j := 0, chd; i < len(txs); i, j = i + 1, j + 1 {
    merkleTree[j] = txs[i]
  }
  l, par := chd + len(txs), halfFloor(chd)
  for chd > 0 {
    for i, j := chd, par; i < l; i, j = i + 2, j + 1 {
      merkleTree[j] = pairHash(merkleTree[i], merkleTree[i + 1])
    }
    chd = halfFloor(chd)
    l, par = chd * 2, halfFloor(chd)
  }
  fmt.Println(merkleTree)
  return merkleTree, nil
}

func MerkleProve(tx int, merkeTree []int) ([]int, error) {
  i := slices.Index(merkeTree, tx)
  if i == -1 {
    return nil, fmt.Errorf("merkle prove: transaction %v not found", tx)
  }
  parentRight := func(i int) int {
    return (i - 1) / 2 + 1
  }
  parentLeft := func(i int) int {
    return (i - 2) / 2 - 1
  }
  merkleProof := make([]int, 0)
  if i % 2 == 0 {
    i--
    merkleProof = append(merkleProof, merkeTree[i])
    merkleProof = append(merkleProof, merkeTree[i + 1])
  } else {
    merkleProof = append(merkleProof, merkeTree[i])
    merkleProof = append(merkleProof, merkeTree[i + 1])
    i++
  }
  for {
    if i % 2 == 0 {
      i = parentLeft(i)
    } else {
      i = parentRight(i)
    }
    merkleProof = append(merkleProof, merkeTree[i])
    if i == 2 || i == 1 {
      break
    }
  }
  fmt.Println(merkleProof)
  return merkleProof, nil
}

func MerkleVerify(tx int, merkleProof []int, merkleRoot int) bool {
  return true
}
