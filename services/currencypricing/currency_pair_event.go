package currencypricing

type CurrencyPairCreatedEvent struct {
  Symbols string
  PipsPosition int
  RatePrecision int
  SampleRate decimal.Decimal
  Comment string
}