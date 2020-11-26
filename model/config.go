package model

import (
	"encoding/json"
	"io"
)

const (
	CONN_SECURITY_NONE     = ""
	CONN_SECURITY_PLAIN    = "PLAIN"
	CONN_SECURITY_TLS      = "TLS"
	CONN_SECURITY_STARTTLS = "STARTTLS"

	PASSWORD_MAXIMUM_LENGTH = 64
	PASSWORD_MINIMUM_LENGTH = 5

	SITENAME_MAX_LENGTH = 30

	SERVICE_SETTINGS_DEFAULT_SITE_URL           = "http://localhost:9065"
	SERVICE_SETTINGS_DEFAULT_TLS_CERT_FILE      = ""
	SERVICE_SETTINGS_DEFAULT_TLS_KEY_FILE       = ""
	SERVICE_SETTINGS_DEFAULT_READ_TIMEOUT       = 300
	SERVICE_SETTINGS_DEFAULT_WRITE_TIMEOUT      = 300
	SERVICE_SETTINGS_DEFAULT_IDLE_TIMEOUT       = 60
	SERVICE_SETTINGS_DEFAULT_MAX_LOGIN_ATTEMPTS = 10
	SERVICE_SETTINGS_DEFAULT_ALLOW_CORS_FROM    = ""
	SERVICE_SETTINGS_DEFAULT_LISTEN_AND_ADDRESS = ":9065"
	SERVICE_SETTINGS_DEFAULT_GFYCAT_API_KEY     = "2_KtH_W5"
	SERVICE_SETTINGS_DEFAULT_GFYCAT_API_SECRET  = "3wLVZPiswc3DnaiaFoLkDvB4X0IV6CpMkj4tf2inJRsBY6-FnkT08zGmppWFgeof"

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

	SHOW_USERNAME          = "username"
	SHOW_NICKNAME_FULLNAME = "nickname_full_name"
	SHOW_FULLNAME          = "full_name"

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

	PERMISSIONS_ALL           = "all"
	PERMISSIONS_CHANNEL_ADMIN = "channel_admin"
	PERMISSIONS_TEAM_ADMIN    = "team_admin"
	PERMISSIONS_SYSTEM_ADMIN  = "system_admin"

	FAKE_SETTING = "********************************"

	RESTRICT_EMOJI_CREATION_ALL          = "all"
	RESTRICT_EMOJI_CREATION_ADMIN        = "admin"
	RESTRICT_EMOJI_CREATION_SYSTEM_ADMIN = "system_admin"

	PERMISSIONS_DELETE_POST_ALL          = "all"
	PERMISSIONS_DELETE_POST_TEAM_ADMIN   = "team_admin"
	PERMISSIONS_DELETE_POST_SYSTEM_ADMIN = "system_admin"

	ALLOW_EDIT_POST_ALWAYS     = "always"
	ALLOW_EDIT_POST_NEVER      = "never"
	ALLOW_EDIT_POST_TIME_LIMIT = "time_limit"

	GROUP_UNREAD_CHANNELS_DISABLED    = "disabled"
	GROUP_UNREAD_CHANNELS_DEFAULT_ON  = "default_on"
	GROUP_UNREAD_CHANNELS_DEFAULT_OFF = "default_off"

	COLLAPSED_THREADS_DISABLED    = "disabled"
	COLLAPSED_THREADS_DEFAULT_ON  = "default_on"
	COLLAPSED_THREADS_DEFAULT_OFF = "default_off"

	EXPERIMENTAL_SETTINGS_DEFAULT_LINK_METADATA_TIMEOUT_MILLISECONDS = 5000

	CLIENT_SIDE_CERT_CHECK_PRIMARY_AUTH   = "primary"
	CLIENT_SIDE_CERT_CHECK_SECONDARY_AUTH = "secondary"

	SUPPORT_SETTINGS_DEFAULT_TERMS_OF_SERVICE_LINK = "https://about.hungknow.com/default-terms/"
	SUPPORT_SETTINGS_DEFAULT_PRIVACY_POLICY_LINK   = "https://about.hungknow.com/default-privacy-policy/"
	SUPPORT_SETTINGS_DEFAULT_ABOUT_LINK            = "https://about.hungknow.com/default-about/"
	SUPPORT_SETTINGS_DEFAULT_HELP_LINK             = "https://about.hungknow.com/default-help/"
	SUPPORT_SETTINGS_DEFAULT_REPORT_A_PROBLEM_LINK = "https://about.hungknow.com/default-report-a-problem/"
	SUPPORT_SETTINGS_DEFAULT_SUPPORT_EMAIL         = "hungbn0112@gmail.com"
	SUPPORT_SETTINGS_DEFAULT_RE_ACCEPTANCE_PERIOD  = 365

	LDAP_SETTINGS_DEFAULT_FIRST_NAME_ATTRIBUTE         = ""
	LDAP_SETTINGS_DEFAULT_LAST_NAME_ATTRIBUTE          = ""
	LDAP_SETTINGS_DEFAULT_EMAIL_ATTRIBUTE              = ""
	LDAP_SETTINGS_DEFAULT_USERNAME_ATTRIBUTE           = ""
	LDAP_SETTINGS_DEFAULT_NICKNAME_ATTRIBUTE           = ""
	LDAP_SETTINGS_DEFAULT_ID_ATTRIBUTE                 = ""
	LDAP_SETTINGS_DEFAULT_POSITION_ATTRIBUTE           = ""
	LDAP_SETTINGS_DEFAULT_LOGIN_FIELD_NAME             = ""
	LDAP_SETTINGS_DEFAULT_GROUP_DISPLAY_NAME_ATTRIBUTE = ""
	LDAP_SETTINGS_DEFAULT_GROUP_ID_ATTRIBUTE           = ""
	LDAP_SETTINGS_DEFAULT_PICTURE_ATTRIBUTE            = ""

	SAML_SETTINGS_DEFAULT_ID_ATTRIBUTE         = ""
	SAML_SETTINGS_DEFAULT_GUEST_ATTRIBUTE      = ""
	SAML_SETTINGS_DEFAULT_ADMIN_ATTRIBUTE      = ""
	SAML_SETTINGS_DEFAULT_FIRST_NAME_ATTRIBUTE = ""
	SAML_SETTINGS_DEFAULT_LAST_NAME_ATTRIBUTE  = ""
	SAML_SETTINGS_DEFAULT_EMAIL_ATTRIBUTE      = ""
	SAML_SETTINGS_DEFAULT_USERNAME_ATTRIBUTE   = ""
	SAML_SETTINGS_DEFAULT_NICKNAME_ATTRIBUTE   = ""
	SAML_SETTINGS_DEFAULT_LOCALE_ATTRIBUTE     = ""
	SAML_SETTINGS_DEFAULT_POSITION_ATTRIBUTE   = ""

	SAML_SETTINGS_SIGNATURE_ALGORITHM_SHA1    = "RSAwithSHA1"
	SAML_SETTINGS_SIGNATURE_ALGORITHM_SHA256  = "RSAwithSHA256"
	SAML_SETTINGS_SIGNATURE_ALGORITHM_SHA512  = "RSAwithSHA512"
	SAML_SETTINGS_DEFAULT_SIGNATURE_ALGORITHM = SAML_SETTINGS_SIGNATURE_ALGORITHM_SHA1

	SAML_SETTINGS_CANONICAL_ALGORITHM_C14N    = "Canonical1.0"
	SAML_SETTINGS_CANONICAL_ALGORITHM_C14N11  = "Canonical1.1"
	SAML_SETTINGS_DEFAULT_CANONICAL_ALGORITHM = SAML_SETTINGS_CANONICAL_ALGORITHM_C14N

	PLUGIN_SETTINGS_DEFAULT_DIRECTORY          = "./plugins"
	PLUGIN_SETTINGS_DEFAULT_CLIENT_DIRECTORY   = "./client/plugins"
	PLUGIN_SETTINGS_DEFAULT_ENABLE_MARKETPLACE = true
	PLUGIN_SETTINGS_DEFAULT_MARKETPLACE_URL    = "https://api.integrations.hungknow.com"
	PLUGIN_SETTINGS_OLD_MARKETPLACE_URL        = "https://marketplace.integrations.hungknow.com"
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
	ExperimentalSettings  ExperimentalSettings
	SupportSettings       SupportSettings
	RateLimitSettings     RateLimitSettings
	FileSettings          FileSettings
	LogSettings           LogSettings
	LdapSettings          LdapSettings
	// ComplianceSettings        ComplianceSettings
	// LocalizationSettings      LocalizationSettings
	SamlSettings   SamlSettings
	PluginSettings PluginSettings
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

	o.LdapSettings.SetDefaults()
	o.SamlSettings.SetDefaults()

	if o.TeamSettings.TeammateNameDisplay == nil {
		o.TeamSettings.TeammateNameDisplay = NewString(SHOW_USERNAME)

		if *o.SamlSettings.Enable || *o.LdapSettings.Enable {
			*o.TeamSettings.TeammateNameDisplay = SHOW_FULLNAME
		}
	}

	o.TeamSettings.SetDefaults()
	o.ServiceSettings.SetDefaults(isUpdate)
	o.PasswordSettings.SetDefaults()
	o.LocalizationSettings.SetDefaults()
	o.SqlSettings.SetDefaults(isUpdate)
	o.PrivacySettings.SetDefaults()
	o.EmailSettings.SetDefaults(isUpdate)
	o.GuestAccountsSettings.SetDefaults()
	o.ExperimentalSettings.SetDefaults()
	o.SupportSettings.SetDefaults()
	o.RateLimitSettings.SetDefaults()
	o.FileSettings.SetDefaults(isUpdate)
	o.LogSettings.SetDefaults()
	o.PluginSettings.SetDefaults(o.LogSettings)
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
	SiteURL                                           *string  `access:"environment,authentication,write_restrictable"`
	WebsocketURL                                      *string  `access:"write_restrictable,cloud_restrictable"`
	LicenseFileLocation                               *string  `access:"write_restrictable,cloud_restrictable"`
	ListenAddress                                     *string  `access:"environment,write_restrictable,cloud_restrictable"`
	ConnectionSecurity                                *string  `access:"environment,write_restrictable,cloud_restrictable"`
	TLSCertFile                                       *string  `access:"environment,write_restrictable,cloud_restrictable"`
	TLSKeyFile                                        *string  `access:"environment,write_restrictable,cloud_restrictable"`
	TLSMinVer                                         *string  `access:"write_restrictable,cloud_restrictable"`
	TLSStrictTransport                                *bool    `access:"write_restrictable,cloud_restrictable"`
	TLSStrictTransportMaxAge                          *int64   `access:"write_restrictable,cloud_restrictable"`
	TLSOverwriteCiphers                               []string `access:"write_restrictable,cloud_restrictable"`
	UseLetsEncrypt                                    *bool    `access:"environment,write_restrictable,cloud_restrictable"`
	LetsEncryptCertificateCacheFile                   *string  `access:"environment,write_restrictable,cloud_restrictable"`
	Forward80To443                                    *bool    `access:"environment,write_restrictable,cloud_restrictable"`
	TrustedProxyIPHeader                              []string `access:"write_restrictable,cloud_restrictable"`
	ReadTimeout                                       *int     `access:"environment,write_restrictable,cloud_restrictable"`
	WriteTimeout                                      *int     `access:"environment,write_restrictable,cloud_restrictable"`
	IdleTimeout                                       *int     `access:"write_restrictable,cloud_restrictable"`
	MaximumLoginAttempts                              *int     `access:"authentication,write_restrictable,cloud_restrictable"`
	GoroutineHealthThreshold                          *int     `access:"write_restrictable,cloud_restrictable"`
	GoogleDeveloperKey                                *string  `access:"site,write_restrictable,cloud_restrictable"`
	EnableOAuthServiceProvider                        *bool    `access:"integrations"`
	EnableIncomingWebhooks                            *bool    `access:"integrations"`
	EnableOutgoingWebhooks                            *bool    `access:"integrations"`
	EnableCommands                                    *bool    `access:"integrations"`
	DEPRECATED_DO_NOT_USE_EnableOnlyAdminIntegrations *bool    `json:"EnableOnlyAdminIntegrations" mapstructure:"EnableOnlyAdminIntegrations"` // This field is deprecated and must not be used.
	EnablePostUsernameOverride                        *bool    `access:"integrations"`
	EnablePostIconOverride                            *bool    `access:"integrations"`
	EnableLinkPreviews                                *bool    `access:"site"`
	EnableTesting                                     *bool    `access:"environment,write_restrictable,cloud_restrictable"`
	EnableDeveloper                                   *bool    `access:"environment,write_restrictable,cloud_restrictable"`
	EnableOpenTracing                                 *bool    `access:"write_restrictable,cloud_restrictable"`
	EnableSecurityFixAlert                            *bool    `access:"environment,write_restrictable,cloud_restrictable"`
	EnableInsecureOutgoingConnections                 *bool    `access:"environment,write_restrictable,cloud_restrictable"`
	AllowedUntrustedInternalConnections               *string  `access:"environment,write_restrictable,cloud_restrictable"`
	EnableMultifactorAuthentication                   *bool    `access:"authentication"`
	EnforceMultifactorAuthentication                  *bool    `access:"authentication"`
	EnableUserAccessTokens                            *bool    `access:"integrations"`
	AllowCorsFrom                                     *string  `access:"integrations,write_restrictable,cloud_restrictable"`
	CorsExposedHeaders                                *string  `access:"integrations,write_restrictable,cloud_restrictable"`
	CorsAllowCredentials                              *bool    `access:"integrations,write_restrictable,cloud_restrictable"`
	CorsDebug                                         *bool    `access:"integrations,write_restrictable,cloud_restrictable"`
	AllowCookiesForSubdomains                         *bool    `access:"write_restrictable,cloud_restrictable"`
	ExtendSessionLengthWithActivity                   *bool    `access:"environment,write_restrictable,cloud_restrictable"`
	SessionLengthWebInDays                            *int     `access:"environment,write_restrictable,cloud_restrictable"`
	SessionLengthMobileInDays                         *int     `access:"environment,write_restrictable,cloud_restrictable"`
	SessionLengthSSOInDays                            *int     `access:"environment,write_restrictable,cloud_restrictable"`
	SessionCacheInMinutes                             *int     `access:"environment,write_restrictable,cloud_restrictable"`
	SessionIdleTimeoutInMinutes                       *int     `access:"environment,write_restrictable,cloud_restrictable"`
	WebsocketSecurePort                               *int     `access:"write_restrictable,cloud_restrictable"`
	WebsocketPort                                     *int     `access:"write_restrictable,cloud_restrictable"`
	WebserverMode                                     *string  `access:"environment,write_restrictable,cloud_restrictable"`
	EnableCustomEmoji                                 *bool    `access:"site"`
	EnableEmojiPicker                                 *bool    `access:"site"`
	EnableGifPicker                                   *bool    `access:"integrations"`
	GfycatApiKey                                      *string  `access:"integrations"`
	GfycatApiSecret                                   *string  `access:"integrations"`
	DEPRECATED_DO_NOT_USE_RestrictCustomEmojiCreation *string  `json:"RestrictCustomEmojiCreation" mapstructure:"RestrictCustomEmojiCreation"` // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_RestrictPostDelete          *string  `json:"RestrictPostDelete" mapstructure:"RestrictPostDelete"`                   // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_AllowEditPost               *string  `json:"AllowEditPost" mapstructure:"AllowEditPost"`                             // This field is deprecated and must not be used.
	PostEditTimeLimit                                 *int     `access:"user_management_permissions"`
	TimeBetweenUserTypingUpdatesMilliseconds          *int64   `access:"experimental,write_restrictable,cloud_restrictable"`
	EnablePostSearch                                  *bool    `access:"write_restrictable,cloud_restrictable"`
	MinimumHashtagLength                              *int     `access:"environment,write_restrictable,cloud_restrictable"`
	EnableUserTypingMessages                          *bool    `access:"experimental,write_restrictable,cloud_restrictable"`
	EnableChannelViewedMessages                       *bool    `access:"experimental,write_restrictable,cloud_restrictable"`
	EnableUserStatuses                                *bool    `access:"write_restrictable,cloud_restrictable"`
	ExperimentalEnableAuthenticationTransfer          *bool    `access:"experimental,write_restrictable,cloud_restrictable"`
	ClusterLogTimeoutMilliseconds                     *int     `access:"write_restrictable,cloud_restrictable"`
	CloseUnusedDirectMessages                         *bool    `access:"experimental"`
	EnablePreviewFeatures                             *bool    `access:"experimental"`
	EnableTutorial                                    *bool    `access:"experimental"`
	ExperimentalEnableDefaultChannelLeaveJoinMessages *bool    `access:"experimental"`
	ExperimentalGroupUnreadChannels                   *string  `access:"experimental"`
	ExperimentalChannelOrganization                   *bool    `access:"experimental"`
	ExperimentalChannelSidebarOrganization            *string  `access:"experimental"`
	ExperimentalDataPrefetch                          *bool    `access:"experimental"`
	DEPRECATED_DO_NOT_USE_ImageProxyType              *string  `json:"ImageProxyType" mapstructure:"ImageProxyType"`       // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_ImageProxyURL               *string  `json:"ImageProxyURL" mapstructure:"ImageProxyURL"`         // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_ImageProxyOptions           *string  `json:"ImageProxyOptions" mapstructure:"ImageProxyOptions"` // This field is deprecated and must not be used.
	EnableAPITeamDeletion                             *bool
	EnableAPIUserDeletion                             *bool
	ExperimentalEnableHardenedMode                    *bool `access:"experimental"`
	DisableLegacyMFA                                  *bool `access:"write_restrictable,cloud_restrictable"`
	ExperimentalStrictCSRFEnforcement                 *bool `access:"experimental,write_restrictable,cloud_restrictable"`
	EnableEmailInvitations                            *bool `access:"authentication"`
	DisableBotsWhenOwnerIsDeactivated                 *bool `access:"integrations,write_restrictable,cloud_restrictable"`
	EnableBotAccountCreation                          *bool `access:"integrations"`
	EnableSVGs                                        *bool `access:"site"`
	EnableLatex                                       *bool `access:"site"`
	EnableAPIChannelDeletion                          *bool
	EnableLocalMode                                   *bool
	LocalModeSocketLocation                           *string
	EnableAWSMetering                                 *bool
	SplitKey                                          *string `access:"environment,write_restrictable"`
	FeatureFlagSyncIntervalSeconds                    *int    `access:"environment,write_restrictable"`
	DebugSplit                                        *bool   `access:"environment,write_restrictable"`
	ThreadAutoFollow                                  *bool   `access:"experimental"`
	CollapsedThreads                                  *string `access:"experimental"`
	ManagedResourcePaths                              *string `access:"environment,write_restrictable,cloud_restrictable"`
}

