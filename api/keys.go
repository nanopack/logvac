package api

import (
	"github.com/nanopack/logvac/authenticator"
	"net/http"
)

func addKey(rw http.ResponseWriter, req *http.Request) {
	err := authenticator.Add(req.Header.Get("X-LOGVAC-KEY"))
	if err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
	}
}

func removeKey(rw http.ResponseWriter, req *http.Request) {
	err := authenticator.Remove(req.Header.Get("X-LOGVAC-KEY"))
	if err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
	}
}
