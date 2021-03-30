package model

import "encoding/json"

type PricingCurrencyPair struct {
  BaseCcy string `json:"base_ccy"`
  QuoteCcy string `json:"quote_ccy"`
}

func (o *PricingCurrencyPair) Symbol() string {
  return o.BaseCcy + o.QuoteCcy
}

func (o *PricingCurrencyPair) ReciprocalSymbol() string {
  return o.QuoteCcy + o.BaseCcy
}

func (o *PricingCurrencyPair) IsReciprocalOf(other *PricingCurrencyPair) bool {
  return o.Symbol() == other.Symbol()
}

func (o *PricingCurrencyPair) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}