package keystore

import "github.com/zalando/go-keyring"

const (
	serviceName = "hardcover-tui"
	userName    = "api-key"
)

func Save(apiKey string) error {
	return keyring.Set(serviceName, userName, apiKey)
}

func Load() (string, error) {
	return keyring.Get(serviceName, userName)
}

func Delete() error {
	return keyring.Delete(serviceName, userName)
}
