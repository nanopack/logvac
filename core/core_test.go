// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//
package logvac_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/jcelliott/lumber"
	"github.com/nanopack/logvac/core"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

var log = lumber.NewConsoleLogger(lumber.TRACE)

func TestBasic(test *testing.T) {
	logVac := logvac.New(log)
	defer logVac.Close()
	called := false

	testDrain := func(l logvac.Logger, msg logvac.Message) {
		called = true
	}

	console := logvac.WriteDrain(os.Stdout)
	logVac.AddDrain("testing", console)
	logVac.AddDrain("fake", testDrain)
	logVac.Publish("what is this?", lumber.DEBUG, "you should see me!")
	assert(test, called, "the drain was not called")
}

func TestBolt(test *testing.T) {
	logVac := logvac.New(log)
	defer logVac.Close()

	db, err := bolt.Open("./test.db", 0600, nil)
	assert(test, err == nil, "failed to create boltDB %v", err)
	defer func() {
		db.Close()
		os.Remove("./test.db")
	}()

	boltArchive := &logvac.BoltArchive{
		DB:            db,
		MaxBucketSize: 10, // store only 10 chunks, this is a test.
	}

	logVac.AddDrain("historical", boltArchive.Write)
	logVac.Publish("app", lumber.DEBUG, "you should see me!")

	// let the other processes finish running
	time.Sleep(100 * time.Millisecond)

	slices, err := boltArchive.Slice("app", 0, 100, lumber.DEBUG)
	assert(test, err == nil, "Slice errored %v", err)
	assert(test, len(slices) == 1, "wrong number of slices %v", slices)

	for i := 0; i < 100; i++ {
		logVac.Publish("app", lumber.DEBUG, fmt.Sprintf("log line:%v", i))
	}

	// let the other processes finish running
	time.Sleep(100 * time.Millisecond)

	slices, err = boltArchive.Slice("app", 0, 100, lumber.DEBUG)
	fmt.Println(slices)
	assert(test, err == nil, "Slice errored %v", err)
	assert(test, len(slices) == 10, "wrong number of slices %v", len(slices))

}

func TestApi(test *testing.T) {
	logVac := logvac.New(log)
	defer logVac.Close()

	db, err := bolt.Open("./test.db", 0600, nil)
	assert(test, err == nil, "failed to create boltDB %v", err)
	defer func() {
		db.Close()
		os.Remove("./test.db")
	}()

	boltArchive := &logvac.BoltArchive{
		DB:            db,
		MaxBucketSize: 10, // store only 10 chunks, this is a test.
	}

	logVac.AddDrain("historical", boltArchive.Write)
	logVac.Publish("app", lumber.INFO, "you should see me!")

	handler := logvac.GenerateArchiveEndpoint(boltArchive)

	go http.ListenAndServe("127.0.0.1:2345", handler)

	// wait for the api to be available
	time.Sleep(time.Millisecond * 10)

	res, err := http.Get("http://127.0.0.1:2345/")
	assert(test, err == nil, "%v", err)
	assert(test, res.StatusCode == 200, "bad response: %v", res)
	resBody, err := ioutil.ReadAll(res.Body)
	assert(test, err == nil, "%v", err)

	result := make([]logvac.Message, 1)
	err = json.Unmarshal(resBody, &result)
	assert(test, err == nil, "%v", err)
	assert(test, len(result) == 1, "wrong number of response lines %v", result)

	fmt.Printf("%v\n", string(resBody))
}

func TestUDPCollector(test *testing.T) {
	logVac := logvac.New(log)
	defer logVac.Close()
	success := false

	testDrain := func(l logvac.Logger, msg logvac.Message) {
		success = true
	}

	logVac.AddDrain("drain", testDrain)

	udpCollector, err := logvac.SyslogUDPStart("app", "127.0.0.1:1234", logVac)
	assert(test, err == nil, "%v", err)
	defer udpCollector.Close()

	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	assert(test, err == nil, "%v", err)

	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	assert(test, err == nil, "%v", err)

	client, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	assert(test, err == nil, "%v", err)
	defer client.Close()

	_, err = client.Write([]byte("<34>Oct 11 22:14:15 mymachine su: 'su root' failed for lonvick on /dev/pts/8"))
	assert(test, err == nil, "%v", err)

	time.Sleep(time.Millisecond * 10)
	assert(test, success, "the message was not received")
}

func TestTCPCollector(test *testing.T) {
	logVac := logvac.New(log)
	defer logVac.Close()
	success := false

	testDrain := func(l logvac.Logger, msg logvac.Message) {
		success = true
	}

	logVac.AddDrain("drain", testDrain)

	tcpCollector, err := logvac.SyslogTCPStart("app", "127.0.0.1:1234", logVac)
	assert(test, err == nil, "%v", err)
	defer tcpCollector.Close()

	client, err := net.Dial("tcp", "127.0.0.1:1234")
	defer client.Close()
	assert(test, err == nil, "%v", err)

	_, err = client.Write([]byte("This is not a standard message\n"))
	assert(test, err == nil, "%v", err)

	time.Sleep(time.Millisecond * 10)
	assert(test, success, "the message was not received")
}

func TestHTTPCollector(test *testing.T) {
	logVac := logvac.New(log)
	defer logVac.Close()
	success := false

	testDrain := func(l logvac.Logger, msg logvac.Message) {
		success = true
	}

	logVac.AddDrain("drain", testDrain)

	go logvac.StartHttpCollector("app", "127.0.0.1:1234", logVac)

	body := bytes.NewReader([]byte("this is a test"))
	res, err := http.Post("http://127.0.0.1:1234/upload", "text/plain", body)
	assert(test, res.StatusCode == 200, "bad response %v", res)
	assert(test, err == nil, "%v", err)
	time.Sleep(time.Millisecond * 10)
	assert(test, success, "the message was not received")
}

func BenchmarkLogvacOne(b *testing.B) {
	benchmarkTest(b, 1)
}

func BenchmarkLogvacTwo(b *testing.B) {
	benchmarkTest(b, 5)
}

func BenchmarkLogvacTen(b *testing.B) {
	benchmarkTest(b, 10)
}

func BenchmarkLogvacOneHundred(b *testing.B) {
	benchmarkTest(b, 100)
}

func benchmarkTest(b *testing.B, listenerCount int) {
	logVac := logvac.New(log)
	defer logVac.Close()

	group := sync.WaitGroup{}
	testDrain := func(l logvac.Logger, msg logvac.Message) {
		group.Done()
	}

	for i := 0; i < listenerCount; i++ {
		logVac.AddDrain(fmt.Sprintf("%v", i), testDrain)
	}

	group.Add(b.N * listenerCount)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logVac.Publish("app", lumber.DEBUG, "testing")
	}
	group.Wait()
}

func assert(test *testing.T, check bool, fmt string, args ...interface{}) {
	if !check {
		test.Logf(fmt, args...)
		test.FailNow()
	}
}
