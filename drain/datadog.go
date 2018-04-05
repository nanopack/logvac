package drain

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/nanopack/logvac/core"
)

// Datadog drain implements the publisher interface for publishing logs to datadog.
type Datadog struct {
	Conn io.WriteCloser // connection to forward logs through
	Key  string         // api key
}

// NewDatadogClient creates a new mist publisher
func NewDatadogClient(key string) (*Datadog, error) {
	_, err := net.ResolveTCPAddr("tcp", "intake.logs.datadoghq.com:10514")
	if err != nil {
		return nil, fmt.Errorf("Failed to resolve datadog address - %s", err.Error())
	}

	conn, err := net.DialTimeout("tcp", "intake.logs.datadoghq.com:10514", 20*time.Second)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial datadog - %s", err.Error())
	}

	go handleServerClose(conn)
	return &Datadog{Conn: conn, Key: key}, nil
}

// Init initializes a connection to mist
func (p Datadog) Init() error {

	// add drain
	logvac.AddDrain("datadog", p.Publish)

	return nil
}

// handleServerClose detects when the server closes a connection then closes it for the client.
func handleServerClose(conn net.Conn) {
	for {
		buff := make([]byte, 1)
		_, err := conn.Read(buff)
		if err == io.EOF {
			conn.Close()
			return
		} else if err != nil {
			return
		}
	}
}

// Publish utilizes mist's Publish to "drain" a log message
func (p Datadog) Publish(msg logvac.Message) {
	msg.PubTries++
	if p.Conn == nil {
		fmt.Println("Redialing datadog")
		conn, err := net.DialTimeout("tcp", "intake.logs.datadoghq.com:10514", 20*time.Second)
		if err != nil {
			if msg.PubTries <= 3 {
				time.Sleep(2 * time.Second)
				p.Publish(msg)
			}
			// return nil, fmt.Errorf("Failed to dial datadog - %s", err.Error())
		}
		p.Conn = conn
		go handleServerClose(conn)
	}

	// ms := append(append([]byte(p.Key+" "), append(addExtra(), msg.Raw[4:]...)...), []byte("\n")...)
	var ms []byte
	if len(msg.Raw) > 4 {
		ms = append(append([]byte(p.Key+" "), msg.Raw[4:]...), []byte("\n")...)
	} else {
		ms = append(append([]byte(p.Key+" "), msg.Raw...), []byte("\n")...)
	}

	fmt.Printf("Sending %s", ms)
	// _, err := p.Conn.Write(append(append([]byte(p.Key+" "), append(addExtra(), msg.Raw[4:]...)...), []byte("\n")...))
	_, err := p.Conn.Write(ms)
	if err != nil {
		fmt.Printf("Failed writing log - %s %d\n", err.Error(), msg.PubTries)
		p.Conn.Close()
		p.Conn = nil
		if msg.PubTries <= 3 {
			time.Sleep(2 * time.Second)
			p.Publish(msg)
		}
	}
}

// Close closes the connection to datadog.
func (p *Datadog) Close() error {
	if p.Conn == nil {
		return nil
	}
	return p.Conn.Close()
}
