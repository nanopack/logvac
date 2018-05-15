// Package drain handles the storing and publishing of logs.
package drain

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

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
		// Close closes the drain
		Close() error
	}
)

var (
	Publisher PublisherDrain            // default publish drain
	Archiver  ArchiverDrain             // default archive drain
	drains    map[string]PublisherDrain // contains the third party drains configured
	drainCfg  map[string]logvac.Drain   // contains the third party drain configuration
	dbDir     string                    // location of db to store drain config
)

func init() {
	drains = make(map[string]PublisherDrain, 0)
	drainCfg = make(map[string]logvac.Drain, 0)
}

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

	err = InitDrains()
	if err != nil {
		return fmt.Errorf("Failed to load drains - %s", err)
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

	dbDir = filepath.Dir(u.Path)

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

// InitDrains loads and configures drains from a config file.
func InitDrains() error {
	drainDB, err := NewBoltArchive(filepath.Join(dbDir, "drains.bolt"))
	if err != nil {
		return fmt.Errorf("Failed to initialize drain db - %s", err)
	}
	defer drainDB.Close()

	tDrains := make(map[string]logvac.Drain, 0)
	err = drainDB.Get("drainConfig", "drains", &tDrains)
	drainDB.Close()
	if err != nil && !strings.Contains(err.Error(), "No bucket found") {
		return fmt.Errorf("Failed to load drain config - %s", err)
	}

	for i := range tDrains {
		err := AddDrain(tDrains[i])
		if err != nil {
			return fmt.Errorf("Failed to load drain 'papertrail' - %s", err)
		}
	}

	return nil
}

// AddDrain starts draining to a third party log service.
func AddDrain(d logvac.Drain) error {

	switch d.Type {
	case "papertrail":
		// if it already exists, close it and create a new one
		if _, ok := drains["papertrail"]; ok {
			drains["papertrail"].Close()
		}
		// pTrail, err := NewPapertrailClient("logs6.papertrailapp.com:19900")
		pTrail, err := NewPapertrailClient(d.URI)
		if err != nil {
			return fmt.Errorf("Failed to create papertrail client - %s", err)
		}
		err = pTrail.Init()
		if err != nil {
			return fmt.Errorf("Papertrail failed to initialize - %s", err)
		}
		drains["papertrail"] = pTrail
		drainCfg["papertrail"] = d
	case "datadog":
		// if it already exists, close it and create a new one
		if _, ok := drains["datadog"]; ok {
			drains["datadog"].Close()
		}
		dDog, err := NewDatadogClient(d.AuthKey)
		if err != nil {
			return fmt.Errorf("Failed to create datadog client - %s", err)
		}
		err = dDog.Init()
		if err != nil {
			return fmt.Errorf("Datadog failed to initialize - %s", err)
		}
		drains["datadog"] = dDog
		drainCfg["datadog"] = d
	default:
		return fmt.Errorf("Drain type not supported")
	}

	// saving drain
	drainDB, err := NewBoltArchive(filepath.Join(dbDir, "drains.bolt"))
	if err != nil {
		return fmt.Errorf("Failed to initialize drain db - %s", err)
	}
	defer drainDB.Close()

	drainDB.Save("drainConfig", "drains", drainCfg)

	return nil
}

// RemoveDrain stops draining to a third party log service.
func RemoveDrain(drainType string) error {
	if _, ok := drains[drainType]; !ok {
		return nil
	}

	err := drains[drainType].Close()
	if err != nil {
		return fmt.Errorf("Drain '%s' failed to close - %s", drainType, err.Error())
	}

	delete(drains, drainType)
	delete(drainCfg, drainType)

	drainDB, err := NewBoltArchive(filepath.Join(dbDir, "drains.bolt"))
	if err != nil {
		return fmt.Errorf("Failed to initialize drain db - %s", err)
	}
	defer drainDB.Close()

	return drainDB.Save("drainConfig", "drains", drainCfg)
}

// GetDrain shows the drain information.
func GetDrain(d logvac.Drain) (*PublisherDrain, error) {
	drain, ok := drains[d.Type]
	if !ok {
		return nil, fmt.Errorf("Drain not found")
	}
	return &drain, nil
}

// ListDrains shows all the drains configured.
func ListDrains() map[string]PublisherDrain {
	return drains
}
