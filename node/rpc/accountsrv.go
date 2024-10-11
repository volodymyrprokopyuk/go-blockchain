package rpc

import (
	"context"
	"fmt"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BalanceChecker interface {
  Balance(acc chain.Address) (uint64, bool)
}

type AccountSrv struct {
  UnimplementedAccountServer
  keyStoreDir string
  balChecker BalanceChecker
}

func NewAccountSrv(
  keyStoreDir string, balChecker BalanceChecker,
) *AccountSrv {
  return &AccountSrv{keyStoreDir: keyStoreDir, balChecker: balChecker}
}

func (s *AccountSrv) AccountCreate(
  _ context.Context, req *AccountCreateReq,
) (*AccountCreateRes, error) {
  pass := []byte(req.Password)
  if len(pass) < 5 {
    return nil, status.Errorf(
      codes.InvalidArgument, "password length is less than 5",
    )
  }
  acc, err := chain.NewAccount()
  if err != nil {
    return nil, status.Errorf(codes.Internal, err.Error())
  }
  err = acc.Write(s.keyStoreDir, pass)
  if err != nil {
    return nil, status.Errorf(codes.Internal, err.Error())
  }
  res := &AccountCreateRes{Address: string(acc.Address())}
  return res, nil
}

func (s *AccountSrv) AccountBalance(
  _ context.Context, req *AccountBalanceReq,
) (*AccountBalanceRes, error) {
  acc := req.Address
  balance, exist := s.balChecker.Balance(chain.Address(acc))
  if !exist {
    return nil, status.Errorf(
      codes.NotFound, fmt.Sprintf(
        "account %v does not exist or has not yet transacted", acc,
      ),
    )
  }
  res := &AccountBalanceRes{Balance: balance}
  return res, nil
}
