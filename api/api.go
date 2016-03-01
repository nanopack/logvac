package api

import (
	"net/http"

	"github.com/gorilla/pat"
	"github.com/nanobox-io/nanoauth"
	"github.com/nanopack/logvac/authenticator"
	"github.com/nanopack/logvac/config"
)

// start the web server with the logvac functions
func Start(collector http.HandlerFunc, retriever http.HandlerFunc) error {
	router := pat.New()

	router.Get("/add-key", handleRequest(addKey))
	router.Get("/remove-key", handleRequest(removeKey))

	router.Post("/", verify(collector))
	router.Get("/", verify(retriever))

	// blocking...
	config.Log.Info("Api Listening on https://%s...", config.ListenHttp)

	return nanoauth.ListenAndServeTLS(config.ListenHttp, config.Token, router, "/")
}

// handleRequest add a bit of logging
func handleRequest(fn http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

		fn(rw, req)

		// must be after fn if ever going to get rw.status (logging still more meaningful)
		config.Log.Trace(`%v - [%v] %v %v %v(%s) - "User-Agent: %s", "X-Nanobox-Token: %s"`,
			req.RemoteAddr, req.Proto, req.Method, req.RequestURI,
			rw.Header().Get("status"), req.Header.Get("Content-Length"),
			req.Header.Get("User-Agent"), req.Header.Get("X-Nanobox-Token"))
	}
}

// verify that the token is allowed throught the authenticator
func verify(fn http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		key := req.Header.Get("X-LOGVAC-KEY")
		if !authenticator.Valid(key) {
			rw.WriteHeader(401)
			return
		}
		fn(rw, req)
	}
}
