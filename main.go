package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/chain/account"
	"github.com/volodymyrprokopyuk/go-blockchain/chain/state"
	"github.com/volodymyrprokopyuk/go-blockchain/chain/store"
)

const (
  keyStoreDir = ".keystore"
  blockStoreDir = ".blockstore"
)

var pwd = []byte("password")

func newAccountSignVerify() error {
  acc, err := account.NewAccount()
  if err != nil {
    return err
  }
  fmt.Printf("account: %v\n", acc.Address())
  err = acc.Write(keyStoreDir, pwd)
  if err != nil {
    return err
  }
  path := filepath.Join(keyStoreDir, string(acc.Address()))
  acc, err = account.Read(path, pwd)
  if err != nil {
    return err
  }
  tx := chain.NewTx(acc.Address(), chain.Address("ben"), 12, 1)
  stx, err := acc.Sign(tx)
  fmt.Printf("%v\n", stx)
  if err != nil {
    return err
  }
  valid, err := account.Verify(stx)
  if err != nil {
    return err
  }
  fmt.Println(valid)
  return nil
}

func writeState() error {
  // create account
  acc, err := account.NewAccount()
  if err != nil {
    return err
  }
  err = acc.Write(keyStoreDir, pwd)
  if err != nil {
    return err
  }
  path := filepath.Join(keyStoreDir, string(acc.Address()))
  acc, err = account.Read(path, pwd)
  if err != nil {
    return err
  }

  // init state
  gen := store.NewGenesis("blockchain", acc.Address(), 1000)
  err = gen.Write(blockStoreDir)
  if err != nil {
    return err
  }
  gen, err = store.ReadGenesis(blockStoreDir)
  if err != nil {
    return err
  }
  sta := state.NewState(gen)
  fmt.Printf("* Initial state (ReadGenesis)\n%v\n", sta)

  for range 2 {
    // send txs
    ben := chain.Address("beneficiary")
    tx := chain.NewTx(acc.Address(), ben, 12, sta.Pending.Nonce(acc.Address()) + 1)
    stx, err := acc.Sign(tx)
    if err != nil {
      return err
    }
    err = sta.Pending.ApplyTx(stx)
    if err != nil {
      return err
    }
    tx = chain.NewTx(acc.Address(), ben, 34, sta.Pending.Nonce(acc.Address()) + 1)
    stx, err = acc.Sign(tx)
    if err != nil {
      return err
    }
    err = sta.Pending.ApplyTx(stx)
    if err != nil {
      return err
    }
    fmt.Printf("* Pending state (ApplyTx)\n%v\n", sta)

    // create block
    clo := sta.Clone()
    blk, err := clo.CreateBlock()
    if err != nil {
      return err
    }
    fmt.Printf("* Block\n%v\n", blk)

    // apply block
    clo = sta.Clone()
    err = clo.ApplyBlock(blk)
    if err != nil {
      return err
    }
    sta.Apply(clo)
    sta.ResetPending()
    fmt.Printf("* Block state (ApplyBlock)\n%v\n", sta)

    // write block
    err = blk.Write(blockStoreDir)
    if err != nil {
      return err
    }
  }
  return nil
}

func readState() error {
  // init state
  gen, err := store.ReadGenesis(blockStoreDir)
  if err != nil {
    return err
  }
  sta := state.NewState(gen)
  fmt.Printf("* Initial state (ReadGenesis)\n%v\n", sta)

  // read blocks
  blocks, closeBlocks, err := store.ReadBlocks(blockStoreDir)
  if err != nil {
    return err
  }
  defer closeBlocks()
  for err, blk := range blocks {
    clo := sta.Clone()
    if err != nil {
      return err
    }
    err = clo.ApplyBlock(blk)
    if err != nil {
      return err
    }
    sta.Apply(clo)
    sta.ResetPending()
  }
  fmt.Printf("* Read state (ReadBlocks)\n%v\n", sta)
  return nil
}

func main() {
  // err := newAccountSignVerify()
  // err := writeState()
  err := readState()

  // cmd := command.ChainCmd()
  // err := cmd.Execute()

  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}
