package node

import (
	"context"
	"fmt"
	"sync"

	"github.com/volodymyrprokopyuk/go-blockchain/chain"
)

type EventStream struct {
  ctx context.Context
  wg *sync.WaitGroup
  chEvent chan chain.Event
  mtx sync.Mutex
  chStreams map[string]chan chain.Event
}

func NewEventStream(
  ctx context.Context, wg *sync.WaitGroup, cap int,
) *EventStream {
  return &EventStream{
    ctx: ctx, wg: wg, chEvent: make(chan chain.Event, cap),
    chStreams: make(map[string]chan chain.Event),
  }
}

func (s *EventStream) PublishEvent(event chain.Event) {
  s.chEvent <- event
}

func (s *EventStream) AddSubscriber(sub string) chan chain.Event {
  fmt.Printf("<~> Stream: %v\n", sub)
  s.mtx.Lock()
  defer s.mtx.Unlock()
  chStream := make(chan chain.Event)
  s.chStreams[sub] = chStream
  return chStream
}

func (s *EventStream) RemoveSubscriber(sub string) {
  fmt.Printf("<~> Unsubscribe: %v\n", sub)
  s.mtx.Lock()
  defer s.mtx.Unlock()
  chStream, exist := s.chStreams[sub]
  if exist {
    close(chStream)
    delete(s.chStreams, sub)
  }
}

func (s *EventStream) StreamEvents() {
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
