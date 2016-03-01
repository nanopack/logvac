package main

import (
	"github.com/boltdb/bolt"
	"github.com/jcelliott/lumber"
	"github.com/nanopack/mist/core"
	"github.com/spf13/cobra"

	"github.com/nanopack/logvac/api"
	"github.com/nanopack/logvac/authenticator"
	"github.com/nanopack/logvac/config"
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
			if configFile != "" {
				config.Setup(configFile)
			}
			serverStart()
		},
	}
	config.AddFlags(&command)

	command.Flags().BoolVarP(&server, "server", "s", false, "Run as server")
	command.Flags().StringVarP(&configFile, "configFile", "", "", "config file location for server")

	command.Execute()
}

func serverStart() {
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt(config.LogLevel))
	authenticator.Setup()
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

	config.Log.Debug("Listening on https://%v udp://%v\n", config.HttpAddress, config.UdpAddress)

	// udpCollector, err := logvac.SyslogTCPStart("app", config.UdpAddress, logVac)
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

	collectHandler := logVac.GenerateHttpCollector("deploy")
	retreiveHandler := logvac.GenerateArchiveEndpoint(DB)

	err = api.Start(collectHandler, retreiveHandler)
	if err != nil {
		panic(err)
	}
}
