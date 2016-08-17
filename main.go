// Simple, lightweight, api-driven log aggregation service with realtime push capabilities and historical persistence.
//
// To start logvac as a server, simply run:
//
//  logvac -s
//
// For more specific usage information, refer to the help doc `logvac -h`:
//
//  Usage:
//    logvac [flags]
//    logvac [command]
//
//  Available Commands:
//    add-token   Add http publish/subscribe authentication token
//    export      Export http publish/subscribe authentication tokens
//    import      Import http publish/subscribe authentication tokens
//
//  Flags:
//    -A, --auth-address string   Address or file location of authentication db. ('boltdb:///var/db/logvac.bolt' or 'postgresql://127.0.0.1') (default "boltdb:///var/db/log-auth.bolt")
//    -c, --config-file string    config file location for server
//    -C, --cors-allow string     Sets the 'Access-Control-Allow-Origin' header (default "*")
//    -d, --db-address string     Log storage address (default "boltdb:///var/db/logvac.bolt")
//    -i, --insecure              Don't use TLS (used for testing)
//    -a, --listen-http string    API listen address (same endpoint for http log collection) (default "127.0.0.1:1234")
//    -t, --listen-tcp string     TCP log collection endpoint (default "127.0.0.1:1235")
//    -u, --listen-udp string     UDP log collection endpoint (default "127.0.0.1:1234")
//    -k, --log-keep string       Age or number of logs to keep per type '{"app":"2w", "deploy": 10}' (int or X(m)in, (h)our,  (d)ay, (w)eek, (y)ear) (default "{\"app\":\"2w\"}")
//    -l, --log-level string      Level at which to log (default "info")
//    -L, --log-type string       Default type to apply to incoming logs (commonly used: app|deploy) (default "app")
//    -p, --pub-address string    Log publisher (mist) address ("mist://127.0.0.1:1445")
//    -P, --pub-auth string       Log publisher (mist) auth token
//    -s, --server                Run as server
//    -T, --token string          Administrative token to add/remove 'X-USER-TOKEN's used to pub/sub via http (default "secret")
//    -v, --version               Print version info and exit
//
package main

import (
	"fmt"
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

	// Logvac provides the logvac cli/server functionality
	Logvac = &cobra.Command{
		Use:               "logvac",
		Short:             "logvac logging server",
		Long:              ``,
		PersistentPreRunE: readConfig,
		PreRunE:           preFlight,
		RunE:              startLogvac,
		SilenceErrors:     true,
		SilenceUsage:      true,
	}

	// version information (populated by go linker)
	// -ldflags="-X main.tag=${tag} -X main.commit=${commit}"
	tag    string
	commit string
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

func readConfig(ccmd *cobra.Command, args []string) error {
	if err := config.ReadConfigFile(configFile); err != nil {
		fmt.Printf("Error: %v\n", err)
		return err
	}
	return nil
}

func preFlight(ccmd *cobra.Command, args []string) error {
	if config.Version {
		fmt.Printf("logvac %s (%s)\n", tag, commit)
		return fmt.Errorf("")
	}

	if !config.Server {
		ccmd.HelpFunc()(ccmd, args)
		return fmt.Errorf("")
	}
	return nil
}

func startLogvac(ccmd *cobra.Command, args []string) error {
	// initialize logger
	lumber.Level(lumber.LvlInt(config.LogLevel)) // for clients using lumber too
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt(config.LogLevel))

	// initialize logvac
	logvac.Init()

	// setup authenticator
	err := authenticator.Init()
	if err != nil {
		config.Log.Fatal("Authenticator failed to initialize - %v", err)
		return err
	}

	// initialize drains
	err = drain.Init()
	if err != nil {
		config.Log.Fatal("Drain failed to initialize - %v", err)
		return err
	}

	// initializes collectors
	err = collector.Init()
	if err != nil {
		config.Log.Fatal("Collector failed to initialize - %v", err)
		return err
	}

	err = api.Start(collector.CollectHandler)
	if err != nil {
		config.Log.Fatal("Api failed to initialize - %v", err)
		return err
	}

	return nil
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
