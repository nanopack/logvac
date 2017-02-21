// Package authenticator handles the management of 'X-USER-TOKEN's, allowing an
// authorized admin to add, remove, and validate keys for http log collection.
package authenticator

import (
	"fmt"
	"io"
	"net/url"

	"github.com/nanopack/logvac/config"
)

type (
	// Authenticatable contains methods all authenticators should have
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

// Init initializes the chosen authenticator
func Init() error {
	u, err := url.Parse(config.AuthAddress)
	if err != nil {
		u, err = url.Parse("boltdb://" + config.AuthAddress)
		if err != nil {
			return fmt.Errorf("Failed to parse auth connection - %s", err)
		}
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

// Add adds an authorized 'X-USER-TOKEN' for http collecting
func Add(token string) error {
	if authenticator == nil {
		return nil
	}
	config.Log.Trace("Adding token: %s...", token)
	return authenticator.add(token)
}

// Remove removes an authorized 'X-USER-TOKEN' for http collecting
func Remove(token string) error {
	if authenticator == nil {
		return nil
	}
	config.Log.Trace("Removing token: %s...", token)
	return authenticator.remove(token)
}

// Valid validates an authorized 'X-USER-TOKEN' for http collecting
func Valid(token string) bool {
	if authenticator == nil {
		return true
	}
	config.Log.Trace("Validating token: %s...", token)
	return authenticator.valid(token)
}

// ExportLogvac exports auth tokens to a `logvac import`able file
func ExportLogvac(exportWriter io.Writer) error {
	// func ExportLogvac() error {
	if authenticator == nil {
		return fmt.Errorf("Authenticator not initialized")
	}
	config.Log.Trace("Exporting tokens...")
	return authenticator.exportLogvac(exportWriter)
}

// ImportLogvac imports auth tokens from a `logvac export`ed file
func ImportLogvac(importReader io.Reader) error {
	if authenticator == nil {
		return nil
	}
	config.Log.Trace("Importing tokens...")
	return authenticator.importLogvac(importReader)
}
