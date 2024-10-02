package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/node"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func nodeCmd(ctx context.Context) *cobra.Command {
  cmd := &cobra.Command{
    Use: "node",
    Short: "Manages the blockchain node",
  }
  cmd.AddCommand(nodeStartCmd(ctx), nodeSubscribeCmd(ctx))
  return cmd
}

func nodeStartCmd(_ context.Context) *cobra.Command {
  cmd := &cobra.Command{
    Use: "start",
    Short: "Starts the blockchain node",
    RunE: func(cmd *cobra.Command, _ []string) error {
      nodeAddr, _ := cmd.Flags().GetString("node")
      reAddr := regexp.MustCompile(`[-\.\w]+:\d+`)
      if !reAddr.MatchString(nodeAddr) {
        return fmt.Errorf("expected --node host:port, got %v", nodeAddr)
      }
      bootstrap, _ := cmd.Flags().GetBool("bootstrap")
      seedAddr, _ := cmd.Flags().GetString("seed")
      if !bootstrap && len(seedAddr) == 0 {
        return fmt.Errorf(
          "either --bootstrap or --seed host:port must be provided",
        )
      }
      if !bootstrap && !reAddr.MatchString(seedAddr) {
        return fmt.Errorf("expected --seed host:port, got %v", seedAddr)
      }
      rePort := regexp.MustCompile(`\d+$`)
      port := rePort.FindString(nodeAddr)
      keyStoreDir, _ := cmd.Flags().GetString("keystore")
      if len(keyStoreDir) == 0 {
        keyStoreDir = ".keystore" + port
      }
      blockStoreDir, _ := cmd.Flags().GetString("blockstore")
      if len(blockStoreDir) == 0 {
        blockStoreDir = ".blockstore" + port
      }
      name, _ := cmd.Flags().GetString("chain")
      authPass, _ := cmd.Flags().GetString("authpass")
      ownerPass, _ := cmd.Flags().GetString("ownerpass")
      balance, _ := cmd.Flags().GetUint64("balance")
      cfg := node.NodeCfg{
        NodeAddr: nodeAddr, Bootstrap: bootstrap, SeedAddr: seedAddr,
        KeyStoreDir: keyStoreDir, BlockStoreDir: blockStoreDir,
        Chain: name, AuthPass: authPass, OwnerPass: ownerPass, Balance: balance,
      }
      nd := node.NewNode(cfg)
      return nd.Start()
    },
  }
  cmd.Flags().Bool("bootstrap", false, "bootstrap node and authority node")
  cmd.Flags().String("seed", "", "seed address host:port")
  cmd.MarkFlagsMutuallyExclusive("bootstrap", "seed")
  cmd.MarkFlagsOneRequired("bootstrap", "seed")
  cmd.Flags().String("keystore", "", "key store directory")
  cmd.Flags().String("blockstore", "", "block store directory")
  cmd.Flags().String("chain", "blockchain", "name of the blockchain")
  cmd.Flags().String("authpass", "", "password for the authority account")
  cmd.Flags().String("ownerpass", "", "password for the genesis account")
  cmd.Flags().Uint64("balance", 0, "initial balance for the genesis account")
  cmd.MarkFlagsRequiredTogether("bootstrap", "authpass")
  cmd.MarkFlagsRequiredTogether("ownerpass", "balance")
  return cmd
}

func grpcStreamSubscribe(
  ctx context.Context, addr string, evTypesStr []string,
) (func(yield func(err error, event chain.Event) bool), func(), error) {
  conn, err := grpc.NewClient(
    addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return nil, nil, err
  }
  close := func() {
    conn.Close()
  }
  cln := rpc.NewNodeClient(conn)
  evTypes := make([]uint64, len(evTypesStr))
  for i, evTypeStr := range evTypesStr {
    evTypes[i] = uint64(chain.NewEventType(evTypeStr))
  }
  req := &rpc.StreamSubscribeReq{EventTypes: evTypes}
  stream, err := cln.StreamSubscribe(ctx, req)
  if err != nil {
    return nil, nil, err
  }
  more := true
  events := func(yield func(err error, event chain.Event) bool) {
    for more {
      res, err := stream.Recv()
      if err == io.EOF {
        return
      }
      if err != nil {
        yield(err, chain.Event{})
        return
      }
      var event chain.Event
      err = json.Unmarshal(res.Event, &event)
      if err != nil {
        yield(err, chain.Event{})
        return
      }
      more = yield(nil, event)
    }
  }
  return events, close, nil
}

func nodeSubscribeCmd(ctx context.Context) *cobra.Command {
  cmd := &cobra.Command{
    Use: "subscribe",
    Short: "Subscribes to selected set of event types",
    RunE: func(cmd *cobra.Command, _ []string) error {
      addr, _ := cmd.Flags().GetString("node")
      evTypesStr, _ := cmd.Flags().GetStringSlice("events")
      events, closeEvents, err := grpcStreamSubscribe(ctx, addr, evTypesStr)
      if err != nil {
        return err
      }
      defer closeEvents()
      for err, event := range events {
        if err != nil {
          return err
        }
        fmt.Printf("%v\n", event)
      }
      return nil
    },
  }
  cmd.Flags().StringSlice("events", []string{"all"}, "event types of interest")
  return cmd
}
