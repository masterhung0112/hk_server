package currencypricing

import "github.com/masterhung0112/hk_server/v5/services/eventstore"


type CurrencyPairAggregate struct {
  eventstore.AggregateBase
}

func CreateCurrencyPairAggregate(symbol string) *CurrencyPairAggregate{
  return &CurrencyPairAggregate{
    AggregateBase: eventstore.AggregateBase{
      ID: "ccyPair" + symbol,

    },
  }
}

func (o *CurrencyPairAggregate) ApplyEvent(event eventstore.EventMessage) {
  switch e := event.Data.(type) {
  case *CurrencyPairCreatedEvent:
    o.Symbol = e.Symbol

  }
}