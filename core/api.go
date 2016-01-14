// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//
package logvac

import (
  "encoding/json"
  "github.com/jcelliott/lumber"
  "net/http"
  "strconv"
)

func GenerateArchiveEndpoint(archive Archive) http.HandlerFunc {
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

    res.WriteHeader(200)
    res.Write(body)

  }
}
