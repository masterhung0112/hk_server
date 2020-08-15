package model

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id            string `json:"id"`
	CreateAt      int64  `json:"create_at,omitempty"`
	UpdateAt      int64  `json:"update_at,omitempty"`
	DeleteAt      int64  `json:"delete_at"`
	Username      string `json:"username"`
	Password      string `json:"password,omitempty"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified,omitempty"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Roles         string `json:"roles"`
}

const (
	USER_EMAIL_MAX_LENGTH     = 128
	USER_NICKNAME_MAX_RUNES   = 64
	USER_POSITION_MAX_RUNES   = 128
	USER_FIRST_NAME_MAX_RUNES = 64
	USER_LAST_NAME_MAX_RUNES  = 64
	USER_AUTH_DATA_MAX_LENGTH = 128
	USER_NAME_MAX_LENGTH      = 64
	USER_NAME_MIN_LENGTH      = 1
	USER_PASSWORD_MAX_LENGTH  = 72
	USER_LOCALE_MAX_LENGTH    = 5
)

// UserFromJson will decode the input and return a User
func UserFromJson(data io.Reader) *User {
	var user *User
	json.NewDecoder(data).Decode(&user)
	return user
}

func (user *User) ToJson() string {
	b, _ := json.Marshal(user)
	return string(b)
}

func NormalizeUsername(username string) string {
	return strings.ToLower(username)
}

func NormalizeEmail(email string) string {
	return strings.ToLower(email)
}

// HashPassword generates a hash using the bcrypt.GenerateFromPassword
func HashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		panic(err)
	}

	return string(hash)
}

// PreSave will set the Id and Username if missing.  It will also fill
// in the CreateAt, UpdateAt times. It will also hash the password. It should
// be run before saving the user to the db.
func (u *User) PreSave() {
	if u.Id == "" {
		u.Id = NewId()
	}

	if u.Username == "" {
		u.Username = NewId()
	}

	// if u.AuthData != nil && *u.AuthData == "" {
	// 	u.AuthData = nil
	// }

	u.Username = SanitizeUnicode(u.Username)
	u.FirstName = SanitizeUnicode(u.FirstName)
	u.LastName = SanitizeUnicode(u.LastName)
	// u.Nickname = SanitizeUnicode(u.Nickname)

	u.Username = NormalizeUsername(u.Username)
	u.Email = NormalizeEmail(u.Email)

	u.CreateAt = GetMillis()
	u.UpdateAt = u.CreateAt

	// u.LastPasswordUpdate = u.CreateAt

	// u.MfaActive = false

	// if u.Locale == "" {
	// 	u.Locale = DEFAULT_LOCALE
	// }

	// if u.Props == nil {
	// 	u.Props = make(map[string]string)
	// }

	// if u.NotifyProps == nil || len(u.NotifyProps) == 0 {
	// 	u.SetDefaultNotifications()
	// }

	// if u.Timezone == nil {
	// 	u.Timezone = timezones.DefaultUserTimezone()
	// }

	if len(u.Password) > 0 {
		u.Password = HashPassword(u.Password)
	}
}

// Generate the AppError for user validation
func InvalidUserError(fieldName string, userId string) *AppError {
	id := fmt.Sprintf("model.user.is_valid.%s.app_error", fieldName)
	details := ""
	if userId != "" {
		details = "user_id=" + userId
	}
	return NewAppError("User.IsValid", id, nil, details, http.StatusBadRequest)
}

var validUsernameChars = regexp.MustCompile(`^[a-z0-9\.\-_]+$`)

var restrictedUsernames = []string{
	"all",
	"channel",
	"matterbot",
	"system",
}

func IsValidUsername(s string) bool {
	if len(s) < USER_NAME_MIN_LENGTH || len(s) > USER_NAME_MAX_LENGTH {
		return false
	}

	if !validUsernameChars.MatchString(s) {
		return false
	}

	for _, restrictedUsername := range restrictedUsernames {
		if s == restrictedUsername {
			return false
		}
	}

	return true
}

// IsValid validates the user and returns an error if it isn't configured
// correctly.
func (u *User) IsValid() *AppError {

	if !IsValidId(u.Id) {
		return InvalidUserError("id", "")
	}

	if u.CreateAt == 0 {
		return InvalidUserError("create_at", u.Id)
	}

	if u.UpdateAt == 0 {
		return InvalidUserError("update_at", u.Id)
	}

	if !IsValidUsername(u.Username) {
		return InvalidUserError("username", u.Id)
	}

	if len(u.Email) > USER_EMAIL_MAX_LENGTH || len(u.Email) == 0 || !IsValidEmail(u.Email) {
		return InvalidUserError("email", u.Id)
	}

	// if utf8.RuneCountInString(u.Nickname) > USER_NICKNAME_MAX_RUNES {
	// 	return InvalidUserError("nickname", u.Id)
	// }

	// if utf8.RuneCountInString(u.Position) > USER_POSITION_MAX_RUNES {
	// 	return InvalidUserError("position", u.Id)
	// }

	if utf8.RuneCountInString(u.FirstName) > USER_FIRST_NAME_MAX_RUNES {
		return InvalidUserError("first_name", u.Id)
	}

	if utf8.RuneCountInString(u.LastName) > USER_LAST_NAME_MAX_RUNES {
		return InvalidUserError("last_name", u.Id)
	}

	// if u.AuthData != nil && len(*u.AuthData) > USER_AUTH_DATA_MAX_LENGTH {
	// 	return InvalidUserError("auth_data", u.Id)
	// }

	// if u.AuthData != nil && len(*u.AuthData) > 0 && len(u.AuthService) == 0 {
	// 	return InvalidUserError("auth_data_type", u.Id)
	// }

	// if len(u.Password) > 0 && u.AuthData != nil && len(*u.AuthData) > 0 {
	// 	return InvalidUserError("auth_data_pwd", u.Id)
	// }

	if len(u.Password) > USER_PASSWORD_MAX_LENGTH {
		return InvalidUserError("password_limit", u.Id)
	}

	// if !IsValidLocale(u.Locale) {
	// 	return InvalidUserError("locale", u.Id)
	// }

	return nil
}
