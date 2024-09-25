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
  mtx sync.Mutex
  chStreams map[string]chan chain.Event
}

func newEventStream(
  ctx context.Context, wg *sync.WaitGroup, cap int,
) *eventStream {
  return &eventStream{
    ctx: ctx, wg: wg, chEvent: make(chan chain.Event, cap),
    chStreams: make(map[string]chan chain.Event),
  }
}

func (s *eventStream) PublishEvent(event chain.Event) {
  s.chEvent <- event
}

func (s *eventStream) AddSubscriber(sub string) chan chain.Event {
  fmt.Printf("* Stream sub: %v\n", sub)
  s.mtx.Lock()
  defer s.mtx.Unlock()
  chStream := make(chan chain.Event)
  s.chStreams[sub] = chStream
  return chStream
}

func (s *eventStream) RemoveSubscriber(sub string) {
  fmt.Printf("* Stream sub remove: %v\n", sub)
  s.mtx.Lock()
  defer s.mtx.Unlock()
  chStream, exist := s.chStreams[sub]
  if exist {
    close(chStream)
    delete(s.chStreams, sub)
  }
}

func (s *eventStream) streamEvents() {
  defer s.wg.Done()
  for {
    select {
    case <- s.ctx.Done():
      return
    case event := <- s.chEvent:
      for _, chStream := range s.chStreams {
        chStream <- event
      }
    }
  }
}
