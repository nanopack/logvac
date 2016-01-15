package authenticator

import (
	"github.com/nanopack/logvac/config"
)

type (
	Authenticator interface {
		Setup(config string) error
		Add(token string) error
		Remove(token string) error
		Valid(token string) bool
	}
)

var (
	availableAuthenticator = map[string]Authenticator{}
	defaultAuthenticator Authenticator
)

func Register(name string, a Authenticator) {
	availableAuthenticator[name] = a
}

func Setup() error {
	authenticator, ok := availableAuthenticator[config.AuthType]
	if ok {
		config.Log.Info("setting up authenticator(%s) config: %s", config.AuthType, config.AuthConfig)
		defaultAuthenticator = authenticator
		return authenticator.Setup(config.AuthConfig)
	}
	return nil
}

func Add(token string) error {
	if defaultAuthenticator == nil {
		return nil
	}
	return defaultAuthenticator.Add(token)
}
func Remove(token string) error {
	if defaultAuthenticator == nil {
		return nil
	}
	return defaultAuthenticator.Remove(token)
}
func Valid(token string) bool {
	if defaultAuthenticator == nil {
		return true
	}
	return defaultAuthenticator.Valid(token)
}
