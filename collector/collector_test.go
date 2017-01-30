// collector_test tests the syslog collectors
// (http collector is tested in api_test)
package collector_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jcelliott/lumber"

	"github.com/nanopack/logvac/api"
	"github.com/nanopack/logvac/authenticator"
	"github.com/nanopack/logvac/collector"
	"github.com/nanopack/logvac/config"
	"github.com/nanopack/logvac/core"
	"github.com/nanopack/logvac/drain"
)

func TestMain(m *testing.M) {
	// clean test dir
	os.RemoveAll("/tmp/syslogTest")

	// manually configure
	initialize()

	// start api
	config.Token = ""
	go api.Start(collector.CollectHandler)
	<-time.After(1 * time.Second)
	rtn := m.Run()

	// clean test dir
	os.RemoveAll("/tmp/syslogTest")

	os.Exit(rtn)
}

// test post logs
func TestPostLogs(t *testing.T) {
	body, err := rest("POST", "/logs", "{\"id\":\"log-test\",\"type\":\"app\",\"message\":\"test log\"}")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if string(body) != "success!\n" {
		t.Errorf("%q doesn't match expected out", body)
		t.FailNow()
	}
	// pause for travis
	time.Sleep(500 * time.Millisecond)
	body, err = rest("POST", "/logs", "another test log")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if string(body) != "success!\n" {
		t.Errorf("%q doesn't match expected out", body)
		t.FailNow()
	}
	// boltdb seems to take some time committing the record (probably the speed/immediate commit tradeoff)
	time.Sleep(time.Second)
}

// test pushing a log via udp
func TestUdp(t *testing.T) {
	ServerAddr, err := net.ResolveUDPAddr("udp", config.ListenUdp)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	client, err := net.DialUDP("udp", nil, ServerAddr)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer client.Close()

	_, err = client.Write([]byte("this would have the tag \"syslog-raw\"")) // for testing fakeParser
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	// pause for travis
	time.Sleep(time.Second)
	_, err = client.Write([]byte("<83>Mar 11 14:13:12 web2 apache[error] ello, your app is broke"))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	// pause for travis
	time.Sleep(time.Second)
}

// test pushing a log via tcp
func TestTcp(t *testing.T) {
	ServerAddr, err := net.ResolveTCPAddr("tcp", config.ListenTcp)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	client, err := net.DialTCP("tcp", nil, ServerAddr)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer client.Close()

	_, err = client.Write([]byte("<83>Mar 11 14:13:12 web2 apache[daemon] serving some sweet stuff\n"))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	// boltdb seems to take some time committing the record (probably the speed/immediate commit tradeoff)
	time.Sleep(time.Second)
}

// test get logs
func TestGetLogs(t *testing.T) {
	body, err := rest("GET", "/logs?type=app", "")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	msg := []logvac.Message{}
	err = json.Unmarshal(body, &msg)
	if err != nil {
		t.Error(fmt.Errorf("Failed to unmarshal - %s", err))
		t.FailNow()
	}

	if len(msg) != 5 || msg[3].Content != "ello, your app is broke" {
		t.Errorf("%q doesn't match expected out", body)
		t.FailNow()
	}
	if msg[4].Content != "serving some sweet stuff" {
		t.Errorf("%q doesn't match expected out", body)
		t.FailNow()
	}
}

// hit api and return response body
func rest(method, route, data string) ([]byte, error) {
	body := bytes.NewBuffer([]byte(data))

	req, _ := http.NewRequest(method, fmt.Sprintf("http://%s%s", config.ListenHttp, route), body)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Unable to %s %s - %s", method, route, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Status '200' expected, got '%d'", res.StatusCode)
	}

	b, _ := ioutil.ReadAll(res.Body)

	return b, nil
}

// manually configure and start internals
func initialize() {
	config.ListenHttp = "127.0.0.1:4234"
	config.ListenTcp = "127.0.0.1:4235"
	config.ListenUdp = "127.0.0.1:4234"
	config.DbAddress = "boltdb:///tmp/syslogTest/logvac.bolt"
	config.AuthAddress = ""
	config.Insecure = true
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt("ERROR"))

	// initialize logvac
	logvac.Init()

	// setup authenticator
	err := authenticator.Init()
	if err != nil {
		config.Log.Fatal("Authenticator failed to initialize - %s", err)
		os.Exit(1)
	}

	// initialize drains
	err = drain.Init()
	if err != nil {
		config.Log.Fatal("Drain failed to initialize - %s", err)
		os.Exit(1)
	}

	// initializes collectors
	err = collector.Init()
	if err != nil {
		config.Log.Fatal("Collector failed to initialize - %s", err)
		os.Exit(1)
	}
}
