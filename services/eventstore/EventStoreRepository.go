package eventstore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EventStore/EventStore-Client-Go/client"
	"github.com/EventStore/EventStore-Client-Go/direction"
	"github.com/EventStore/EventStore-Client-Go/messages"
	"github.com/masterhung0112/hk_server/v5/shared/mlog"
)

const (
  WritePageSize = 500
  ReadPageSize = 500
)

type IEventStoreRepository interface {
  GetById(context context.Context, id string, aggregate IAggregate) error
  SaveAsync(aggregate IAggregate, extraHeaders ...map[string]string) (int, error)
}

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
}

func (o *EventStoreRepository) SaveAsync(context context.Context, aggregate IAggregate, extraHeaders ...map[string]string) (uint64, error) {
  streamName := aggregate.Identifier()
  mlog.Info("Saving aggretate from Event Store", mlog.String("streamName", streamName))

  pendingEvents := aggregate.GetPendingEvents()
  originalVersion := aggregate.Version - len(pendingEvents)

  eventBatches := GetEventBatches(eventsToSave)

  if len(eventBatches) == 1 {
    // If just one batch write them straight to the Event Store
    writeResult, err := o.eventStoreClient.AppendToStream(context, streamName, originalVersion, eventBatches[0])
    if err != nil {
      return 0, err
    }
    return writeResult.NextExpectedVersion, nil
  } else {
    nextVersion := originalVersion
    // If we have more events to save than can be done in one batch according to the WritePageSize, then we need to save them in a transaction to ensure atomicity
    for _, eventBatch := range eventBatches {
      writeResult, err := o.eventStoreClient.AppendToStream(context, streamName, originalVersion, eventBatch)
      if err != nil {
        return originalVersion, err
      }
      nextVersion = writeResult.NextExpectedVersion
    }

    aggregate.ClearPendingEvents()
    return nextVersion, nil
  }
}

func ToProposedEvent(eventId string, eventType string, jsonData string, headers map[string]string) messages.ProposedEvent {
    // var eventHeaders = new Dictionary<string, string>(headers)
    // {
    //     {MetadataKeys.EventClrTypeHeader, evnt.GetType().AssemblyQualifiedName}
    // };

    return messages.ProposedEvent{
      EventID: eventId,
      EventType: eventType,
      ContentType: "application/json",
      Data: jsonData,
      UserMetadata: json.Marshal(headers),
    }
}