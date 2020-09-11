package model

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const (
  ME                                 = "me"

  DEFAULT_LOCALE          = "en"
	USER_AUTH_SERVICE_EMAIL = "email"
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

// Options for counting users
type UserCountOptions struct {
	// Should include users that are bots
	IncludeBotAccounts bool
	// Should include deleted users (of any type)
	IncludeDeleted bool
	// Exclude regular users
	ExcludeRegularUsers bool
	// Only include users on a specific team. "" for any team.
	TeamId string
	// Only include users on a specific channel. "" for any channel.
	ChannelId string
	// Restrict to search in a list of teams and channels
	ViewRestrictions *ViewUsersRestrictions
	// Only include users matching any of the given system wide roles.
	Roles []string
	// Only include users matching any of the given channel roles, must be used with ChannelId.
	ChannelRoles []string
	// Only include users matching any of the given team roles, must be used with TeamId.
	TeamRoles []string
}

type ViewUsersRestrictions struct {
	Teams    []string
	Channels []string
}

func (r *ViewUsersRestrictions) Hash() string {
	if r == nil {
		return ""
	}
	ids := append(r.Teams, r.Channels...)
	sort.Strings(ids)
	hash := sha256.New()
	hash.Write([]byte(strings.Join(ids, "")))
	return fmt.Sprintf("%x", hash.Sum(nil))
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


func (u *User) MakeNonNil() {
  //TODO: Open
	// if u.Props == nil {
	// 	u.Props = make(map[string]string)
	// }

	// if u.NotifyProps == nil {
	// 	u.NotifyProps = make(map[string]string)
	// }
}

// Remove any private data from the user object
func (u *User) Sanitize(options map[string]bool) {
  u.Password = ""
  //TODO: Open
	// u.AuthData = NewString("")
	// u.MfaSecret = ""

	if len(options) != 0 && !options["email"] {
		u.Email = ""
	}
	if len(options) != 0 && !options["fullname"] {
		u.FirstName = ""
		u.LastName = ""
  }
  //TODO: Open
	if len(options) != 0 && !options["passwordupdate"] {
		// u.LastPasswordUpdate = 0
	}
	if len(options) != 0 && !options["authservice"] {
		// u.AuthService = ""
	}
}

func (u *User) GetRoles() []string {
	return strings.Fields(u.Roles)
}

func (u *User) ClearNonProfileFields() {
  u.Password = ""
  //TODO: Open this
	// u.AuthData = NewString("")
	// u.MfaSecret = ""
	u.EmailVerified = false
	// u.AllowMarketing = false
	// u.NotifyProps = StringMap{}
	// u.LastPasswordUpdate = 0
	// u.FailedAttempts = 0
}

func (u *User) SanitizeProfile(options map[string]bool) {
	u.ClearNonProfileFields()

	u.Sanitize(options)
}