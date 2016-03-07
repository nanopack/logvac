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
func GenerateHttpCollector() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}
		// todo make sure we parse the (expected) json body
		var msg logvac.Message
		err = json.Unmarshal(body, &msg)
		if err != nil {
			// todo: keep body as "message" and make up other Message.bits
			w.WriteHeader(400)
			return
		}

		if msg.Type == "" {
			msg.Type = config.MsgType
		}

		config.Log.Trace("Message: %+v", msg)
		logvac.WriteMessage(msg)
		// logvac.Publish(header, logLevel, string(body))
	}
}

func GenerateArchiveEndpoint(archive drain.ArchiverDrain) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()

		host := query.Get("id")
		tag := query.Get("tag")

		kind := query.Get("type")
		if kind == "" {
			kind = "app"
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
