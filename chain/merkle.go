package chain

import (
	"fmt"
	"math"
	"slices"
)

type position int

const (
  Left position = 1
  Right position = 2
)

type Proof[H comparable] struct {
  Hash H `json:"hash"`
  Pos position `json:"pos"`
}

func newProof[H comparable](hash H, pos position) Proof[H] {
  return Proof[H]{Hash: hash, Pos: pos}
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
  l := int(math.Pow(2, math.Ceil(math.Log2(float64(len(htxs)))) + 1) - 1)
  merkleTree := make([]H, l)
  chd := l / 2
  for i, j := 0, chd; i < len(htxs); i, j = i + 1, j + 1 {
    merkleTree[j] = htxs[i]
  }
  l, par := chd * 2, chd / 2
  for chd > 0 {
    for i, j := chd, par; i < l; i, j = i + 2, j + 1 {
      merkleTree[j] = pairHash(merkleTree[i], merkleTree[i + 1])
    }
    chd /= 2
    l, par = chd * 2, chd / 2
  }
  return merkleTree, nil
}

func MerkleProve[H comparable](txh H, merkleTree []H) ([]Proof[H], error) {
  if len(merkleTree) == 0 {
    return nil, fmt.Errorf("merkle prove: empty merkle tree")
  }
  start := len(merkleTree) / 2
  i := slices.Index(merkleTree[start:], txh)
  if i == -1 {
    return nil, fmt.Errorf("merkle prove: transaction %v not found", txh)
  }
  i += start
  if len(merkleTree) == 1 {
    return []Proof[H]{newProof(merkleTree[0], Left)}, nil
  }
  if len(merkleTree) == 3 {
    return []Proof[H]{
      newProof(merkleTree[1], Left), newProof(merkleTree[2], Right),
    }, nil
  }
  merkleProof := make([]Proof[H], 0)
  var nilHash H
  if i % 2 == 0 {
    merkleProof = append(merkleProof, newProof(merkleTree[i - 1], Left))
    merkleProof = append(merkleProof, newProof(merkleTree[i], Right))
    i--
  } else {
    merkleProof = append(merkleProof, newProof(merkleTree[i], Left))
    hash := merkleTree[i + 1]
    if hash != nilHash {
      merkleProof = append(merkleProof, newProof(hash, Right))
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
        merkleProof = append(merkleProof, newProof(hash, Right))
      } else {
        merkleProof = append(merkleProof, newProof(hash, Left))
      }
    }
    if i == 2 || i == 1 {
      break
    }
  }
  return merkleProof, nil
}

func MerkleVerify[H comparable](
  txh H, merkleProof []Proof[H], merkleRoot H, pairHash func(H, H) H,
) bool {
  i := slices.IndexFunc(merkleProof, func(proof Proof[H]) bool {
    return proof.Hash == txh
  })
  if i == -1 {
    return false
  }
  hash := merkleProof[0].Hash
  for i := 1; i < len(merkleProof); i++ {
    proof := merkleProof[i]
    if proof.Pos == Left {
      hash = pairHash(proof.Hash, hash)
    } else {
      hash = pairHash(hash, proof.Hash)
    }
  }
  return hash == merkleRoot
}
