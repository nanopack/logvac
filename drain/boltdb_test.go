// drain_test tests the archiver and publisher drain
// functionality (writing, reading, cleaning)
package drain_test

import (
	"os"
	"testing"
	"time"

	"github.com/jcelliott/lumber"

	"github.com/nanopack/logvac/config"
	"github.com/nanopack/logvac/core"
	"github.com/nanopack/logvac/drain"
)

func TestMain(m *testing.M) {
	// clean test dir
	os.RemoveAll("/tmp/boltdbTest")

	// manually configure
	err := initialize()
	if err != nil {
		os.Exit(1)
	}

	rtn := m.Run()

	// clean test dir
	os.RemoveAll("/tmp/boltdbTest")

	os.Exit(rtn)
}

// Test writing and getting data
func TestWrite(t *testing.T) {
	// create test messages
	messages := []logvac.Message{
		logvac.Message{
			Time:     time.Now(),
			UTime:    time.Now().UnixNano(),
			Id:       "myhost",
			Tag:      "test[bolt]",
			Type:     "app",
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
	drain.Archiver.Write(messages[0])
	drain.Archiver.Write(messages[1])

	// test successful write
	appMsgs, err := drain.Archiver.Slice("app", "", "", 0, 0, 100, 0)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	// compare written message to original
	if len(appMsgs) != 1 || appMsgs[0].Content != messages[0].Content {
		t.Errorf("%q doesn't match expected out", appMsgs)
		t.FailNow()
	}
}

// Test expiring/cleanup of data
func TestExpire(t *testing.T) {
	go drain.Archiver.Expire()
	time.Sleep(2 * time.Second)

	// finish expire loop
	drain.Archiver.(*drain.BoltArchive).Done <- true

	// test successful clean
	appMsgs, err := drain.Archiver.Slice("app", "", "", 0, 0, 100, 0)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	// compare written message to original
	if len(appMsgs) != 0 {
		t.Errorf("%q doesn't match expected out", appMsgs)
		t.FailNow()
	}

	// test successful clean
	depMsgs, err := drain.Archiver.Slice("deploy", "", "", 0, 0, 100, 0)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	// compare written message to original
	if len(depMsgs) != 0 {
		t.Errorf("%q doesn't match expected out", depMsgs)
		t.FailNow()
	}

	drain.Archiver.(*drain.BoltArchive).Close()

}

// manually configure and start internals
func initialize() error {
	var err error
	config.CleanFreq = 1
	config.LogKeep = `{"app": "1s", "deploy":0}`
	config.LogKeep = `{"app": "1s", "deploy":0, "a":"1m", "aa":"1h", "b":"1d", "c":"1w", "d":"1y", "e":"1"}`
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt("ERROR"))

	// initialize logvac
	logvac.Init()

	// initialize archiver
	// Doing broke db
	config.DbAddress = "~!@#$%^&*()"
	drain.Init()

	// Doing file db
	config.DbAddress = "file:///tmp/boltdbTest/logvac.bolt"
	drain.Init()
	drain.Archiver.(*drain.BoltArchive).Close()

	// Doing no db
	config.DbAddress = "/tmp/boltdbTest/logvac.bolt"
	drain.Init()
	drain.Archiver.(*drain.BoltArchive).Close()

	// Doing bolt db
	config.DbAddress = "boltdb:///tmp/boltdbTest/logvac.bolt"
	drain.Init()

	return err
}
