package currencypricing

type PriceSource struct {
  PriceGenerators map[string]IPriceGenerator
  MarketAdapters []IMarketDataAdapter
}

func NewPriceSource() (*PriceSource, error){

  return nil, nil
}