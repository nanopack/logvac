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
	tokenName  string

	exportCommand = &cobra.Command{
		Use:   "export",
		Short: "Export http publish/subscribe authentication tokens",
		Long:  ``,

		Run: exportLogvac,
	}

	importCommand = &cobra.Command{
		Use:   "import",
		Short: "Import http publish/subscribe authentication tokens",
		Long:  ``,

		Run: importLogvac,
	}

	addKeyCommand = &cobra.Command{
		Use:   "add-token",
		Short: "Add http publish/subscribe authentication token",
		Long:  ``,

		Run: addKey,
	}

	Logvac = &cobra.Command{
		Use:   "logvac",
		Short: "logvac logging server",
		Long:  ``,

		Run: startLogvac,
	}
)

func main() {
	Logvac.Flags().StringVarP(&configFile, "config-file", "c", "", "config file location for server")
	Logvac.AddCommand(exportCommand)
	Logvac.AddCommand(importCommand)
	Logvac.AddCommand(addKeyCommand)

	config.AddFlags(Logvac)
	exportCommand.Flags().StringVarP(&portFile, "file", "f", "", "Export file location")
	importCommand.Flags().StringVarP(&portFile, "file", "f", "", "Import file location")
	addKeyCommand.Flags().StringVarP(&tokenName, "token", "t", "", "Authentication token for http publish/subscribe")

	Logvac.Execute()
}

// func serverStart() {
func startLogvac(ccmd *cobra.Command, args []string) {
	if err := config.ReadConfigFile(configFile); err != nil {
		config.Log.Fatal("Failed to read config - %v", err)
		os.Exit(1)
	}

	if !config.Server {
		ccmd.HelpFunc()(ccmd, args)
		return
	}
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

func addKey(ccmd *cobra.Command, args []string) {
	err := authenticator.Init()
	if err != nil {
		config.Log.Fatal("Authenticator failed to initialize - %v", err)
		os.Exit(1)
	}

	err = authenticator.Add(tokenName)
	if err != nil {
		config.Log.Fatal("Failed to add token '%v' - %v", tokenName, err)
	}
}
