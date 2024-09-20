package rpc

import (
	"context"
	"fmt"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

type AccountBalancer interface {
  Balance(acc chain.Address) uint64
}

type AccountSrv struct {
  UnimplementedAccountServer
  keyStoreDir string
  accBalancer AccountBalancer
}

func NewAccountSrv(
  keyStoreDir string, accBalancer AccountBalancer,
) *AccountSrv {
  return &AccountSrv{keyStoreDir: keyStoreDir, accBalancer: accBalancer}
}

func (s *AccountSrv) AccountCreate(
  _ context.Context, req *AccountCreateReq,
) (*AccountCreateRes, error) {
  pass := []byte(req.Password)
  if len(pass) < 5 {
    return nil, fmt.Errorf("password length is less than 5")
  }
  acc, err := chain.NewAccount()
  if err != nil {
    return nil, err
  }
  err = acc.Write(s.keyStoreDir, pass)
  if err != nil {
    return nil, err
  }
  res := &AccountCreateRes{Address: string(acc.Address())}
  return res, nil
}

func (s *AccountSrv) AccountBalance(
  _ context.Context, req *AccountBalanceReq,
) (*AccountBalanceRes, error) {
  acc := req.Address
  balance := s.accBalancer.Balance(chain.Address(acc))
  res := &AccountBalanceRes{Balance: balance}
  return res, nil
}
