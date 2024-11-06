package chain

import (
	"fmt"
	"math"
	"slices"
)

type place int

const (
  left place = 1
  right place = 2
)

type proofStep[H comparable] struct {
  hash H
  place place
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

func MerkleProve[H comparable](tx H, merkleTree []H) ([]proofStep[H], error) {
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
    return []proofStep[H]{{hash: merkleTree[0], place: left}}, nil
  }
  if len(merkleTree) == 3 {
    return []proofStep[H]{
      {hash: merkleTree[1], place: left}, {hash: merkleTree[2], place: right},
    }, nil
  }
  merkleProof := make([]proofStep[H], 0)
  addStep := func(hash H, place place) {
    merkleProof = append(merkleProof, proofStep[H]{hash: hash, place: place})
  }
  var nilHash H
  if i % 2 == 0 {
    addStep(merkleTree[i - 1], left)
    addStep(merkleTree[i], right)
    i--
  } else {
    addStep(merkleTree[i], left)
    hash := merkleTree[i + 1]
    if hash != nilHash {
      addStep(hash, right)
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
    if hash != nilHash {
      if i % 2 == 0 {
        addStep(hash, right)
      } else {
        addStep(hash, left)
      }
    }
    if i == 2 || i == 1 {
      break
    }
  }
  return merkleProof, nil
}

func MerkleVerify[H comparable](
  tx H, merkleProof []proofStep[H], merkleRoot H, pairHash func(H, H) H,
) bool {
  i := slices.IndexFunc(merkleProof, func(step proofStep[H]) bool {
    return step.hash == tx
  })
  if i == -1 {
    return false
  }
  hash := merkleProof[0].hash
  for i := 1; i < len(merkleProof); i++ {
    step := merkleProof[i]
    if step.place == left {
      hash = pairHash(step.hash, hash)
    } else {
      hash = pairHash(hash, step.hash)
    }
  }
  return hash == merkleRoot
}



// func MerkleProve2[H comparable](tx H, merkleTree []H) ([]H, error) {
//   if len(merkleTree) == 0 {
//     return nil, fmt.Errorf("merkle prove: empty merkle tree")
//   }
//   start := int(math.Floor(float64(len(merkleTree) / 2)))
//   i := slices.Index(merkleTree[start:], tx)
//   if i == -1 {
//     return nil, fmt.Errorf("merkle prove: transaction %v not found", tx)
//   }
//   i += start
//   if len(merkleTree) == 1 {
//     return []H{merkleTree[0]}, nil
//   }
//   if len(merkleTree) == 3 {
//     return []H{merkleTree[1], merkleTree[2]}, nil
//   }
//   stk, que := make([]H, 0), make([]H, 0)
//   var nilHash H
//   if i % 2 == 0 {
//     stk = append(stk, merkleTree[i - 1])
//     que = append(que, merkleTree[i])
//     i--
//   } else {
//     stk = append(stk, merkleTree[i])
//     hash := merkleTree[i + 1]
//     if hash != nilHash {
//       que = append(que, hash)
//     }
//     i++
//   }
//   for {
//     if i % 2 == 0 {
//       i = (i - 2) / 2
//     } else {
//       i = (i - 1) / 2
//     }
//     if i % 2 == 0 {
//       i--
//     } else {
//       i++
//     }
//     hash := merkleTree[i]
//     if hash != nilHash {
//       if i % 2 == 0 {
//         que = append(que, hash)
//       } else {
//         stk = append(stk, hash)
//       }
//     }
//     if i == 2 || i == 1 {
//       break
//     }
//   }
//   merkleProof := make([]H, 0, len(stk) + len(que))
//   slices.Reverse(stk)
//   merkleProof = append(merkleProof, stk...)
//   merkleProof = append(merkleProof, que...)
//   return merkleProof, nil
// }

// func MerkleVerify2[H comparable](
//   tx H, merkleProof []H, merkleRoot H, pairHash func(H, H) H,
// ) bool {
//   i := slices.Index(merkleProof, tx)
//   if i == -1 {
//     return false
//   }
//   hash := merkleProof[0]
//   for i := 1; i < len(merkleProof); i++ {
//     hash = pairHash(hash, merkleProof[i])
//   }
//   return hash == merkleRoot
// }
