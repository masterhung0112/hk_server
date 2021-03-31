package currencypricing

type PriceSource struct {
  PriceGenerators map[string]IPriceGenerator
  MarketAdapters []IMarketDataAdapter
}