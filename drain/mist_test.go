package drain_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jcelliott/lumber"
	"github.com/nanopack/mist/server"

	"github.com/nanopack/logvac/config"
	"github.com/nanopack/logvac/core"
	"github.com/nanopack/logvac/drain"
)

var (
	line       chan string
	stop       chan bool
	PubAddress = "127.0.0.1:2445"
)

// Test writing and getting data
func TestPublish(t *testing.T) {
	lumber.Level(lumber.LvlInt("fatal"))
	err := mistInitialize()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// create test messages
	messages := []logvac.Message{
		logvac.Message{
			Time:     time.Now(),
			UTime:    time.Now().UnixNano(),
			Id:       "myhost",
			Tag:      "test[bolt]",
			Type:     "",
			Priority: 4,
			Content:  "This is a test message",
		},
		logvac.Message{
			Time:     time.Now(),
			UTime:    time.Now().UnixNano(),
			Id:       "myhost",
			Tag:      "test[expire]",
			Type:     "deploy",
			Priority: 4,
			Content:  "This is another test message",
		},
	}

	// write test messages
	drain.Publisher.Publish(messages[0])
	drain.Publisher.Publish(messages[1])
}

// manually configure and start internals
func mistInitialize() error {
	var err error
	config.CleanFreq = 1
	config.LogKeep = `{"app": "1s", "deploy":0}`
	config.LogKeep = `{"app": "1s", "deploy":0, "a":"1m", "aa":"1h", "b":"1d", "c":"1w", "d":"1y", "e":"1"}`
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt("ERROR"))
	config.DbAddress = "boltdb:///tmp/boltdbTest/logvac.bolt"

	server.StartTCP(PubAddress, nil)

	// initialize logvac
	logvac.Init()

	// initialize publisher
	// Doing broke publisher
	config.PubAddress = "~!@#$%^&*()"
	drain.Init()
	drain.Archiver.(*drain.BoltArchive).Close()

	// Doing schemeless publisher
	config.PubAddress = "127.0.0.1:2445"
	drain.Init()
	drain.Archiver.(*drain.BoltArchive).Close()
	drain.Publisher.(*drain.Mist).Close()

	// Doing real publisher
	config.PubAddress = "mist://127.0.0.1:2445"
	err = drain.Init()

	return err
}
