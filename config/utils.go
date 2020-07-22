package config

import (
	"github.com/masterhung0112/go_server/model"

)

// desanitize replaces fake settings with their actual values.
func desanitize(actual, target *model.Config) {
    //TODO: Uncomment out

	// if target.LdapSettings.BindPassword != nil && *target.LdapSettings.BindPassword == model.FAKE_SETTING {
	// 	*target.LdapSettings.BindPassword = *actual.LdapSettings.BindPassword
	// }

	// if *target.FileSettings.PublicLinkSalt == model.FAKE_SETTING {
	// 	*target.FileSettings.PublicLinkSalt = *actual.FileSettings.PublicLinkSalt
	// }
	// if *target.FileSettings.AmazonS3SecretAccessKey == model.FAKE_SETTING {
	// 	target.FileSettings.AmazonS3SecretAccessKey = actual.FileSettings.AmazonS3SecretAccessKey
	// }

	// if *target.EmailSettings.SMTPPassword == model.FAKE_SETTING {
	// 	target.EmailSettings.SMTPPassword = actual.EmailSettings.SMTPPassword
	// }

	// if *target.GitLabSettings.Secret == model.FAKE_SETTING {
	// 	target.GitLabSettings.Secret = actual.GitLabSettings.Secret
	// }

	// if *target.SqlSettings.DataSource == model.FAKE_SETTING {
	// 	*target.SqlSettings.DataSource = *actual.SqlSettings.DataSource
	// }
	// if *target.SqlSettings.AtRestEncryptKey == model.FAKE_SETTING {
	// 	target.SqlSettings.AtRestEncryptKey = actual.SqlSettings.AtRestEncryptKey
	// }

	// if *target.ElasticsearchSettings.Password == model.FAKE_SETTING {
	// 	*target.ElasticsearchSettings.Password = *actual.ElasticsearchSettings.Password
	// }

	// target.SqlSettings.DataSourceReplicas = make([]string, len(actual.SqlSettings.DataSourceReplicas))
	// for i := range target.SqlSettings.DataSourceReplicas {
	// 	target.SqlSettings.DataSourceReplicas[i] = actual.SqlSettings.DataSourceReplicas[i]
	// }

	// target.SqlSettings.DataSourceSearchReplicas = make([]string, len(actual.SqlSettings.DataSourceSearchReplicas))
	// for i := range target.SqlSettings.DataSourceSearchReplicas {
	// 	target.SqlSettings.DataSourceSearchReplicas[i] = actual.SqlSettings.DataSourceSearchReplicas[i]
	// }
}

// fixConfig patches invalid or missing data in the configuration, returning true if changed.
func fixConfig(cfg *model.Config) bool {
  changed := false

  //TODO: Uncomment out

	// Ensure SiteURL has no trailing slash.
	// if strings.HasSuffix(*cfg.ServiceSettings.SiteURL, "/") {
	// 	*cfg.ServiceSettings.SiteURL = strings.TrimRight(*cfg.ServiceSettings.SiteURL, "/")
	// 	changed = true
	// }

	// Ensure the directory for a local file store has a trailing slash.
	// if *cfg.FileSettings.DriverName == model.IMAGE_DRIVER_LOCAL {
	// 	if !strings.HasSuffix(*cfg.FileSettings.Directory, "/") {
	// 		*cfg.FileSettings.Directory += "/"
	// 		changed = true
	// 	}
	// }

	// if FixInvalidLocales(cfg) {
	// 	changed = true
	// }

	return changed
}
