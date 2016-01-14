package main

import (
	"io/ioutil"
	"encoding/json"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/spf13/cobra"
	"github.com/gorilla/pat"
	"github.com/jcelliott/lumber"
	"github.com/nanopack/mist/core"

	"github.com/nanopack/logvac/core"
)

var configFile string

func main() {
	server := true
	command := cobra.Command{
		Use:   "logvac",
		Short: "logvac logging server",
		Long:  ``,
		Run: func(ccmd *cobra.Command, args []string) {
			if !server {
				ccmd.HelpFunc()(ccmd, args)
				return
			}
			serverStart()
		},
	}
	command.Flags().BoolVarP(&server, "server", "s", false, "Run as server")
	command.Flags().StringVarP(&configFile, "configFile", "", "","config file location for server")

	command.Execute()
}

func serverStart() {
	
	config := map[string]string{
		"httpAddress": ":1234",
		"udpAddress":  ":1234",
		"mistAddress": "127.0.0.1:1234",
		"logLevel":    "info",
		"dbPath":      "/tmp/logvac.bolt",
	}

	log := lumber.NewConsoleLogger(lumber.LvlInt(config["logLevel"]))
	log.Prefix("[logtap]")

	if configFile != "" {
		b, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Error("unalbe to read file: %v", err)
		} else {
			if err := json.Unmarshal(b, &config); err != nil {
				log.Error("unable to parse json config: %v", err)
			}
		}
	}

	// update the log leve incase it was re configured
	log.Level(lumber.LvlInt(config["logLevel"]))

	mist, err := mist.NewRemoteClient(config["mistAddress"])
	if err != nil {
		panic(err)
	}

	log.Debug("[BOLTDB]Opening at %v\n", config["dbPath"])
	db, err := bolt.Open(config["dbPath"], 0600, nil)
	if err != nil {
		panic(err)
	}

	DB := &logvac.BoltArchive{
		DB:            db,
		MaxBucketSize: 10000, // this should be configurable
	}

	log.Debug("Listening on http://%v udp://%v\n", config["httpAddress"], config["udpAddress"])
	logVac := logvac.New(log)

	udpCollector, err := logvac.SyslogUDPStart("app", config["udpAddress"], logVac)
	defer udpCollector.Close()
	if err != nil {
		panic(err)
	}

	logVac.AddDrain("mist", logvac.PublishDrain(mist))

	logVac.AddDrain("historical", DB.Write)

	collectHandler := logvac.GenerateHttpCollector("deploy", logVac)
	retreiveHandler := logvac.GenerateArchiveEndpoint(DB)

	router := pat.New()

	router.Post("/", collectHandler)
	router.Get("/", retreiveHandler)

	err = http.ListenAndServe(config["httpAddress"], router)
	if err != nil {
		panic(err)
	}
}
