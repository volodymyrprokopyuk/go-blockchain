package account

import (
	"context"

	"github.com/volodymyrprokopyuk/go-blockchain/account"
)

const (
  keystoreDir = ".keystore"
  blockstoreDir = ".blockstore"
)

type AccountService struct { }

func (s *AccountService) Create(
  ctx context.Context, req *AccountCreateReq,
) (*AccountCreateRes, error) {
  acc, err := account.NewAccount()
  if err != nil {
    return nil, err
  }
  err = acc.Write(keystoreDir, []byte(req.Password))
  if err != nil {
    return nil, err
  }
  res := &AccountCreateRes{Address: string(acc.Address())}
  return res, nil
}
