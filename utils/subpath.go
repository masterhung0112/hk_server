package utils

import (
	"github.com/masterhung0112/go_server/model"
	"github.com/pkg/errors"
	"net/url"
	"path"
)

func GetSubpathFromConfig(config *model.Config) (string, error) {
	if config == nil {
		return "", errors.New("no config provided")
	} else if config.ServiceSettings.SiteURL == nil {
		return "/", nil
	}

	u, err := url.Parse(*config.ServiceSettings.SiteURL)
	if err != nil {
		return "", errors.Wrap(err, "Failed to parse SiteURL from config")
	}

	if u.Path == "" {
		return "/", nil
	}

	return path.Clean(u.Path), nil
}
