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
	authenticator          Authenticator
)

func Register(name string, a Authenticator) {
	availableAuthenticator[name] = a
}

func Setup() error {
	// todo: refactor to parse authType/config and switch default to none
	// would get rid of `Register`, although clever
	var ok bool
	authenticator, ok = availableAuthenticator[config.AuthType]
	if ok {
		config.Log.Debug("Authenticator(%s) config: %s initializing...", config.AuthType, config.AuthConfig)
		return authenticator.Setup(config.AuthConfig)
	}
	return nil
}

func Add(token string) error {
	if authenticator == nil {
		return nil
	}
	return authenticator.Add(token)
}
func Remove(token string) error {
	if authenticator == nil {
		return nil
	}
	return authenticator.Remove(token)
}
func Valid(token string) bool {
	if authenticator == nil {
		return true
	}
	return authenticator.Valid(token)
}
