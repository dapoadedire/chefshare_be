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
    // Common administrative and system-related names
    "admin": {}, "root": {}, "support": {}, "system": {}, "administrator": {}, "superuser": {},
    "webmaster": {}, "hostmaster": {}, "postmaster": {}, "daemon": {}, "nobody": {},
    "staff": {}, "security": {}, "billing": {}, "management": {},

    // Technical and network-related terms
    "localhost": {}, "127.0.0.1": {}, // IPv4 loopback address
    "::1": {}, // IPv6 loopback address
    "ftp": {}, "ssh": {}, "smtp": {}, "http": {}, "https": {}, "www": {}, "mail": {},
    "ns1": {}, "ns2": {}, // Common nameserver prefixes
    "server": {}, "client": {}, "router": {}, "gateway": {}, "network": {},
    "proxy": {}, "firewall": {}, "backend": {}, "frontend": {},

    // Potentially confusing or problematic names
    "null": {}, "test": {}, "example": {}, "info": {}, "contact": {}, "anonymous": {},
    "guest": {}, "public": {}, "user": {}, "users": {}, "me": {}, "you": {}, "myself": {},

    // API and application-specific names
    "api": {}, "developer": {}, "dev": {}, "app": {}, "service": {}, "bot": {},
    "webhook": {}, "callback": {},

    // Names that could be confused with database or internal system terms
    "select": {}, "insert": {}, "update": {}, "delete": {},
    "account": {}, "accounts": {}, "profile": {}, "profiles": {}, "data": {}, "database": {},
    "schema": {}, "table": {}, "column": {}, "row": {}, "index": {}, "query": {},

    // Names that could interfere with routing or reserved paths
    "login": {}, "logout": {}, "register": {}, "signup": {}, "signin": {}, "signout": {},
    "password": {}, "reset": {}, "forgot": {}, "settings": {}, "dashboard": {}, "home": {},
    "about": {}, "help": {}, "faq": {}, "privacy": {}, "terms": {}, "legal": {}, "blog": {},
    "forum": {}, "community": {}, "feed": {}, "feeds": {}, "explore": {}, "discover": {},
    "articles": {}, "posts": {}, "pages": {}, "files": {}, "upload": {}, "download": {},
    "assets": {}, "images": {}, "videos": {}, "audio": {}, "docs": {}, "documentation": {},

    // Potentially sensitive, religious, or highly opinionated terms
    // Use caution and consider your application's audience and legal implications.
    "god": {}, "devil": {}, "jesus": {}, "buddha": {}, "allah": {}, "satan": {},
    "creator": {}, "master": {}, "slave": {}, // Note: "slave" might have problematic connotations
    "administer": {}, "commander": {}, "official": {},

    // Names of your own application/company (if applicable)
    // "yourcompanyname": {}, "yourappname": {}, // Replace with your actual names

    // Single character names (often used for testing or can be confusing)
    "a": {}, "b": {}, // ... "z"
    "1": {}, "2": {}, // ... "0" (for common numbers)

    // Common placeholders and generic terms
    "username": {}, "email": {}, "undefined": {}, "nil": {},
    "current": {}, "new": {}, "old": {}, "archive": {}, "archives": {}, "search": {},
    "subscribe": {}, "unsubscribe": {}, "reply": {}, "replies": {}, "message": {},
    "messages": {}, "notification": {}, "notifications": {}, "alert": {}, "alerts": {},
    "debug": {}, "testuser": {}, "temp": {}, "tmp": {}, "demo": {}, "draft": {},
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
