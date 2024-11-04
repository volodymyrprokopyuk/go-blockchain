package chain

import (
	"fmt"
	"math"
	"strconv"
)

func MerkleHash(txs []int) ([]int, error) {
  if len(txs) == 0 {
    return nil, fmt.Errorf("merkle hash: empty transaction list")
  }
  halfFloor := func(v int) int {
    return int(math.Floor(float64(v / 2)))
  }
  l := int(math.Pow(2, math.Ceil(math.Log2(float64(len(txs)))) + 1) - 1)
  mt := make([]int, l)
  chd := halfFloor(l)
  for i, j := 0, chd; i < len(txs); i, j = i + 1, j + 1 {
    mt[j] = txs[i]
  }
  l, par := chd + len(txs), halfFloor(chd)
  for chd > 0 {
    for i, j := chd, par; i < l; i, j = i + 2, j + 1 {
      if mt[i + 1] != 0 {
        mt[j], _ = strconv.Atoi(fmt.Sprintf("%d%d", mt[i], mt[i + 1]))
      } else {
        mt[j] = mt[i]
      }
    }
    chd = halfFloor(chd)
    l, par = chd * 2, halfFloor(chd)
  }
  fmt.Println(mt)
  return mt, nil
}
