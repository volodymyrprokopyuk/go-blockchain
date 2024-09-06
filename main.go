package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/account"
	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/state"
	"github.com/volodymyrprokopyuk/go-blockchain/store"
)

const (
  keystoreDir = ".keystore"
  blockstoreDir = ".blockstore"
)

var pwd = []byte("password")

// func sendTx(acc account.Account)

func useAccount() error {
  acc, err := account.NewAccount()
  if err != nil {
    return err
  }
  fmt.Printf("%v\n", acc.Address())
  err = acc.Write(keystoreDir, pwd)
  if err != nil {
    return err
  }
  path := filepath.Join(keystoreDir, string(acc.Address()))
  acc, err = account.Read(path, pwd)
  if err != nil {
    return err
  }
  tx := chain.Tx{
    From: acc.Address(), To: chain.Address("to"),
    Value: 12, Nonce: 0, Time: time.Now(),
  }
  stx, err := acc.Sign(tx)
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

func signAndApply(tx chain.Tx, acc account.Account, sta *state.State) error {
  stx, err := acc.Sign(tx)
  if err != nil {
    return err
  }
  return sta.Pending.ApplyTx(stx)
}

func useState() error {
  // init state
  addr := chain.Address("9338ccb4ac74594f1f84ce6b46403350a55fc0340cd1c4814af7b6aea765ab4b")
  // gen := store.NewGenesis("Blockchain", addr, 1000)
  // err := gen.Write(blockstoreDir)
  // if err != nil {
  //   return err
  // }
  gen, err := store.ReadGenesis(blockstoreDir)
  if err != nil {
    return err
  }
  sta := state.NewState(gen)
  fmt.Printf("* Init state\n%v\n", sta)

  // read account
  path := filepath.Join(keystoreDir, string(addr))
  acc, err := account.Read(path, pwd)
  if err != nil {
    return err
  }

  // send txs
  to := chain.Address("beneficiary")
  tx := chain.Tx{
    From: addr, To: to, Value: 12,
    Nonce: sta.Pending.Nonce(addr) + 1, Time: time.Now(),
  }
  err = signAndApply(tx, acc, sta)
  if err != nil {
    return err
  }
  tx = chain.Tx{
    From: addr, To: to, Value: 34,
    Nonce: sta.Pending.Nonce(addr) + 1, Time: time.Now(),
  }
  err = signAndApply(tx, acc, sta)
  if err != nil {
    return err
  }
  fmt.Printf("* State\n%v\n", sta)

  // create block
  cloSta := sta.Clone()
  blk, err := cloSta.CreateBlock()
  if err != nil {
    return err
  }
  fmt.Printf("* Clone state\n%v\n", cloSta)
  fmt.Printf("* Block\n%v\n", blk)

  // apply block
  // cloSta = sta.Clone()
  // err = cloSta.ApplyBlock(blk)
  // if err != nil {
  //   return err
  // }
  // sta.Apply(cloSta)
  // fmt.Printf("* State\n%v\n", sta)

  // write block
  // return blk.Write(blockstoreDir)

  // read block store
  // cloSta := sta.Clone()
  // blocks, closeStore, err := store.ReadBlocks(blockstoreDir)
  // if err != nil {
  //   return err
  // }
  // defer closeStore()
  // for err, blk := range blocks {
  //   if err != nil {
  //     return err
  //   }
  //   err = cloSta.ApplyBlock(blk)
  //   if err != nil {
  //     return err
  //   }
  // }
  // sta.Apply(cloSta)
  // fmt.Printf("* State\n%v\n", sta)
  return nil
}

func main() {
  // err := useAccount()
  err := useState()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}
