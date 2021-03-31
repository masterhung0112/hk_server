package model

import "fmt"

type PricingMarketData struct {
  CurrencyPair CurrencyPair
  SampleRate float64
  Date int64
  Source string
}

func (o *PricingMarketData) ToString() string {
  return fmt.Sprintf("%s | %f | %d | %s ", o.CurrencyPair.Symbol(), o.SampleRate, o.Date, o.Source)
}