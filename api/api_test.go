// api_test tests the api, from posting logs, to getting them (no-auth configured)
package api_test

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
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

var (
	secureHttp   string
	insecureHttp string
)

func TestMain(m *testing.M) {
	// clean test dir
	os.RemoveAll("/tmp/apiTest")

	// manually configure
	initialize()

	// start insecure api
	go api.Start(collector.CollectHandler)
	time.Sleep(time.Second)
	// start secure api
	config.Insecure = false
	config.ListenHttp = secureHttp
	go api.Start(collector.CollectHandler)
	<-time.After(time.Second)
	rtn := m.Run()

	// clean test dir
	os.RemoveAll("/tmp/apiTest")

	os.Exit(rtn)
}

// test cors
func TestCors(t *testing.T) {
	body, err := irest("OPTIONS", "/?type=app&id=log-test&start=0&limit=1", "")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if string(body) != "success!\n" {
		t.Errorf("%q doesn't match expected out", body)
		t.FailNow()
	}
}

// test adding an auth token
func TestAddToken(t *testing.T) {
	body, err := rest("GET", "/add-token", "")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if string(body) != "success!\n" {
		t.Errorf("%q doesn't match expected out", body)
		t.FailNow()
	}
}

// test post logs
func TestPostLogs(t *testing.T) {
	// secure
	body, err := rest("POST", "/", "{\"id\":\"log-test\",\"type\":\"app\",\"message\":\"test log\"}")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if string(body) != "success!\n" {
		t.Errorf("%q doesn't match expected out", body)
		t.FailNow()
	}
	// insecure
	body, err = irest("POST", "/", "{\"id\":\"log-test\",\"type\":\"app\",\"message\":\"test log\"}")
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
	body, err := irest("GET", "/?type=app&id=log-test&start=0&limit=1", "")
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
	if len(msg) != 1 || msg[0].Content != "test log" {
		t.Errorf("%q doesn't match expected out", body)
	}
	_, err = irest("GET", "/", "")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	_, err = irest("GET", "/?start=word", "")
	if err == nil || strings.Contains(err.Error(), "bad start offset") {
		t.Error("bad start is too forgiving")
		t.FailNow()
	}
	_, err = irest("GET", "/?limit=word", "")
	if err == nil || strings.Contains(err.Error(), "bad limit") {
		t.Error("bad limit is too forgiving")
		t.FailNow()
	}
	_, err = irest("GET", "/?level=word", "")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}

// test removing an auth token
func TestRemoveToken(t *testing.T) {
	body, err := rest("GET", "/remove-token", "")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if string(body) != "success!\n" {
		t.Errorf("%q doesn't match expected out", body)
		t.FailNow()
	}
}

// hit api and return response body
func irest(method, route, data string) ([]byte, error) {
	body := bytes.NewBuffer([]byte(data))

	req, _ := http.NewRequest(method, fmt.Sprintf("http://%s%s", insecureHttp, route), body)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Unable to %v %v - %v", method, route, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Status '200' expected, got '%d'", res.StatusCode)
	}

	b, _ := ioutil.ReadAll(res.Body)

	return b, nil
}

func rest(method, route, data string) ([]byte, error) {
	body := bytes.NewBuffer([]byte(data))

	req, _ := http.NewRequest(method, fmt.Sprintf("https://%s%s", secureHttp, route), body)
	req.Header.Add("X-ADMIN-TOKEN", "secret")
	req.Header.Add("X-AUTH-TOKEN", "user")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Unable to %v %v - %v", method, route, err)
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
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	secureHttp = "127.0.0.1:2236"
	insecureHttp = "127.0.0.1:2234"
	config.Insecure = true
	config.ListenHttp = "127.0.0.1:2234"
	config.ListenTcp = "127.0.0.1:2235"
	config.ListenUdp = "127.0.0.1:2234"
	config.DbAddress = "boltdb:///tmp/apiTest/logvac.bolt"
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
