package chain_test

import (
	"os"
	"testing"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

func TestGenesisWriteReadSignGenVerifyGen(t *testing.T) {
  defer os.RemoveAll(keyStoreDir)
  defer os.RemoveAll(blockStoreDir)
  auth, err := chain.NewAccount()
  if err != nil {
    t.Fatal(err)
  }
  err = auth.Write(blockStoreDir, []byte(authPass))
  if err != nil {
    t.Fatal(err)
  }
  acc, err := chain.NewAccount()
  if err != nil {
    t.Fatal(err)
  }
  err = acc.Write(blockStoreDir, []byte(ownerPass))
  if err != nil {
    t.Fatal(err)
  }
  gen := chain.NewGenesis(chainName, auth.Address(), acc.Address(), ownerBalance)
  sgen, err := auth.SignGen(gen)
  if err != nil {
    t.Fatal(err)
  }
  err = sgen.Write(blockStoreDir)
  if err != nil {
    t.Fatal(err)
  }
  sgen, err = chain.ReadGenesis(blockStoreDir)
  if err != nil {
    t.Fatal(err)
  }
  valid, err := chain.VerifyGen(sgen)
  if err != nil {
    t.Fatal(err)
  }
  if !valid {
    t.Errorf("invalid genesis signature")
  }
}
