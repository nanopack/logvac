// Package drain handles the storing and publishing of logs.
package drain

import (
	"fmt"
	"net/url"

	"github.com/nanopack/logvac/config"
	"github.com/nanopack/logvac/core"
)

type (
	// Archiver defines a storage type drain
	ArchiverDrain interface {
		// Init initializes the archiver drain
		Init() error
		// Slice returns a slice of logs based on the name, offset, limit, and log-level
		Slice(name, host string, tag []string, offset, end, limit int64, level int) ([]logvac.Message, error)
		// Write writes the message to database
		Write(msg logvac.Message)
		// Expire cleans up old logs
		Expire()
	}

	// Publisher defines a pub-sub type drain
	PublisherDrain interface {
		// Init initializes the publish drain
		Init() error
		// Publish publishes the tagged data
		Publish(msg logvac.Message)
	}
)

var (
	Publisher PublisherDrain // default publish drain
	Archiver  ArchiverDrain  // default archive drain
)

// Init initializes the archiver and publisher drains if configured
func Init() error {
	// initialize archiver
	err := archiveInit()
	if err != nil {
		return fmt.Errorf("Failed to initialize archiver - %s", err)
	}
	config.Log.Info("Archiving drain '%s' initialized", config.DbAddress)

	// initialize publisher (if not empty)
	if config.PubAddress != "" {
		err = publishInit()
		if err != nil {
			return fmt.Errorf("Failed to initialize publisher - %s", err)
		}
		config.Log.Info("Publishing drain '%s' initialized", config.PubAddress)
	}

	return nil
}

func archiveInit() error {
	u, err := url.Parse(config.DbAddress)
	if err != nil {
		u, err = url.Parse("boltdb://" + config.DbAddress)
		if err != nil {
			return fmt.Errorf("Failed to parse db connection - %s", err)
		}
	}
	switch u.Scheme {
	case "boltdb":
		// todo: use `dirname DbAddress` and create a 'db' for each log-type
		Archiver, err = NewBoltArchive(u.Path)
		if err != nil {
			return err
		}
	case "file":
		Archiver, err = NewBoltArchive(u.Path)
		if err != nil {
			return err
		}
	// case "elasticsearch":
	// 	Archiver, err = NewElasticArchive(config.DbAddress)
	// 	if err != nil {
	// 		return err
	// 	}
	default:
		Archiver, err = NewBoltArchive(u.Path)
		if err != nil {
			return err
		}
	}
	// initialize Archiver
	err = Archiver.Init()
	if err != nil {
		return err
	}
	// start cleanup goroutine
	go Archiver.Expire()
	return nil
}

func publishInit() error {
	u, err := url.Parse(config.PubAddress)
	if err != nil {
		u, err = url.Parse("mist://" + config.PubAddress) // hard requirement for scheme in go 1.8 (with ip:port only)
		if err != nil {
			return fmt.Errorf("Failed to parse publisher connection - %s", err)
		}
	}
	switch u.Scheme {
	case "mist":
		Publisher, err = NewMistClient(u.Host)
		if err != nil {
			return err
		}
	case "": // url.Parse requires scheme
		Publisher, err = NewMistClient(config.PubAddress)
		if err != nil {
			return err
		}
	default:
		Publisher, err = NewMistClient(u.Host)
		if err != nil {
			return err
		}
	}
	// initialize publisher
	err = Publisher.Init()
	if err != nil {
		return err
	}
	return nil
}
