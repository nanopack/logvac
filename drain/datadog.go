package drain

// import (
// 	"fmt"
// 	"net"
// 	"time"
// 
// 	"github.com/DataDog/datadog-agent/pkg/logs/sender"
// 
// 	"github.com/nanopack/logvac/core"
// )
// 
// // Datadog drain implements the publisher interface for publishing logs to datadog.
// type Datadog struct {
// 	connManager *sender.ConnectionManager
// 	Conn        net.Conn
// 	Key         string // datadog api key
// }
// 
// // NewDatadogClient creates a new mist publisher
// func NewDatadogClient(key string) (*Datadog, error) {
// 	_, err := net.ResolveTCPAddr("tcp", "intake.logs.datadoghq.com:10514")
// 	if err != nil {
// 		return nil, fmt.Errorf("Failed to resolve datadog address - %s", err.Error())
// 	}
// 
// 	cm := sender.NewConnectionManager("intake.logs.datadoghq.com", 10514, true)
// 	conn := cm.NewConnection()
// 
// 	return &Datadog{Key: key, connManager: cm, Conn: conn}, nil
// }
// 
// // Init initializes a connection to mist
// func (p Datadog) Init() error {
// 
// 	// add drain
// 	logvac.AddDrain("datadog", p.Publish)
// 
// 	return nil
// }
// 
// // Publish utilizes mist's Publish to "drain" a log message
// func (p *Datadog) Publish(msg logvac.Message) {
// 	msg.PubTries++
// 
// 	if p.Conn == nil {
// 		fmt.Println("Redialing datadog")
// 		p.Conn = p.connManager.NewConnection() // doesn't block (don't goroutine call to Publish)
// 	}
// 
// 	var ms []byte
// 	if len(msg.Raw) > 4 {
// 		ms = append(append([]byte(p.Key+" "), msg.Raw[4:]...), []byte("\n")...)
// 	} else {
// 		ms = append(append([]byte(p.Key+" "), msg.Raw...), []byte("\n")...)
// 	}
// 
// 	_, err := p.Conn.Write(ms)
// 	if err != nil {
// 		fmt.Printf("Failed writing log - %s %d\n", err.Error(), msg.PubTries)
// 		p.connManager.CloseConnection(p.Conn)
// 		p.Conn = nil
// 		if msg.PubTries <= 3 {
// 			time.Sleep(2 * time.Second)
// 			p.Publish(msg)
// 		}
// 	}
// }
// 
// // Close closes the connection to datadog.
// func (p *Datadog) Close() error {
// 	p.connManager.CloseConnection(p.Conn)
// 	return nil
// }
