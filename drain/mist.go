package drain

import (
	"encoding/json"

	mistCore "github.com/nanopack/mist/clients"

	"github.com/nanopack/logvac/config"
	"github.com/nanopack/logvac/core"
)

type pthinger interface {
	Publish(tags []string, data string) error
}
type Mist struct {
	mist pthinger
}

func NewMistClient(address string) (*Mist, error) {
	c, err := mistCore.New(address)
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
	config.Log.Trace("Mist publisher publishing...")
	return m.mist.Publish(tag, data)
}

// publishDrain returns a Drain
func publishDrain(pubDrain PublisherDrain) logvac.Drain {
	return func(msg logvac.Message) {
		tags := []string{"log", msg.Type, msg.Id, msg.Tag}
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
		pubDrain.Publish(tags, string(data))
	}
}
