package utils

import (
	"github.com/masterhung0112/go_server/model"
	"strings"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsPasswordValidWithSettings(t *testing.T) {
	for name, tc := range map[string]struct {
		Password      string
		Settings      *model.PasswordSettings
		ExpectedError string
	}{
		"Short": {
			Password: strings.Repeat("x", 3),
			Settings: &model.PasswordSettings{
				MinimumLength: model.NewInt(3),
				Lowercase:     model.NewBool(false),
				Uppercase:     model.NewBool(false),
				Number:        model.NewBool(false),
				Symbol:        model.NewBool(false),
			},
		},
		"MissingLower": {
			Password: "AAAAAAAAAAASD123!@#",
			Settings: &model.PasswordSettings{
				Lowercase: model.NewBool(true),
				Uppercase: model.NewBool(false),
				Number:    model.NewBool(false),
				Symbol:    model.NewBool(false),
			},
			ExpectedError: "model.user.is_valid.pwd_lowercase.app_error",
		},
	} {
		tc.Settings.SetDefaults()
		t.Run(name, func(t *testing.T) {
			if err := IsPasswordValidWithSettings(tc.Password, tc.Settings); tc.ExpectedError == "" {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, tc.ExpectedError, err.Id)
			}
		})
	}
}
