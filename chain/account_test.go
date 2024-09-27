package chain

import (
	"os"
	"path/filepath"
	"testing"
)

const keyStoreDir = ".keystoretest"
var ownerPass = []byte("password")

func TestAccountWriteReadSignTxVerifyTx(t *testing.T) {
  acc, err := NewAccount()
  if err != nil {
    t.Fatal(err)
  }
  err = acc.Write(keyStoreDir, ownerPass)
  if err != nil {
    t.Fatal(err)
  }
  defer os.RemoveAll(keyStoreDir)
  path := filepath.Join(keyStoreDir, string(acc.Address()))
  acc, err = ReadAccount(path, ownerPass)
  if err != nil {
    t.Fatal(err)
  }
  tx := NewTx(acc.Address(), Address("to"), 12, 1)
  stx, err := acc.SignTx(tx)
  if err != nil {
    t.Fatal(err)
  }
  valid, err := VerifyTx(stx)
  if err != nil {
    t.Fatal(err)
  }
  if !valid {
    t.Errorf("invalid tx signature")
  }
}
