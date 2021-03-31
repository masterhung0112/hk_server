package eventstore

import (
	"github.com/EventStore/EventStore-Client-Go/messages"
)

type AggregateEvent interface {
  ToJson() string
  EventType() string
}

type IAggregate interface {
  Version() uint64
  Identifier() string
  ApplyEvent(event messages.RecordedEvent)
  GetPendingEvents() []AggregateEvent
  ClearPendingEvents()
}

type AggregateBase struct {
  pendingEvents []AggregateEvent
  version int64
}

func (o *AggregateBase) Version() int {
  return -1
}

func (o *AggregateBase) Identifier() string {
  return ""
}

func (o *AggregateBase) ApplyEvent(event interface{}) {

}

func (o *AggregateBase) ClearPendingEvents() {
  o.pendingEvents = nil
}