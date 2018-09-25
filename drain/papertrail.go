package drain

import (
	"fmt"
	"io"
	"net"

	"github.com/nanopack/logvac/config"
	"github.com/nanopack/logvac/core"
)

// Papertrail drain implements the publisher interface for publishing logs to papertrail.
type Papertrail struct {
	ID		string				 // the app id or name
	Conn 	io.WriteCloser // connection to forward logs through
}

// NewPapertrailClient creates a new mist publisher
func NewPapertrailClient(uri, id string) (*Papertrail, error) {
	addr, err := net.ResolveUDPAddr("udp", uri)
	if err != nil {
		return nil, fmt.Errorf("Failed to resolve papertrail address - %s", err.Error())
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial papertrail - %s", err.Error())
	}
	
	config.Log.Info("Connection to papertrail endpoint established")

	return &Papertrail{Conn: conn, ID: id}, nil
}

// Init initializes a connection to mist
func (p *Papertrail) Init() error {

	// add drain
	logvac.AddDrain("papertrail", p.Publish)

	return nil
}

// Publish utilizes mist's Publish to "drain" a log message
func (p *Papertrail) Publish(msg logvac.Message) {
	date := fmt.Sprintf("%s %02d %02d:%02d:%02d", 
		msg.Time.Month().String()[:3],
		msg.Time.Day(),
		msg.Time.Hour(),
		msg.Time.Minute(),
		msg.Time.Second())
	id := fmt.Sprintf("%s.%s", p.ID, msg.Id)
	tag := msg.Tag[0]
	
	// the final message
	message := fmt.Sprintf("<%d>%s %s %s: %s\n", 
		msg.Priority, date, id, tag, msg.Content)
	
	config.Log.Info("%s", message)
	p.Conn.Write([]byte(message))
}

// Close closes the connection to papertrail.
func (p *Papertrail) Close() error {
	if p.Conn == nil {
		return nil
	}
	return p.Conn.Close()
}
