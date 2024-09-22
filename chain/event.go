package chain

import (
	"encoding/json"
	"fmt"
)

type EventType uint32

const (
  EvAll EventType = 0
  EvTx EventType = 1
  EvBlock EventType = 2
)

func (t EventType) String() string {
  switch t {
  case EvTx:
    return "Tx"
  case EvBlock:
    return "Block"
  default:
    return "Event"
  }
}

type Event struct {
  Type EventType `json:"type"`
  Action string `json:"action"`
  Body []byte `json:"body"`
}

func NewEvent(evType EventType, action string, body []byte) Event {
  return Event{Type: evType, Action: action, Body: body}
}

func (e Event) String() string {
  switch e.Type {
  case EvTx:
    var tx SigTx
    err := json.Unmarshal(e.Body, &tx)
    if err != nil {
      return err.Error()
    }
    return fmt.Sprintf("%v %v\n%v", e.Type, e.Action, tx)
  case EvBlock:
    var blk Block
    err := json.Unmarshal(e.Body, &blk)
    if err != nil {
      return err.Error()
    }
    return fmt.Sprintf("%v %v\n%v", e.Type, e.Action, blk)
  default:
    return fmt.Sprintf("Unsupported EventType %v", e.Type)
  }
}

type EventStreamer interface {
  PublishEvent(event Event)
}
