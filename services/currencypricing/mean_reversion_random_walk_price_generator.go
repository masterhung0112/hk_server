package currencypricing

import (
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

  CurrencyPair model.CurrencyPair
  EffectiveDate time.Time
  SourceName string

  PriceChanges chan model.SpotPriceDto
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
    CurrencyPair: currencyPair,
    EffectiveDate: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
    reversion: reversionCoefficient,
    vol: volatility.Div(power),
    initial: initial,
    precision: precision,
    halfSpreadPercentage: decimal.NewFromInt(randomValue).Div(power).Div(initial),
    previousMid: initial,
    SourceName: HardCodedSourceName,
  }, nil
}

func (o *MeanReversionRandomWalkPriceGenerator) UpdateWalkPrice() {
  random := decimal.NewFromInt32(rand.Int31())

  // _previousMid += _reversion * (_initial - _previousMid) + random * _vol;
  o.previousMid.Add(o.reversion.Mul(o.initial.Sub(o.previousMid)).Add(random.Mul((o.vol))))

  oneDecimal := decimal.NewFromInt(1)

  o.PriceChanges <- model.SpotPriceDto{
    Symbol: o.CurrencyPair.Symbol(),
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