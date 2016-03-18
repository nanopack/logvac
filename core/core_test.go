// logvac_test tests the core drain functionality (adding, writing, removing)
package logvac_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/jcelliott/lumber"

	"github.com/nanopack/logvac/config"
	"github.com/nanopack/logvac/core"
)

func TestMain(m *testing.M) {
	// manually configure
	err := initialize()
	if err != nil {
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// Test adding and writing to a drain
func TestAddDrain(t *testing.T) {
	// create a buffer to "drain" to
	buf := &bytes.Buffer{}

	// add buffer drain
	logvac.AddDrain("test", writeDrain(buf))

	// create test message
	msg := logvac.Message{
		Time:     time.Now(),
		UTime:    time.Now().UnixNano(),
		Id:       "myhost",
		Tag:      "test[drains]",
		Type:     "app",
		Priority: 4,
		Content:  "This is quite important",
	}

	// write test message
	logvac.WriteMessage(msg)
	time.Sleep(time.Millisecond)

	// ensure write succeeded
	// read from buffer
	r, err := buf.ReadBytes('\n')
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// convert readbytes to Message
	rMsg := logvac.Message{}
	err = json.Unmarshal(r, &rMsg)
	if err != nil {
		t.Error(fmt.Errorf("Failed to unmarshal - %v", err))
		t.FailNow()
	}

	// compare "drained" message to original
	if rMsg.Content != msg.Content {
		t.Errorf("%q doesn't match expected out", rMsg.Content)
		t.FailNow()
	}
}

// Test removing a drain
func TestRemoveDrain(t *testing.T) {
	tag := "null"
	drain := func(msg logvac.Message) {
		return
	}
	logvac.AddDrain(tag, drain)
	logvac.RemoveDrain(tag)
}

// Test closing the logvac instance
func TestClose(t *testing.T) {
	logvac.Close()
	time.Sleep(time.Second)
}

// writeDrain creates a drain from an io.Writer
func writeDrain(writer io.Writer) logvac.Drain {
	return func(msg logvac.Message) {
		data, err := json.Marshal(msg)
		if err != nil {
			config.Log.Error("writeDrain failed to marshal message")
			return
		}
		writer.Write(append(data, '\n'))
	}
}

// manually configure and start internals
func initialize() error {
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt("ERROR"))

	// initialize logvac
	return logvac.Init()
}
