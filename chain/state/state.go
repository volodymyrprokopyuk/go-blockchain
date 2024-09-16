package state

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

type State struct {
  mtx sync.RWMutex
  balances map[chain.Address]uint64
  nonces map[chain.Address]uint64
  lastBlock chain.Block
  genesisHash chain.Hash
  txs map[chain.Hash]chain.SigTx
  Pending *State
}

func (s *State) Nonce(acc chain.Address) uint64 {
  s.mtx.RLock()
  defer s.mtx.RUnlock()
  return s.nonces[acc]
}

func (s *State) LastBlock() chain.Block {
  s.mtx.RLock()
  defer s.mtx.RUnlock()
  return s.lastBlock
}

func NewState(gen chain.SigGenesis) *State {
  return &State{
    balances: maps.Clone(gen.Balances),
    nonces: make(map[chain.Address]uint64),
    genesisHash: gen.Hash(),
    txs: make(map[chain.Hash]chain.SigTx),
    Pending: &State{
      balances: maps.Clone(gen.Balances),
      nonces: make(map[chain.Address]uint64),
      genesisHash: gen.Hash(),
      txs: make(map[chain.Hash]chain.SigTx),
    },
  }
}

func (s *State) Clone() *State {
  s.mtx.RLock()
  defer s.mtx.RUnlock()
  return &State{
    balances: maps.Clone(s.balances),
    nonces: maps.Clone(s.nonces),
    lastBlock: s.lastBlock,
    genesisHash: s.genesisHash,
    txs: maps.Clone(s.txs),
    Pending: &State{
      txs: maps.Clone(s.Pending.txs),
    },
  }
}

func (s *State) Apply(sta *State) {
  s.mtx.Lock()
  defer s.mtx.Unlock()
  s.balances = sta.balances
  s.nonces = sta.nonces
  s.lastBlock = sta.lastBlock
  s.Pending.balances = maps.Clone(s.balances)
  s.Pending.nonces = maps.Clone(s.nonces)
  for _, tx := range sta.lastBlock.Txs {
    delete(s.Pending.txs, tx.Hash())
  }
}

func (s *State) String() string {
  s.mtx.RLock()
  defer s.mtx.RUnlock()
  var bld strings.Builder
  bld.WriteString("Balances\n")
  for acc, bal := range s.balances {
    bld.WriteString(fmt.Sprintf("  %.7s: %29d\n", acc, bal))
  }
  bld.WriteString("Nonces\n")
  for acc, nonce := range s.nonces {
    bld.WriteString(fmt.Sprintf("  %.7s: %35d\n", acc, nonce))
  }
  bld.WriteString("Last block\n")
  bld.WriteString(fmt.Sprintf("  %v", s.lastBlock))
  if s.Pending != nil && len(s.Pending.txs) > 0 {
    bld.WriteString("Pending txs\n")
    for _, tx := range s.Pending.txs {
      bld.WriteString(fmt.Sprintf("  %v\n", tx))
    }
  }
  if s.Pending != nil && len(s.Pending.balances) > 0 {
    bld.WriteString("Pending balances\n")
    for acc, bal := range s.Pending.balances {
      bld.WriteString(fmt.Sprintf("  %.7s: %29d\n", acc, bal))
    }
  }
  if s.Pending != nil && len(s.Pending.nonces) > 0 {
    bld.WriteString("Pending nonces\n")
    for acc, nonce := range s.Pending.nonces {
      bld.WriteString(fmt.Sprintf("  %.7s: %35d\n", acc, nonce))
    }
  }
  return bld.String()
}

func (s *State) ApplyTx(stx chain.SigTx) error {
  s.mtx.Lock()
  defer s.mtx.Unlock()
  hash := stx.Hash()
  valid, err := chain.VerifyTx(stx)
  if err != nil {
    return err
  }
  if !valid {
    return fmt.Errorf("%.7s: invalid signature", hash)
  }
  if stx.Nonce != s.nonces[stx.From] + 1 {
    return fmt.Errorf("%.7s: invalid nonce", hash)
  }
  if s.balances[stx.From] < stx.Value {
    return fmt.Errorf("%.7s: insufficient funds", hash)
  }
  s.balances[stx.From] -= stx.Value
  s.balances[stx.To] += stx.Value
  s.nonces[stx.From]++
  s.txs[hash] = stx
  return nil
}

func (s *State) CreateBlock() chain.Block {
  // no need to lock/unlock as CreateBlock is always executed on a clone
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
  txs := make([]chain.SigTx, 0, len(pndTxs))
  for _, tx := range pndTxs {
    err := s.ApplyTx(tx)
    if err != nil {
      fmt.Printf("REJECTED %v\n", err)
      continue
    }
    txs = append(txs, tx)
  }
  if s.lastBlock.Number == 0 {
    return chain.NewBlock(s.lastBlock.Number + 1, s.genesisHash, txs)
  }
  return chain.NewBlock(s.lastBlock.Number + 1, s.lastBlock.Hash(), txs)
}

func (s *State) ApplyBlock(blk chain.Block) error {
  // no need to lock/unlock as ApplyBlock is always executed on a clone
  if blk.Number != s.lastBlock.Number + 1 {
    return fmt.Errorf("%.7s: invalid block number", blk.Hash())
  }
  hash := s.lastBlock.Hash()
  if blk.Number == 1 {
    hash = s.genesisHash
  }
  if blk.Parent != hash {
    return fmt.Errorf("%.7s: invalid parent hash", blk.Hash())
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
