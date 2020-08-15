package model

import (
	"bytes"
  "net/http"
  "strings"
  "io"
  "io/ioutil"
  "time"
  "crypto/rand"
  "encoding/base32"
  "encoding/json"

  "github.com/pborman/uuid"
)

const (
  LOWERCASE_LETTERS = "abcdefghijklmnopqrstuvwxyz"
  UPPERCASE_LETTERS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
  NUMBERS           = "0123456789"
  SYMBOLS           = " !\"\\#$%&'()*+,-./:;<=>?@[]^_`|~"
)

type StringInterface map[string]interface{}
type StringMap map[string]string
type StringArray []string

func (sa StringArray) Equals(input StringArray) bool {

	if len(sa) != len(input) {
		return false
	}

	for index := range sa {

		if sa[index] != input[index] {
			return false
		}
	}

	return true
}

var encoding = base32.NewEncoding("ybndrfg8ejkmcpqxot1uwisza345h769")

// NewId is a globally unique identifier.  It is a [A-Z0-9] string 26
// characters long.  It is a UUID version 4 Guid that is zbased32 encoded
// with the padding stripped off.
func NewId() string {
	var b bytes.Buffer
	encoder := base32.NewEncoder(encoding, &b)
	encoder.Write(uuid.NewRandom())
	encoder.Close()
	b.Truncate(26) // removes the '==' padding
	return b.String()
}

func NewAppError(where string, id string, params map[string]interface{}, details string, status int) *AppError {
  ap := &AppError{}
  ap.Id = id
  // ap.params = params
  ap.Message = id
  ap.Where = where
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

func ArrayToJson(objmap []string) string {
	b, _ := json.Marshal(objmap)
	return string(b)
}

func ArrayFromJson(data io.Reader) []string {
	decoder := json.NewDecoder(data)

	var objmap []string
	if err := decoder.Decode(&objmap); err != nil {
		return make([]string, 0)
	} else {
		return objmap
	}
}

func ArrayFromInterface(data interface{}) []string {
	stringArray := []string{}

	dataArray, ok := data.([]interface{})
	if !ok {
		return stringArray
	}

	for _, v := range dataArray {
		if str, ok := v.(string); ok {
			stringArray = append(stringArray, str)
		}
	}

	return stringArray
}

func StringInterfaceToJson(objmap map[string]interface{}) string {
	b, _ := json.Marshal(objmap)
	return string(b)
}

func StringInterfaceFromJson(data io.Reader) map[string]interface{} {
	decoder := json.NewDecoder(data)

	var objmap map[string]interface{}
	if err := decoder.Decode(&objmap); err != nil {
		return make(map[string]interface{})
	} else {
		return objmap
	}
}

func StringToJson(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func StringFromJson(data io.Reader) string {
	decoder := json.NewDecoder(data)

	var s string
	if err := decoder.Decode(&s); err != nil {
		return ""
	} else {
		return s
	}
}