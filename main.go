// -*- mode: go; tab-width: 2; indent-tabs-mode: 1; st-rulers: [70] -*-
// vim: ts=4 sw=4 ft=lua noet
//--------------------------------------------------------------------
// @author Daniel Barney <daniel@nanobox.io>
// Copyright (C) Pagoda Box, Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly
// prohibited. Proprietary and confidential
//
// @doc
//
// @end
// Created :   24 August 2015 by Daniel Barney <daniel@nanobox.io>
//--------------------------------------------------------------------
package main

import (
	"bitbucket.org/nanobox/na-api"
	"github.com/boltdb/bolt"
	"github.com/jcelliott/lumber"
	"github.com/pagodabox/golang-mist"
	"github.com/pagodabox/nanobox-config"
	"github.com/pagodabox/nanobox-logtap"
	logtapApi "github.com/pagodabox/nanobox-logtap/api"
	"github.com/pagodabox/nanobox-logtap/archive"
	"github.com/pagodabox/nanobox-logtap/collector"
	"github.com/pagodabox/nanobox-logtap/drain"
	"os"
	"strings"
)

func main() {
	configFile := ""
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		configFile = os.Args[1]
	}

	defaults := map[string]string{
		"httpAddress": ":1234",
		"udpAddress":  ":1234",
		"mistAddress": "127.0.0.1:1234",
		"logLevel":    "info",
		"dbPath":      "/tmp/logtap.bolt",
	}

	config.Load(defaults, configFile)
	config := config.Config

	level := lumber.LvlInt(config["logLevel"])
	log := lumber.NewConsoleLogger(level)
	log.Prefix("[logtap]")

	mist, err := mist.NewRemoteClient(config["mistAddress"])
	if err != nil {
		panic(err)
	}

	log.Debug("[BOLTDB]Opening at %v\n", config["dbPath"])
	db, err := bolt.Open(config["dbPath"], 0600, nil)
	if err != nil {
		panic(err)
	}

	DB := &archive.BoltArchive{
		DB:            db,
		MaxBucketSize: 10000, // this should be configurable
	}

	log.Debug("Listening on http://%v udp://%v\n", config["httpAddress"], config["udpAddress"])
	logTap := logtap.New(log)

	udpCollector, err := collector.SyslogUDPStart("app", config["udpAddress"], logTap)
	defer udpCollector.Close()
	if err != nil {
		panic(err)
	}

	collectHandler := collector.GenerateHttpCollector("deploy", logTap)
	api.Router.Post("/", api.TraceRequest(collectHandler))

	logTap.AddDrain("mist", drain.AdaptPublisher(mist))

	logTap.AddDrain("historical", DB.Write)

	retreiveHandler := logtapApi.GenerateArchiveEndpoint(DB)
	api.Router.Get("/", api.TraceRequest(retreiveHandler))

	api.Start(config["httpAddress"])
}
