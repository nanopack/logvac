package logvac

import (
	"sync"
	"time"

	"github.com/nanopack/logvac/config"
)

type (
	// Logger is a simple interface that's designed to be intentionally generic to
	// allow many different types of Logger's to satisfy its interface
	Logger interface {
		Fatal(string, ...interface{})
		Error(string, ...interface{})
		Warn(string, ...interface{})
		Info(string, ...interface{})
		Debug(string, ...interface{})
		Trace(string, ...interface{})
	}

	Message struct {
		Time     time.Time `json:"time"`
		UTime    int64     `json:"utime"`
		Id       string    `json:"id"`  // ignoreifempty?
		Tag      string    `json:"tag"` // ignoreifempty? // []string?
		Type     string    `json:"type"`
		Priority int       `json:"priority"`
		Content  string    `json:"message"`
	}

	Logvac struct {
		drains map[string]drainChannels
	}

	// Drain is a function that "drains a Message"
	Drain func(Message)

	drainChannels struct {
		send chan Message
		done chan bool
	}
)

var Vac Logvac

// Initializes a logvac object
func Init() error {
	Vac = Logvac{
		drains: make(map[string]drainChannels),
	}
	config.Log.Debug("Logvac initialized")
	return nil
}

// Close logvac and remove all drains
func Close() {
	Vac.close()
}

func (l *Logvac) close() {
	for tag := range l.drains {
		l.removeDrain(tag)
	}
}

// AddDrain adds a drain to the listeners and sets its logger
func AddDrain(tag string, drain Drain) {
	Vac.addDrain(tag, drain)
}

func (l *Logvac) addDrain(tag string, drain Drain) {
	channels := drainChannels{
		done: make(chan bool),
		send: make(chan Message),
	}

	go func() {
		for {
			select {
			case <-channels.done:
				return
			case msg := <-channels.send:
				// todo: ensure mist plays nice with goroutine
				go drain(msg)
			}
		}
	}()

	l.drains[tag] = channels
}

// RemoveDrain drops a drain
func RemoveDrain(tag string) {
	Vac.removeDrain(tag)
}

func (l *Logvac) removeDrain(tag string) {
	drain, ok := l.drains[tag]
	if ok {
		close(drain.done)
		delete(l.drains, tag)
	}
}

// WriteMessage broadcasts to all drains in seperate go routines
// Returns once all drains have received the message, but may not have processed
// the message yet
func WriteMessage(msg Message) {
	Vac.writeMessage(msg)
}

func (l *Logvac) writeMessage(msg Message) {
	// config.Log.Trace("Writing message - %v...", msg)
	group := sync.WaitGroup{}
	for _, drain := range l.drains {
		group.Add(1)
		go func(myDrain drainChannels) {
			select {
			case <-myDrain.done:
			case myDrain.send <- msg:
			}
			group.Done()
		}(drain)
	}
	group.Wait()
}
