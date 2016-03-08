package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/jcelliott/lumber"
	"github.com/spf13/cobra"
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
	LogKeep  = `{"app":"2w"}` // LogType and expire (X(m)in, (h)our,  (d)ay, (w)eek, (y)ear) (1, 10, 100 == keep up to that many)
	LogType  = "app"
	LogLevel = "info"
	Token    = "secret"
	Log      lumber.Logger
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
}

// todo: use viper
func Setup(configFile string) {
	Log.Prefix("[logvac]")
	config := map[string]string{
		"listenHttp": ListenHttp,
		"listenUdp":  ListenUdp,
		"pubAddress": PubAddress,
		"logLevel":   LogLevel,
		"dbAddress":  DbAddress,
		"token":      Token,
	}

	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		Log.Error("unalbe to read file: %v", err)
		return
	}
	if err := json.Unmarshal(b, &config); err != nil {
		Log.Error("unable to parse json config: %v", err)
		return
	}
	ListenHttp = config["listenHttp"]
	ListenUdp = config["listenUdp"]
	PubAddress = config["pubAddress"]
	LogLevel = config["logLevel"]
	DbAddress = config["dbAddress"]
	Token = config["token"]
}
