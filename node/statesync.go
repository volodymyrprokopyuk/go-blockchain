package node

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type stateSync struct {
  cfg NodeCfg
  ctx context.Context
  state *chain.State
  peerDisc *peerDiscovery
}

func newStateSync(
  ctx context.Context, cfg NodeCfg, peerDisc *peerDiscovery,
) *stateSync {
  return &stateSync{ctx: ctx, cfg: cfg, peerDisc: peerDisc}
}

func (s *stateSync) createGenesis() (chain.SigGenesis, error) {
  authPass := []byte(s.cfg.AuthPass)
  if len(authPass) < 5 {
    return chain.SigGenesis{}, fmt.Errorf("authpass length is less than 5")
  }
  auth, err := chain.NewAccount()
  if err != nil {
    return chain.SigGenesis{}, err
  }
  err = auth.Write(s.cfg.KeyStoreDir, authPass)
  if err != nil {
    return chain.SigGenesis{}, err
  }
  ownerPass := []byte(s.cfg.OwnerPass)
  if len(ownerPass) < 5 {
    return chain.SigGenesis{}, fmt.Errorf("ownerpass length is less than 5")
  }
  if s.cfg.Balance == 0 {
    return chain.SigGenesis{}, fmt.Errorf("balance must be positive")
  }
  acc, err := chain.NewAccount()
  if err != nil {
    return chain.SigGenesis{}, err
  }
  err = acc.Write(s.cfg.KeyStoreDir, ownerPass)
  s.cfg.OwnerPass = "erase"
  if err != nil {
    return chain.SigGenesis{}, err
  }
  gen := chain.NewGenesis(
    s.cfg.Chain, auth.Address(), acc.Address(), s.cfg.Balance,
  )
  sgen, err := auth.SignGen(gen)
  if err != nil {
    return chain.SigGenesis{}, err
  }
  err = sgen.Write(s.cfg.BlockStoreDir)
  if err != nil {
    return chain.SigGenesis{}, err
  }
  return sgen, nil
}

func (s *stateSync) grpcGenesisSync() ([]byte, error) {
  conn, err := grpc.NewClient(
    s.cfg.SeedAddr, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return nil, err
  }
  defer conn.Close()
  cln := rpc.NewBlockClient(conn)
  req := &rpc.GenesisSyncReq{}
  res, err := cln.GenesisSync(s.ctx, req)
  if err != nil {
    return nil, err
  }
  return res.Genesis, nil
}

func (s *stateSync) syncGenesis() (chain.SigGenesis, error) {
  jgen, err := s.grpcGenesisSync()
  if err != nil {
    return chain.SigGenesis{}, err
  }
  var gen chain.SigGenesis
  err = json.Unmarshal(jgen, &gen)
  if err != nil {
    return chain.SigGenesis{}, err
  }
  valid, err := chain.VerifyGen(gen)
  if err != nil {
    return chain.SigGenesis{}, err
  }
  if !valid {
    return chain.SigGenesis{}, fmt.Errorf("invalid genesis signature")
  }
  err = gen.Write(s.cfg.BlockStoreDir)
  if err != nil {
    return chain.SigGenesis{}, err
  }
  return gen, nil
}

func (s *stateSync) readBlocks() error {
  blocks, closeBlocks, err := chain.ReadBlocks(s.cfg.BlockStoreDir)
  if err != nil {
    return err
  }
  defer closeBlocks()
  for err, blk := range blocks {
    if err != nil {
      return err
    }
    clone := s.state.Clone()
    err = clone.ApplyBlock(blk)
    if err != nil {
      return err
    }
    s.state.Apply(clone)
  }
  return nil
}

func (s *stateSync) grpcBlockSync(peer string) (
  func(yield (func(err error, jblk []byte) bool)), func(), error,
) {
  conn, err := grpc.NewClient(
    peer, grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return nil, nil, err
  }
  close := func() {
    conn.Close()
  }
  cln := rpc.NewBlockClient(conn)
  req := &rpc.BlockSyncReq{Number: s.state.LastBlock().Number + 1}
  stream, err := cln.BlockSync(s.ctx, req)
  if err != nil {
    return nil, nil, err
  }
  more := true
  blocks := func(yield func(err error, jblk []byte) bool) {
    for more {
      res, err := stream.Recv()
      if err == io.EOF {
        return
      }
      if err != nil {
        yield(err, nil)
        return
      }
      more = yield(nil, res.Block)
    }
  }
  return blocks, close, nil
}

func (s *stateSync) syncBlocks() error {
  for _, peer := range s.peerDisc.Peers() {
    blocks, closeBlocks, err := s.grpcBlockSync(peer)
    if err != nil {
      return err
    }
    defer closeBlocks()
    for err, jblk := range blocks {
      if err != nil {
        return err
      }
      var blk chain.SigBlock
      err = json.Unmarshal(jblk, &blk)
      if err != nil {
        return err
      }
      clone := s.state.Clone()
      err = clone.ApplyBlock(blk)
      if err != nil {
        return err
      }
      s.state.Apply(clone)
      err = blk.Write(s.cfg.BlockStoreDir)
      if err != nil {
        return err
      }
    }
  }
  return nil
}

func (s *stateSync) syncState() (*chain.State, error) {
  gen, err := chain.ReadGenesis(s.cfg.BlockStoreDir)
  if err != nil {
    if s.cfg.Bootstrap {
      gen, err = s.createGenesis()
      if err != nil {
        return nil, err
      }
    } else {
      gen, err = s.syncGenesis()
      if err != nil {
        return nil, err
      }
    }
  }
  valid, err := chain.VerifyGen(gen)
  if err != nil {
    return nil, err
  }
  if !valid {
    return nil, fmt.Errorf("invalid genesis signature")
  }
  s.state = chain.NewState(gen)
  err = chain.InitBlockStore(s.cfg.BlockStoreDir)
  if err != nil {
    return nil, err
  }
  err = s.readBlocks()
  if err != nil {
    return nil, err
  }
  err = s.syncBlocks()
  if err != nil {
    return nil, err
  }
  fmt.Printf("== Sync state\n%v", s.state)
  return s.state, nil
}
