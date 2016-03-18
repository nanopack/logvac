// authenticator handles the management of 'X-AUTH-TOKEN's, allowing an
// authorized user to add, remove, and validate keys for http log collection.
package authenticator

import (
	"fmt"
	"io"
	"net/url"

	"github.com/nanopack/logvac/config"
)

type (
	Authenticatable interface {
		initialize() error
		add(token string) error
		remove(token string) error
		valid(token string) bool

		exportLogvac(exportWriter io.Writer) error
		importLogvac(importReader io.Reader) error
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
	config.Log.Info("Authenticator '%s' initialized", config.AuthAddress)
	return nil
}

// Add adds an authorized 'X-AUTH-TOKEN' for http collecting
func Add(token string) error {
	if authenticator == nil {
		return nil
	}
	config.Log.Trace("Adding token: %v...", token)
	return authenticator.add(token)
}

// Remove removes an authorized 'X-AUTH-TOKEN' for http collecting
func Remove(token string) error {
	if authenticator == nil {
		return nil
	}
	config.Log.Trace("Removing token: %v...", token)
	return authenticator.remove(token)
}

// Valid validates an authorized 'X-AUTH-TOKEN' for http collecting
func Valid(token string) bool {
	if authenticator == nil {
		return true
	}
	config.Log.Trace("Validating token: %v...", token)
	return authenticator.valid(token)
}

// Export exports auth tokens to a `logvac import`able file
func ExportLogvac(exportWriter io.Writer) error {
	// func ExportLogvac() error {
	if authenticator == nil {
		return fmt.Errorf("Authenticator not initialized")
	}
	config.Log.Trace("Exporting tokens...")
	return authenticator.exportLogvac(exportWriter)
}

// Import imports auth tokens from a `logvac export`ed file
func ImportLogvac(importReader io.Reader) error {
	if authenticator == nil {
		return nil
	}
	config.Log.Trace("Importing tokens...")
	return authenticator.importLogvac(importReader)
}
