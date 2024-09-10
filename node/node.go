package node

import (
	"fmt"
	"net"

	"github.com/volodymyrprokopyuk/go-blockchain/blockchain/state"
	"github.com/volodymyrprokopyuk/go-blockchain/blockchain/store"
	"github.com/volodymyrprokopyuk/go-blockchain/node/raccount"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rstore"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rtx"
	"google.golang.org/grpc"
)

type Node struct {
  keyStoreDir string
  blockStoreDir string
  nodeAddr string
  state *state.State
}

func NewNode(keyStoreDir string, blockStoreDir string, nodeAddr string) *Node {
  return &Node{
    keyStoreDir: keyStoreDir, blockStoreDir: blockStoreDir, nodeAddr: nodeAddr,
  }
}

func (n *Node) Start() error {
  err := n.readState()
  if err != nil {
    return err
  }
  return n.servegRPC()
}

func (n *Node) readState() error {
  gen, err := store.ReadGenesis(n.blockStoreDir)
  if err != nil {
    fmt.Printf(
      `warning: genesis not found
  > chain store init
  > chain node start
`)
  } else {
    n.state = state.NewState(gen)
    if err != nil {
      return err
    }
    fmt.Printf("* Initial state (ReadGenesis)\n%v\n", n.state)
  }
  return nil
}

func (n *Node) servegRPC() error {
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
  txSrv := rtx.NewTxSrv(n.keyStoreDir, n.state)
  rtx.RegisterTxServer(grpcSrv, txSrv)
  return grpcSrv.Serve(lis)
}
