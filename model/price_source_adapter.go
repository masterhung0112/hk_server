package model

import "fmt"

type PricingMarketData struct {
  CurrencyPair CurrencyPair
  SampleRate string
  Date int64
  Source string
}

func (o *PricingMarketData) ToString() string {
  return fmt.Sprintf("%s | %s | %d | %s ", o.CurrencyPair.Symbol(), o.SampleRate, o.Date, o.Source)
}