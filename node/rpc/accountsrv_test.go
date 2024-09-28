package rpc_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const (
  keyStoreDir = ".keystoretest"
  blockStoreDir = ".blockstoretest"
  chainName = "testblockchain"
  authPass = "password"
  ownerPass = "password"
  ownerBalance = 1000
)

func createGenesis() (chain.SigGenesis, error) {
  auth, err := chain.NewAccount()
  if err != nil {
    return chain.SigGenesis{}, err
  }
  err = auth.Write(blockStoreDir, []byte(authPass))
  if err != nil {
    return chain.SigGenesis{}, err
  }
  acc, err := chain.NewAccount()
  if err != nil {
    return chain.SigGenesis{}, err
  }
  err = acc.Write(blockStoreDir, []byte(ownerPass))
  if err != nil {
    return chain.SigGenesis{}, err
  }
  gen := chain.NewGenesis(chainName, auth.Address(), acc.Address(), ownerBalance)
  sgen, err := auth.SignGen(gen)
  if err != nil {
    return chain.SigGenesis{}, err
  }
  err = sgen.Write(blockStoreDir)
  if err != nil {
    return chain.SigGenesis{}, err
  }
  return sgen, nil
}

func grpcClientConn(
  t *testing.T, grpcRegisterSrv func(grpcSrv *grpc.Server),
) *grpc.ClientConn {
  // server
  lis := bufconn.Listen(1024 * 1024)
  grpcSrv := grpc.NewServer()
  grpcRegisterSrv(grpcSrv)
  go func() {
    err := grpcSrv.Serve(lis)
    if err != nil {
      fmt.Println(err)
    }
  }()
  // client
  resolver.SetDefaultScheme("passthrough")
  conn, err := grpc.NewClient(
    "bufnet",
    grpc.WithContextDialer(
      func(ctx context.Context, _ string) (net.Conn, error) {
        return lis.DialContext(ctx)
      },
    ),
    grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    t.Fatal(err)
  }
  t.Cleanup(func() {
    lis.Close()
    grpcSrv.GracefulStop()
    conn.Close()
  })
  return conn
}

func TestAccountCreate(t *testing.T) {
  conn := grpcClientConn(t, func(grpcSrv *grpc.Server) {
    acc := rpc.NewAccountSrv(keyStoreDir, nil)
    rpc.RegisterAccountServer(grpcSrv, acc)
  })
  defer os.RemoveAll(keyStoreDir)
  cln := rpc.NewAccountClient(conn)
  req := &rpc.AccountCreateReq{Password: ownerPass}
  res, err := cln.AccountCreate(context.Background(), req)
  if err != nil {
    t.Fatal(err)
  }
  exp, got := 64, res.Address
  if len(got) != exp {
    t.Errorf("invalid account address: expected length %v, got %v", exp, got)
  }
}

func TestAccountBalance(t *testing.T) {
  gen, err := createGenesis()
  if err != nil {
    t.Fatal(err)
  }
  state := chain.NewState(gen)
  conn := grpcClientConn(t, func(grpcSrv *grpc.Server) {
    acc := rpc.NewAccountSrv(keyStoreDir, state)
    rpc.RegisterAccountServer(grpcSrv, acc)
  })
  defer os.RemoveAll(keyStoreDir)
  defer os.RemoveAll(blockStoreDir)
  var ownAcc chain.Address
  var ownBalance uint64
  for acc, bal := range gen.Balances {
    ownAcc, ownBalance = acc, bal
    break
  }
  ctx := context.Background()
  cln := rpc.NewAccountClient(conn)
  t.Run("balance exists", func(t *testing.T) {
    req := &rpc.AccountBalanceReq{Address: string(ownAcc)}
    res, err := cln.AccountBalance(ctx, req)
    if err != nil {
      t.Fatal(err)
    }
    got, exp := res.Balance, ownBalance
    if got != exp {
      t.Errorf("invalid balance: expected %v, got %v", exp, got)
    }
  })
  t.Run("balance does not exist", func(t *testing.T) {
    req := &rpc.AccountBalanceReq{Address: "non-existing"}
    _, err := cln.AccountBalance(ctx, req)
    if err == nil {
      t.Fatalf("non-existing account exists: expected error, got none")
    }
    got, exp := status.Code(err), codes.NotFound
    if got != exp {
      t.Errorf("wrong error: expected %v, got %v", got, exp)
    }
  })
}
