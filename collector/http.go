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
	"strings"
	"time"

	"github.com/jcelliott/lumber"

	"github.com/nanopack/logvac/config"
	"github.com/nanopack/logvac/core"
	"github.com/nanopack/logvac/drain"
)

type Http struct{}

// create and return a http handler that can be dropped into an api.
func GenerateHttpCollector() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			res.WriteHeader(500)
			return
		}

		var msg logvac.Message
		err = json.Unmarshal(body, &msg)
		if err != nil {
			if !strings.Contains(err.Error(), "invalid character") {
				res.WriteHeader(400)
				res.Write([]byte(err.Error()))
				return
			}

			// keep body as "message" and make up priority
			msg.Content = string(body)
			msg.Priority = 2
			msg.Type = "http-raw" // todo: default to MsgType instead?
		}

		if msg.Type == "" {
			msg.Type = config.MsgType
		}
		msg.Time = time.Now()
		msg.UTime = msg.Time.UnixNano()

		config.Log.Trace("Message: %+v", msg)
		logvac.WriteMessage(msg)

		res.WriteHeader(200)
		res.Write([]byte("success!\n"))
	}
}

func GenerateArchiveEndpoint(archive drain.ArchiverDrain) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()

		host := query.Get("id")
		tag := query.Get("tag")

		kind := query.Get("type")
		if kind == "" {
			kind = config.MsgType //"app"
		}
		start := query.Get("start")
		if start == "" {
			start = "0"
		}
		limit := query.Get("limit")
		if limit == "" {
			limit = "100"
		}
		level := query.Get("level")
		if level == "" {
			level = "TRACE"
		}
		config.Log.Trace("type: %v, start: %v, limit: %v, level: %v, id: %v, tag: %v", kind, start, limit, level, host, tag)
		logLevel := lumber.LvlInt(level)
		realOffset, err := strconv.Atoi(start)
		if err != nil {
			res.WriteHeader(500)
			res.Write([]byte("bad start offset"))
			return
		}
		realLimit, err := strconv.Atoi(limit)
		if err != nil {
			res.WriteHeader(500)
			res.Write([]byte("bad limit"))
			return
		}
		slices, err := archive.Slice(kind, host, tag, uint64(realOffset), uint64(realLimit), logLevel)
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

		res.WriteHeader(200)
		res.Write(append(body, byte('\n')))
	}
}
