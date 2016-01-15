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
	config.Log.Info("Api Listening on %s", config.HttpAddress)

	return nanoauth.ListenAndServeTLS(config.HttpAddress, config.Token, router, "/")
}

// handleRequest add a bit of logging 
func handleRequest(fn http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

		config.Log.Debug(`
Request:
--------------------------------------------------------------------------------
%+v

`, req)

		//
		fn(rw, req)

		config.Log.Debug(`
Response:
--------------------------------------------------------------------------------
%+v

`, rw)
	}
}

// verify that the token is allowed throught the authenticator
func verify(fn http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		key := req.Header.Get("X-LOGVAC-KEY")
		if !authenticator.Valid(key) {
			rw.WriteHeader(404)
			return
		}
		fn(rw, req)
	}
}
