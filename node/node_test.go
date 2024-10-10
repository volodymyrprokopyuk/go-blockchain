package node_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/node"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestNodeStart(t *testing.T) {
  defer os.RemoveAll(bootKeyStoreDir)
  defer os.RemoveAll(bootBlockStoreDir)
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()
  wg := new(sync.WaitGroup)
  // Configure the bootstrap node
  nodeCfg := node.NodeCfg{
    NodeAddr: bootAddr, Bootstrap: true,
    KeyStoreDir: bootKeyStoreDir, BlockStoreDir: bootBlockStoreDir,
    Chain: chainName, AuthPass: authPass,
    OwnerPass: ownerPass, Balance: ownerBalance,
    Period: 100 * time.Millisecond,
  }
  nd := node.NewNode(nodeCfg)
  // Start the bootstrap node in a separate goroutine
  wg.Add(1)
  go func() {
    defer wg.Done()
    err := nd.Start()
    if err != nil {
      fmt.Println(err)
    }
  }()
  // Wait for the bootstrap node to start
  time.Sleep(100 * time.Millisecond)
  // Set up a gRPC client connection with the bootstrap node
  conn, err := grpc.NewClient(
    bootAddr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    t.Fatal(err)
  }
  defer conn.Close()
  // Send several transactions to the bootstrap node in a separate goroutine
  go func() {
    // Get the initial owner account and its balance from the genesis
    gen, err := chain.ReadGenesis(bootBlockStoreDir)
    if err != nil {
      fmt.Println(err)
      return
    }
    ownerAcc, _ := genesisAccount(gen)
    // Re-create the initial owner account from the genesis
    path := filepath.Join(bootKeyStoreDir, string(ownerAcc))
    acc, err := chain.ReadAccount(path, []byte(ownerPass))
    if err != nil {
      fmt.Println(err)
      return
    }
    // Create the gRPC transaction client
    txCln := rpc.NewTxClient(conn)
    // Start sending transaction to the bootstrap node
    for i, value := range []uint64{12, 34} {
      // Create and sign a new transaction
      tx := chain.NewTx(
        acc.Address(), chain.Address("to"), value, uint64(i + 1),
      )
      stx, err := acc.SignTx(tx)
      if err != nil {
        fmt.Println(err)
        return
      }
      // Encode the signed transaction
      jtx, err := json.Marshal(stx)
      if err != nil {
        fmt.Println(err)
        return
      }
      // Call the gRPC TxSend method to the the signed encoded transaction
      req := &rpc.TxSendReq{Tx: jtx}
      _, err = txCln.TxSend(ctx, req)
      if err != nil {
        fmt.Println(err)
        return
      }
    }
  }()
  // Subscribe to the event stream of the bootstrap node and verify that
  // received events are correct
  subscribeAndVerifyEvents(t, ctx, conn)
  // Stop gracefully the node
  nd.GracefulStop()
  wg.Wait()
}
