package utils

import (
	"net/url"
	"regexp"
	"strings"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+(?:_[a-zA-Z0-9]+)*$`)

func IsValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	return usernameRegex.MatchString(username)
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

var reservedUsernames = map[string]struct{}{
	"admin": {}, "root": {}, "support": {}, "null": {}, "contact": {}, "api": {}, "system": {},
}

func IsReservedUsername(username string) bool {
	_, exists := reservedUsernames[strings.ToLower(username)]
	return exists
}

func ContainsNumberAndSymbol(s string) bool {
	hasNum := false
	hasSym := false
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			hasNum = true
		case strings.ContainsRune("!@#$%^&*()-_=+[]{}|;:',.<>?/`~", r):
			hasSym = true
		}
	}
	return hasNum && hasSym
}

func IsValidURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	return err == nil
}

// IsNumeric checks if a string contains only numeric characters
func IsNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
