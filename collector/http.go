// Copyright (c) 2016 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

package collector

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/jcelliott/lumber"

	"github.com/nanopack/logvac/config"
	"github.com/nanopack/logvac/core"
	"github.com/nanopack/logvac/drain"
)

type Http struct{}

// create and return a http handler that can be dropped into an api.
func GenerateHttpCollector(kind string) http.HandlerFunc {
	headerName := "X-" + kind + "-Id"
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}
		var level string
		if r.Header.Get("X-Log-Level") == "" {
			level = "INFO"
		}
		logLevel := lumber.LvlInt(level)
		header := r.Header.Get(headerName)
		if header == "" {
			header = kind
		}
		config.Log.Trace("Header: %v, LogLevel: %v, Body: %v", header, logLevel, string(body))
		logvac.Publish(header, logLevel, string(body))
	}
}

func GenerateArchiveEndpoint(archive drain.ArchiverDrain) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()
		name := query.Get("kind")
		if name == "" {
			name = "app"
		}
		offset := query.Get("offset")
		if offset == "" {
			offset = "0"
		}
		limit := query.Get("limit")
		if limit == "" {
			limit = "100"
		}
		level := query.Get("level")
		if level == "" {
			level = "INFO"
		}
		config.Log.Trace("name: %v, offset: %v, limit: %v, level: %v", name, offset, limit, level)
		logLevel := lumber.LvlInt(level)
		realOffset, err := strconv.Atoi(offset)
		if err != nil {
			res.WriteHeader(500)
			res.Write([]byte("bad offset"))
			return
		}
		realLimit, err := strconv.Atoi(limit)
		if err != nil {
			res.WriteHeader(500)
			res.Write([]byte("bad limit"))
			return
		}
		slices, err := archive.Slice(name, uint64(realOffset), uint64(realLimit), logLevel)
		if err != nil {
			res.WriteHeader(500)
			res.Write([]byte(err.Error()))
			return
		}
		body, err := json.Marshal(slices)
		if err != nil {
			res.WriteHeader(500)
			res.Write([]byte(err.Error()))
			return
		}
		config.Log.Trace("Body: %s", string(body))

		res.WriteHeader(200)
		res.Write(append(body, byte('\n')))
	}
}
