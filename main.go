package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/volodymyrprokopyuk/go-blockchain/account"
	"github.com/volodymyrprokopyuk/go-blockchain/state"
)

const (
  keystoreDir = ".keystore"
  blockstoreDir = ".blockstore"
)

func useAccount() error {
  acc, err := account.NewAccount()
  if err != nil {
    return err
  }
  fmt.Printf("%v\n", acc.Address())
  pwd := []byte("password")
  err = acc.Write(keystoreDir, pwd)
  if err != nil {
    return err
  }
  path := filepath.Join(keystoreDir, string(acc.Address()))
  acc, err = account.ReadAccount(path, pwd)
  if err != nil {
    return err
  }
  msg := []byte("abc")
  sig, err := acc.Sign(msg)
  if err != nil {
    return err
  }
  valid, err := account.VerifySig(sig, msg, acc.Address())
  if err != nil {
    return err
  }
  fmt.Println(valid)
  return nil
}

func useState() error {
  addr := account.Address("daf5c55e75fc98b19e9cc790c99d0d631ba8fcc026dc36a2bca1944bc5abd236")
  gen := state.NewGenesis("Blockchain", addr, 1000)
  return gen.Write(blockstoreDir)
}

func main() {
  // err := useAccount()
  err := useState()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}
