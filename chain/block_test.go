package chain_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

func TestBlockSignBlockWriteReadVerifyBlock(t *testing.T) {
  defer os.RemoveAll(keyStoreDir)
  defer os.RemoveAll(blockStoreDir)
  // Create and persist the genesis
  gen, err := createGenesis()
  if err != nil {
    t.Fatal(err)
  }
  // Re-create the authority account
  path := filepath.Join(keyStoreDir, string(gen.Authority))
  auth, err := chain.ReadAccount(path, []byte(authPass))
  if err != nil {
    t.Fatal(err)
  }
  // Re-create the initial owner account
  ownerAcc, _ := genesisAccount(gen)
  path = filepath.Join(keyStoreDir, string(ownerAcc))
  acc, err := chain.ReadAccount(path, []byte(ownerPass))
  if err != nil {
    t.Fatal(err)
  }
  // Create and sign a transaction
  tx := chain.NewTx(chain.Address("from"), chain.Address("to"), 12, 1)
  stx, err := acc.SignTx(tx)
  if err != nil {
    t.Fatal(err)
  }
  // Create and sign a block
  txs := make([]chain.SigTx, 0, 1)
  txs = append(txs, stx)
  blk := chain.NewBlock(1, gen.Hash(), txs)
  sblk, err := auth.SignBlock(blk)
  if err != nil {
    t.Fatal(err)
  }
  // Persist the signed block
  err = sblk.Write(blockStoreDir)
  if err != nil {
    t.Fatal(err)
  }
  // Re-create the signed block
  blocks, closeBlocs, err := chain.ReadBlocks(blockStoreDir)
  if err != nil {
    t.Fatal(err)
  }
  defer closeBlocs()
  for err, blk := range blocks {
    if err != nil {
      t.Fatal(err)
    }
    // Verify the signature of the signed block
    valid, err := chain.VerifyBlock(blk, auth.Address())
    if err != nil {
      t.Fatal(err)
    }
    if !valid {
      t.Errorf("invalid block signature")
    }
    break
  }
}
