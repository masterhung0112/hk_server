package model

const (
  PASSWORD_MAXIMUM_LENGTH = 64
  PASSWORD_MINIMUM_LENGTH = 5
)

type Config struct {
  PasswordSettings PasswordSettings
}

func (c *Config) SetDefaults() {

  c.PasswordSettings.SetDefaults()
}

type PasswordSettings struct {
  MinimumLength *int
  Lowercase     *bool
  Number        *bool
  Uppercase     *bool
  Symbol        *bool
}

func (s *PasswordSettings) SetDefaults() {
  if s.MinimumLength == nil {
    s.MinimumLength = NewInt(10)
  }

  if s.Lowercase == nil {
    s.Lowercase = NewBool(true)
  }

  if s.Number == nil {
    s.Number = NewBool(true)
  }

  if s.Uppercase == nil {
    s.Uppercase = NewBool(true)
  }

  if s.Symbol == nil {
    s.Symbol = NewBool(true)
  }
}