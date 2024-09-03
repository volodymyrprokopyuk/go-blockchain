package state

import (
	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/store"
)

type State struct {
  balances map[chain.Address]uint
  nonces map[chain.Address]uint
  newTxs map[chain.Hash]chain.Tx
  arcTxs map[chain.Hash]chain.Tx
}

func NewState(gen store.Genesis) *State {
  n := 100
  sta := &State{
    balances: make(map[chain.Address]uint, n),
    nonces: make(map[chain.Address]uint, n),
    newTxs: make(map[chain.Hash]chain.Tx, n),
    arcTxs: make(map[chain.Hash]chain.Tx, n),
  }
  for addr, amount := range gen.Balances {
    sta.balances[addr] = amount
    sta.nonces[addr] = 0
  }
  return sta
}
