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
  // Create a new account
  acc, err := chain.NewAccount()
  if err != nil {
    t.Fatal(err)
  }
  // Persist the account
  err = acc.Write(keyStoreDir, []byte(ownerPass))
  if err != nil {
    t.Fatal(err)
  }
  // Re-create the persisted account
  path := filepath.Join(keyStoreDir, string(acc.Address()))
  acc, err = chain.ReadAccount(path, []byte(ownerPass))
  if err != nil {
    t.Fatal(err)
  }
  // Create and sign a transaction
  tx := chain.NewTx(acc.Address(), chain.Address("to"), 12, 1)
  stx, err := acc.SignTx(tx)
  if err != nil {
    t.Fatal(err)
  }
  // Verify the signature of the signed transaction
  valid, err := chain.VerifyTx(stx)
  if err != nil {
    t.Fatal(err)
  }
  if !valid {
    t.Errorf("invalid transaction signature")
  }
}
