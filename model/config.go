package model

import (
	"io"
	"encoding/json"
)

const (
	CONN_SECURITY_NONE     = ""
	CONN_SECURITY_PLAIN    = "PLAIN"
	CONN_SECURITY_TLS      = "TLS"
	CONN_SECURITY_STARTTLS = "STARTTLS"

	PASSWORD_MAXIMUM_LENGTH                     = 64
	PASSWORD_MINIMUM_LENGTH                     = 5
	SERVICE_SETTINGS_DEFAULT_SITE_URL           = "http://localhost:9065"
	SERVICE_SETTINGS_DEFAULT_LISTEN_AND_ADDRESS = ":9065"

	DATABASE_DRIVER_SQLITE   = "sqlite3"
	DATABASE_DRIVER_MYSQL    = "mysql"
	DATABASE_DRIVER_POSTGRES = "postgres"

	IMAGE_DRIVER_LOCAL = "local"
	IMAGE_DRIVER_S3    = "s3"

	MINIO_ACCESS_KEY = "minioaccesskey"
	MINIO_SECRET_KEY = "miniosecretkey"
	MINIO_BUCKET     = "hungknow-test"

	FILE_SETTINGS_DEFAULT_DIRECTORY = "./data/"

	SQL_SETTINGS_DEFAULT_DATA_SOURCE = "hkuser:mostest@tcp(localhost:7306)/hungknow_test?charset=utf8mb4,utf8&readTimeout=30s&writeTimeout=30s"

	TEAM_SETTINGS_DEFAULT_SITE_NAME                = "HungKnow"
	TEAM_SETTINGS_DEFAULT_MAX_USERS_PER_TEAM       = 50
	TEAM_SETTINGS_DEFAULT_CUSTOM_BRAND_TEXT        = ""
	TEAM_SETTINGS_DEFAULT_CUSTOM_DESCRIPTION_TEXT  = ""
	TEAM_SETTINGS_DEFAULT_USER_STATUS_AWAY_TIMEOUT = 300

	DIRECT_MESSAGE_ANY  = "any"
	DIRECT_MESSAGE_TEAM = "team"

	LOCAL_MODE_SOCKET_PATH = "/var/tmp/hungknow_local.socket"

	EMAIL_SETTINGS_DEFAULT_FEEDBACK_ORGANIZATION = ""

	GENERIC_NO_CHANNEL_NOTIFICATION = "generic_no_channel"
	GENERIC_NOTIFICATION            = "generic"
	GENERIC_NOTIFICATION_SERVER     = "https://push-test.mattermost.com"
	MM_SUPPORT_ADDRESS              = "support@hungknow.com"
	FULL_NOTIFICATION               = "full"
	ID_LOADED_NOTIFICATION          = "id_loaded"

	EMAIL_BATCHING_BUFFER_SIZE = 256
	EMAIL_BATCHING_INTERVAL    = 30

	EMAIL_NOTIFICATION_CONTENTS_FULL    = "full"
	EMAIL_NOTIFICATION_CONTENTS_GENERIC = "generic"

	SERVICE_SETTINGS_DEFAULT_MAX_LOGIN_ATTEMPTS = 10

	PERMISSIONS_ALL           = "all"
	PERMISSIONS_CHANNEL_ADMIN = "channel_admin"
	PERMISSIONS_TEAM_ADMIN    = "team_admin"
	PERMISSIONS_SYSTEM_ADMIN  = "system_admin"

	ALLOW_EDIT_POST_ALWAYS     = "always"
	ALLOW_EDIT_POST_NEVER      = "never"
  ALLOW_EDIT_POST_TIME_LIMIT = "time_limit"

  FAKE_SETTING = "********************************"

  EXPERIMENTAL_SETTINGS_DEFAULT_LINK_METADATA_TIMEOUT_MILLISECONDS = 5000

  CLIENT_SIDE_CERT_CHECK_PRIMARY_AUTH   = "primary"
	CLIENT_SIDE_CERT_CHECK_SECONDARY_AUTH = "secondary"
)

