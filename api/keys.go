package api

import (
	"net/http"
	"github.com/nanopack/logvac/authenticator"
)

func addKey(rw http.ResponseWriter, req *http.Request) {
	err := authenticator.Add(req.Header.Get("X-LOGVAC-KEY"))
	if err != nil {
		rw.WriteHeader(404)
	}
}

func removeKey(rw http.ResponseWriter, req *http.Request) {
	err := authenticator.Remove(req.Header.Get("X-LOGVAC-KEY"))
	if err != nil {
		rw.WriteHeader(404)
	}
}