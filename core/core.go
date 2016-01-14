package logvac

import (
	"github.com/jcelliott/lumber"
	"sync"
	"time"
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

	Archive interface {
		Slice(name string, offset, limit uint64, level int) ([]Message, error)
	}

	Drain func(Logger, Message)

	Message struct {
		Type     string
		Time     time.Time `json:"time"`
		Priority int       `json:"priority"`
		Content  string    `json:"content"`
	}

	Logvac struct {
		log    Logger
		drains map[string]drainChannels
	}

	drainChannels struct {
		send chan Message
		done chan bool
	}
)

// Establishes a new logtap object
// and makes sure it has a logger
func New(log Logger) *Logvac {
	if log == nil {
		log = lumber.NewConsoleLogger(lumber.ERROR)
	}
	return &Logvac{
		log:    log,
		drains: make(map[string]drainChannels),
	}
}

// Close logtap and remove all drains
func (l *Logvac) Close() {
	for tag := range l.drains {
		l.RemoveDrain(tag)
	}
}

// AddDrain addes a drain to the listeners and sets its logger
func (l *Logvac) AddDrain(tag string, drain Drain) {
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
				drain(l.log, msg)
			}
		}
	}()

	l.drains[tag] = channels
}

// RemoveDrain drops a drain
func (l *Logvac) RemoveDrain(tag string) {
	drain, ok := l.drains[tag]
	if ok {
		close(drain.done)
		delete(l.drains, tag)
	}
}

func (l *Logvac) Publish(kind string, priority int, content string) {
	m := Message{
		Type:     kind,
		Time:     time.Now(),
		Priority: priority,
		Content:  content,
	}
	l.WriteMessage(m)
}

// WriteMessage broadcasts to all drains in seperate go routines
// Returns once all drains have received the message, but may not have processed
// the message yet
func (l *Logvac) WriteMessage(msg Message) {
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
