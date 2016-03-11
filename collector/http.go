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
	"strings"
	"time"

	"github.com/nanopack/logvac/config"
	"github.com/nanopack/logvac/core"
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
			msg.Type = "http-raw" // todo: default to LogType instead?
		}

		if msg.Type == "" {
			msg.Type = config.LogType
		}
		msg.Time = time.Now()
		msg.UTime = msg.Time.UnixNano()

		// config.Log.Trace("Message: %q", msg)
		logvac.WriteMessage(msg)

		res.WriteHeader(200)
		res.Write([]byte("success!\n"))
	}
}
