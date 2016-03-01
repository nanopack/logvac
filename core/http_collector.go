package logvac

import (
	"io/ioutil"
	"net/http"

	"github.com/jcelliott/lumber"

	"github.com/nanopack/logvac/config"
)

// create and return a http handler that can be dropped into an api.
func (l *Logvac) GenerateHttpCollector(kind string) http.HandlerFunc {
	headerName := "X-" + kind + "-Id"
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}
		var level string
		if r.Header.Get("X-Log-Level") == "" {
			level = "INFO"
		}
		logLevel := lumber.LvlInt(level)
		header := r.Header.Get(headerName)
		if header == "" {
			header = kind
		}
		config.Log.Trace("Header: %v, LogLevel: %v, Body: %v", header, logLevel, string(body))
		l.Publish(header, logLevel, string(body))
	}
}
