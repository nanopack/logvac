package api

import (
	"net/http"

	"github.com/nanopack/logvac/authenticator"
)

func addKey(rw http.ResponseWriter, req *http.Request) {
	err := authenticator.Add(req.Header.Get("X-AUTH-TOKEN"))
	if err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
	}
	rw.WriteHeader(200)
	rw.Write([]byte("success!\n"))
}

func removeKey(rw http.ResponseWriter, req *http.Request) {
	err := authenticator.Remove(req.Header.Get("X-AUTH-TOKEN"))
	if err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
	}
	rw.WriteHeader(200)
	rw.Write([]byte("success!\n"))
}
