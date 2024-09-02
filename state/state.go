package state

import "github.com/volodymyrprokopyuk/go-blockchain/account"

type State struct {
  balances map[account.Address]uint
  nonces map[account.Address]uint
}