func (s *ServiceSettings) SetDefaults(isUpdate bool) {
	if s.EnableEmailInvitations == nil {
		// If the site URL is also not present then assume this is a clean install
		if s.SiteURL == nil {
			s.EnableEmailInvitations = NewBool(false)
		} else {
			s.EnableEmailInvitations = NewBool(true)
		}
	}

	if s.SiteURL == nil {
		if s.EnableDeveloper != nil && *s.EnableDeveloper {
			s.SiteURL = NewString(SERVICE_SETTINGS_DEFAULT_SITE_URL)
		} else {
			s.SiteURL = NewString("")
		}
	}

	if s.WebsocketURL == nil {
		s.WebsocketURL = NewString("")
	}

	if s.LicenseFileLocation == nil {
		s.LicenseFileLocation = NewString("")
	}

	if s.ListenAddress == nil {
		s.ListenAddress = NewString(SERVICE_SETTINGS_DEFAULT_LISTEN_AND_ADDRESS)
	}

	if s.EnableLinkPreviews == nil {
		s.EnableLinkPreviews = NewBool(true)
	}

	if s.EnableTesting == nil {
		s.EnableTesting = NewBool(false)
	}

	if s.EnableDeveloper == nil {
		s.EnableDeveloper = NewBool(false)
	}

	if s.EnableOpenTracing == nil {
		s.EnableOpenTracing = NewBool(false)
	}

	if s.EnableSecurityFixAlert == nil {
		s.EnableSecurityFixAlert = NewBool(true)
	}

	if s.EnableInsecureOutgoingConnections == nil {
		s.EnableInsecureOutgoingConnections = NewBool(false)
	}

	if s.AllowedUntrustedInternalConnections == nil {
		s.AllowedUntrustedInternalConnections = NewString("")
	}

	if s.EnableMultifactorAuthentication == nil {
		s.EnableMultifactorAuthentication = NewBool(false)
	}

	if s.EnforceMultifactorAuthentication == nil {
		s.EnforceMultifactorAuthentication = NewBool(false)
	}

	if s.EnableUserAccessTokens == nil {
		s.EnableUserAccessTokens = NewBool(false)
	}

	if s.GoroutineHealthThreshold == nil {
		s.GoroutineHealthThreshold = NewInt(-1)
	}

	if s.GoogleDeveloperKey == nil {
		s.GoogleDeveloperKey = NewString("")
	}

	if s.EnableOAuthServiceProvider == nil {
		s.EnableOAuthServiceProvider = NewBool(false)
	}

	if s.EnableIncomingWebhooks == nil {
		s.EnableIncomingWebhooks = NewBool(true)
	}

	if s.EnableOutgoingWebhooks == nil {
		s.EnableOutgoingWebhooks = NewBool(true)
	}

	if s.ConnectionSecurity == nil {
		s.ConnectionSecurity = NewString("")
	}

	if s.TLSKeyFile == nil {
		s.TLSKeyFile = NewString(SERVICE_SETTINGS_DEFAULT_TLS_KEY_FILE)
	}

	if s.TLSCertFile == nil {
		s.TLSCertFile = NewString(SERVICE_SETTINGS_DEFAULT_TLS_CERT_FILE)
	}

	if s.TLSMinVer == nil {
		s.TLSMinVer = NewString("1.2")
	}

	if s.TLSStrictTransport == nil {
		s.TLSStrictTransport = NewBool(false)
	}

	if s.TLSStrictTransportMaxAge == nil {
		s.TLSStrictTransportMaxAge = NewInt64(63072000)
	}

	if s.TLSOverwriteCiphers == nil {
		s.TLSOverwriteCiphers = []string{}
	}

	if s.UseLetsEncrypt == nil {
		s.UseLetsEncrypt = NewBool(false)
	}

	if s.LetsEncryptCertificateCacheFile == nil {
		s.LetsEncryptCertificateCacheFile = NewString("./config/letsencrypt.cache")
	}

	if s.ReadTimeout == nil {
		s.ReadTimeout = NewInt(SERVICE_SETTINGS_DEFAULT_READ_TIMEOUT)
	}

	if s.WriteTimeout == nil {
		s.WriteTimeout = NewInt(SERVICE_SETTINGS_DEFAULT_WRITE_TIMEOUT)
	}

	if s.IdleTimeout == nil {
		s.IdleTimeout = NewInt(SERVICE_SETTINGS_DEFAULT_IDLE_TIMEOUT)
	}

	if s.MaximumLoginAttempts == nil {
		s.MaximumLoginAttempts = NewInt(SERVICE_SETTINGS_DEFAULT_MAX_LOGIN_ATTEMPTS)
	}

	if s.Forward80To443 == nil {
		s.Forward80To443 = NewBool(false)
	}

	if isUpdate {
		// When updating an existing configuration, ensure that defaults are set.
		if s.TrustedProxyIPHeader == nil {
			s.TrustedProxyIPHeader = []string{HEADER_FORWARDED, HEADER_REAL_IP}
		}
	} else {
		// When generating a blank configuration, leave the list empty.
		s.TrustedProxyIPHeader = []string{}
	}

	if s.TimeBetweenUserTypingUpdatesMilliseconds == nil {
		s.TimeBetweenUserTypingUpdatesMilliseconds = NewInt64(5000)
	}

	if s.EnablePostSearch == nil {
		s.EnablePostSearch = NewBool(true)
	}

	if s.MinimumHashtagLength == nil {
		s.MinimumHashtagLength = NewInt(3)
	}

	if s.EnableUserTypingMessages == nil {
		s.EnableUserTypingMessages = NewBool(true)
	}

	if s.EnableChannelViewedMessages == nil {
		s.EnableChannelViewedMessages = NewBool(true)
	}

	if s.EnableUserStatuses == nil {
		s.EnableUserStatuses = NewBool(true)
	}

	if s.ClusterLogTimeoutMilliseconds == nil {
		s.ClusterLogTimeoutMilliseconds = NewInt(2000)
	}

	if s.CloseUnusedDirectMessages == nil {
		s.CloseUnusedDirectMessages = NewBool(false)
	}

	if s.EnableTutorial == nil {
		s.EnableTutorial = NewBool(true)
	}

	// Must be manually enabled for existing installations.
	if s.ExtendSessionLengthWithActivity == nil {
		s.ExtendSessionLengthWithActivity = NewBool(!isUpdate)
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

	if s.SessionCacheInMinutes == nil {
		s.SessionCacheInMinutes = NewInt(10)
	}

	if s.SessionIdleTimeoutInMinutes == nil {
		s.SessionIdleTimeoutInMinutes = NewInt(43200)
	}

	if s.EnableCommands == nil {
		s.EnableCommands = NewBool(true)
	}

	if s.DEPRECATED_DO_NOT_USE_EnableOnlyAdminIntegrations == nil {
		s.DEPRECATED_DO_NOT_USE_EnableOnlyAdminIntegrations = NewBool(true)
	}

	if s.EnablePostUsernameOverride == nil {
		s.EnablePostUsernameOverride = NewBool(false)
	}

	if s.EnablePostIconOverride == nil {
		s.EnablePostIconOverride = NewBool(false)
	}

	if s.WebsocketPort == nil {
		s.WebsocketPort = NewInt(80)
	}

	if s.WebsocketSecurePort == nil {
		s.WebsocketSecurePort = NewInt(443)
	}

	if s.AllowCorsFrom == nil {
		s.AllowCorsFrom = NewString(SERVICE_SETTINGS_DEFAULT_ALLOW_CORS_FROM)
	}

	if s.CorsExposedHeaders == nil {
		s.CorsExposedHeaders = NewString("")
	}

	if s.CorsAllowCredentials == nil {
		s.CorsAllowCredentials = NewBool(false)
	}

	if s.CorsDebug == nil {
		s.CorsDebug = NewBool(false)
	}

	if s.AllowCookiesForSubdomains == nil {
		s.AllowCookiesForSubdomains = NewBool(false)
	}

	if s.WebserverMode == nil {
		s.WebserverMode = NewString("gzip")
	} else if *s.WebserverMode == "regular" {
		*s.WebserverMode = "gzip"
	}

	if s.EnableCustomEmoji == nil {
		s.EnableCustomEmoji = NewBool(true)
	}

	if s.EnableEmojiPicker == nil {
		s.EnableEmojiPicker = NewBool(true)
	}

	if s.EnableGifPicker == nil {
		s.EnableGifPicker = NewBool(true)
	}

	if s.GfycatApiKey == nil || *s.GfycatApiKey == "" {
		s.GfycatApiKey = NewString(SERVICE_SETTINGS_DEFAULT_GFYCAT_API_KEY)
	}

	if s.GfycatApiSecret == nil || *s.GfycatApiSecret == "" {
		s.GfycatApiSecret = NewString(SERVICE_SETTINGS_DEFAULT_GFYCAT_API_SECRET)
	}

	if s.DEPRECATED_DO_NOT_USE_RestrictCustomEmojiCreation == nil {
		s.DEPRECATED_DO_NOT_USE_RestrictCustomEmojiCreation = NewString(RESTRICT_EMOJI_CREATION_ALL)
	}

	if s.DEPRECATED_DO_NOT_USE_RestrictPostDelete == nil {
		s.DEPRECATED_DO_NOT_USE_RestrictPostDelete = NewString(PERMISSIONS_DELETE_POST_ALL)
	}

	if s.DEPRECATED_DO_NOT_USE_AllowEditPost == nil {
		s.DEPRECATED_DO_NOT_USE_AllowEditPost = NewString(ALLOW_EDIT_POST_ALWAYS)
	}

	if s.ExperimentalEnableAuthenticationTransfer == nil {
		s.ExperimentalEnableAuthenticationTransfer = NewBool(true)
	}

	if s.PostEditTimeLimit == nil {
		s.PostEditTimeLimit = NewInt(-1)
	}

	if s.EnablePreviewFeatures == nil {
		s.EnablePreviewFeatures = NewBool(true)
	}

	if s.ExperimentalEnableDefaultChannelLeaveJoinMessages == nil {
		s.ExperimentalEnableDefaultChannelLeaveJoinMessages = NewBool(true)
	}

	if s.ExperimentalGroupUnreadChannels == nil {
		s.ExperimentalGroupUnreadChannels = NewString(GROUP_UNREAD_CHANNELS_DISABLED)
	} else if *s.ExperimentalGroupUnreadChannels == "0" {
		s.ExperimentalGroupUnreadChannels = NewString(GROUP_UNREAD_CHANNELS_DISABLED)
	} else if *s.ExperimentalGroupUnreadChannels == "1" {
		s.ExperimentalGroupUnreadChannels = NewString(GROUP_UNREAD_CHANNELS_DEFAULT_ON)
	}

	if s.ExperimentalChannelOrganization == nil {
		experimentalUnreadEnabled := *s.ExperimentalGroupUnreadChannels != GROUP_UNREAD_CHANNELS_DISABLED
		s.ExperimentalChannelOrganization = NewBool(experimentalUnreadEnabled)
	}

	if s.ExperimentalChannelSidebarOrganization == nil {
		s.ExperimentalChannelSidebarOrganization = NewString("disabled")
	}

	if s.ExperimentalDataPrefetch == nil {
		s.ExperimentalDataPrefetch = NewBool(true)
	}

	if s.DEPRECATED_DO_NOT_USE_ImageProxyType == nil {
		s.DEPRECATED_DO_NOT_USE_ImageProxyType = NewString("")
	}

	if s.DEPRECATED_DO_NOT_USE_ImageProxyURL == nil {
		s.DEPRECATED_DO_NOT_USE_ImageProxyURL = NewString("")
	}

	if s.DEPRECATED_DO_NOT_USE_ImageProxyOptions == nil {
		s.DEPRECATED_DO_NOT_USE_ImageProxyOptions = NewString("")
	}

	if s.EnableAPITeamDeletion == nil {
		s.EnableAPITeamDeletion = NewBool(false)
	}

	if s.EnableAPIUserDeletion == nil {
		s.EnableAPIUserDeletion = NewBool(false)
	}

	if s.EnableAPIChannelDeletion == nil {
		s.EnableAPIChannelDeletion = NewBool(false)
	}

	if s.ExperimentalEnableHardenedMode == nil {
		s.ExperimentalEnableHardenedMode = NewBool(false)
	}

	if s.DisableLegacyMFA == nil {
		s.DisableLegacyMFA = NewBool(!isUpdate)
	}

	if s.ExperimentalStrictCSRFEnforcement == nil {
		s.ExperimentalStrictCSRFEnforcement = NewBool(false)
	}

	if s.DisableBotsWhenOwnerIsDeactivated == nil {
		s.DisableBotsWhenOwnerIsDeactivated = NewBool(true)
	}

	if s.EnableBotAccountCreation == nil {
		s.EnableBotAccountCreation = NewBool(false)
	}

	if s.EnableSVGs == nil {
		if isUpdate {
			s.EnableSVGs = NewBool(true)
		} else {
			s.EnableSVGs = NewBool(false)
		}
	}

	if s.EnableLatex == nil {
		if isUpdate {
			s.EnableLatex = NewBool(true)
		} else {
			s.EnableLatex = NewBool(false)
		}
	}

	if s.EnableLocalMode == nil {
		s.EnableLocalMode = NewBool(false)
	}

	if s.LocalModeSocketLocation == nil {
		s.LocalModeSocketLocation = NewString(LOCAL_MODE_SOCKET_PATH)
	}

	if s.EnableAWSMetering == nil {
		s.EnableAWSMetering = NewBool(false)
	}

	if s.SplitKey == nil {
		s.SplitKey = NewString("")
	}

	if s.FeatureFlagSyncIntervalSeconds == nil {
		s.FeatureFlagSyncIntervalSeconds = NewInt(30)
	}

	if s.DebugSplit == nil {
		s.DebugSplit = NewBool(false)
	}

	if s.ThreadAutoFollow == nil {
		s.ThreadAutoFollow = NewBool(true)
	}

	if s.CollapsedThreads == nil {
		s.CollapsedThreads = NewString(COLLAPSED_THREADS_DISABLED)
	}

	if s.ManagedResourcePaths == nil {
		s.ManagedResourcePaths = NewString("")
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

type SamlSettings struct {
	// Basic
	Enable                        *bool `access:"authentication"`
	EnableSyncWithLdap            *bool `access:"authentication"`
	EnableSyncWithLdapIncludeAuth *bool `access:"authentication"`
	IgnoreGuestsLdapSync          *bool `access:"authentication"`

	Verify      *bool `access:"authentication"`
	Encrypt     *bool `access:"authentication"`
	SignRequest *bool `access:"authentication"`

	IdpUrl                      *string `access:"authentication"`
	IdpDescriptorUrl            *string `access:"authentication"`
	IdpMetadataUrl              *string `access:"authentication"`
	ServiceProviderIdentifier   *string `access:"authentication"`
	AssertionConsumerServiceURL *string `access:"authentication"`

	SignatureAlgorithm *string `access:"authentication"`
	CanonicalAlgorithm *string `access:"authentication"`

	ScopingIDPProviderId *string `access:"authentication"`
	ScopingIDPName       *string `access:"authentication"`

	IdpCertificateFile    *string `access:"authentication"`
	PublicCertificateFile *string `access:"authentication"`
	PrivateKeyFile        *string `access:"authentication"`

	// User Mapping
	IdAttribute          *string `access:"authentication"`
	GuestAttribute       *string `access:"authentication"`
	EnableAdminAttribute *bool
	AdminAttribute       *string
	FirstNameAttribute   *string `access:"authentication"`
	LastNameAttribute    *string `access:"authentication"`
	EmailAttribute       *string `access:"authentication"`
	UsernameAttribute    *string `access:"authentication"`
	NicknameAttribute    *string `access:"authentication"`
	LocaleAttribute      *string `access:"authentication"`
	PositionAttribute    *string `access:"authentication"`

	LoginButtonText *string `access:"authentication"`

	LoginButtonColor       *string `access:"authentication"`
	LoginButtonBorderColor *string `access:"authentication"`
	LoginButtonTextColor   *string `access:"authentication"`
}

func (s *SamlSettings) SetDefaults() {
	if s.Enable == nil {
		s.Enable = NewBool(false)
	}

	if s.EnableSyncWithLdap == nil {
		s.EnableSyncWithLdap = NewBool(false)
	}

	if s.EnableSyncWithLdapIncludeAuth == nil {
		s.EnableSyncWithLdapIncludeAuth = NewBool(false)
	}

	if s.IgnoreGuestsLdapSync == nil {
		s.IgnoreGuestsLdapSync = NewBool(false)
	}

	if s.EnableAdminAttribute == nil {
		s.EnableAdminAttribute = NewBool(false)
	}

	if s.Verify == nil {
		s.Verify = NewBool(true)
	}

	if s.Encrypt == nil {
		s.Encrypt = NewBool(true)
	}

	if s.SignRequest == nil {
		s.SignRequest = NewBool(false)
	}

	if s.SignatureAlgorithm == nil {
		s.SignatureAlgorithm = NewString(SAML_SETTINGS_DEFAULT_SIGNATURE_ALGORITHM)
	}

	if s.CanonicalAlgorithm == nil {
		s.CanonicalAlgorithm = NewString(SAML_SETTINGS_DEFAULT_CANONICAL_ALGORITHM)
	}

	if s.IdpUrl == nil {
		s.IdpUrl = NewString("")
	}

	if s.IdpDescriptorUrl == nil {
		s.IdpDescriptorUrl = NewString("")
	}

	if s.ServiceProviderIdentifier == nil {
		if s.IdpDescriptorUrl != nil {
			s.ServiceProviderIdentifier = NewString(*s.IdpDescriptorUrl)
		} else {
			s.ServiceProviderIdentifier = NewString("")
		}
	}

	if s.IdpMetadataUrl == nil {
		s.IdpMetadataUrl = NewString("")
	}

	if s.IdpCertificateFile == nil {
		s.IdpCertificateFile = NewString("")
	}

	if s.PublicCertificateFile == nil {
		s.PublicCertificateFile = NewString("")
	}

	if s.PrivateKeyFile == nil {
		s.PrivateKeyFile = NewString("")
	}

	if s.AssertionConsumerServiceURL == nil {
		s.AssertionConsumerServiceURL = NewString("")
	}

	if s.ScopingIDPProviderId == nil {
		s.ScopingIDPProviderId = NewString("")
	}

	if s.ScopingIDPName == nil {
		s.ScopingIDPName = NewString("")
	}

	if s.LoginButtonText == nil || *s.LoginButtonText == "" {
		s.LoginButtonText = NewString(USER_AUTH_SERVICE_SAML_TEXT)
	}

	if s.IdAttribute == nil {
		s.IdAttribute = NewString(SAML_SETTINGS_DEFAULT_ID_ATTRIBUTE)
	}

	if s.GuestAttribute == nil {
		s.GuestAttribute = NewString(SAML_SETTINGS_DEFAULT_GUEST_ATTRIBUTE)
	}
	if s.AdminAttribute == nil {
		s.AdminAttribute = NewString(SAML_SETTINGS_DEFAULT_ADMIN_ATTRIBUTE)
	}
	if s.FirstNameAttribute == nil {
		s.FirstNameAttribute = NewString(SAML_SETTINGS_DEFAULT_FIRST_NAME_ATTRIBUTE)
	}

	if s.LastNameAttribute == nil {
		s.LastNameAttribute = NewString(SAML_SETTINGS_DEFAULT_LAST_NAME_ATTRIBUTE)
	}

	if s.EmailAttribute == nil {
		s.EmailAttribute = NewString(SAML_SETTINGS_DEFAULT_EMAIL_ATTRIBUTE)
	}

	if s.UsernameAttribute == nil {
		s.UsernameAttribute = NewString(SAML_SETTINGS_DEFAULT_USERNAME_ATTRIBUTE)
	}

	if s.NicknameAttribute == nil {
		s.NicknameAttribute = NewString(SAML_SETTINGS_DEFAULT_NICKNAME_ATTRIBUTE)
	}

	if s.PositionAttribute == nil {
		s.PositionAttribute = NewString(SAML_SETTINGS_DEFAULT_POSITION_ATTRIBUTE)
	}

	if s.LocaleAttribute == nil {
		s.LocaleAttribute = NewString(SAML_SETTINGS_DEFAULT_LOCALE_ATTRIBUTE)
	}

	if s.LoginButtonColor == nil {
		s.LoginButtonColor = NewString("#34a28b")
	}

	if s.LoginButtonBorderColor == nil {
		s.LoginButtonBorderColor = NewString("#2389D7")
	}

	if s.LoginButtonTextColor == nil {
		s.LoginButtonTextColor = NewString("#ffffff")
	}
}

type SqlSettings struct {
	DriverName               *string
	DataSource               *string `restricted:"true"`
	DataSourceReplicas       []string
	DataSourceSearchReplicas []string `access:"environment,write_restrictable,cloud_restrictable"`
	QueryTimeout             *int     `restricted:"true"`
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

	if s.DataSourceSearchReplicas == nil {
		s.DataSourceSearchReplicas = []string{}
	}

	if s.QueryTimeout == nil {
		s.QueryTimeout = NewInt(30)
	}
}

type FileSettings struct {
	EnableFileAttachments *bool   `access:"site,cloud_restrictable"`
	EnableMobileUpload    *bool   `access:"site,cloud_restrictable"`
	EnableMobileDownload  *bool   `access:"site,cloud_restrictable"`
	MaxFileSize           *int64  `access:"environment,cloud_restrictable"`
	Directory             *string `restricted:"true"`
	DriverName            *string `restricted:"true"`
	EnablePublicLink      *bool   `access:"site,cloud_restrictable"`
	PublicLinkSalt        *string `access:"site,cloud_restrictable"`
	InitialFont           *string `access:"environment,cloud_restrictable"`
	S3AccessKeyId         *string `restricted:"true"`
	S3SecretAccessKey     *string `restricted:"true"`
	S3Bucket              *string `restricted:"true"`
	S3PathPrefix          *string `restricted:"true"`
	S3Region              *string `restricted:"true"`
	S3Endpoint            *string `restricted:"true"`
	S3SSL                 *bool   `restricted:"true"`
	S3SignV2              *bool   `restricted:"true"`
	S3SSE                 *bool   `restricted:"true"`
	S3Trace               *bool   `restricted:"true"`
}

func (s *FileSettings) SetDefaults(isUpdate bool) {
	if s.EnableFileAttachments == nil {
		s.EnableFileAttachments = NewBool(true)
	}

	if s.EnableMobileUpload == nil {
		s.EnableMobileUpload = NewBool(true)
	}

	if s.EnableMobileDownload == nil {
		s.EnableMobileDownload = NewBool(true)
	}

	if s.MaxFileSize == nil {
		s.MaxFileSize = NewInt64(52428800) // 50 MB
	}

	if s.DriverName == nil {
		s.DriverName = NewString(IMAGE_DRIVER_LOCAL)
	}

	if s.Directory == nil || *s.Directory == "" {
		s.Directory = NewString(FILE_SETTINGS_DEFAULT_DIRECTORY)
	}

	if s.EnablePublicLink == nil {
		s.EnablePublicLink = NewBool(false)
	}

	if isUpdate {
		// When updating an existing configuration, ensure link salt has been specified.
		if s.PublicLinkSalt == nil || len(*s.PublicLinkSalt) == 0 {
			s.PublicLinkSalt = NewString(NewRandomString(32))
		}
	} else {
		// When generating a blank configuration, leave link salt empty to be generated on server start.
		s.PublicLinkSalt = NewString("")
	}

	if s.InitialFont == nil {
		// Defaults to "nunito-bold.ttf"
		s.InitialFont = NewString("nunito-bold.ttf")
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
	SiteName                                                  *string  `access:"site"`
	MaxUsersPerTeam                                           *int     `access:"site"`
	DEPRECATED_DO_NOT_USE_EnableTeamCreation                  *bool    `json:"EnableTeamCreation" mapstructure:"EnableTeamCreation"` // This field is deprecated and must not be used.
	EnableUserCreation                                        *bool    `access:"authentication"`
	EnableOpenServer                                          *bool    `access:"authentication"`
	EnableUserDeactivation                                    *bool    `access:"experimental"`
	RestrictCreationToDomains                                 *string  `access:"authentication"`
	EnableCustomBrand                                         *bool    `access:"site"`
	CustomBrandText                                           *string  `access:"site"`
	CustomDescriptionText                                     *string  `access:"site"`
	RestrictDirectMessage                                     *string  `access:"site"`
	DEPRECATED_DO_NOT_USE_RestrictTeamInvite                  *string  `json:"RestrictTeamInvite" mapstructure:"RestrictTeamInvite"`                                   // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_RestrictPublicChannelManagement     *string  `json:"RestrictPublicChannelManagement" mapstructure:"RestrictPublicChannelManagement"`         // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManagement    *string  `json:"RestrictPrivateChannelManagement" mapstructure:"RestrictPrivateChannelManagement"`       // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_RestrictPublicChannelCreation       *string  `json:"RestrictPublicChannelCreation" mapstructure:"RestrictPublicChannelCreation"`             // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_RestrictPrivateChannelCreation      *string  `json:"RestrictPrivateChannelCreation" mapstructure:"RestrictPrivateChannelCreation"`           // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_RestrictPublicChannelDeletion       *string  `json:"RestrictPublicChannelDeletion" mapstructure:"RestrictPublicChannelDeletion"`             // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_RestrictPrivateChannelDeletion      *string  `json:"RestrictPrivateChannelDeletion" mapstructure:"RestrictPrivateChannelDeletion"`           // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManageMembers *string  `json:"RestrictPrivateChannelManageMembers" mapstructure:"RestrictPrivateChannelManageMembers"` // This field is deprecated and must not be used.
	EnableXToLeaveChannelsFromLHS                             *bool    `access:"experimental"`
	UserStatusAwayTimeout                                     *int64   `access:"experimental"`
	MaxChannelsPerTeam                                        *int64   `access:"site"`
	MaxNotificationsPerChannel                                *int64   `access:"environment"`
	EnableConfirmNotificationsToChannel                       *bool    `access:"site"`
	TeammateNameDisplay                                       *string  `access:"site"`
	ExperimentalViewArchivedChannels                          *bool    `access:"experimental,site"`
	ExperimentalEnableAutomaticReplies                        *bool    `access:"experimental"`
	ExperimentalHideTownSquareinLHS                           *bool    `access:"experimental"`
	ExperimentalTownSquareIsReadOnly                          *bool    `access:"experimental"`
	LockTeammateNameDisplay                                   *bool    `access:"site"`
	ExperimentalPrimaryTeam                                   *string  `access:"experimental"`
	ExperimentalDefaultChannels                               []string `access:"experimental"`
}

func (s *TeamSettings) SetDefaults() {

	if s.SiteName == nil || *s.SiteName == "" {
		s.SiteName = NewString(TEAM_SETTINGS_DEFAULT_SITE_NAME)
	}

	if s.MaxUsersPerTeam == nil {
		s.MaxUsersPerTeam = NewInt(TEAM_SETTINGS_DEFAULT_MAX_USERS_PER_TEAM)
	}

	if s.DEPRECATED_DO_NOT_USE_EnableTeamCreation == nil {
		s.DEPRECATED_DO_NOT_USE_EnableTeamCreation = NewBool(true)
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

	if s.DEPRECATED_DO_NOT_USE_RestrictTeamInvite == nil {
		s.DEPRECATED_DO_NOT_USE_RestrictTeamInvite = NewString(PERMISSIONS_ALL)
	}

	if s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelManagement == nil {
		s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelManagement = NewString(PERMISSIONS_ALL)
	}

	if s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManagement == nil {
		s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManagement = NewString(PERMISSIONS_ALL)
	}

	if s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelCreation == nil {
		s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelCreation = new(string)
		// If this setting does not exist, assume migration from <3.6, so use management setting as default.
		if *s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelManagement == PERMISSIONS_CHANNEL_ADMIN {
			*s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelCreation = PERMISSIONS_TEAM_ADMIN
		} else {
			*s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelCreation = *s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelManagement
		}
	}

	if s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelCreation == nil {
		// If this setting does not exist, assume migration from <3.6, so use management setting as default.
		if *s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManagement == PERMISSIONS_CHANNEL_ADMIN {
			s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelCreation = NewString(PERMISSIONS_TEAM_ADMIN)
		} else {
			s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelCreation = NewString(*s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManagement)
		}
	}

	if s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelDeletion == nil {
		// If this setting does not exist, assume migration from <3.6, so use management setting as default.
		s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelDeletion = NewString(*s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelManagement)
	}

	if s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelDeletion == nil {
		// If this setting does not exist, assume migration from <3.6, so use management setting as default.
		s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelDeletion = NewString(*s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManagement)
	}

	if s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManageMembers == nil {
		s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManageMembers = NewString(PERMISSIONS_ALL)
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

	if s.DEPRECATED_DO_NOT_USE_EnableTeamCreation == nil {
		s.DEPRECATED_DO_NOT_USE_EnableTeamCreation = NewBool(true)
	}

	if s.EnableUserCreation == nil {
		s.EnableUserCreation = NewBool(true)
	}

	if s.ExperimentalViewArchivedChannels == nil {
		s.ExperimentalViewArchivedChannels = NewBool(true)
	}

	if s.LockTeammateNameDisplay == nil {
		s.LockTeammateNameDisplay = NewBool(false)
	}
}

type ClientRequirements struct {
	AndroidLatestVersion string `access:"write_restrictable,cloud_restrictable"`
	AndroidMinVersion    string `access:"write_restrictable,cloud_restrictable"`
	DesktopLatestVersion string `access:"write_restrictable,cloud_restrictable"`
	DesktopMinVersion    string `access:"write_restrictable,cloud_restrictable"`
	IosLatestVersion     string `access:"write_restrictable,cloud_restrictable"`
	IosMinVersion        string `access:"write_restrictable,cloud_restrictable"`
}

type LdapSettings struct {
	// Basic
	Enable             *bool   `access:"authentication"`
	EnableSync         *bool   `access:"authentication"`
	LdapServer         *string `access:"authentication"`
	LdapPort           *int    `access:"authentication"`
	ConnectionSecurity *string `access:"authentication"`
	BaseDN             *string `access:"authentication"`
	BindUsername       *string `access:"authentication"`
	BindPassword       *string `access:"authentication"`

	// Filtering
	UserFilter        *string `access:"authentication"`
	GroupFilter       *string `access:"authentication"`
	GuestFilter       *string `access:"authentication"`
	EnableAdminFilter *bool
	AdminFilter       *string

	// Group Mapping
	GroupDisplayNameAttribute *string `access:"authentication"`
	GroupIdAttribute          *string `access:"authentication"`

	// User Mapping
	FirstNameAttribute *string `access:"authentication"`
	LastNameAttribute  *string `access:"authentication"`
	EmailAttribute     *string `access:"authentication"`
	UsernameAttribute  *string `access:"authentication"`
	NicknameAttribute  *string `access:"authentication"`
	IdAttribute        *string `access:"authentication"`
	PositionAttribute  *string `access:"authentication"`
	LoginIdAttribute   *string `access:"authentication"`
	PictureAttribute   *string `access:"authentication"`

	// Synchronization
	SyncIntervalMinutes *int `access:"authentication"`

	// Advanced
	SkipCertificateVerification *bool   `access:"authentication"`
	PublicCertificateFile       *string `access:"authentication"`
	PrivateKeyFile              *string `access:"authentication"`
	QueryTimeout                *int    `access:"authentication"`
	MaxPageSize                 *int    `access:"authentication"`

	// Customization
	LoginFieldName *string `access:"authentication"`

	LoginButtonColor       *string `access:"authentication"`
	LoginButtonBorderColor *string `access:"authentication"`
	LoginButtonTextColor   *string `access:"authentication"`

	Trace *bool `access:"authentication"`
}

func (s *LdapSettings) SetDefaults() {
	if s.Enable == nil {
		s.Enable = NewBool(false)
	}

	// When unset should default to LDAP Enabled
	if s.EnableSync == nil {
		s.EnableSync = NewBool(*s.Enable)
	}

	if s.EnableAdminFilter == nil {
		s.EnableAdminFilter = NewBool(false)
	}

	if s.LdapServer == nil {
		s.LdapServer = NewString("")
	}

	if s.LdapPort == nil {
		s.LdapPort = NewInt(389)
	}

	if s.ConnectionSecurity == nil {
		s.ConnectionSecurity = NewString("")
	}

	if s.PublicCertificateFile == nil {
		s.PublicCertificateFile = NewString("")
	}

	if s.PrivateKeyFile == nil {
		s.PrivateKeyFile = NewString("")
	}

	if s.BaseDN == nil {
		s.BaseDN = NewString("")
	}

	if s.BindUsername == nil {
		s.BindUsername = NewString("")
	}

	if s.BindPassword == nil {
		s.BindPassword = NewString("")
	}

	if s.UserFilter == nil {
		s.UserFilter = NewString("")
	}

	if s.GuestFilter == nil {
		s.GuestFilter = NewString("")
	}

	if s.AdminFilter == nil {
		s.AdminFilter = NewString("")
	}

	if s.GroupFilter == nil {
		s.GroupFilter = NewString("")
	}

	if s.GroupDisplayNameAttribute == nil {
		s.GroupDisplayNameAttribute = NewString(LDAP_SETTINGS_DEFAULT_GROUP_DISPLAY_NAME_ATTRIBUTE)
	}

	if s.GroupIdAttribute == nil {
		s.GroupIdAttribute = NewString(LDAP_SETTINGS_DEFAULT_GROUP_ID_ATTRIBUTE)
	}

	if s.FirstNameAttribute == nil {
		s.FirstNameAttribute = NewString(LDAP_SETTINGS_DEFAULT_FIRST_NAME_ATTRIBUTE)
	}

	if s.LastNameAttribute == nil {
		s.LastNameAttribute = NewString(LDAP_SETTINGS_DEFAULT_LAST_NAME_ATTRIBUTE)
	}

	if s.EmailAttribute == nil {
		s.EmailAttribute = NewString(LDAP_SETTINGS_DEFAULT_EMAIL_ATTRIBUTE)
	}

	if s.UsernameAttribute == nil {
		s.UsernameAttribute = NewString(LDAP_SETTINGS_DEFAULT_USERNAME_ATTRIBUTE)
	}

	if s.NicknameAttribute == nil {
		s.NicknameAttribute = NewString(LDAP_SETTINGS_DEFAULT_NICKNAME_ATTRIBUTE)
	}

	if s.IdAttribute == nil {
		s.IdAttribute = NewString(LDAP_SETTINGS_DEFAULT_ID_ATTRIBUTE)
	}

	if s.PositionAttribute == nil {
		s.PositionAttribute = NewString(LDAP_SETTINGS_DEFAULT_POSITION_ATTRIBUTE)
	}

	if s.PictureAttribute == nil {
		s.PictureAttribute = NewString(LDAP_SETTINGS_DEFAULT_PICTURE_ATTRIBUTE)
	}

	// For those upgrading to the version when LoginIdAttribute was added
	// they need IdAttribute == LoginIdAttribute not to break
	if s.LoginIdAttribute == nil {
		s.LoginIdAttribute = s.IdAttribute
	}

	if s.SyncIntervalMinutes == nil {
		s.SyncIntervalMinutes = NewInt(60)
	}

	if s.SkipCertificateVerification == nil {
		s.SkipCertificateVerification = NewBool(false)
	}

	if s.QueryTimeout == nil {
		s.QueryTimeout = NewInt(60)
	}

	if s.MaxPageSize == nil {
		s.MaxPageSize = NewInt(0)
	}

	if s.LoginFieldName == nil {
		s.LoginFieldName = NewString(LDAP_SETTINGS_DEFAULT_LOGIN_FIELD_NAME)
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

	if s.Trace == nil {
		s.Trace = NewBool(false)
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
	if o.LdapSettings.BindPassword != nil && len(*o.LdapSettings.BindPassword) > 0 {
		*o.LdapSettings.BindPassword = FAKE_SETTING
	}

	*o.FileSettings.PublicLinkSalt = FAKE_SETTING

	if len(*o.FileSettings.S3SecretAccessKey) > 0 {
		*o.FileSettings.S3SecretAccessKey = FAKE_SETTING
	}

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

type SupportSettings struct {
	TermsOfServiceLink                     *string `access:"site,write_restrictable"`
	PrivacyPolicyLink                      *string `access:"site,write_restrictable"`
	AboutLink                              *string `access:"site,write_restrictable"`
	HelpLink                               *string `access:"site,write_restrictable"`
	ReportAProblemLink                     *string `access:"site,write_restrictable"`
	SupportEmail                           *string `access:"site"`
	CustomTermsOfServiceEnabled            *bool   `access:"compliance"`
	CustomTermsOfServiceReAcceptancePeriod *int    `access:"compliance"`
	EnableAskCommunityLink                 *bool   `access:"site"`
}

func (s *SupportSettings) SetDefaults() {
	if !IsSafeLink(s.TermsOfServiceLink) {
		*s.TermsOfServiceLink = SUPPORT_SETTINGS_DEFAULT_TERMS_OF_SERVICE_LINK
	}

	if s.TermsOfServiceLink == nil {
		s.TermsOfServiceLink = NewString(SUPPORT_SETTINGS_DEFAULT_TERMS_OF_SERVICE_LINK)
	}

	if !IsSafeLink(s.PrivacyPolicyLink) {
		*s.PrivacyPolicyLink = ""
	}

	if s.PrivacyPolicyLink == nil {
		s.PrivacyPolicyLink = NewString(SUPPORT_SETTINGS_DEFAULT_PRIVACY_POLICY_LINK)
	}

	if !IsSafeLink(s.AboutLink) {
		*s.AboutLink = ""
	}

	if s.AboutLink == nil {
		s.AboutLink = NewString(SUPPORT_SETTINGS_DEFAULT_ABOUT_LINK)
	}

	if !IsSafeLink(s.HelpLink) {
		*s.HelpLink = ""
	}

	if s.HelpLink == nil {
		s.HelpLink = NewString(SUPPORT_SETTINGS_DEFAULT_HELP_LINK)
	}

	if !IsSafeLink(s.ReportAProblemLink) {
		*s.ReportAProblemLink = ""
	}

	if s.ReportAProblemLink == nil {
		s.ReportAProblemLink = NewString(SUPPORT_SETTINGS_DEFAULT_REPORT_A_PROBLEM_LINK)
	}

	if s.SupportEmail == nil {
		s.SupportEmail = NewString(SUPPORT_SETTINGS_DEFAULT_SUPPORT_EMAIL)
	}

	if s.CustomTermsOfServiceEnabled == nil {
		s.CustomTermsOfServiceEnabled = NewBool(false)
	}

	if s.CustomTermsOfServiceReAcceptancePeriod == nil {
		s.CustomTermsOfServiceReAcceptancePeriod = NewInt(SUPPORT_SETTINGS_DEFAULT_RE_ACCEPTANCE_PERIOD)
	}

	if s.EnableAskCommunityLink == nil {
		s.EnableAskCommunityLink = NewBool(true)
	}
}

type RateLimitSettings struct {
	Enable           *bool  `access:"environment,write_restrictable"`
	PerSec           *int   `access:"environment,write_restrictable"`
	MaxBurst         *int   `access:"environment,write_restrictable"`
	MemoryStoreSize  *int   `access:"environment,write_restrictable"`
	VaryByRemoteAddr *bool  `access:"environment,write_restrictable"`
	VaryByUser       *bool  `access:"environment,write_restrictable"`
	VaryByHeader     string `access:"environment,write_restrictable"`
}

func (s *RateLimitSettings) SetDefaults() {
	if s.Enable == nil {
		s.Enable = NewBool(false)
	}

	if s.PerSec == nil {
		s.PerSec = NewInt(10)
	}

	if s.MaxBurst == nil {
		s.MaxBurst = NewInt(100)
	}

	if s.MemoryStoreSize == nil {
		s.MemoryStoreSize = NewInt(10000)
	}

	if s.VaryByRemoteAddr == nil {
		s.VaryByRemoteAddr = NewBool(true)
	}

	if s.VaryByUser == nil {
		s.VaryByUser = NewBool(false)
	}
}

type LogSettings struct {
	EnableConsole          *bool   `access:"environment,write_restrictable"`
	ConsoleLevel           *string `access:"environment,write_restrictable"`
	ConsoleJson            *bool   `access:"environment,write_restrictable"`
	EnableFile             *bool   `access:"environment,write_restrictable"`
	FileLevel              *string `access:"environment,write_restrictable"`
	FileJson               *bool   `access:"environment,write_restrictable"`
	FileLocation           *string `access:"environment,write_restrictable"`
	EnableWebhookDebugging *bool   `access:"environment,write_restrictable"`
	EnableDiagnostics      *bool   `access:"environment,write_restrictable"`
	EnableSentry           *bool   `access:"environment,write_restrictable"`
	AdvancedLoggingConfig  *string `access:"environment,write_restrictable"`
}

func (s *LogSettings) SetDefaults() {
	if s.EnableConsole == nil {
		s.EnableConsole = NewBool(true)
	}

	if s.ConsoleLevel == nil {
		s.ConsoleLevel = NewString("DEBUG")
	}

	if s.EnableFile == nil {
		s.EnableFile = NewBool(true)
	}

	if s.FileLevel == nil {
		s.FileLevel = NewString("INFO")
	}

	if s.FileLocation == nil {
		s.FileLocation = NewString("")
	}

	if s.EnableWebhookDebugging == nil {
		s.EnableWebhookDebugging = NewBool(true)
	}

	if s.EnableDiagnostics == nil {
		s.EnableDiagnostics = NewBool(true)
	}

	if s.EnableSentry == nil {
		s.EnableSentry = NewBool(*s.EnableDiagnostics)
	}

	if s.ConsoleJson == nil {
		s.ConsoleJson = NewBool(true)
	}

	if s.FileJson == nil {
		s.FileJson = NewBool(true)
	}

	if s.AdvancedLoggingConfig == nil {
		s.AdvancedLoggingConfig = NewString("")
	}
}

type PluginState struct {
	Enable bool
}

type PluginSettings struct {
	Enable                      *bool                             `access:"plugins,write_restrictable"`
	EnableUploads               *bool                             `access:"plugins,write_restrictable"`
	AllowInsecureDownloadUrl    *bool                             `access:"plugins,write_restrictable"`
	EnableHealthCheck           *bool                             `access:"plugins,write_restrictable"`
	Directory                   *string                           `access:"plugins,write_restrictable"`
	ClientDirectory             *string                           `access:"plugins,write_restrictable"`
	Plugins                     map[string]map[string]interface{} `access:"plugins"`
	PluginStates                map[string]*PluginState           `access:"plugins"`
	EnableMarketplace           *bool                             `access:"plugins,write_restrictable"`
	EnableRemoteMarketplace     *bool                             `access:"plugins,write_restrictable"`
	AutomaticPrepackagedPlugins *bool                             `access:"plugins,write_restrictable"`
	RequirePluginSignature      *bool                             `access:"plugins,write_restrictable"`
	MarketplaceUrl              *string                           `access:"plugins,write_restrictable"`
	SignaturePublicKeyFiles     []string                          `access:"plugins,write_restrictable"`
}

func (s *PluginSettings) SetDefaults(ls LogSettings) {
	if s.Enable == nil {
		s.Enable = NewBool(true)
	}

	if s.EnableUploads == nil {
		s.EnableUploads = NewBool(false)
	}

	if s.AllowInsecureDownloadUrl == nil {
		s.AllowInsecureDownloadUrl = NewBool(false)
	}

	if s.EnableHealthCheck == nil {
		s.EnableHealthCheck = NewBool(true)
	}

	if s.Directory == nil || *s.Directory == "" {
		s.Directory = NewString(PLUGIN_SETTINGS_DEFAULT_DIRECTORY)
	}

	if s.ClientDirectory == nil || *s.ClientDirectory == "" {
		s.ClientDirectory = NewString(PLUGIN_SETTINGS_DEFAULT_CLIENT_DIRECTORY)
	}

	if s.Plugins == nil {
		s.Plugins = make(map[string]map[string]interface{})
	}

	if s.PluginStates == nil {
		s.PluginStates = make(map[string]*PluginState)
	}

	if s.PluginStates["com.mattermost.nps"] == nil {
		// Enable the NPS plugin by default if diagnostics are enabled
		s.PluginStates["com.mattermost.nps"] = &PluginState{Enable: ls.EnableDiagnostics == nil || *ls.EnableDiagnostics}
	}

	if s.PluginStates["com.mattermost.plugin-incident-response"] == nil {
		// Enable the incident response plugin by default
		s.PluginStates["com.mattermost.plugin-incident-response"] = &PluginState{Enable: true}
	}

	if s.EnableMarketplace == nil {
		s.EnableMarketplace = NewBool(PLUGIN_SETTINGS_DEFAULT_ENABLE_MARKETPLACE)
	}

	if s.EnableRemoteMarketplace == nil {
		s.EnableRemoteMarketplace = NewBool(true)
	}

	if s.AutomaticPrepackagedPlugins == nil {
		s.AutomaticPrepackagedPlugins = NewBool(true)
	}

	if s.MarketplaceUrl == nil || *s.MarketplaceUrl == "" || *s.MarketplaceUrl == PLUGIN_SETTINGS_OLD_MARKETPLACE_URL {
		s.MarketplaceUrl = NewString(PLUGIN_SETTINGS_DEFAULT_MARKETPLACE_URL)
	}

	if s.RequirePluginSignature == nil {
		s.RequirePluginSignature = NewBool(false)
	}

	if s.SignaturePublicKeyFiles == nil {
		s.SignaturePublicKeyFiles = []string{}
	}
}
