package node

import (
	"fmt"
	"net"

	"github.com/volodymyrprokopyuk/go-blockchain/node/raccount"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rstore"
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
  accSrv := raccount.NewAccountSrv(n.keyStoreDir)
  raccount.RegisterAccountServer(grpcSrv, accSrv)
  stoSrv := rstore.NewStoreSrv(n.keyStoreDir, n.blockStoreDir)
  rstore.RegisterStoreServer(grpcSrv, stoSrv)
  return grpcSrv.Serve(lis)
}
