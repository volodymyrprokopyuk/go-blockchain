package chain

import (
	"fmt"
	"math"
	"slices"
)

type Hstr string

func (s Hstr) Hash() string {
  return string(s)
}

// func (s Hstr) String() string {
//   if s == "" {
//     return "_"
//   }
//   return string(s)
// }

func StrTypeHash(s string) string {
  return s
}

func StrPairHash(l, r string) string {
  if r != "" {
    return l + r
  } else {
    return l
  }
}

type Hasher[H comparable] interface {
  Hash() H
}

func MerkleHash[T any, H comparable](
  txs []T, typeHash func(T) H, pairHash func(H, H) H,
) ([]H, error) {
  if len(txs) == 0 {
    return nil, fmt.Errorf("merkle hash: empty transaction list")
  }
  htxs := make([]H, len(txs))
  for i, tx := range txs {
    htxs[i] = typeHash(tx)
  }
  halfFloor := func(i int) int {
    return int(math.Floor(float64(i / 2)))
  }
  l := int(math.Pow(2, math.Ceil(math.Log2(float64(len(htxs)))) + 1) - 1)
  merkleTree := make([]H, l)
  chd := halfFloor(l)
  for i, j := 0, chd; i < len(htxs); i, j = i + 1, j + 1 {
    merkleTree[j] = htxs[i]
  }
  l, par := chd * 2, halfFloor(chd)
  for chd > 0 {
    for i, j := chd, par; i < l; i, j = i + 2, j + 1 {
      merkleTree[j] = pairHash(merkleTree[i], merkleTree[i + 1])
    }
    chd = halfFloor(chd)
    l, par = chd * 2, halfFloor(chd)
  }
  return merkleTree, nil
}

func hashPair(l, r string) string {
  if r != "_" {
    return l + r
  } else {
    return l
  }
}

func MerkleHashStr(txs []string) ([]string, error) {
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

func MerkleProveStr(tx string, merkleTree []string) ([]string, error) {
  if len(merkleTree) == 0 {
    return nil, fmt.Errorf("merkle prove: empty merkle tree")
  }
  start := int(math.Floor(float64(len(merkleTree) / 2)))
  i := slices.Index(merkleTree[start:], tx)
  if i == -1 {
    return nil, fmt.Errorf("merkle prove: transaction %v not found", tx)
  }
  i += start
  if len(merkleTree) == 1 {
    return []string{merkleTree[0]}, nil
  }
  if len(merkleTree) == 3 {
    return []string{merkleTree[1], merkleTree[2]}, nil
  }
  stk, que := make([]string, 0), make([]string, 0)
  if i % 2 == 0 {
    stk = append(stk, merkleTree[i - 1])
    que = append(que, merkleTree[i])
    i--
  } else {
    stk = append(stk, merkleTree[i])
    hash := merkleTree[i + 1]
    if hash != "_" {
      que = append(que, hash)
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
      if i % 2 == 0 {
        que = append(que, hash)
      } else {
        stk = append(stk, hash)
      }
    }
    if i == 2 || i == 1 {
      break
    }
  }
  merkleProof := make([]string, 0, len(stk) + len(que))
  slices.Reverse(stk)
  merkleProof = append(merkleProof, stk...)
  merkleProof = append(merkleProof, que...)
  return merkleProof, nil
}

func MerkleVerifyStr(tx string, merkleProof []string, merkleRoot string) bool {
  i := slices.Index(merkleProof, tx)
  if i == -1 {
    return false
  }
  hash := merkleProof[0]
  for i := 1; i < len(merkleProof); i++ {
    hash = hashPair(hash, merkleProof[i])
  }
  return hash == merkleRoot
}
