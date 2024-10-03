package node

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

const (
  nodeAddr = "localhost:1122"
  seedAddr = "localhost:1123"
  bootKeyStoreDir = ".keystoreboot"
  bootBlockStoreDir = ".blockstoreboot"
  keyStoreDir = ".keystoretest"
  blockStoreDir = ".blockstoretest"
  chainName = "testblockchaint"
  authPass = "password"
  ownerPass = "password"
  ownerBalance = 1000
)

func genesisAccount(gen chain.SigGenesis) (chain.Address, uint64) {
  for acc, bal := range gen.Balances {
    return acc, bal
  }
  return "", 0
}

func TestStateSync(t *testing.T) {
  defer os.RemoveAll(bootKeyStoreDir)
  defer os.RemoveAll(bootBlockStoreDir)
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()
  wg := new(sync.WaitGroup)
  peerDiscCfg := peerDiscoveryCfg{nodeAddr: nodeAddr, bootstrap: true}
  peerDisc := newPeerDiscovery(ctx, wg, peerDiscCfg)
  nodeCfg := NodeCfg{
    NodeAddr: nodeAddr, Bootstrap: true,
    KeyStoreDir: bootBlockStoreDir, BlockStoreDir: bootBlockStoreDir,
    Chain: chainName, AuthPass: authPass,
    OwnerPass: ownerPass, Balance: ownerBalance,
  }
  bootStateSync := newStateSync(ctx, nodeCfg, peerDisc)
  bootState, err := bootStateSync.syncState()
  if err != nil {
    t.Fatal(err)
  }
  gen, err := chain.ReadGenesis(bootBlockStoreDir)
  if err != nil {
    t.Fatal(err)
  }
  ownerAcc, ownerBal := genesisAccount(gen)
  got, exist := bootState.Balance(ownerAcc)
  if !exist {
    t.Fatalf("balance does not exist")
  }
  exp := ownerBal
  if got != exp {
    t.Errorf("Invalid balance: expected %v, got %v", exp, got)
  }
  wg.Wait()
}
