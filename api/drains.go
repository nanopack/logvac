package api

import (
	"fmt"
	"net/http"

	"github.com/nanopack/logvac/core"
	"github.com/nanopack/logvac/drain"
)

func addDrain(rw http.ResponseWriter, req *http.Request) {
	drainer := logvac.Drain{}

	err := parseBody(req, &drainer)
	if err != nil {
		rw.WriteHeader(400)
		rw.Write([]byte(fmt.Sprintf("Failed to parse drain - %s", err.Error())))
		return
	}

	err = drain.AddDrain(drainer)
	if err != nil {
		rw.WriteHeader(400)
		rw.Write([]byte(fmt.Sprintf("Failed to add drain - %s", err.Error())))
		return
	}

	rw.WriteHeader(200)
	rw.Write([]byte("success!\n"))
}

func updateDrain(rw http.ResponseWriter, req *http.Request) {
	drainType := req.URL.Query().Get(":drainType")

	drainer := logvac.Drain{}

	err := parseBody(req, &drainer)
	if err != nil {
		rw.WriteHeader(400)
		rw.Write([]byte(fmt.Sprintf("Failed to parse drain - %s", err.Error())))
		return
	}

	if drainer.Type != drainType {
		err = drain.RemoveDrain(drainType)
		if err != nil {
			rw.WriteHeader(400)
			rw.Write([]byte(fmt.Sprintf("Failed to remove drain - %s", err.Error())))
			return
		}
	}

	err = drain.AddDrain(drainer)
	if err != nil {
		rw.WriteHeader(400)
		rw.Write([]byte(fmt.Sprintf("Failed to add drain - %s", err.Error())))
		return
	}

	rw.WriteHeader(200)
	rw.Write([]byte("success!\n"))
}

func deleteDrain(rw http.ResponseWriter, req *http.Request) {
	drainType := req.URL.Query().Get(":drainType")

	err := drain.RemoveDrain(drainType)
	if err != nil {
		rw.WriteHeader(400)
		rw.Write([]byte(fmt.Sprintf("Failed to remove drain - %s", err.Error())))
		return
	}

	rw.WriteHeader(200)
	rw.Write([]byte("success!\n"))
}

func listDrains(rw http.ResponseWriter, req *http.Request) {
	drains := drain.ListDrains()

	rw.WriteHeader(200)
	rw.Write([]byte(fmt.Sprintf("%q\n", drains)))
}
