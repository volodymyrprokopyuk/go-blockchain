package state

import (
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/account"
	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/store"
)

type State struct {
  balances map[chain.Address]uint
  nonces map[chain.Address]uint
  lastBlock store.Block
  pndTxs map[chain.Hash]chain.SigTx
  pndBals map[chain.Address]uint
  pndNces map[chain.Address]uint
}

func NewState(gen store.Genesis) *State {
  n := 100
  sta := &State{
    balances: make(map[chain.Address]uint, n),
    nonces: make(map[chain.Address]uint, n),
    pndTxs: make(map[chain.Hash]chain.SigTx, n),
    pndBals: make(map[chain.Address]uint, n),
    pndNces: make(map[chain.Address]uint, n),
  }
  for addr, amount := range gen.Balances {
    sta.balances[addr] = amount
    sta.pndBals[addr] = amount
  }
  return sta
}

func (s *State) String() string {
  var bld strings.Builder
  bld.WriteString("Balances\n")
  for addr, amount := range s.balances {
    bld.WriteString(fmt.Sprintf("  %.7s: %20d\n", addr, amount))
  }
  bld.WriteString("Nonces\n")
  bld.WriteString("Last block\n")
  bld.WriteString(fmt.Sprintf("  %s\n", s.lastBlock))
  for addr, nonce := range s.nonces {
    bld.WriteString(fmt.Sprintf("  %.7s: %5d\n", addr, nonce))
  }
  if len(s.pndTxs) > 0 {
    bld.WriteString("Pending txs\n")
    for _, tx := range s.pndTxs {
      bld.WriteString(fmt.Sprintf("  %s\n", tx))
    }
  }
  if len(s.pndBals) > 0 {
    bld.WriteString("Pending balances\n")
    for addr, amount := range s.pndBals {
      bld.WriteString(fmt.Sprintf("  %.7s: %20d\n", addr, amount))
    }
  }
  if len(s.pndNces) > 0 {
    bld.WriteString("Pending nonces\n")
    for addr, nonce := range s.pndNces {
      bld.WriteString(fmt.Sprintf("  %.7s: %26d\n", addr, nonce))
    }
  }
  return bld.String()
}

func (s *State) clone() *State {
  return &State{
    balances: maps.Clone(s.balances),
    nonces: maps.Clone(s.nonces),
    lastBlock: s.lastBlock,
  }
}

func (s *State) Send(from account.Account, to chain.Address, value uint) error {
  addr := from.Address()
  if s.pndBals[addr] < value {
    return fmt.Errorf("%v insufficient funds", addr)
  }
  nonce := s.pndNces[addr] + 1
  tx := chain.Tx{
    From: addr, To: to, Value: value, Nonce: nonce, Time: time.Now(),
  }
  stx, err := from.Sign(tx)
  if err != nil {
    return err
  }
  hash, err := stx.Hash()
  if err != nil {
    return err
  }
  s.pndTxs[hash] = stx
  s.pndBals[addr] -= value
  s.pndBals[to] += value
  s.pndNces[addr] = nonce
  return nil
}
