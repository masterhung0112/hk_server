package eventstore

import (
  "github.com/EventStore/EventStore-Client-Go/messages"
)

type IAggregate interface {
  Version() int
  Identifier() string
  ApplyEvent(event messages.RecordedEvent)
  GetPendingEvents() []interface{}
  ClearPendingEvents()
}

type AggregateBase struct {
  pendingEvents []interface{}
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
  o.pendingEvents = []interface{}
}