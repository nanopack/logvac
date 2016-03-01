// authenticator handles the management of 'X-LOGVAC-KEY's, allowing an
// authorized user to add, remove, and validate keys for http log collection.
package authenticator

import (
	"fmt"
	"net/url"

	"github.com/nanopack/logvac/config"
)

type (
	Authenticatable interface {
		initialize() error
		add(token string) error
		remove(token string) error
		valid(token string) bool
	}
)

var (
	authenticator Authenticatable
)

func Init() error {
	var err error
	var u *url.URL
	u, err = url.Parse(config.AuthAddress)
	if err != nil {
		return fmt.Errorf("Failed to parse db connection - %v", err)
	}
	switch u.Scheme {
	case "boltdb":
		authenticator, err = NewBoltDb(u.Path)
		if err != nil {
			return err
		}
	case "file":
		authenticator, err = NewBoltDb(u.Path)
		if err != nil {
			return err
		}
	case "postgresql":
		authenticator, err = NewPgDb(config.AuthAddress)
		if err != nil {
			return err
		}
	default:
		authenticator = nil
		config.Log.Debug("Authenticator not initialized")
		return nil
	}
	err = authenticator.initialize()
	if err != nil {
		return err
	}
	config.Log.Debug("Authenticator '%s' initialized", config.AuthAddress)
	return nil
}

// Add adds an authorized 'X-LOGVAC-KEY' for http collecting
func Add(token string) error {
	if authenticator == nil {
		return nil
	}
	config.Log.Trace("Adding token: %v...", token)
	return authenticator.add(token)
}

// Remove removes an authorized 'X-LOGVAC-KEY' for http collecting
func Remove(token string) error {
	if authenticator == nil {
		return nil
	}
	config.Log.Trace("Removing token: %v...", token)
	return authenticator.remove(token)
}

// Valid validates an authorized 'X-LOGVAC-KEY' for http collecting
func Valid(token string) bool {
	if authenticator == nil {
		return true
	}
	config.Log.Trace("Validating token: %v...", token)
	return authenticator.valid(token)
}
