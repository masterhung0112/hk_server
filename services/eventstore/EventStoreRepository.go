package eventstore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EventStore/EventStore-Client-Go/client"
	"github.com/EventStore/EventStore-Client-Go/direction"
	"github.com/EventStore/EventStore-Client-Go/messages"
	"github.com/EventStore/EventStore-Client-Go/streamrevision"
	"github.com/gofrs/uuid"
	"github.com/masterhung0112/hk_server/v5/model"
	"github.com/masterhung0112/hk_server/v5/shared/mlog"
)

const (
  WritePageSize = 500
  ReadPageSize = 500
)

type IEventStoreRepository interface {
  GetById(context context.Context, id string, aggregate IAggregate) error
  SaveAsync(context context.Context, aggregate IAggregate, extraHeaders map[string]string) (uint64, error)
}

/**
 * Implementation
 */
type EventStoreRepository struct {
  eventStoreClient client.Client
}

func (o *EventStoreRepository) GetById(context context.Context, id string, aggregate IAggregate) error {
  streamName := fmt.Sprintf("%s%s", aggregate.Identifier, id)
  mlog.Info("Loading aggretate from Event Store", mlog.String("streamName", streamName))

  var eventNumber uint64 = 0
  for {
    recordedEventMsgs, err := o.eventStoreClient.ReadStreamEvents(context, direction.Forwards, streamName, eventNumber, ReadPageSize, false)
    if err != nil {
      return err
    }

    for _, recordedEventMsg := range recordedEventMsgs {
      aggregate.ApplyEvent(recordedEventMsg)
      eventNumber = recordedEventMsg.EventNumber
    }

    // End of stream
    if len(recordedEventMsgs) != ReadPageSize {
      break
    }
  }
  return nil
}

func (o *EventStoreRepository) SaveAsync(context context.Context, aggregate IAggregate, extraHeaders map[string]string) (uint64, error) {
  streamName := aggregate.Identifier()
  mlog.Info("Saving aggretate from Event Store", mlog.String("streamName", streamName))

  pendingEvents := aggregate.GetPendingEvents()
  var originalVersion uint64 = aggregate.Version() - uint64(len(pendingEvents))
  eventsToSave := []messages.ProposedEvent{}
  for _, pendingEvent := range pendingEvents {
    newEvent, err := ToProposedEvent(model.NewId(), pendingEvent, extraHeaders)
    if err != nil {
      return originalVersion, err
    }
    eventsToSave = append(eventsToSave, newEvent)
  }

  // eventBatches := eventsToSave// GetEventBatches(eventsToSave)

  writeResult, err := o.eventStoreClient.AppendToStream(context, streamName, streamrevision.NewStreamRevision(originalVersion), eventsToSave)
  if err != nil {
    return originalVersion, err
  }
  aggregate.ClearPendingEvents()

  return writeResult.NextExpectedVersion, nil
  // if len(eventBatches) == 1 {
  //   // If just one batch write them straight to the Event Store
  //   writeResult, err := o.eventStoreClient.AppendToStream(context, streamName, originalVersion, eventBatches[0])
  //   if err != nil {
  //     return 0, err
  //   }
  //   return writeResult.NextExpectedVersion, nil
  // } else {
  //   nextVersion := originalVersion
  //   // If we have more events to save than can be done in one batch according to the WritePageSize, then we need to save them in a transaction to ensure atomicity
  //   for _, eventBatch := range eventBatches {
  //     writeResult, err := o.eventStoreClient.AppendToStream(context, streamName, originalVersion, eventBatch)
  //     if err != nil {
  //       return originalVersion, err
  //     }
  //     nextVersion = writeResult.NextExpectedVersion
  //   }

  //   aggregate.ClearPendingEvents()
  //   return nextVersion, nil
  // }
}

func ToProposedEvent(eventId string, event AggregateEvent, headers map[string]string) (messages.ProposedEvent, error) {
    eventUuid, err := uuid.FromString(eventId)
    if err != nil {
      return messages.ProposedEvent{}, err
    }
    headerJson, err := json.Marshal(headers)
    if err != nil {
      return messages.ProposedEvent{}, err
    }
    return messages.ProposedEvent{
      EventID: eventUuid,
      EventType: event.EventType(),
      ContentType: "application/json",
      Data: []byte(event.ToJson()),
      UserMetadata: headerJson,
    }, nil
}