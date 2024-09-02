package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/volodymyrprokopyuk/go-blockchain/account"
)

func account() error {
  acc, err := wallet.NewAccount()
  if err != nil {
    return err
  }
  fmt.Printf("%v\n", acc.Address())
  dir := ".wallet"
  pwd := []byte("password")
  err = acc.Write(dir, pwd)
  if err != nil {
    return err
  }
  path := filepath.Join(dir, string(acc.Address()))
  acc, err = wallet.ReadAccount(path, pwd)
  if err != nil {
    return err
  }
  msg := []byte("abc")
  sig, err := acc.Sign(msg)
  if err != nil {
    return err
  }
  valid, err := wallet.VerifySig(sig, msg, acc.Address())
  if err != nil {
    return err
  }
  fmt.Println(valid)
  return nil
}

func main() {
  err := account()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}