const ConfigAccessTagWriteRestrictable = "write_restrictable"

type Config struct {
	ServiceSettings       ServiceSettings
	TeamSettings          TeamSettings
	PasswordSettings      PasswordSettings
	LocalizationSettings  LocalizationSettings
	SqlSettings           SqlSettings
	PrivacySettings       PrivacySettings
	EmailSettings         EmailSettings
  GuestAccountsSettings GuestAccountsSettings
  ExperimentalSettings      ExperimentalSettings
}

func ConfigFromJson(data io.Reader) *Config {
	var o *Config
	json.NewDecoder(data).Decode(&o)
	return o
}

// isUpdate detects a pre-existing config based on whether SiteURL has been changed
func (o *Config) isUpdate() bool {
	return o.ServiceSettings.SiteURL != nil
}

func (o *Config) SetDefaults() {
	isUpdate := o.isUpdate()

	o.TeamSettings.SetDefaults()
	o.ServiceSettings.SetDefaults(isUpdate)
	o.PasswordSettings.SetDefaults()
	o.LocalizationSettings.SetDefaults()
	o.SqlSettings.SetDefaults(isUpdate)
	o.PrivacySettings.SetDefaults()
	o.EmailSettings.SetDefaults(isUpdate)
  o.GuestAccountsSettings.SetDefaults()
  o.ExperimentalSettings.SetDefaults()
}

func (o *Config) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func (o *Config) Clone() *Config {
	var ret Config
	if err := json.Unmarshal([]byte(o.ToJson()), &ret); err != nil {
		panic(err)
	}
	return &ret
}

