package eventstore

type EventMessage interface {
  ToJson() string
  EventType() string
}

type IAggregateHandler interface {
  Version() uint64
  GetID() string
  ApplyEvent(event EventMessage)
  GetPendingEvents() []EventMessage
  ClearPendingEvents()
  // ID      string
	// Type    string
	// Version int
	// Changes []Event
}

// AggregateBase contains the basic info
// that all aggregates should have
type AggregateBase struct {
  ID      string
	Type    string
	Version int
	Changes []EventMessage
}

func NewAggregateBase(id string) AggregateBase {
	return AggregateBase{
		ID:      id,
		Changes: []EventMessage{},
		Version: -1,
	}
}

// IncrementVersion ads 1 to the current version
func (o *AggregateBase) IncrementVersion() {
	o.Version++
}

func (o *AggregateBase) GetID() string {
  return o.ID
}

func (o *AggregateBase) ApplyEvent(aggregate IAggregateHandler, event EventMessage, commit bool) {
  // apply the event itself
	aggregate.ApplyEvent(event)

  o.IncrementVersion()

  if commit {
		// event.Version = o.Version
		_, event.Type = GetTypeName(event.Data)
		b.Changes = append(b.Changes, event)
	}
}

func (o *AggregateBase) ClearPendingEvents() {
  o.Changes = []EventMessage{}
}