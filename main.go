package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/volodymyrprokopyuk/go-blockchain/cli"
)

func main() {
  ctx, cancel := signal.NotifyContext(
    context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL,
  )
  defer cancel()
  cmd := cli.ChainCmd(ctx)
  err := cmd.Execute()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}
