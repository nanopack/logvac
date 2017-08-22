package drain

import (
	"encoding/json"
	"fmt"

	mistCore "github.com/nanopack/mist/clients"

	"github.com/nanopack/logvac/config"
	"github.com/nanopack/logvac/core"
)

type pthinger interface {
	Publish(tags []string, data string) error
	Close()
}

// Mist is a mist publisher
type Mist struct {
	address string   // address for redialing
	mist    pthinger // mistCore.tcpClient
}

// NewMistClient creates a new mist publisher
func NewMistClient(address string) (*Mist, error) {
	c, err := mistCore.New(address, config.PubAuth)
	if err != nil {
		return nil, err
	}

	m := Mist{
		address: address,
		mist:    c,
	}

	return &m, nil
}

// Init initializes a connection to mist
func (m Mist) Init() error {

	// add drain
	logvac.AddDrain("mist", m.Publish)

	return nil
}

// Publish utilizes mist's Publish to "drain" a log message
func (m *Mist) Publish(msg logvac.Message) {
	// id is included in the tags
	tags := []string{"log", msg.Type}
	tags = append(tags, msg.Tag...)
	// remove blank tags
cleanTags:
	for i := range tags {
		if tags[i] == "" {
			tags = append(tags[:i], tags[i+1:]...)
			goto cleanTags
		}
	}

	severities := []string{"trace", "debug", "info", "warn", "error", "fatal"}
	tags = append(tags, severities[:((msg.Priority+1)%6)]...)
	data, err := json.Marshal(msg)
	if err != nil {
		config.Log.Error("PublishDrain failed to marshal message")
		return
	}
	config.Log.Trace("Mist publisher publishing...")
	// todo: still possible to lose a message (until connection dies)
	err = m.mist.Publish(tags, string(data))
	if err != nil {
		// re-establish connection and publish
		err2 := m.retryPublish(tags, string(data))
		if err2 != nil {
			config.Log.Error("Failed to Publish - %s - %s", err, err2)
		}
	}
}

func (m *Mist) retryPublish(tags []string, data string) error {
	c, err := mistCore.New(m.address, config.PubAuth)
	if err != nil {
		return fmt.Errorf("Failed to redial mist - %s", err)
	}

	m.mist = c

	return m.mist.Publish(tags, data)
}

// Close cleanly closes the mist client
func (m *Mist) Close() error {
	m.mist.Close()
	return nil
}
