package main

import (
	"fmt"
	"os"

	"github.com/volodymyrprokopyuk/go-blockchain/wallet"
)

func main() {
  err := wallet.SignVerify()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}
