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
		Init() error
		// Slice returns a slice of logs based on the name, offset, limit, and log-level
		Slice(name, host, tag string, offset, limit uint64, level int) ([]logvac.Message, error)
		// Write writes the message to database
		Write(msg logvac.Message)
		Expire()
	}

	// Publisher defines a pub-sub type drain
	PublisherDrain interface {
		Init() error
		// Publish publishes the tagged data
		Publish(tag []string, data string) error
	}
)

var (
	publisher PublisherDrain
	Archiver  ArchiverDrain
)

func Init() error {
	// initialize archiver
	err := archiveInit()
	if err != nil {
		return fmt.Errorf("Failed to initialize archiver - %v", err)
	}
	config.Log.Info("Archiving drain '%s' initialized", config.DbAddress)

	// initialize publisher (if not empty)
	if config.PubAddress != "" {
		err = publishInit()
		if err != nil {
			return fmt.Errorf("Failed to initialize publisher - %v", err)
		}
		config.Log.Info("Publishing drain '%s' initialized", config.PubAddress)
	}

	return nil
}

func archiveInit() error {
	u, err := url.Parse(config.DbAddress)
	if err != nil {
		return fmt.Errorf("Failed to parse db connection - %v", err)
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
		return fmt.Errorf("Failed to parse publisher connection - %v", err)
	}
	switch u.Scheme {
	case "mist":
		publisher, err = NewMistClient(config.PubAddress)
		if err != nil {
			return err
		}
	default:
		publisher, err = NewMistClient(config.PubAddress)
		if err != nil {
			return err
		}
	}
	// initialize publisher
	err = publisher.Init()
	if err != nil {
		return err
	}
	return nil
}
