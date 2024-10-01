package chain_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

func TestApplyTx(t *testing.T) {
  defer os.RemoveAll(keyStoreDir)
  defer os.RemoveAll(blockStoreDir)
  // Create and persist the genesis
  gen, err := createGenesis()
  if err != nil {
    t.Fatal(err)
  }
  // Create the state
  state := chain.NewState(gen)
  pending := state.Pending
  // Lookup the initial owner account address and balance
  ownerAcc, ownerBal := genesisAccount(gen)
  // Re-create the initial owner account
  path := filepath.Join(keyStoreDir, string(ownerAcc))
  acc, err := chain.ReadAccount(path, []byte(ownerPass))
  if err != nil {
    t.Fatal(err)
  }
  cases := []struct{
    name string
    value, nonceInc uint64
    err error
  }{
    {"valid tx 1", 12, 1, nil},
    {"invalid nonce error", 99, 0, fmt.Errorf("invalid nonce")},
    {"valid tx 2", 34, 1, nil},
  }
  for _, c := range cases {
    t.Run(c.name, func(t *testing.T) {
      // Create and sign a transaction
      tx := chain.NewTx(
        acc.Address(), chain.Address("to"), c.value,
        pending.Nonce(acc.Address()) + c.nonceInc,
      )
      stx, err := acc.SignTx(tx)
      if err != nil {
        t.Fatal(err)
      }
      // Apply the transaction to the pending state
      err = pending.ApplyTx(stx)
      if c.err == nil && err != nil {
        t.Error(err)
      }
      if c.err != nil && err == nil {
        t.Errorf("expected invalid nonce error, got none")
      }
    })
  }
  // Lookup the balance of the initial owner
  got, exist := pending.Balance(acc.Address())
  exp := ownerBal - 12 - 34
  if !exist {
    t.Fatalf("balance does not exist")
  }
  if got != exp {
    t.Errorf("invalid balance: expected %v, got %v", exp, got)
  }
  t.Run("insufficient funds error", func(t *testing.T) {
    // Create and sign a transaction
    tx := chain.NewTx(
      acc.Address(), chain.Address("to"), 1000, pending.Nonce(acc.Address()) + 1,
    )
    stx, err := acc.SignTx(tx)
    if err != nil {
      t.Fatal(err)
    }
    // Apply the transaction to the pending state
    err = pending.ApplyTx(stx)
    if err == nil {
      t.Errorf("expected insufficient funds error, got none")
    }
  })
  t.Run("invalid signature error", func(t *testing.T) {
    // Create a new account different from the sender account
    acc2, err := createAccount()
    if err != nil {
      t.Fatal(err)
    }
    // Create and sign a transaction with the new account
    tx := chain.NewTx(
      acc.Address(), chain.Address("to"), 12, pending.Nonce(acc.Address()) + 1,
    )
    stx, err := acc2.SignTx(tx)
    if err != nil {
      t.Fatal(err)
    }
    // Apply the transaction to the pending state
    err = pending.ApplyTx(stx)
    if err == nil {
      t.Errorf("expected invalid signature error, got none")
    }
  })
}

func TestApplyBlock(t *testing.T) {
  defer os.RemoveAll(keyStoreDir)
  defer os.RemoveAll(blockStoreDir)
  // Create and persist the genesis
  gen, err := createGenesis()
  if err != nil {
    t.Fatal(err)
  }
  // Create the state
  state := chain.NewState(gen)
  pending := state.Pending
  // Lookup the initial owner account address and balance
  ownerAcc, ownerBal := genesisAccount(gen)
  // Re-create the initial owner account
  path := filepath.Join(keyStoreDir, string(ownerAcc))
  acc, err := chain.ReadAccount(path, []byte(ownerPass))
  if err != nil {
    t.Fatal(err)
  }
  // Re-create the authority account
  path = filepath.Join(keyStoreDir, string(gen.Authority))
  auth, err := chain.ReadAccount(path, []byte(authPass))
  if err != nil {
    t.Fatal(err)
  }
  for _, value := range []uint64{12, 1000, 34} {
    // Create and sign a transaction
    tx := chain.NewTx(
      acc.Address(), chain.Address("to"), value,
      pending.Nonce(acc.Address()) + 1,
    )
    stx, err := acc.SignTx(tx)
    if err != nil {
      t.Fatal(err)
    }
    // Apply the transaction ot the pending state
    err = pending.ApplyTx(stx)
    if err != nil {
      fmt.Println(err)
    }
  }
  // Create a new block on the cloned state
  clone := state.Clone()
  blk, err := clone.CreateBlock(auth)
  if err != nil {
    t.Fatal(err)
  }
  // Apply the new block to the cloned state
  clone = state.Clone()
  err = clone.ApplyBlock(blk)
  if err != nil {
    t.Fatal(err)
  }
  // Apply the cloned state with the new block updates to the confirmed state
  state.Apply(clone)
  // Lookup the balance of the initial owner
  got, exist := state.Balance(acc.Address())
  exp := ownerBal - 12 - 34
  if !exist {
    t.Fatalf("balance does not exist")
  }
  if got != exp {
    t.Errorf("invalid balance: expected %v, got %v", exp, got)
  }
}
