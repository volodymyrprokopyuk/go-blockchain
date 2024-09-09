package node

import (
	"fmt"
	"net"

	"github.com/volodymyrprokopyuk/go-blockchain/node/account"
	"google.golang.org/grpc"
)

type Node struct {
  keyStoreDir string
  blockStoreDir string
  nodeAddr string
}

func NewNode(keyStoreDir string, blockStoreDir string, nodeAddr string) *Node {
  return &Node{
    keyStoreDir: keyStoreDir, blockStoreDir: blockStoreDir, nodeAddr: nodeAddr,
  }
}

func (n *Node) Start() error {
  return n.serve()
}

func (n *Node) serve() error {
  lis, err := net.Listen("tcp", n.nodeAddr)
  if err != nil {
    return err
  }
  defer lis.Close()
  fmt.Printf("* gRPC listening on %v\n", n.nodeAddr)
  grpcSrv := grpc.NewServer()
  accSrv := account.NewAccountSrv(n.keyStoreDir)
  account.RegisterAccountServer(grpcSrv, accSrv)
  return grpcSrv.Serve(lis)
}
