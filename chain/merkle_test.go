package chain_test

import (
	"testing"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

func TestMerkleHash(t *testing.T) {
  chain.MerkleHash([]int{1, 2, 3, 4, 5, 6, 7, 8})
  chain.MerkleHash([]int{1, 2, 3, 4, 5, 6, 7})
  chain.MerkleHash([]int{1, 2, 3, 4, 5, 6})
  chain.MerkleHash([]int{1, 2, 3, 4, 5})
  chain.MerkleHash([]int{1, 2, 3, 4})
  chain.MerkleHash([]int{1, 2, 3})
  chain.MerkleHash([]int{1, 2})
  chain.MerkleHash([]int{1})
  chain.MerkleHash([]int{})
  chain.MerkleHash([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14})
}
