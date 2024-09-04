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

func useState() error {
  addr := chain.Address("9338ccb4ac74594f1f84ce6b46403350a55fc0340cd1c4814af7b6aea765ab4b")
  gen := store.NewGenesis("Blockchain", addr, 1000)
  err := gen.Write(blockstoreDir)
  if err != nil {
    return err
  }
  gen, err = store.ReadGenesis(blockstoreDir)
  if err != nil {
    return err
  }
  sta := state.NewState(gen)
  path := filepath.Join(keystoreDir, string(addr))
  acc, err := account.Read(path, pwd)
  if err != nil {
    return err
  }
  err = sta.Send(acc, chain.Address("recipient"), 123)
  if err != nil {
    return err
  }
  err = sta.Send(acc, chain.Address("recipient"), 456)
  if err != nil {
    return err
  }
  fmt.Printf("* State\n%v\n", sta)
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
