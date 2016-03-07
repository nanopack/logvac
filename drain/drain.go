package drain

import (
	// "encoding/json"
	// "fmt"
	// "io"

	// "github.com/jcelliott/lumber"

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
	var err error
	// parse db-url if scheme bolt, use boltdb, default to bolt
	if true {
		Archiver, err = NewBoltArchive(config.DbAddress)
		if err != nil {
			return err
		}
	}
	// initialize Archiver
	err = Archiver.Init()
	if err != nil {
		return err
	}
	config.Log.Debug("Archiving drain added")

	// initialize publisher (if not empty)
	if config.PubAddress != "" {
		// parse pub-url if scheme mist? use mist, default to mist
		if true {
			publisher, err = NewMistClient(config.PubAddress)
			if err != nil {
				return err
			}
			// publisher = &Mist{}
		}
		err = publisher.Init()
		if err != nil {
			return err
		}
		config.Log.Debug("Publishing drain added")
	}

	return nil
}
