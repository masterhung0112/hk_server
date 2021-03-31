package currencypricing

import (
	"github.com/masterhung0112/hk_server/v5/model"
	"github.com/shopspring/decimal"
)
type MeanReversionRandomWalkPriceGenerator struct {

}


func NewMeanReversionRandomWalkPriceGenerator(currencyPair model.CurrencyPair, initial decimal.Decimal, precision int, reversionCoefficient decimal.Decimal, volatility decimal.Decimal) (*MeanReversionRandomWalkPriceGenerator, error) {
  return nil, nil
}

func (o *MeanReversionRandomWalkPriceGenerator) UpdateWalkPrice() {

}