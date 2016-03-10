package config

import (
	"path/filepath"

	"github.com/jcelliott/lumber"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// collectors
	ListenHttp = "127.0.0.1:1234"
	ListenUdp  = "127.0.0.1:1234"
	ListenTcp  = "127.0.0.1:1235"

	// drains
	PubAddress = ""
	DbAddress  = "boltdb:///tmp/logvac.bolt"

	// authenticator
	AuthAddress = "" // address or file location of auth backend ('boltdb:///var/db/logvac.bolt' or 'postgresql://127.0.0.1')

	// other
	LogKeep  = `{"app":"2w"}` // LogType and expire (X(m)in, (h)our,  (d)ay, (w)eek, (y)ear) (1, 10, 100 == keep up to that many) // todo: maybe map[string]interface
	LogType  = "app"
	LogLevel = "info"
	Token    = "secret"
	Log      lumber.Logger
	Server   = false
)

func AddFlags(cmd *cobra.Command) {
	// collectors
	cmd.Flags().StringVarP(&ListenHttp, "listen-http", "a", ListenHttp, "API listen address (same endpoint for http log collection)")
	cmd.Flags().StringVarP(&ListenUdp, "listen-udp", "u", ListenUdp, "UDP log collection endpoint")
	cmd.Flags().StringVarP(&ListenTcp, "listen-tcp", "t", ListenTcp, "TCP log collection endpoint")

	// drains
	cmd.Flags().StringVarP(&PubAddress, "pub-address", "p", PubAddress, "Log publisher (mist) address")
	cmd.Flags().StringVarP(&DbAddress, "db-address", "d", DbAddress, "Log storage address")

	// authenticator
	cmd.PersistentFlags().StringVarP(&AuthAddress, "auth-address", "A", AuthAddress, "Address or file location of authentication db. ('boltdb:///var/db/logvac.bolt' or 'postgresql://127.0.0.1')")

	// other
	cmd.Flags().StringVarP(&LogKeep, "log-keep", "k", LogKeep, "Age or number of logs to keep per type `{\"app\":\"2w\", \"deploy\": 10}` (int or X(m)in, (h)our,  (d)ay, (w)eek, (y)ear)")
	cmd.Flags().StringVarP(&LogLevel, "log-level", "l", LogLevel, "Level at which to log")
	cmd.Flags().StringVarP(&LogType, "log-type", "L", LogType, "Default type to apply to incoming logs (commonly used: app|deploy)")
	cmd.Flags().StringVarP(&Token, "token", "T", Token, "Token security")
	cmd.Flags().BoolVarP(&Server, "server", "s", Server, "Run as server")

	Log = lumber.NewConsoleLogger(lumber.LvlInt(LogLevel))
}

func ReadConfigFile(configFile string) error {
	if configFile == "" {
		return nil
	}

	// Set defaults to whatever might be there already
	viper.SetDefault("listen-http", ListenHttp)
	viper.SetDefault("listen-udp", ListenUdp)
	viper.SetDefault("listen-tcp", ListenTcp)
	viper.SetDefault("pub-address", PubAddress)
	viper.SetDefault("db-address", DbAddress)
	viper.SetDefault("auth-address", AuthAddress)
	viper.SetDefault("log-keep", LogKeep)
	viper.SetDefault("log-level", LogLevel)
	viper.SetDefault("log-type", LogType)
	viper.SetDefault("token", Token)
	viper.SetDefault("server", Server)

	filename := filepath.Base(configFile)
	viper.SetConfigName(filename[:len(filename)-len(filepath.Ext(filename))])
	viper.AddConfigPath(filepath.Dir(configFile))

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	// Set values. Config file will override commandline
	ListenHttp = viper.GetString("listen-http")
	ListenUdp = viper.GetString("listen-udp")
	ListenTcp = viper.GetString("listen-tcp")
	PubAddress = viper.GetString("pub-address")
	DbAddress = viper.GetString("db-address")
	AuthAddress = viper.GetString("auth-address")
	LogKeep = viper.GetString("log-keep")
	LogLevel = viper.GetString("log-level")
	LogType = viper.GetString("log-type")
	Token = viper.GetString("token")
	Server = viper.GetBool("server")

	return nil
}
