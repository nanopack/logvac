package main

import (
	"github.com/boltdb/bolt"
	"github.com/spf13/cobra"
	"github.com/nanopack/mist/core"

	"github.com/nanopack/logvac/api"
	"github.com/nanopack/logvac/core"
	"github.com/nanopack/logvac/config"
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
			if configFile != "" {
				config.Setup(configFile)
			}
			serverStart()
		},
	}
	config.AddFlags(&command)
	command.Flags().BoolVarP(&server, "server", "s", false, "Run as server")
	command.Flags().StringVarP(&configFile, "configFile", "", "","config file location for server")

	command.Execute()
}

func serverStart() {
	logVac := logvac.New(config.Log)

	config.Log.Debug("[BOLTDB]Opening at %v\n", config.DbPath)
	db, err := bolt.Open(config.DbPath, 0600, nil)
	if err != nil {
		panic(err)
	}

	DB := &logvac.BoltArchive{
		DB:            db,
		MaxBucketSize: 10000, // this should be configurable
	}

	config.Log.Debug("Listening on http://%v udp://%v\n", config.HttpAddress, config.UdpAddress)

	udpCollector, err := logvac.SyslogUDPStart("app", config.UdpAddress, logVac)
	defer udpCollector.Close()
	if err != nil {
		panic(err)
	}

	if config.MistAddress != "" {
		mist, err := mist.NewRemoteClient(config.MistAddress)
		if err != nil {
			panic(err)
		}
		logVac.AddDrain("mist", logvac.PublishDrain(mist))
	}


	logVac.AddDrain("historical", DB.Write)

	collectHandler := logvac.GenerateHttpCollector("deploy", logVac)
	retreiveHandler := logvac.GenerateArchiveEndpoint(DB)


	err = api.Start(collectHandler, retreiveHandler)
	if err != nil {
		panic(err)
	}
}
