package model

import (
	"encoding/json"
)

const (
  CONN_SECURITY_TLS = "TLS"

  PASSWORD_MAXIMUM_LENGTH = 64
  PASSWORD_MINIMUM_LENGTH = 5
  SERVICE_SETTINGS_DEFAULT_SITE_URL           = "http://localhost:8065"
  SERVICE_SETTINGS_DEFAULT_LISTEN_AND_ADDRESS = ":8065"
  DEFAULT_LOCALE = "en"

  DATABASE_DRIVER_SQLITE   = "sqlite3"
	DATABASE_DRIVER_MYSQL    = "mysql"
  DATABASE_DRIVER_POSTGRES = "postgres"

  IMAGE_DRIVER_LOCAL = "local"
  IMAGE_DRIVER_S3    = "s3"

  FILE_SETTINGS_DEFAULT_DIRECTORY = "./data/"

  SQL_SETTINGS_DEFAULT_DATA_SOURCE = "mmuser:mostest@tcp(localhost:3306)/mattermost_test?charset=utf8mb4,utf8&readTimeout=30s&writeTimeout=30s"
)

type Config struct {
  ServiceSettings           ServiceSettings
  PasswordSettings PasswordSettings
  LocalizationSettings LocalizationSettings
  SqlSettings SqlSettings
}

// isUpdate detects a pre-existing config based on whether SiteURL has been changed
func (o *Config) isUpdate() bool {
	return o.ServiceSettings.SiteURL != nil
}

func (o *Config) SetDefaults() {
	isUpdate := o.isUpdate()

  o.ServiceSettings.SetDefaults(isUpdate)
  o.PasswordSettings.SetDefaults()
  o.LocalizationSettings.SetDefaults()
  o.SqlSettings.SetDefaults(isUpdate)
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
  SiteURL                                           *string  `restricted:"true"`
  ConnectionSecurity                                *string  `restricted:"true"`
  ListenAddress                                     *string  `restricted:"true"`
  EnableDeveloper                                   *bool   `restricted:"true"`
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
  DriverName *string
  DataSource                  *string  `restricted:"true"`
  DataSourceReplicas          []string
  QueryTimeout                *int     `restricted:"true"`
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