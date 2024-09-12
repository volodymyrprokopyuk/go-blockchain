package node

import (
	"fmt"
	"net"
	"os"

	"github.com/volodymyrprokopyuk/go-blockchain/chain/state"
	"github.com/volodymyrprokopyuk/go-blockchain/chain/store"
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
    fmt.Println("warning: genesis not found: > bcn store init, then restart")
    return nil
  }
  n.state = state.NewState(gen)
  err = n.state.ReadBlocks(n.blockStoreDir)
  if err != nil {
    if _, assert := err.(*os.PathError); !assert {
      return err
    }
    fmt.Println("warning: blocks not yet created")
  }
  fmt.Printf("* Read state (ReadBlocks)\n%v\n", n.state)
  return nil
}

func (n *Node) servegRPC() error {
  lis, err := net.Listen("tcp", n.nodeAddr)
  if err != nil {
    return err
  }
  defer lis.Close()
  fmt.Printf("* gRPC listening on %v\n", n.nodeAddr)
  srv := grpc.NewServer()
  acc := raccount.NewAccountSrv(n.keyStoreDir)
  raccount.RegisterAccountServer(srv, acc)
  sto := rstore.NewStoreSrv(n.keyStoreDir, n.blockStoreDir)
  rstore.RegisterStoreServer(srv, sto)
  tx := rtx.NewTxSrv(n.keyStoreDir, n.state)
  rtx.RegisterTxServer(srv, tx)
  return srv.Serve(lis)
}
