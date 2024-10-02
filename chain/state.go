package chain

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
)

type State struct {
  mtx sync.RWMutex
  authority Address
  balances map[Address]uint64
  nonces map[Address]uint64
  lastBlock SigBlock
  genesisHash Hash
  txs map[Hash]SigTx
  Pending *State
}

func NewState(gen SigGenesis) *State {
  return &State{
    authority: gen.Authority,
    balances: maps.Clone(gen.Balances),
    nonces: make(map[Address]uint64),
    genesisHash: gen.Hash(),
    txs: make(map[Hash]SigTx),
    Pending: &State{
      authority: gen.Authority,
      balances: maps.Clone(gen.Balances),
      nonces: make(map[Address]uint64),
      genesisHash: gen.Hash(),
      txs: make(map[Hash]SigTx),
    },
  }
}

func (s *State) Clone() *State {
  s.mtx.RLock()
  defer s.mtx.RUnlock()
  return &State{
    authority: s.authority,
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

func (s *State) Apply(clone *State) {
  s.mtx.Lock()
  defer s.mtx.Unlock()
  s.balances = clone.balances
  s.nonces = clone.nonces
  s.lastBlock = clone.lastBlock
  s.Pending.balances = maps.Clone(s.balances)
  s.Pending.nonces = maps.Clone(s.nonces)
  for _, tx := range clone.lastBlock.Txs {
    delete(s.Pending.txs, tx.Hash())
  }
}

func (s *State) Authority() Address {
  return s.authority
}

func (s *State) Balance(acc Address) (uint64, bool) {
  s.mtx.RLock()
  defer s.mtx.RUnlock()
  balance, exist := s.balances[acc]
  return balance, exist
}

func (s *State) Nonce(acc Address) uint64 {
  s.mtx.RLock()
  defer s.mtx.RUnlock()
  return s.nonces[acc]
}

func (s *State) LastBlock() SigBlock {
  s.mtx.RLock()
  defer s.mtx.RUnlock()
  return s.lastBlock
}

func (s *State) String() string {
  s.mtx.RLock()
  defer s.mtx.RUnlock()
  var bld strings.Builder
  bld.WriteString("* Balances and nonces\n")
  format := "acc %.7s:                    %8d %8d\n"
  for acc, bal := range s.balances {
    nonce := s.nonces[acc]
    bld.WriteString(fmt.Sprintf(format, acc, bal, nonce))
  }
  bld.WriteString("* Last block\n")
  bld.WriteString(fmt.Sprintf("%v", s.lastBlock))
  if s.Pending != nil && len(s.Pending.txs) > 0 {
    bld.WriteString("* Pending txs\n")
    for _, tx := range s.Pending.txs {
      bld.WriteString(fmt.Sprintf("%v\n", tx))
    }
  }
  if s.Pending != nil && len(s.Pending.balances) > 0 {
    bld.WriteString("* Pending balances and nonces\n")
    for acc, bal := range s.Pending.balances {
      nonce := s.Pending.nonces[acc]
      bld.WriteString(fmt.Sprintf(format, acc, bal, nonce))
    }
  }
  return bld.String()
}

func (s *State) ApplyTx(tx SigTx) error {
  s.mtx.Lock()
  defer s.mtx.Unlock()
  valid, err := VerifyTx(tx)
  if err != nil {
    return err
  }
  if !valid {
    return fmt.Errorf("tx: invalid transaction signature\n%v\n", tx)
  }
  if tx.Nonce != s.nonces[tx.From] + 1 {
    return fmt.Errorf("tx: invalid transaction nonce\n%v\n", tx)
  }
  if s.balances[tx.From] < tx.Value {
    return fmt.Errorf("tx: insufficient account funds\n%v\n", tx)
  }
  s.balances[tx.From] -= tx.Value
  s.balances[tx.To] += tx.Value
  s.nonces[tx.From]++
  s.txs[tx.Hash()] = tx
  return nil
}

func (s *State) CreateBlock(authority Account) (SigBlock, error) {
  // The is no need to lock/unlock as the CreateBlock is always executed on the
  // cloned state
  pndTxs := make([]SigTx, 0, len(s.Pending.txs))
  for _, tx := range s.Pending.txs {
    pndTxs = append(pndTxs, tx)
  }
  slices.SortFunc(pndTxs, func(a, b SigTx) int {
    cmp := strings.Compare(string(a.From), string(b.From))
    if cmp != 0 {
      return cmp
    }
    return int(a.Nonce) - int(b.Nonce)
  })
  txs := make([]SigTx, 0, len(pndTxs))
  for _, tx := range pndTxs {
    err := s.ApplyTx(tx)
    if err != nil {
      fmt.Printf("tx: rejected: %v\n", err)
      continue
    }
    txs = append(txs, tx)
  }
  var blk Block
  if s.lastBlock.Number == 0 {
    blk = NewBlock(s.lastBlock.Number + 1, s.genesisHash, txs)
  } else {
    blk = NewBlock(s.lastBlock.Number + 1, s.lastBlock.Hash(), txs)
  }
  return authority.SignBlock(blk)
}

func (s *State) ApplyBlock(blk SigBlock) error {
  // The is no need to lock/unlock as the CreateBlock is always executed on the
  // cloned state
  valid, err := VerifyBlock(blk, s.authority)
  if err != nil {
    return err
  }
  if !valid {
    return fmt.Errorf("blk: invalid block signature\n%v", blk)
  }
  if blk.Number != s.lastBlock.Number + 1 {
    return fmt.Errorf("blk: invalid block number\n%v", blk)
  }
  var hash Hash
  if blk.Number == 1 {
    hash = s.genesisHash
  } else {
    hash = s.lastBlock.Hash()
  }
  if blk.Parent != hash {
    return fmt.Errorf("blk: invalid parent hash\n%v", blk)
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

func (s *State) ApplyBlockToState(blk SigBlock) error {
  clone := s.Clone()
  err := clone.ApplyBlock(blk)
  if err != nil {
    return err
  }
  s.Apply(clone)
  fmt.Printf("== Block state\n%v", s)
  return nil
}
