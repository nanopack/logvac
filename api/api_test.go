package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	os.RemoveAll("/tmp/logvacTest")

	// manually configure
	initialize()

	// start api
	go api.Start(collector.CollectHandler, collector.RetreiveHandler)
	<-time.After(1 * time.Second)
	rtn := m.Run()

	// clean test dir
	os.RemoveAll("/tmp/logvacTest")

	os.Exit(rtn)
}

// test post logs
func TestPostLogs(t *testing.T) {
	body, err := rest("POST", "/", "{\"id\":\"log-test\",\"type\":\"app\",\"message\":\"test log\"}")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if string(body) != "success!\n" {
		t.Errorf("%q doesn't match expected out", body)
		t.FailNow()
	}
	// boltdb seems to take some time committing the record (probably the speed/immediate commit tradeoff)
	time.Sleep(500 * time.Millisecond)
}

// test get logs
func TestGetLogs(t *testing.T) {
	body, err := rest("GET", "/?type=app&id=log-test&start=0&limit=1", "")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	msg := []logvac.Message{}
	err = json.Unmarshal(body, &msg)
	if err != nil {
		t.Error(fmt.Errorf("Failed to unmarshal - %v", err))
		t.FailNow()
	}

	if msg[0].Content != "test log" {
		t.Errorf("%q doesn't match expected out", body)
	}
}

// hit api and return response body
func rest(method, route, data string) ([]byte, error) {
	body := bytes.NewBuffer([]byte(data))

	req, _ := http.NewRequest(method, fmt.Sprintf("http://%s%s", config.ListenHttp, route), body)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Unable to %v %v - %v", method, route, err)
	}
	defer res.Body.Close()

	b, _ := ioutil.ReadAll(res.Body)

	return b, nil
}

// manually configure and start internals
func initialize() {
	config.Insecure = true
	config.ListenHttp = "127.0.0.1:2234"
	config.ListenTcp = "127.0.0.1:2235"
	config.ListenUdp = "127.0.0.1:2234"
	config.DbAddress = "boltdb:///tmp/logvacTest/logvac.bolt"
	config.AuthAddress = ""
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt("ERROR"))

	// initialize logvac
	logvac.Init()

	// setup authenticator
	err := authenticator.Init()
	if err != nil {
		config.Log.Fatal("Authenticator failed to initialize - %v", err)
		os.Exit(1)
	}

	// initialize drains
	err = drain.Init()
	if err != nil {
		config.Log.Fatal("Drain failed to initialize - %v", err)
		os.Exit(1)
	}

	// initializes collectors
	err = collector.Init()
	if err != nil {
		config.Log.Fatal("Collector failed to initialize - %v", err)
		os.Exit(1)
	}
}
