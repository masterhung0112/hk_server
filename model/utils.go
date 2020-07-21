package model

import (
	"net/http"
	"strings"
	"io"
	"io/ioutil"
  "time"
  "crypto/rand"
  "encoding/base32"
  "encoding/json"
)

const (
  LOWERCASE_LETTERS = "abcdefghijklmnopqrstuvwxyz"
  UPPERCASE_LETTERS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
  NUMBERS           = "0123456789"
  SYMBOLS           = " !\"\\#$%&'()*+,-./:;<=>?@[]^_`|~"
)

var encoding = base32.NewEncoding("ybndrfg8ejkmcpqxot1uwisza345h769")

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

func NewRandomString(length int) string {
  data := make([]byte, 1+(length*5/8))
  rand.Read(data)
  return encoding.EncodeToString(data)[:length]
}


func GetMillis() int64 {
  return time.Now().UnixNano() / int64(time.Millisecond)
}

// MapToJson converts a map to a json string
func MapToJson(objmap map[string]string) string {
	b, _ := json.Marshal(objmap)
	return string(b)
}

func AppErrorFromJson(data io.Reader) *AppError {
	str := ""
	bytes, rerr := ioutil.ReadAll(data)
	if rerr != nil {
		str = rerr.Error()
	} else {
		str = string(bytes)
	}

	decoder := json.NewDecoder(strings.NewReader(str))
	var er AppError
	err := decoder.Decode(&er)
	if err == nil {
		return &er
	} else {
		return NewAppError("AppErrorFromJson", "model.utils.decode_json.app_error", nil, "body: "+str, http.StatusInternalServerError)
	}
}