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
	// todo: needs update
	AuthType   = "none" // type of backend to check against ('boltdb' or 'postgresql')
	AuthConfig = ""     // address or file location of auth backend

	// other
	LogLevel = "info"
	Token    = "secret"
	Log      lumber.Logger
)

func AddFlags(cmd *cobra.Command) {
	// collectors
	cmd.PersistentFlags().StringVarP(&ListenHttp, "listen-http", "", ListenHttp, "API listen address (same endpoint for http log collection)")
	cmd.PersistentFlags().StringVarP(&ListenUdp, "listen-udp", "", ListenUdp, "UDP log collection endpoint")
	cmd.PersistentFlags().StringVarP(&ListenTcp, "listen-tcp", "", ListenTcp, "TCP log collection endpoint")

	// drains
	cmd.PersistentFlags().StringVarP(&PubAddress, "pub-address", "", PubAddress, "Log publisher (mist) address")
	cmd.PersistentFlags().StringVarP(&DbAddress, "db-address", "", DbAddress, "Log storage address")

	// authenticator
	cmd.PersistentFlags().StringVarP(&AuthType, "auth-type", "", AuthType, "Type of backend to authenticate against ('boltdb' or 'postgresql')")
	cmd.PersistentFlags().StringVarP(&AuthConfig, "auth-config", "", AuthConfig, "Address or file location of auth-type")

	// other
	cmd.PersistentFlags().StringVarP(&LogLevel, "log-level", "", LogLevel, "LogLevel")
	cmd.PersistentFlags().StringVarP(&Token, "token", "", Token, "Token security")
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
