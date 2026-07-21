package keystore

import (
	"strings"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "hardcover-tui"
	userName    = "api-key"
)

// FormatToken cleans and formats an API token to ensure it has a single "Bearer " prefix.
func FormatToken(apiKey string) string {
	token := strings.TrimSpace(apiKey)
	if token == "" {
		return ""
	}
	for {
		lower := strings.ToLower(token)
		if strings.HasPrefix(lower, "bearer ") {
			token = strings.TrimSpace(token[7:])
		} else if strings.HasPrefix(lower, "bearer") {
			token = strings.TrimSpace(token[6:])
		} else {
			break
		}
	}
	if token == "" {
		return ""
	}
	return "Bearer " + token
}

func Save(apiKey string) error {
	formatted := FormatToken(apiKey)
	return keyring.Set(serviceName, userName, formatted)
}

func Load() (string, error) {
	token, err := keyring.Get(serviceName, userName)
	if err != nil {
		return "", err
	}
	return FormatToken(token), nil
}

func Delete() error {
	return keyring.Delete(serviceName, userName)
}

