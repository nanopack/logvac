package drain

import (
	"encoding/json"

	mistCore "github.com/nanopack/mist/core"

	"github.com/nanopack/logvac/config"
	"github.com/nanopack/logvac/core"
)

type Mist struct {
	mist mistCore.Client
}

func NewMistClient(address string) (*Mist, error) {
	c, err := mistCore.NewRemoteClient(address)
	if err != nil {
		return nil, err
	}

	m := Mist{
		mist: c,
	}

	return &m, nil
}

// Init initializes a connection to mist
func (m *Mist) Init() error {

	// add drain
	logvac.AddDrain("mist", publishDrain(m))

	return nil
}

// Publish is a wrapper for the mist client Publish function
func (m *Mist) Publish(tag []string, data string) error {
	return m.mist.Publish(tag, data)
}

// publishDrain returns a Drain
func publishDrain(pubDrain PublisherDrain) logvac.Drain {
	return func(msg logvac.Message) {
		tags := []string{"log", msg.Type}
		severities := []string{"trace", "debug", "info", "warn", "error", "fatal"}
		tags = append(tags, severities[:((msg.Priority+1)%6)]...)
		data, err := json.Marshal(msg)
		if err != nil {
			config.Log.Error("PublishDrain failed to marshal message")
			return
		}
		pubDrain.Publish(tags, string(data))
	}
}
