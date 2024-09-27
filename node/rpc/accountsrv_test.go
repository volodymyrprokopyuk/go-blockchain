package rpc_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/test/bufconn"
)

const keyStoreDir = ".keystoretest"
var ownerPass = "password"

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
