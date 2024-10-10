package chain_test

import (
	"testing"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

func TestTxSignTxVerifyTx(t *testing.T) {
  // Create a new account
  acc, err := chain.NewAccount()
  if err != nil {
    t.Fatal(err)
  }
  // Create and sign a transaction
  tx := chain.NewTx(acc.Address(), chain.Address("to"), 12, 1)
  stx, err := acc.SignTx(tx)
  if err != nil {
    t.Fatal(err)
  }
  // Verify that the signature of the signed transaction is valid
  valid, err := chain.VerifyTx(stx)
  if err != nil {
    t.Fatal(err)
  }
  if !valid {
    t.Errorf("invalid transaction signature")
  }
}
