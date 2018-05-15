package drain_test

import (
	"bytes"
	"testing"

	"github.com/nanopack/logvac/core"
	"github.com/nanopack/logvac/drain"
)

// Test adding drain.
func TestPTrailInit(t *testing.T) {
	trailTest := &drain.Papertrail{}
	trailTest.Close()
	trailTest.Init()
}

// Test writing and reading data, as well as closing.
func TestPTrailPublish(t *testing.T) {
	var b WriteCloseBuffer
	trailTest := &drain.Papertrail{Conn: &b}
	if trailTest.Conn == nil {
		t.Fatal("Failed to create a thing")
	}

	msg := logvac.Message{Raw: []byte("This is a message\n")}

	trailTest.Publish(msg)
	if b.String() != string(msg.Raw) {
		t.Fatalf("Failed to publish - '%s'", b.String())
	}
	trailTest.Close()
}

type WriteCloseBuffer struct {
	bytes.Buffer
}

func (cb WriteCloseBuffer) Close() error {
	return nil
}
