package model

const (
  LOWERCASE_LETTERS = "abcdefghijklmnopqrstuvwxyz"
  UPPERCASE_LETTERS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
  NUMBERS           = "0123456789"
  SYMBOLS           = " !\"\\#$%&'()*+,-./:;<=>?@[]^_`|~"
)

func NewAppError(where string, id string, params map[string]interface{}, details string, status int) *AppError {
  ap := &AppError{}
  ap.Id = id
  // ap.params = params
  ap.Message = id
  // ap.Where = where
  ap.DetailedError = details
  ap.StatusCode = status
  // ap.IsOAuth = false
  // ap.Translate(translateFunc)
  return ap
}
