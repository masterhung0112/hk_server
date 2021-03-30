package model

import "fmt"

type MarketData struct {
  CurrencyPair CurrencyPair
  SampleRate float64
  Date int64
  Source string
}

func (o *MarketData) ToString() string {
  return fmt.Sprintf("%s | %f | %d | %s ", o.CurrencyPair.Symbol(), o.SampleRate, o.Date, o.Source)
}