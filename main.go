package main

import (
	"io"
	"os"
	"syscall"

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
	command.Flags().StringVarP(&configFile, "config-file", "c", "", "config file location for server")

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
	err := authenticator.Init()
	if err != nil {
		config.Log.Fatal("Authenticator failed to initialize - %v", err)
		os.Exit(1)
	}

	var exportWriter io.Writer
	if portFile != "" {
		exportWriter, err = os.Create(portFile)
		if err != nil {
			config.Log.Fatal("Failed to open file - %v", err)
		}
	} else {
		exportWriter = os.NewFile(uintptr(syscall.Stdout), "/dev/stdout") // stdout
	}

	err = authenticator.ExportLogvac(exportWriter)
	if err != nil {
		config.Log.Fatal("Failed to export - %v", err)
	}
}

func importLogvac(ccmd *cobra.Command, args []string) {
	err := authenticator.Init()
	if err != nil {
		config.Log.Fatal("Authenticator failed to initialize - %v", err)
		os.Exit(1)
	}

	var importReader io.Reader
	if portFile != "" {
		importReader, err = os.Open(portFile)
		if err != nil {
			config.Log.Fatal("Failed to open file - %v", err)
		}
	} else {
		importReader = os.NewFile(uintptr(syscall.Stdin), "/dev/stdin") // stdin
	}

	err = authenticator.ImportLogvac(importReader)
	if err != nil {
		config.Log.Fatal("Failed to import - %v", err)
	}
}
