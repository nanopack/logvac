package drain

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/nanopack/logvac/core"
)

// Datadog drain implements the publisher interface for publishing logs to datadog.
type Datadog struct {
	connManager *ConnectionManager
	Key         string         // api key
}

// NewDatadogClient creates a new mist publisher
func NewDatadogClient(key string) (*Datadog, error) {
	_, err := net.ResolveTCPAddr("tcp", "intake.logs.datadoghq.com:10514")
	if err != nil {
		return nil, fmt.Errorf("Failed to resolve datadog address - %s", err.Error())
	}

	cm := NewConnectionManager("intake.logs.datadoghq.com", 10514)
	conn := cm.NewConnection()
	cm.conn = conn

	return &Datadog{Key: key, connManager: cm}, nil
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
func (p *Datadog) Publish(msg logvac.Message) {
	msg.PubTries++

	if p.connManager.conn == nil {
		fmt.Println("Redialing datadog")
		p.connManager.conn = p.connManager.NewConnection() // blocks until a new conn is ready
	}

	var ms []byte
	if len(msg.Raw) > 4 {
		ms = append(append([]byte(p.Key+" "), msg.Raw[4:]...), []byte("\n")...)
	} else {
		ms = append(append([]byte(p.Key+" "), msg.Raw...), []byte("\n")...)
	}

	_, err := p.connManager.conn.Write(ms)
	if err != nil {
		fmt.Printf("Failed writing log - %s %d\n", err.Error(), msg.PubTries)
		p.connManager.CloseConnection(p.connManager.conn)
		p.connManager.conn = nil
		if msg.PubTries <= 3 {
			time.Sleep(2 * time.Second)
			p.Publish(msg)
		}
	}
}

// Close closes the connection to datadog.
func (p *Datadog) Close() error {
	if p.connManager.conn == nil {
		return nil
	}
	return p.connManager.conn.Close()
}

// Adapted from datadog-log-agent
const (
	backoffSleepTimeUnit = 2  // in seconds
	maxBackoffSleepTime  = 30 // in seconds
	timeout              = 20 * time.Second
)

// A ConnectionManager manages connections
type ConnectionManager struct {
	connectionString string
	serverName       string

	mutex   sync.Mutex
	retries int

	conn net.Conn
}

// NewConnectionManager returns an initialized ConnectionManager
func NewConnectionManager(ddUrl string, ddPort int) *ConnectionManager {
	return &ConnectionManager{
		connectionString: fmt.Sprintf("%s:%d", ddUrl, ddPort),
		serverName:       ddUrl,

		mutex: sync.Mutex{},
	}
}

// NewConnection returns an initialized connection to the intake.
// It blocks until a connection is available
func (cm *ConnectionManager) NewConnection() net.Conn {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for {
		if cm.conn != nil {
			return cm.conn
		}
		fmt.Println("Connecting to the backend:", cm.connectionString)
		cm.retries += 1
		outConn, err := net.DialTimeout("tcp", cm.connectionString, timeout)
		if err != nil {
			fmt.Println(err)
			cm.backoff()
			continue
		}

		cm.retries = 0
		go cm.handleServerClose(outConn)
		return outConn
	}
}

// CloseConnection closes a connection on the client side
func (cm *ConnectionManager) CloseConnection(conn io.Closer) {
	conn.Close()
}

// handleServerClose lets the connection manager detect when a connection
// has been closed by the server, and closes it for the client.
func (cm *ConnectionManager) handleServerClose(conn net.Conn) {
	for {
		buff := make([]byte, 1)
		_, err := conn.Read(buff)
		if err == io.EOF {
			cm.CloseConnection(conn)
			return
		} else if err != nil {
			fmt.Println(err)
			return
		}
	}
}

// backoff lets the connection mananger sleep a bit
func (cm *ConnectionManager) backoff() {
	backoffDuration := backoffSleepTimeUnit * cm.retries
	if backoffDuration > maxBackoffSleepTime {
		backoffDuration = maxBackoffSleepTime
	}
	timer := time.NewTimer(time.Second * time.Duration(backoffDuration))
	<-timer.C
}
