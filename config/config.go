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
	DbAddress  = "file:///tmp/logvac.bolt"

	// authenticator
	AuthAddress = "" // address or file location of auth backend ('boltdb:///var/db/logvac.bolt' or 'postgresql://127.0.0.1')

	// other
	LogLevel = "info"
	Token    = "secret"
	Log      lumber.Logger
)

func AddFlags(cmd *cobra.Command) {
	// collectors
	cmd.Flags().StringVarP(&ListenHttp, "listen-http", "", ListenHttp, "API listen address (same endpoint for http log collection)")
	cmd.Flags().StringVarP(&ListenUdp, "listen-udp", "", ListenUdp, "UDP log collection endpoint")
	cmd.Flags().StringVarP(&ListenTcp, "listen-tcp", "", ListenTcp, "TCP log collection endpoint")

	// drains
	cmd.Flags().StringVarP(&PubAddress, "pub-address", "", PubAddress, "Log publisher (mist) address")
	cmd.Flags().StringVarP(&DbAddress, "db-address", "", DbAddress, "Log storage address")

	// authenticator
	cmd.Flags().StringVarP(&AuthAddress, "auth-address", "", AuthAddress, "Address or file location of authentication db. ('boltdb:///var/db/logvac.bolt' or 'postgresql://127.0.0.1')")

	// other
	cmd.Flags().StringVarP(&LogLevel, "log-level", "", LogLevel, "LogLevel")
	cmd.Flags().StringVarP(&Token, "token", "", Token, "Token security")
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
