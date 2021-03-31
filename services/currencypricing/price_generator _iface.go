package currencypricing

import (
	"context"

	"github.com/masterhung0112/hk_server/v5/model"
)

type IPriceGenerator interface {
  CurrencyPair() model.CurrencyPair
  EffectiveDate() int64
  SourceName() string
  SampleRate() float64
  UpdateInitialValue(newValue float64, effectDatae int64, sourceName string)
  UpdateWalkPrice()
  PriceChanges() chan model.SpotPriceDto
}

type IPricingService interface {
  GetPriceUpdates(context context.Context, request model.GetSpotStreamRequestDto) chan model.SpotPriceDto
  GetAllPriceUpdates() chan model.SpotPriceDto
}

type IMarketDataAdapter interface {
  RequestUriString() string
  GetMarketData() chan model.PricingMarketData
}