func (o *Config) IsValid() *AppError {
	//TODO: Uncomment out
	// if len(*o.ServiceSettings.SiteURL) == 0 && *o.EmailSettings.EnableEmailBatching {
	// 	return NewAppError("Config.IsValid", "model.config.is_valid.site_url_email_batching.app_error", nil, "", http.StatusBadRequest)
	// }

	// if *o.ClusterSettings.Enable && *o.EmailSettings.EnableEmailBatching {
	// 	return NewAppError("Config.IsValid", "model.config.is_valid.cluster_email_batching.app_error", nil, "", http.StatusBadRequest)
	// }

	// if len(*o.ServiceSettings.SiteURL) == 0 && *o.ServiceSettings.AllowCookiesForSubdomains {
	// 	return NewAppError("Config.IsValid", "model.config.is_valid.allow_cookies_for_subdomains.app_error", nil, "", http.StatusBadRequest)
	// }

	// if err := o.TeamSettings.isValid(); err != nil {
	// 	return err
	// }

	// if err := o.SqlSettings.isValid(); err != nil {
	// 	return err
	// }

	// if err := o.FileSettings.isValid(); err != nil {
	// 	return err
	// }

	// if err := o.EmailSettings.isValid(); err != nil {
	// 	return err
	// }

	// if err := o.LdapSettings.isValid(); err != nil {
	// 	return err
	// }

	// if err := o.SamlSettings.isValid(); err != nil {
	// 	return err
	// }

	// if *o.PasswordSettings.MinimumLength < PASSWORD_MINIMUM_LENGTH || *o.PasswordSettings.MinimumLength > PASSWORD_MAXIMUM_LENGTH {
	// 	return NewAppError("Config.IsValid", "model.config.is_valid.password_length.app_error", map[string]interface{}{"MinLength": PASSWORD_MINIMUM_LENGTH, "MaxLength": PASSWORD_MAXIMUM_LENGTH}, "", http.StatusBadRequest)
	// }

	// if err := o.RateLimitSettings.isValid(); err != nil {
	// 	return err
	// }

	// if err := o.ServiceSettings.isValid(); err != nil {
	// 	return err
	// }

	// if err := o.ElasticsearchSettings.isValid(); err != nil {
	// 	return err
	// }

	// if err := o.BleveSettings.isValid(); err != nil {
	// 	return err
	// }

	// if err := o.DataRetentionSettings.isValid(); err != nil {
	// 	return err
	// }

	// if err := o.LocalizationSettings.isValid(); err != nil {
	// 	return err
	// }

	// if err := o.MessageExportSettings.isValid(o.FileSettings); err != nil {
	// 	return err
	// }

	// if err := o.DisplaySettings.isValid(); err != nil {
	// 	return err
	// }

	// if err := o.ImageProxySettings.isValid(); err != nil {
	// 	return err
	// }
	return nil
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

type ServiceSettings struct {
	SiteURL                         *string `restricted:"true"`
	ConnectionSecurity              *string `restricted:"true"`
	ListenAddress                   *string `restricted:"true"`
	EnableDeveloper                 *bool   `restricted:"true"`
	LocalModeSocketLocation         *string
	ExtendSessionLengthWithActivity *bool `access:"environment,write_restrictable"`
	SessionIdleTimeoutInMinutes     *int  `access:"environment,write_restrictable"`
	EnableUserAccessTokens          *bool `access:"integrations"`
	MaximumLoginAttempts            *int  `access:"authentication,write_restrictable"`
	SessionLengthWebInDays          *int  `access:"environment,write_restrictable"`
	AllowCookiesForSubdomains       *bool `access:"write_restrictable"`
	SessionLengthMobileInDays       *int  `access:"environment,write_restrictable"`
	SessionLengthSSOInDays          *int  `access:"environment,write_restrictable"`
}

func (s *ServiceSettings) SetDefaults(isUpdate bool) {
	if s.ConnectionSecurity == nil {
		s.ConnectionSecurity = NewString("")
	}

	if s.SiteURL == nil {
		if s.EnableDeveloper != nil && *s.EnableDeveloper {
			s.SiteURL = NewString(SERVICE_SETTINGS_DEFAULT_SITE_URL)
		} else {
			s.SiteURL = NewString("")
		}
	}

	if s.EnableDeveloper == nil {
		s.EnableDeveloper = NewBool(false)
	}

	if s.ListenAddress == nil {
		s.ListenAddress = NewString(SERVICE_SETTINGS_DEFAULT_LISTEN_AND_ADDRESS)
	}

	if s.LocalModeSocketLocation == nil {
		s.LocalModeSocketLocation = NewString(LOCAL_MODE_SOCKET_PATH)
	}

	// Must be manually enabled for existing installations.
	if s.ExtendSessionLengthWithActivity == nil {
		s.ExtendSessionLengthWithActivity = NewBool(!isUpdate)
	}

	if s.SessionIdleTimeoutInMinutes == nil {
		s.SessionIdleTimeoutInMinutes = NewInt(43200)
	}

	if s.EnableUserAccessTokens == nil {
		s.EnableUserAccessTokens = NewBool(false)
	}

	if s.MaximumLoginAttempts == nil {
		s.MaximumLoginAttempts = NewInt(SERVICE_SETTINGS_DEFAULT_MAX_LOGIN_ATTEMPTS)
	}

	if s.SessionLengthWebInDays == nil {
		if isUpdate {
			s.SessionLengthWebInDays = NewInt(180)
		} else {
			s.SessionLengthWebInDays = NewInt(30)
		}
	}

	if s.SessionLengthMobileInDays == nil {
		if isUpdate {
			s.SessionLengthMobileInDays = NewInt(180)
		} else {
			s.SessionLengthMobileInDays = NewInt(30)
		}
	}

	if s.SessionLengthSSOInDays == nil {
		s.SessionLengthSSOInDays = NewInt(30)
	}

	if s.AllowCookiesForSubdomains == nil {
		s.AllowCookiesForSubdomains = NewBool(false)
	}
}

type LocalizationSettings struct {
	DefaultServerLocale *string
	DefaultClientLocale *string
	AvailableLocales    *string
}

func (s *LocalizationSettings) SetDefaults() {
	if s.DefaultServerLocale == nil {
		s.DefaultServerLocale = NewString(DEFAULT_LOCALE)
	}

	if s.DefaultClientLocale == nil {
		s.DefaultClientLocale = NewString(DEFAULT_LOCALE)
	}

	if s.AvailableLocales == nil {
		s.AvailableLocales = NewString("")
	}
}

type SqlSettings struct {
	DriverName         *string
	DataSource         *string `restricted:"true"`
	DataSourceReplicas []string
	QueryTimeout       *int `restricted:"true"`
}

func (s *SqlSettings) SetDefaults(isUpdate bool) {
	if s.DriverName == nil {
		s.DriverName = NewString(DATABASE_DRIVER_MYSQL)
	}

	if s.DataSource == nil {
		s.DataSource = NewString(SQL_SETTINGS_DEFAULT_DATA_SOURCE)
	}

	if s.DataSourceReplicas == nil {
		s.DataSourceReplicas = []string{}
	}

	if s.QueryTimeout == nil {
		s.QueryTimeout = NewInt(30)
	}
}

type FileSettings struct {
	Directory         *string `restricted:"true"`
	DriverName        *string `restricted:"true"`
	S3AccessKeyId     *string `restricted:"true"`
	S3SecretAccessKey *string `restricted:"true"`
	S3Bucket          *string `restricted:"true"`
	S3PathPrefix      *string `restricted:"true"`
	S3Region          *string `restricted:"true"`
	S3Endpoint        *string `restricted:"true"`
	S3SSL             *bool   `restricted:"true"`
	S3SignV2          *bool   `restricted:"true"`
	S3SSE             *bool   `restricted:"true"`
	S3Trace           *bool   `restricted:"true"`
}

func (s *FileSettings) SetDefaults(isUpdate bool) {
	if s.DriverName == nil {
		s.DriverName = NewString(IMAGE_DRIVER_LOCAL)
	}

	if s.Directory == nil || *s.Directory == "" {
		s.Directory = NewString(FILE_SETTINGS_DEFAULT_DIRECTORY)
	}

	if s.S3AccessKeyId == nil {
		s.S3AccessKeyId = NewString("")
	}

	if s.S3SecretAccessKey == nil {
		s.S3SecretAccessKey = NewString("")
	}

	if s.S3Bucket == nil {
		s.S3Bucket = NewString("")
	}

	if s.S3PathPrefix == nil {
		s.S3PathPrefix = NewString("")
	}

	if s.S3Region == nil {
		s.S3Region = NewString("")
	}

	if s.S3Endpoint == nil || len(*s.S3Endpoint) == 0 {
		// Defaults to "s3.amazonaws.com"
		s.S3Endpoint = NewString("s3.amazonaws.com")
	}

	if s.S3SSL == nil {
		s.S3SSL = NewBool(true) // Secure by default.
	}

	if s.S3SignV2 == nil {
		s.S3SignV2 = new(bool)
		// Signature v2 is not enabled by default.
	}

	if s.S3SSE == nil {
		s.S3SSE = NewBool(false) // Not Encrypted by default.
	}

	if s.S3Trace == nil {
		s.S3Trace = NewBool(false)
	}
}

type TeamSettings struct {
	SiteName                                 *string  `access:"site"`
	MaxUsersPerTeam                          *int     `access:"site"`
	DEPRECATED_DO_NOT_USE_EnableTeamCreation *bool    `json:"EnableTeamCreation" mapstructure:"EnableTeamCreation"` // This field is deprecated and must not be used.
	EnableUserCreation                       *bool    `access:"authentication"`
	EnableOpenServer                         *bool    `access:"authentication"`
	EnableUserDeactivation                   *bool    `access:"experimental"`
	RestrictCreationToDomains                *string  `access:"authentication"`
	EnableCustomBrand                        *bool    `access:"site"`
	CustomBrandText                          *string  `access:"site"`
	CustomDescriptionText                    *string  `access:"site"`
	RestrictDirectMessage                    *string  `access:"site"`
	EnableXToLeaveChannelsFromLHS            *bool    `access:"experimental"`
	UserStatusAwayTimeout                    *int64   `access:"experimental"`
	MaxChannelsPerTeam                       *int64   `access:"site"`
	MaxNotificationsPerChannel               *int64   `access:"environment"`
	EnableConfirmNotificationsToChannel      *bool    `access:"site"`
	TeammateNameDisplay                      *string  `access:"site"`
	ExperimentalViewArchivedChannels         *bool    `access:"experimental,site"`
	ExperimentalEnableAutomaticReplies       *bool    `access:"experimental"`
	ExperimentalHideTownSquareinLHS          *bool    `access:"experimental"`
	ExperimentalTownSquareIsReadOnly         *bool    `access:"experimental"`
	LockTeammateNameDisplay                  *bool    `access:"site"`
	ExperimentalPrimaryTeam                  *string  `access:"experimental"`
	ExperimentalDefaultChannels              []string `access:"experimental"`
}

func (s *TeamSettings) SetDefaults() {

	if s.SiteName == nil || *s.SiteName == "" {
		s.SiteName = NewString(TEAM_SETTINGS_DEFAULT_SITE_NAME)
	}

	if s.MaxUsersPerTeam == nil {
		s.MaxUsersPerTeam = NewInt(TEAM_SETTINGS_DEFAULT_MAX_USERS_PER_TEAM)
	}

	if s.EnableUserCreation == nil {
		s.EnableUserCreation = NewBool(true)
	}

	if s.EnableOpenServer == nil {
		s.EnableOpenServer = NewBool(false)
	}

	if s.RestrictCreationToDomains == nil {
		s.RestrictCreationToDomains = NewString("")
	}

	if s.EnableCustomBrand == nil {
		s.EnableCustomBrand = NewBool(false)
	}

	if s.EnableUserDeactivation == nil {
		s.EnableUserDeactivation = NewBool(false)
	}

	if s.CustomBrandText == nil {
		s.CustomBrandText = NewString(TEAM_SETTINGS_DEFAULT_CUSTOM_BRAND_TEXT)
	}

	if s.CustomDescriptionText == nil {
		s.CustomDescriptionText = NewString(TEAM_SETTINGS_DEFAULT_CUSTOM_DESCRIPTION_TEXT)
	}

	if s.RestrictDirectMessage == nil {
		s.RestrictDirectMessage = NewString(DIRECT_MESSAGE_ANY)
	}

	if s.EnableXToLeaveChannelsFromLHS == nil {
		s.EnableXToLeaveChannelsFromLHS = NewBool(false)
	}

	if s.UserStatusAwayTimeout == nil {
		s.UserStatusAwayTimeout = NewInt64(TEAM_SETTINGS_DEFAULT_USER_STATUS_AWAY_TIMEOUT)
	}

	if s.MaxChannelsPerTeam == nil {
		s.MaxChannelsPerTeam = NewInt64(2000)
	}

	if s.MaxNotificationsPerChannel == nil {
		s.MaxNotificationsPerChannel = NewInt64(1000)
	}

	if s.EnableConfirmNotificationsToChannel == nil {
		s.EnableConfirmNotificationsToChannel = NewBool(true)
	}

	if s.ExperimentalEnableAutomaticReplies == nil {
		s.ExperimentalEnableAutomaticReplies = NewBool(false)
	}

	if s.ExperimentalHideTownSquareinLHS == nil {
		s.ExperimentalHideTownSquareinLHS = NewBool(false)
	}

	if s.ExperimentalTownSquareIsReadOnly == nil {
		s.ExperimentalTownSquareIsReadOnly = NewBool(false)
	}

	if s.ExperimentalPrimaryTeam == nil {
		s.ExperimentalPrimaryTeam = NewString("")
	}

	if s.ExperimentalDefaultChannels == nil {
		s.ExperimentalDefaultChannels = []string{}
	}

	if s.EnableUserCreation == nil {
		s.EnableUserCreation = NewBool(true)
	}

	if s.ExperimentalViewArchivedChannels == nil {
		s.ExperimentalViewArchivedChannels = NewBool(false)
	}

	if s.LockTeammateNameDisplay == nil {
		s.LockTeammateNameDisplay = NewBool(false)
	}
}

func (o *Config) GetSanitizeOptions() map[string]bool {
	options := map[string]bool{}
	options["fullname"] = *o.PrivacySettings.ShowFullName
	options["email"] = *o.PrivacySettings.ShowEmailAddress

	return options
}

type PrivacySettings struct {
	ShowEmailAddress *bool `access:"site"`
	ShowFullName     *bool `access:"site"`
}

func (s *PrivacySettings) SetDefaults() {
	if s.ShowEmailAddress == nil {
		s.ShowEmailAddress = NewBool(true)
	}

	if s.ShowFullName == nil {
		s.ShowFullName = NewBool(true)
	}
}

type EmailSettings struct {
	EnableSignUpWithEmail             *bool   `access:"authentication"`
	EnableSignInWithEmail             *bool   `access:"authentication"`
	EnableSignInWithUsername          *bool   `access:"authentication"`
	SendEmailNotifications            *bool   `access:"site"`
	UseChannelInEmailNotifications    *bool   `access:"experimental"`
	RequireEmailVerification          *bool   `access:"authentication"`
	FeedbackName                      *string `access:"site"`
	FeedbackEmail                     *string `access:"site"`
	ReplyToAddress                    *string `access:"site"`
	FeedbackOrganization              *string `access:"site"`
	EnableSMTPAuth                    *bool   `access:"environment,write_restrictable"`
	SMTPUsername                      *string `access:"environment,write_restrictable"`
	SMTPPassword                      *string `access:"environment,write_restrictable"`
	SMTPServer                        *string `access:"environment,write_restrictable"`
	SMTPPort                          *string `access:"environment,write_restrictable"`
	SMTPServerTimeout                 *int
	ConnectionSecurity                *string `access:"environment,write_restrictable"`
	SendPushNotifications             *bool   `access:"environment"`
	PushNotificationServer            *string `access:"environment"`
	PushNotificationContents          *string `access:"site"`
	PushNotificationBuffer            *int
	EnableEmailBatching               *bool   `access:"site"`
	EmailBatchingBufferSize           *int    `access:"experimental"`
	EmailBatchingInterval             *int    `access:"experimental"`
	EnablePreviewModeBanner           *bool   `access:"site"`
	SkipServerCertificateVerification *bool   `access:"environment,write_restrictable"`
	EmailNotificationContentsType     *string `access:"site"`
	LoginButtonColor                  *string `access:"experimental"`
	LoginButtonBorderColor            *string `access:"experimental"`
	LoginButtonTextColor              *string `access:"experimental"`
}

func (s *EmailSettings) SetDefaults(isUpdate bool) {
	if s.EnableSignUpWithEmail == nil {
		s.EnableSignUpWithEmail = NewBool(true)
	}

	if s.EnableSignInWithEmail == nil {
		s.EnableSignInWithEmail = NewBool(*s.EnableSignUpWithEmail)
	}

	if s.EnableSignInWithUsername == nil {
		s.EnableSignInWithUsername = NewBool(true)
	}

	if s.SendEmailNotifications == nil {
		s.SendEmailNotifications = NewBool(true)
	}

	if s.UseChannelInEmailNotifications == nil {
		s.UseChannelInEmailNotifications = NewBool(false)
	}

	if s.RequireEmailVerification == nil {
		s.RequireEmailVerification = NewBool(false)
	}

	if s.FeedbackName == nil {
		s.FeedbackName = NewString("")
	}

	if s.FeedbackEmail == nil {
		s.FeedbackEmail = NewString("test@example.com")
	}

	if s.ReplyToAddress == nil {
		s.ReplyToAddress = NewString("test@example.com")
	}

	if s.FeedbackOrganization == nil {
		s.FeedbackOrganization = NewString(EMAIL_SETTINGS_DEFAULT_FEEDBACK_ORGANIZATION)
	}

	if s.EnableSMTPAuth == nil {
		if s.ConnectionSecurity == nil || *s.ConnectionSecurity == CONN_SECURITY_NONE {
			s.EnableSMTPAuth = NewBool(false)
		} else {
			s.EnableSMTPAuth = NewBool(true)
		}
	}

	if s.SMTPUsername == nil {
		s.SMTPUsername = NewString("")
	}

	if s.SMTPPassword == nil {
		s.SMTPPassword = NewString("")
	}

	if s.SMTPServer == nil || len(*s.SMTPServer) == 0 {
		s.SMTPServer = NewString("localhost")
	}

	if s.SMTPPort == nil || len(*s.SMTPPort) == 0 {
		s.SMTPPort = NewString("10025")
	}

	if s.SMTPServerTimeout == nil || *s.SMTPServerTimeout == 0 {
		s.SMTPServerTimeout = NewInt(10)
	}

	if s.ConnectionSecurity == nil || *s.ConnectionSecurity == CONN_SECURITY_PLAIN {
		s.ConnectionSecurity = NewString(CONN_SECURITY_NONE)
	}

	if s.SendPushNotifications == nil {
		s.SendPushNotifications = NewBool(!isUpdate)
	}

	if s.PushNotificationServer == nil {
		if isUpdate {
			s.PushNotificationServer = NewString("")
		} else {
			s.PushNotificationServer = NewString(GENERIC_NOTIFICATION_SERVER)
		}
	}

	if s.PushNotificationContents == nil {
		s.PushNotificationContents = NewString(FULL_NOTIFICATION)
	}

	if s.PushNotificationBuffer == nil {
		s.PushNotificationBuffer = NewInt(1000)
	}

	if s.EnableEmailBatching == nil {
		s.EnableEmailBatching = NewBool(false)
	}

	if s.EmailBatchingBufferSize == nil {
		s.EmailBatchingBufferSize = NewInt(EMAIL_BATCHING_BUFFER_SIZE)
	}

	if s.EmailBatchingInterval == nil {
		s.EmailBatchingInterval = NewInt(EMAIL_BATCHING_INTERVAL)
	}

	if s.EnablePreviewModeBanner == nil {
		s.EnablePreviewModeBanner = NewBool(true)
	}

	if s.EnableSMTPAuth == nil {
		if *s.ConnectionSecurity == CONN_SECURITY_NONE {
			s.EnableSMTPAuth = NewBool(false)
		} else {
			s.EnableSMTPAuth = NewBool(true)
		}
	}

	if *s.ConnectionSecurity == CONN_SECURITY_PLAIN {
		*s.ConnectionSecurity = CONN_SECURITY_NONE
	}

	if s.SkipServerCertificateVerification == nil {
		s.SkipServerCertificateVerification = NewBool(false)
	}

	if s.EmailNotificationContentsType == nil {
		s.EmailNotificationContentsType = NewString(EMAIL_NOTIFICATION_CONTENTS_FULL)
	}

	if s.LoginButtonColor == nil {
		s.LoginButtonColor = NewString("#0000")
	}

	if s.LoginButtonBorderColor == nil {
		s.LoginButtonBorderColor = NewString("#2389D7")
	}

	if s.LoginButtonTextColor == nil {
		s.LoginButtonTextColor = NewString("#2389D7")
	}
}

type GuestAccountsSettings struct {
	Enable                           *bool   `access:"authentication"`
	AllowEmailAccounts               *bool   `access:"authentication"`
	EnforceMultifactorAuthentication *bool   `access:"authentication"`
	RestrictCreationToDomains        *string `access:"authentication"`
}

func (s *GuestAccountsSettings) SetDefaults() {
	if s.Enable == nil {
		s.Enable = NewBool(false)
	}

	if s.AllowEmailAccounts == nil {
		s.AllowEmailAccounts = NewBool(true)
	}

	if s.EnforceMultifactorAuthentication == nil {
		s.EnforceMultifactorAuthentication = NewBool(false)
	}

	if s.RestrictCreationToDomains == nil {
		s.RestrictCreationToDomains = NewString("")
	}
}

func (o *Config) Sanitize() {
	// if o.LdapSettings.BindPassword != nil && len(*o.LdapSettings.BindPassword) > 0 {
	// 	*o.LdapSettings.BindPassword = FAKE_SETTING
	// }

	// *o.FileSettings.PublicLinkSalt = FAKE_SETTING

	// if len(*o.FileSettings.AmazonS3SecretAccessKey) > 0 {
	// 	*o.FileSettings.AmazonS3SecretAccessKey = FAKE_SETTING
	// }

	if o.EmailSettings.SMTPPassword != nil && len(*o.EmailSettings.SMTPPassword) > 0 {
		*o.EmailSettings.SMTPPassword = FAKE_SETTING
	}

	// if len(*o.GitLabSettings.Secret) > 0 {
	// 	*o.GitLabSettings.Secret = FAKE_SETTING
	// }

	// if o.GoogleSettings.Secret != nil && len(*o.GoogleSettings.Secret) > 0 {
	// 	*o.GoogleSettings.Secret = FAKE_SETTING
	// }

	// if o.Office365Settings.Secret != nil && len(*o.Office365Settings.Secret) > 0 {
	// 	*o.Office365Settings.Secret = FAKE_SETTING
	// }

	*o.SqlSettings.DataSource = FAKE_SETTING
	// *o.SqlSettings.AtRestEncryptKey = FAKE_SETTING

	// *o.ElasticsearchSettings.Password = FAKE_SETTING

	for i := range o.SqlSettings.DataSourceReplicas {
		o.SqlSettings.DataSourceReplicas[i] = FAKE_SETTING
	}

	// for i := range o.SqlSettings.DataSourceSearchReplicas {
	// 	o.SqlSettings.DataSourceSearchReplicas[i] = FAKE_SETTING
	// }

	// if o.MessageExportSettings.GlobalRelaySettings.SmtpPassword != nil && len(*o.MessageExportSettings.GlobalRelaySettings.SmtpPassword) > 0 {
	// 	*o.MessageExportSettings.GlobalRelaySettings.SmtpPassword = FAKE_SETTING
	// }

	// if o.ServiceSettings.GfycatApiSecret != nil && len(*o.ServiceSettings.GfycatApiSecret) > 0 {
	// 	*o.ServiceSettings.GfycatApiSecret = FAKE_SETTING
	// }
}

type ExperimentalSettings struct {
	ClientSideCertEnable            *bool   `access:"experimental"`
	ClientSideCertCheck             *string `access:"experimental"`
	EnableClickToReply              *bool   `access:"experimental,write_restrictable"`
	LinkMetadataTimeoutMilliseconds *int64  `access:"experimental,write_restrictable"`
	RestrictSystemAdmin             *bool   `access:"experimental,write_restrictable"`
	UseNewSAMLLibrary               *bool   `access:"experimental"`
	CloudUserLimit                  *int64  `access:"experimental,write_restrictable"`
	CloudBilling                    *bool   `access:"experimental,write_restrictable"`
}

func (s *ExperimentalSettings) SetDefaults() {
	if s.ClientSideCertEnable == nil {
		s.ClientSideCertEnable = NewBool(false)
	}

	if s.ClientSideCertCheck == nil {
		s.ClientSideCertCheck = NewString(CLIENT_SIDE_CERT_CHECK_SECONDARY_AUTH)
	}

	if s.EnableClickToReply == nil {
		s.EnableClickToReply = NewBool(false)
	}

	if s.LinkMetadataTimeoutMilliseconds == nil {
		s.LinkMetadataTimeoutMilliseconds = NewInt64(EXPERIMENTAL_SETTINGS_DEFAULT_LINK_METADATA_TIMEOUT_MILLISECONDS)
	}

	if s.RestrictSystemAdmin == nil {
		s.RestrictSystemAdmin = NewBool(false)
	}

	if s.CloudUserLimit == nil {
		// User limit 0 is treated as no limit
		s.CloudUserLimit = NewInt64(0)
	}

	if s.CloudBilling == nil {
		s.CloudBilling = NewBool(false)
	}

	if s.UseNewSAMLLibrary == nil {
		s.UseNewSAMLLibrary = NewBool(false)
	}
}