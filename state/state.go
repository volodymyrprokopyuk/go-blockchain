package state

import (
	"fmt"
	"maps"
	"slices"
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
  sta := &State{
    balances: make(map[chain.Address]uint, 10),
    nonces: make(map[chain.Address]uint, 10),
    pndTxs: make(map[chain.Hash]chain.SigTx, 10),
    pndBals: make(map[chain.Address]uint, 10),
    pndNces: make(map[chain.Address]uint, 10),
  }
  for addr, amount := range gen.Balances {
    sta.balances[addr] = amount
    sta.pndBals[addr] = amount
  }
  return sta
}

func (s *State) Clone() *State {
  return &State{
    balances: maps.Clone(s.balances),
    nonces: maps.Clone(s.nonces),
    lastBlock: s.lastBlock,
    pndTxs: maps.Clone(s.pndTxs),
  }
}

func (s *State) Apply(sta *State) {
  s.balances = sta.balances
  s.nonces = sta.nonces
  s.lastBlock = sta.lastBlock
  s.pndTxs = make(map[chain.Hash]chain.SigTx, 10)
  s.pndBals = maps.Clone(s.balances)
  s.pndNces = maps.Clone(s.nonces)
}

func (s *State) String() string {
  var bld strings.Builder
  bld.WriteString("Balances\n")
  for addr, amount := range s.balances {
    bld.WriteString(fmt.Sprintf("  %.7s: %20d\n", addr, amount))
  }
  bld.WriteString("Nonces\n")
  for addr, nonce := range s.nonces {
    bld.WriteString(fmt.Sprintf("  %.7s: %26d\n", addr, nonce))
  }
  bld.WriteString("Last block\n")
  bld.WriteString(fmt.Sprintf("  %s\n", s.lastBlock))
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

func (s *State) Send(from account.Account, to chain.Address, value uint) error {
  addr := from.Address()
  if s.pndBals[addr] < value {
    return fmt.Errorf("insufficient funds %.7s", addr)
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

func (s *State) applyTx(tx chain.SigTx) error {
  hash, err := tx.Hash()
  if err != nil {
    return err
  }
  valid, err := account.Verify(tx)
  if err != nil {
    return err
  }
  if !valid {
    return fmt.Errorf("invalid signature %.7s", hash)
  }
  if s.balances[tx.From] < tx.Value {
    return fmt.Errorf("insufficient funds %.7s", hash)
  }
  if tx.Nonce != s.nonces[tx.From] + 1 {
    return fmt.Errorf("invalid nonce %.7s", hash)
  }
  s.balances[tx.From] -= tx.Value
  s.balances[tx.To] += tx.Value
  s.nonces[tx.From]++
  return nil
}

func (s *State) CreateBlock() (store.Block, error) {
  pndTxs := make([]chain.SigTx, 0, len(s.pndTxs))
  for _, tx := range s.pndTxs {
    pndTxs = append(pndTxs, tx)
  }
  slices.SortFunc(pndTxs, func(a, b chain.SigTx) int {
    cmp := strings.Compare(string(a.From), string(b.From))
    if cmp != 0 {
      return cmp
    }
    return int(a.Nonce) - int(b.Nonce)
  })
  vldTxs := make([]chain.SigTx, 0, len(s.pndTxs))
  for _, tx := range pndTxs {
    err := s.applyTx(tx)
    if err != nil {
      fmt.Printf("REJECTED %s\n", err)
      continue
    }
    vldTxs = append(vldTxs, tx)
  }
  if len(vldTxs) == 0 {
    return store.Block{}, fmt.Errorf("none of txs is valid")
  }
  blk := store.Block{
    Number: s.lastBlock.Number + 1, Parent: s.lastBlock.Parent,
    Time: time.Now(), Txs: vldTxs,
  }
  return blk, nil
}

func (s *State) ApplyBlock(blk store.Block) error {
  hash, err := blk.Hash()
  if err != nil {
    return err
  }
  if blk.Number != s.lastBlock.Number + 1 {
    return fmt.Errorf("invalid block number %.7s", hash)
  }
  lstHash, err := s.lastBlock.Hash()
  if err != nil {
    return err
  }
  if blk.Parent != lstHash {
    return fmt.Errorf("invalid parent hash %.7s", hash)
  }
  for _, tx := range blk.Txs {
    err := s.applyTx(tx)
    if err != nil {
      return err
    }
  }
  s.lastBlock = blk
  return nil
}
