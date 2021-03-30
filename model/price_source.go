package model

import "encoding/json"

type CurrencyPair struct {
  BaseCcy string `json:"base_ccy"`
  QuoteCcy string `json:"quote_ccy"`
}

func (o *CurrencyPair) Symbol() string {
  return o.BaseCcy + o.QuoteCcy
}

func (o *CurrencyPair) ReciprocalSymbol() string {
  return o.QuoteCcy + o.BaseCcy
}

func (o *CurrencyPair) IsReciprocalOf(other *CurrencyPair) bool {
  return o.Symbol() == other.Symbol()
}

func (o *CurrencyPair) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}