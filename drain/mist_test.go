package drain_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jcelliott/lumber"
	"github.com/nanopack/mist/core"

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
	drain.Publisher.(*drain.Mist).Publish(messages[0])
	drain.Publisher.(*drain.Mist).Publish(messages[1])

	// get "handled" message from fake mist
	msg := <-line
	mistMsg := &mist.Message{}
	err = json.Unmarshal([]byte(msg), &mistMsg)
	if err != nil {
		t.Error(fmt.Errorf("Failed to unmarshal - %v", err))
		t.FailNow()
	}

	weMsg := &logvac.Message{}
	err = json.Unmarshal([]byte(mistMsg.Data), &weMsg)
	if err != nil {
		t.Error(fmt.Errorf("Failed to unmarshal - %v", err))
		t.FailNow()
	}

	if weMsg.Content != messages[0].Content {
		t.Errorf("%q doesn't match expected out", weMsg)
		t.FailNow()
	}

	// get second "handled" message from fake mist
	msg = <-line
	err = json.Unmarshal([]byte(msg), &mistMsg)
	if err != nil {
		t.Error(fmt.Errorf("Failed to unmarshal - %v", err))
		t.FailNow()
	}

	err = json.Unmarshal([]byte(mistMsg.Data), &weMsg)
	if err != nil {
		t.Error(fmt.Errorf("Failed to unmarshal - %v", err))
		t.FailNow()
	}
	if weMsg.Content != messages[1].Content {
		t.Errorf("%q doesn't match expected out", weMsg)
		t.FailNow()
	}
}

// manually configure and start internals
func mistInitialize() error {
	var err error
	drain.CleanFreq = 1
	config.LogKeep = `{"app": "1s", "deploy":0}`
	config.LogKeep = `{"app": "1s", "deploy":0, "a":"1m", "aa":"1h", "b":"1d", "c":"1w", "d":"1y", "e":"1"}`
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt("ERROR"))
	config.DbAddress = "boltdb:///tmp/boltdbTest/logvac.bolt"

	err = fakeMistStart()
	if err != nil {
		return err
	}

	// initialize logvac
	logvac.Init()

	// initialize publisher
	// Doing broke publisher
	config.PubAddress = "~!@#$%^&*()"
	drain.Init()
	drain.Archiver.(*drain.BoltArchive).Close()

	config.PubAddress = "127.0.0.1:2445"
	drain.Init()
	drain.Archiver.(*drain.BoltArchive).Close()

	// Doing real publisher
	config.PubAddress = "mist://127.0.0.1:2445"
	err = drain.Init()
	return err
}

func fakeMistStart() error {
	line = make(chan string)
	stop = make(chan bool)

	serverSocket, err := net.Listen("tcp", PubAddress)
	if err != nil {
		return err
	}

	// listen and handle
	go func() {
		for {
			select {
			default:
				conn, err := serverSocket.Accept()
				if err != nil {
					return
				}
				// handleConnection
				go func(c net.Conn) {
					r := bufio.NewReader(c)
					for {
						words, err := r.ReadString('\n')
						if err != nil && err != io.EOF {
							return
						}

						line <- strings.TrimSuffix(words, "\n")
					}
				}(conn)
			case <-stop:
				return
			}
		}
	}()

	return nil
}
