package chain

import (
	"encoding/json"
	"fmt"
)

type EventType uint64

const (
  EvAll EventType = 0
  EvTx EventType = 1
  EvBlock EventType = 2
)

func NewEventType(eventStr string) EventType {
  switch eventStr {
  case "all":
    return EvAll
  case "tx":
    return EvTx
  case "blk", "block":
    return EvBlock
  default:
    panic(fmt.Sprintf("unsupported event type: %v", eventStr))
  }
}

func (t EventType) String() string {
  switch t {
  case EvTx:
    return "tx"
  case EvBlock:
    return "blk"
  default:
    return "ev"
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
    var blk SigBlock
    err := json.Unmarshal(e.Body, &blk)
    if err != nil {
      return err.Error()
    }
    return fmt.Sprintf("%v %v\n%v", e.Type, e.Action, blk)
  default:
    return fmt.Sprintf("Unsupported EventType %v", e.Type)
  }
}

type EventPublisher interface {
  PublishEvent(event Event)
}
