package state

import (
	"fmt"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/account"
	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/store"
)

type State struct {
  balances map[chain.Address]uint
  nonces map[chain.Address]uint
  newTxs map[chain.Hash]chain.SigTx
  arcTxs map[chain.Hash]chain.SigTx
}

func NewState(gen store.Genesis) *State {
  n := 100
  sta := &State{
    balances: make(map[chain.Address]uint, n),
    nonces: make(map[chain.Address]uint, n),
    newTxs: make(map[chain.Hash]chain.SigTx, n),
    arcTxs: make(map[chain.Hash]chain.SigTx, n),
  }
  for addr, amount := range gen.Balances {
    sta.balances[addr] = amount
    sta.nonces[addr] = 0
  }
  return sta
}

func (s *State) Send(from account.Account, to chain.Address, value uint) error {
  addr := from.Address()
  if s.balances[addr] < value {
    return fmt.Errorf("%v insufficient funds", addr)
  }
  s.nonces[addr]++
  tx := chain.Tx{
    From: addr, To: to, Value: value, Nonce: s.nonces[addr], Time: time.Now(),
  }
  stx, err := from.Sign(tx)
  if err != nil {
    return err
  }
  hash, err := stx.Hash()
  if err != nil {
    return err
  }
  s.newTxs[hash] = stx
  return nil
}
