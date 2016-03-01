package main

import (
	"github.com/jcelliott/lumber"
	"github.com/spf13/cobra"

	"github.com/nanopack/logvac/api"
	"github.com/nanopack/logvac/authenticator"
	"github.com/nanopack/logvac/collector"
	"github.com/nanopack/logvac/config"
	"github.com/nanopack/logvac/core"
	"github.com/nanopack/logvac/drain"
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
	// initialize logger
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt(config.LogLevel))

	// setup authenticator
	authenticator.Setup()

	// initialize logvac
	logvac.Init()

	// initialize drains
	err := drain.Init()
	if err != nil {
		panic(err)
	}

	// initializes collectors
	err = collector.Init()
	if err != nil {
		panic(err)
	}

	err = api.Start(collector.CollectHandler, collector.RetreiveHandler)
	if err != nil {
		panic(err)
	}
}
