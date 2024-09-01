package main

import (
	"fmt"
	"os"

	"github.com/volodymyrprokopyuk/go-blockchain/wallet"
)

func account() error {
  acc, err := wallet.NewAccount()
  if err != nil {
    return err
  }
  fmt.Println(acc.Address())
  jsnPrv, err := acc.Encode()
  if err != nil {
    return err
  }
  fmt.Printf("%s\n", jsnPrv)
  acc, err = wallet.DecodeAccount(jsnPrv)
  if err != nil {
    return err
  }
  fmt.Println(acc.Address())
  return nil
}

func main() {
  // err := wallet.SignVerify()
  err := account()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}
