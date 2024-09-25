package node

import (
	"context"
	"fmt"
	"sync"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

type eventStream struct {
  ctx context.Context
  wg *sync.WaitGroup
  chEvent chan chain.Event
}

func newEventStream(
  ctx context.Context, wg *sync.WaitGroup, cap int,
) *eventStream {
  return &eventStream{ctx: ctx, wg: wg, chEvent: make(chan chain.Event, cap)}
}

func (s *eventStream) PublishEvent(event chain.Event) {
  s.chEvent <- event
}

func (s *eventStream) streamEvents() {
  defer s.wg.Done()
  for {
    select {
    case <- s.ctx.Done():
      return
    case event := <- s.chEvent:
      fmt.Printf("=> %v\n", event)
    }
  }
}
