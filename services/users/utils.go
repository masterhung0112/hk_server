// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package users

import (
	"strings"

	"github.com/masterhung0112/hk_server/v5/model"
)

// checkUserDomain checks that a user's email domain matches a list of space-delimited domains as a string.
func checkUserDomain(user *model.User, domains string) bool {
	return checkEmailDomain(user.Email, domains)
}

// checkEmailDomain checks that an email domain matches a list of space-delimited domains as a string.
func checkEmailDomain(email string, domains string) bool {
	if domains == "" {
		return true
	}

	domainArray := strings.Fields(strings.TrimSpace(strings.ToLower(strings.Replace(strings.Replace(domains, "@", " ", -1), ",", " ", -1))))

	for _, d := range domainArray {
		if strings.HasSuffix(strings.ToLower(email), "@"+d) {
			return true
		}
	}

	return false
}

func (us *UserService) sanitizeProfiles(users []*model.User, asAdmin bool) []*model.User {
	for _, u := range users {
		us.SanitizeProfile(u, asAdmin)
	}

	return users
}

func (us *UserService) SanitizeProfile(user *model.User, asAdmin bool) {
	options := us.GetSanitizeOptions(asAdmin)

	user.SanitizeProfile(options)
}

func (us *UserService) GetSanitizeOptions(asAdmin bool) map[string]bool {
	options := us.config().GetSanitizeOptions()
	if asAdmin {
		options["email"] = true
		options["fullname"] = true
		options["authservice"] = true
	}
	return options
}

// IsUsernameTaken checks if the username is already used by another user. Return false if the username is invalid.
func (us *UserService) IsUsernameTaken(name string) bool {
	if !model.IsValidUsername(name) {
		return false
	}

	if _, err := us.store.GetByUsername(name); err != nil {
		return false
	}

	return true
}
