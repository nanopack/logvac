package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/pat"
	"github.com/jcelliott/lumber"
	"github.com/nanobox-io/golang-nanoauth"

	"github.com/nanopack/logvac/authenticator"
	"github.com/nanopack/logvac/config"
	"github.com/nanopack/logvac/drain"
)

// start the web server with the logvac functions
func Start(collector http.HandlerFunc) error {
	retriever := GenerateArchiveEndpoint(drain.Archiver)

	router := pat.New()

	router.Get("/add-token", handleRequest(addKey))
	router.Get("/remove-token", handleRequest(removeKey))
	router.Add("OPTIONS", "/", handleRequest(cors))

	router.Post("/", verify(handleRequest(collector)))
	router.Get("/", verify(handleRequest(retriever)))

	// blocking...
	if config.Insecure {
		config.Log.Info("Api Listening on http://%s...", config.ListenHttp)
		return http.ListenAndServe(config.ListenHttp, router)
	}
	config.Log.Info("Api Listening on https://%s...", config.ListenHttp)
	cert, _ := nanoauth.Generate("nanobox.io")
	auth := nanoauth.Auth{
		Header:      "X-ADMIN-TOKEN",
		Certificate: cert,
	}
	return auth.ListenAndServeTLS(config.ListenHttp, config.Token, router, "/")
}

func cors(rw http.ResponseWriter, req *http.Request) {
	// todo: allow something more specific than "*"
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Methods", "GET, POST")
	rw.Header().Set("Access-Control-Allow-Headers", "X-ADMIN-TOKEN, X-AUTH-TOKEN")
	rw.WriteHeader(200)
	rw.Write([]byte("success!\n"))
}

// handleRequest add a bit of logging
func handleRequest(fn http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		// todo: allow something more specific than "*"
		rw.Header().Set("Access-Control-Allow-Origin", "*")
		rw.Header().Set("Access-Control-Allow-Methods", "GET, POST")
		rw.Header().Set("Access-Control-Allow-Headers", "X-ADMIN-TOKEN, X-AUTH-TOKEN")

		fn(rw, req)

		// must be after req returns
		getStatus := func(trw http.ResponseWriter) string {
			r, _ := regexp.Compile("status:([0-9]*)")
			return r.FindStringSubmatch(fmt.Sprintf("%+v", trw))[1]
		}

		getWrote := func(trw http.ResponseWriter) string {
			r, _ := regexp.Compile("written:([0-9]*)")
			return r.FindStringSubmatch(fmt.Sprintf("%+v", trw))[1]
		}

		config.Log.Debug(`%v - [%v] %v %v %v(%s) - "User-Agent: %s"`,
			req.RemoteAddr, req.Proto, req.Method, req.RequestURI,
			getStatus(rw), getWrote(rw), // %v(%s)
			req.Header.Get("User-Agent"))
	}
}

// verify that the token is allowed throught the authenticator
func verify(fn http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		key := req.Header.Get("X-AUTH-TOKEN")
		// allow browsers to authenticate/fetch logs
		if key == "" {
			query := req.URL.Query()
			key = query.Get("auth")
		}
		if !authenticator.Valid(key) {
			rw.WriteHeader(401)
			return
		}
		fn(rw, req)
	}
}

// note: javascript number precision may cause unexpected results (missing logs within 100 nanosecond window)
func GenerateArchiveEndpoint(archive drain.ArchiverDrain) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()

		host := query.Get("id")
		tag := query.Get("tag")

		kind := query.Get("type")
		if kind == "" {
			kind = config.LogType // "app"
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
		realOffset, err := strconv.ParseInt(start, 0, 64)
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
		slices, err := archive.Slice(kind, host, tag, realOffset, int64(realLimit), logLevel)
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
