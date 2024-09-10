package raccount

import (
	"context"
	"fmt"

	"github.com/volodymyrprokopyuk/go-blockchain/blockchain/account"
)

type AccountSrv struct {
  UnimplementedAccountServer
  keyStoreDir string
}

func NewAccountSrv(keyStoreDir string) *AccountSrv {
  return &AccountSrv{keyStoreDir: keyStoreDir}
}

func (s *AccountSrv) AccountCreate(
  _ context.Context, req *AccountCreateReq,
) (*AccountCreateRes, error) {
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
  res := &AccountCreateRes{Address: string(acc.Address())}
  return res, nil
}
