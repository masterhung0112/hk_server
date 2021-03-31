package currencypricing

import (
	"context"
	"time"

	"github.com/masterhung0112/hk_server/v5/model"
	"github.com/shopspring/decimal"
)

type IPriceGenerator interface {
  CurrencyPair() model.CurrencyPair
  EffectiveDate() time.Time
  SourceName() string
  SampleRate() decimal.Decimal
  UpdateInitialValue(newValue decimal.Decimal, effectDate time.Time, sourceName string)
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