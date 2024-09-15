package main

import (
	"fmt"
	"os"

	"github.com/volodymyrprokopyuk/go-blockchain/command"
)

func main() {
  cmd := command.ChainCmd()
  err := cmd.Execute()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}
