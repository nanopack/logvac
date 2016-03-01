package main

import (
	"fmt"
	"os"

	"github.com/jcelliott/lumber"
	"github.com/spf13/cobra"

	"github.com/nanopack/logvac/api"
	"github.com/nanopack/logvac/authenticator"
	"github.com/nanopack/logvac/collector"
	"github.com/nanopack/logvac/config"
	"github.com/nanopack/logvac/core"
	"github.com/nanopack/logvac/drain"
)

var (
	configFile string
	portFile   string

	exportCommand = &cobra.Command{
		Use:   "export",
		Short: "Export authenticators",
		Long:  ``,

		Run: exportLogvac,
	}

	importCommand = &cobra.Command{
		Use:   "import",
		Short: "Import authenticators",
		Long:  ``,

		Run: importLogvac,
	}
)

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

	exportCommand.Flags().StringVarP(&portFile, "file", "f", "", "Export file location")
	importCommand.Flags().StringVarP(&portFile, "file", "f", "", "Import file location")

	command.AddCommand(exportCommand)
	command.AddCommand(importCommand)

	// initialize logger *(only benefit is export/import command) todo: just don't use config.Log
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt(config.LogLevel))

	command.Execute()
}

func serverStart() {
	// initialize logger
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt(config.LogLevel))

	// initialize logvac
	logvac.Init()

	// setup authenticator
	err := authenticator.Init()
	if err != nil {
		config.Log.Fatal("Authenticator failed to initialize - %v", err)
		os.Exit(1)
	}

	// initialize drains
	err = drain.Init()
	if err != nil {
		config.Log.Fatal("Drain failed to initialize - %v", err)
		os.Exit(1)
	}

	// initializes collectors
	err = collector.Init()
	if err != nil {
		config.Log.Fatal("Collector failed to initialize - %v", err)
		os.Exit(1)
	}

	err = api.Start(collector.CollectHandler, collector.RetreiveHandler)
	if err != nil {
		config.Log.Fatal("Api failed to initialize - %v", err)
		os.Exit(1)
	}
}

func exportLogvac(ccmd *cobra.Command, args []string) {
	// authenticator.Export()
	fmt.Printf("File to export: %v\n", portFile)
	config.Log.Debug("File to export: %v\n", portFile)
	return
}

func importLogvac(ccmd *cobra.Command, args []string) {
	// authenticator.Import()
	fmt.Printf("File to import: %v\n", portFile)
	config.Log.Debug("File to import: %v\n", portFile)
	return
}
