package chain_test

import (
	"os"
	"testing"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

func createAccount() (chain.Account, error) {
  acc, err := chain.NewAccount()
  if err != nil {
    return chain.Account{}, err
  }
  err = acc.Write(keyStoreDir, []byte(ownerPass))
  if err != nil {
    return chain.Account{}, err
  }
  return acc, nil
}

func createGenesis() (chain.SigGenesis, error) {
  auth, err := createAccount()
  if err != nil {
    return chain.SigGenesis{}, err
  }
  acc, err := createAccount()
  if err != nil {
    return chain.SigGenesis{}, err
  }
  gen := chain.NewGenesis(chainName, auth.Address(), acc.Address(), ownerBalance)
  sgen, err := auth.SignGen(gen)
  if err != nil {
    return chain.SigGenesis{}, err
  }
  err = sgen.Write(blockStoreDir)
  if err != nil {
    return chain.SigGenesis{}, err
  }
  return sgen, nil
}

func genesisAccount(gen chain.SigGenesis) (chain.Address, uint64) {
  for acc, bal := range gen.Balances {
    return acc, bal
  }
  return "", 0
}

func TestGenesisWriteReadSignGenVerifyGen(t *testing.T) {
  defer os.RemoveAll(keyStoreDir)
  defer os.RemoveAll(blockStoreDir)
  // Create and persist the authority account to sign the genesis and proposed
  // blocks
  auth, err := createAccount()
  if err != nil {
    t.Fatal(err)
  }
  // Create and persist the initial owner account to hold the initial balance of
  // the blockchain
  acc, err := createAccount()
  if err != nil {
    t.Fatal(err)
  }
  // Create and persist the genesis
  gen := chain.NewGenesis(chainName, auth.Address(), acc.Address(), ownerBalance)
  sgen, err := auth.SignGen(gen)
  if err != nil {
    t.Fatal(err)
  }
  err = sgen.Write(blockStoreDir)
  if err != nil {
    t.Fatal(err)
  }
  // Re-create the persisted genesis
  sgen, err = chain.ReadGenesis(blockStoreDir)
  if err != nil {
    t.Fatal(err)
  }
  // Verify that the signature of the persisted genesis is valid
  valid, err := chain.VerifyGen(sgen)
  if err != nil {
    t.Fatal(err)
  }
  if !valid {
    t.Errorf("invalid genesis signature")
  }
}
