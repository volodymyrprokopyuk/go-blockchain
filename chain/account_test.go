package chain_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

const (
  keyStoreDir = ".keystoretest"
  blockStoreDir = ".keystoretest"
  chainName = "testblockchain"
  authPass = "password"
  ownerPass = "password"
  ownerBalance = 1000
)

func TestAccountWriteReadSignTxVerifyTx(t *testing.T) {
  defer os.RemoveAll(keyStoreDir)
  acc, err := chain.NewAccount()
  if err != nil {
    t.Fatal(err)
  }
  err = acc.Write(keyStoreDir, []byte(ownerPass))
  if err != nil {
    t.Fatal(err)
  }
  path := filepath.Join(keyStoreDir, string(acc.Address()))
  acc, err = chain.ReadAccount(path, []byte(ownerPass))
  if err != nil {
    t.Fatal(err)
  }
  tx := chain.NewTx(acc.Address(), chain.Address("to"), 12, 1)
  stx, err := acc.SignTx(tx)
  if err != nil {
    t.Fatal(err)
  }
  valid, err := chain.VerifyTx(stx)
  if err != nil {
    t.Fatal(err)
  }
  if !valid {
    t.Errorf("invalid transaction signature")
  }
}
