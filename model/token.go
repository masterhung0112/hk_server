package model

const (
  TOKEN_SIZE            = 64
  MAX_TOKEN_EXPIRY_TIME = 1000 * 60 * 60 * 48 // 48 hour
  TOKEN_TYPE_OAUTH      = "oauth"
)

type Token struct {
  Token     string
  CreatedAt int64
  Type      string
  Extra     string
}

func NewToken(tokentype, extra string) *Token {
  return &Token{
    Token:    NewRandomString(TOKEN_SIZE),
    CreatedAt:  GetMillis(),
    Type:       tokentype,
    Extra:      extra,
  }
}
