package currencypricing

import (
	"fmt"
	"strings"
	"time"

	"github.com/masterhung0112/hk_server/v5/shared/mlog"
	"github.com/shopspring/decimal"
)

type PriceSource struct {
  PriceGenerators map[string]IPriceGenerator
  MarketAdapters []IMarketDataAdapter
}

func NewPriceSource() (*PriceSource, error){

  return nil, nil
}

func (o *PriceSource) RefreshMarketRates() {
  for _, marketAdapter := range o.MarketAdapters {
    marketAdapterChan := marketAdapter.GetMarketData()

    for marketData := range marketAdapterChan {
      // Any item older than 10 minutes is considered available to update,
      // this way if the preceeding adapters did not update the rate then perhaps the next adapter will
      if priceGenerator, ok := o.PriceGenerators[marketData.CurrencyPair.Symbol()]; ok {
        //  (DateTime.UtcNow - priceGenerator.EffectiveDate).TotalMinutes > 10)
        if sampleRateDecimal, err := decimal.NewFromString(marketData.SampleRate); err != nil {
          priceGenerator.UpdateInitialValue(sampleRateDecimal, time.Unix(marketData.Date, 0), marketData.Source)
        } else {
          mlog.Error("Invalid Sample Rate", mlog.String("SampleRate", marketData.SampleRate))
        }
      }
    }
  }
  o.ComputeMissingReciprocals()
}

// Currency pairs are typically listed only as Major/Minor CCY codes.
// This method computes the reciprocal rate for missing Minor/Majors
func (o *PriceSource) ComputeMissingReciprocals() {
  var missing = make([]IPriceGenerator, 5)
  for _, priceGenerator := range o.PriceGenerators {
    if priceGenerator.SourceName() == HardCodedSourceName || strings.Contains(priceGenerator.SourceName(), "1/") {
      missing = append(missing, priceGenerator)
    }
  }
  for _, missingGenerator := range missing {
    if priceGenerator, ok := o.PriceGenerators[missingGenerator.CurrencyPair().ReciprocalSymbol()]; ok {
      if priceGenerator.SourceName() != HardCodedSourceName {
        priceGenerator.UpdateInitialValue(decimal.NewFromInt(1).Div(priceGenerator.SampleRate()), priceGenerator.EffectiveDate(), fmt.Sprintf("1/ %s", priceGenerator.SourceName()))
      }
    }
  }
}