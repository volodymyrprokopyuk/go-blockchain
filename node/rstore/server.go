package rstore

import (
	"context"
	"fmt"

	"github.com/volodymyrprokopyuk/go-blockchain/account"
	"github.com/volodymyrprokopyuk/go-blockchain/store"
)

type StoreSrv struct {
  UnimplementedStoreServer
  keyStoreDir string
  blockStoreDir string
}

func NewStoreSrv(keyStoreDir, blockStoreDir string) *StoreSrv {
  return &StoreSrv{keyStoreDir: keyStoreDir, blockStoreDir: blockStoreDir}
}

func (s *StoreSrv) StoreInit(
  _ context.Context, req *StoreInitReq,
) (*StoreInitRes, error) {
  pwd := []byte(req.Password)
  if len(pwd) < 5 {
    return nil, fmt.Errorf("password length is less than 5")
  }
  acc, err := account.NewAccount()
  if err != nil {
    return nil, err
  }
  err = acc.Write(s.keyStoreDir, pwd)
  if err != nil {
    return nil, err
  }
  gen := store.NewGenesis(req.Chain, acc.Address(), uint(req.Balance))
  err = gen.Write(s.blockStoreDir)
  if err != nil {
    return nil, err
  }
  res := &StoreInitRes{Address: string(acc.Address())}
  return res, nil
}
