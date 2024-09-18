package node

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
	"github.com/volodymyrprokopyuk/go-blockchain/chain/account"
	"github.com/volodymyrprokopyuk/go-blockchain/chain/state"
	"github.com/volodymyrprokopyuk/go-blockchain/node/rblock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type stateSync struct {
  cfg NodeCfg
  ctx context.Context
  dis *discovery
  state *state.State
}

func newStateSync(ctx context.Context, cfg NodeCfg, dis *discovery) *stateSync {
  return &stateSync{ctx: ctx, cfg: cfg, dis: dis}
}

func (s *stateSync) createGenesis() (chain.SigGenesis, error) {
  pass := []byte(s.cfg.Password)
  if len(pass) < 5 {
    return chain.SigGenesis{}, fmt.Errorf("password length is less than 5")
  }
  if s.cfg.Balance == 0 {
    return chain.SigGenesis{}, fmt.Errorf("balance must be positive")
  }
  acc, err := account.NewAccount()
  if err != nil {
    return chain.SigGenesis{}, err
  }
  err = acc.Write(s.cfg.KeyStoreDir, pass)
  s.cfg.Password = "erase"
  if err != nil {
    return chain.SigGenesis{}, err
  }
  gen := chain.NewGenesis(s.cfg.Chain, acc.Address(), s.cfg.Balance)
  sgen, err := acc.SignGen(gen)
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
  cln := rblock.NewBlockClient(conn)
  req := &rblock.GenesisSyncReq{}
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
    if _, assert := err.(*os.PathError); !assert {
      return err
    }
    fmt.Println("warning: blocks not yet created")
    return nil
  }
  defer closeBlocks()
  for err, blk := range blocks {
    if err != nil {
      return err
    }
    clo := s.state.Clone()
    err = clo.ApplyBlock(blk)
    if err != nil {
      return err
    }
    s.state.Apply(clo)
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
  cln := rblock.NewBlockClient(conn)
  req := &rblock.BlockSyncReq{Number: s.state.LastBlock().Number + 1}
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
  for _, peer := range s.dis.Peers() {
    blocks, closeBlocks, err := s.grpcBlockSync(peer)
    if err != nil {
      return err
    }
    defer closeBlocks()
    for err, jblk := range blocks {
      if err != nil {
        return err
      }
      blk, err := chain.UnmarshalBlockBytes(jblk)
      if err != nil {
        return err
      }
      clo := s.state.Clone()
      err = clo.ApplyBlock(blk)
      if err != nil {
        return err
      }
      s.state.Apply(clo)
      err = blk.Write(s.cfg.BlockStoreDir)
      if err != nil {
        return err
      }
    }
  }
  return nil
}

func (s *stateSync) syncState() (*state.State, error) {
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
  s.state = state.NewState(gen)
  err = s.readBlocks()
  if err != nil {
    return nil, err
  }
  err = s.syncBlocks()
  if err != nil {
    return nil, err
  }
  fmt.Printf("* Sync state (SyncBlocks)\n%v\n", s.state)
  return s.state, nil
}
