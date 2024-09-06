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
  txs map[chain.Hash]chain.SigTx
  Pending *State
}

func (s *State) Nonce(addr chain.Address) uint {
  return s.nonces[addr]
}

func NewState(gen store.Genesis) *State {
  return &State{
    balances: maps.Clone(gen.Balances),
    nonces: make(map[chain.Address]uint),
    txs: make(map[chain.Hash]chain.SigTx),
    Pending: &State{
      balances: maps.Clone(gen.Balances),
      nonces: make(map[chain.Address]uint),
      txs: make(map[chain.Hash]chain.SigTx),
    },
  }
}

func (s *State) Clone2() *State {
  sta := &State{
    balances: maps.Clone(s.balances),
    nonces: maps.Clone(s.nonces),
    lastBlock: s.lastBlock,
    txs: maps.Clone(s.txs),
  }
  sta.Pending = &State{txs: maps.Clone(s.Pending.txs)}
  return sta
}

func (s *State) Clone() *State {
  return &State{
    balances: maps.Clone(s.balances),
    nonces: maps.Clone(s.nonces),
    lastBlock: s.lastBlock,
    txs: maps.Clone(s.txs),
    Pending: &State{
      txs: maps.Clone(s.Pending.txs),
    },
  }
}

func (s *State) ResetPending() {
  s.Pending = &State{
    balances: maps.Clone(s.balances),
    nonces: maps.Clone(s.nonces),
    txs: make(map[chain.Hash]chain.SigTx),
  }
}

func (s *State) Apply(sta *State) {
  s.balances = sta.balances
  s.nonces = sta.nonces
  s.lastBlock = sta.lastBlock
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
  if s.Pending != nil && len(s.Pending.txs) > 0 {
    bld.WriteString("Pending txs\n")
    for _, tx := range s.Pending.txs {
      bld.WriteString(fmt.Sprintf("  %s\n", tx))
    }
  }
  if s.Pending != nil && len(s.Pending.balances) > 0 {
    bld.WriteString("Pending balances\n")
    for addr, amount := range s.Pending.balances {
      bld.WriteString(fmt.Sprintf("  %.7s: %20d\n", addr, amount))
    }
  }
  if s.Pending != nil && len(s.Pending.nonces) > 0 {
    bld.WriteString("Pending nonces\n")
    for addr, nonce := range s.Pending.nonces {
      bld.WriteString(fmt.Sprintf("  %.7s: %26d\n", addr, nonce))
    }
  }
  return bld.String()
}

func (s *State) ApplyTx(tx chain.SigTx) error {
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
  if tx.Nonce != s.nonces[tx.From] + 1 {
    return fmt.Errorf("invalid nonce %.7s", hash)
  }
  if s.balances[tx.From] < tx.Value {
    return fmt.Errorf("insufficient funds %.7s", hash)
  }
  s.balances[tx.From] -= tx.Value
  s.balances[tx.To] += tx.Value
  s.nonces[tx.From]++
  s.txs[hash] = tx
  return nil
}

func (s *State) CreateBlock() (store.Block, error) {
  pndTxs := make([]chain.SigTx, 0, len(s.Pending.txs))
  for _, tx := range s.Pending.txs {
    pndTxs = append(pndTxs, tx)
  }
  slices.SortFunc(pndTxs, func(a, b chain.SigTx) int {
    cmp := strings.Compare(string(a.From), string(b.From))
    if cmp != 0 {
      return cmp
    }
    return int(a.Nonce) - int(b.Nonce)
  })
  vldTxs := make([]chain.SigTx, 0, len(pndTxs))
  for _, tx := range pndTxs {
    err := s.ApplyTx(tx)
    if err != nil {
      fmt.Printf("REJECTED %s\n", err)
      continue
    }
    vldTxs = append(vldTxs, tx)
  }
  if len(vldTxs) == 0 {
    return store.Block{}, fmt.Errorf("none of txs is valid")
  }
  hash, err := s.lastBlock.Hash()
  if err != nil {
    return store.Block{}, err
  }
  blk := store.Block{
    Number: s.lastBlock.Number + 1, Parent: hash, Time: time.Now(), Txs: vldTxs,
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
    err := s.ApplyTx(tx)
    if err != nil {
      return err
    }
  }
  s.lastBlock = blk
  return nil
}
