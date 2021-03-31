package currencypricing

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/masterhung0112/hk_server/v5/model"
	"github.com/shopspring/decimal"
)

const HardCodedSourceName = "hard-coded"

type MeanReversionRandomWalkPriceGenerator struct {
  halfSpreadPercentage decimal.Decimal
  reversion decimal.Decimal
  vol decimal.Decimal
  previousMid decimal.Decimal
  initial decimal.Decimal
  precision decimal.Decimal

  currencyPair model.CurrencyPair
  effectiveDate time.Time
  sourceName string

  priceChanges chan model.SpotPriceDto
}

func NewMeanReversionRandomWalkPriceGenerator(currencyPair model.CurrencyPair, initial decimal.Decimal, precision decimal.Decimal, reversionCoefficient decimal.Decimal, volatility decimal.Decimal) (*MeanReversionRandomWalkPriceGenerator, error) {
  if reversionCoefficient.IsNegative(){
    reversionCoefficient, _ = decimal.NewFromString("0.001")
  }
  if volatility.IsNegative() {
    volatility = decimal.NewFromInt(5)
  }

  power := decimal.NewFromInt(10).Pow(precision)

  randomValue := int64(rand.Intn(16 - 2 + 1) + 2)
  return &MeanReversionRandomWalkPriceGenerator{
    currencyPair: currencyPair,
    effectiveDate: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
    reversion: reversionCoefficient,
    vol: volatility.Div(power),
    initial: initial,
    precision: precision,
    halfSpreadPercentage: decimal.NewFromInt(randomValue).Div(power).Div(initial),
    previousMid: initial,
    sourceName: HardCodedSourceName,
  }, nil
}

func (o *MeanReversionRandomWalkPriceGenerator) UpdateWalkPrice() {
  random := decimal.NewFromInt32(rand.Int31())

  // _previousMid += _reversion * (_initial - _previousMid) + random * _vol;
  o.previousMid.Add(o.reversion.Mul(o.initial.Sub(o.previousMid)).Add(random.Mul((o.vol))))

  oneDecimal := decimal.NewFromInt(1)

  o.priceChanges <- model.SpotPriceDto{
    Symbol: o.currencyPair.Symbol(),
    ValueDate: time.Now().AddDate(0, 0, 14).Unix(),
    Mid: o.format(o.previousMid).String(),
    Ask: o.format(o.previousMid.Mul(oneDecimal.Add(o.halfSpreadPercentage))).String(),
    Bid: o.format(o.previousMid.Mul(oneDecimal.Sub(o.halfSpreadPercentage))).String(),
    CreationTimestamp: time.Now().Unix(),
  }
}

func (o *MeanReversionRandomWalkPriceGenerator) format(price decimal.Decimal) decimal.Decimal {
  power := decimal.NewFromInt(10).Pow(o.precision)
  return price.Mul(power).Floor().Div(power)
}

func (o *MeanReversionRandomWalkPriceGenerator) String() string {
  return fmt.Sprintf("%s|%s|%s|%s", o.currencyPair.Symbol(), o.effectiveDate.String(), o.initial.String(), o.sourceName)
}

/**
 * IPriceGenerator Interface
 */
func (o *MeanReversionRandomWalkPriceGenerator) CurrencyPair() model.CurrencyPair {
  return o.currencyPair
}

func (o *MeanReversionRandomWalkPriceGenerator) EffectiveDate() time.Time {
  return o.effectiveDate
}

func (o *MeanReversionRandomWalkPriceGenerator) SourceName() string {
  return o.sourceName
}

func (o *MeanReversionRandomWalkPriceGenerator) SampleRate() decimal.Decimal {
  return o.previousMid
}

func (o *MeanReversionRandomWalkPriceGenerator) UpdateInitialValue(newValue decimal.Decimal, effectDate time.Time, sourceName string) {
  o.initial = newValue
  o.previousMid = newValue
  o.effectiveDate = effectDate
  o.sourceName = sourceName
  o.UpdateWalkPrice()
}

func (o *MeanReversionRandomWalkPriceGenerator) PriceChanges() chan model.SpotPriceDto {
  return o.priceChanges
}

