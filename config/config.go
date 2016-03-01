package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/jcelliott/lumber"
	"github.com/spf13/cobra"
)

var (
	HttpAddress = ":1234"
	UdpAddress  = ":1234"
	TcpAddress  = ":1235"
	MistAddress = ""
	LogLevel    = "info"
	DbPath      = "/tmp/logvac.bolt"
	AuthType    = "none"
	AuthConfig  = ""
	Token       = "secret"
	Log         lumber.Logger
)

func AddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&HttpAddress, "http-address", "", HttpAddress, "[server] HttpListenAddress")
	cmd.PersistentFlags().StringVarP(&UdpAddress, "udp-address", "", UdpAddress, "[server] UDPListenAddress")
	cmd.PersistentFlags().StringVarP(&UdpAddress, "tcp-address", "", TcpAddress, "[server] TCPListenAddress")
	cmd.PersistentFlags().StringVarP(&MistAddress, "mist-address", "", MistAddress, "[server] MistAddress")
	cmd.PersistentFlags().StringVarP(&LogLevel, "log-level", "", LogLevel, "[server] LogLevel")
	cmd.PersistentFlags().StringVarP(&DbPath, "db-path", "", DbPath, "[server] DbPath")
	cmd.PersistentFlags().StringVarP(&AuthType, "auth-type", "", AuthType, "[server] AuthType")
	cmd.PersistentFlags().StringVarP(&AuthConfig, "auth-config", "", AuthConfig, "[server] AuthConfig")
	cmd.PersistentFlags().StringVarP(&Token, "token", "", Token, "Token security")
}

func Setup(configFile string) {
	Log.Prefix("[logvac]")
	config := map[string]string{
		"httpAddress": HttpAddress,
		"udpAddress":  UdpAddress,
		"mistAddress": MistAddress,
		"logLevel":    LogLevel,
		"dbPath":      DbPath,
		"token":       Token,
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
	HttpAddress = config["httpAddress"]
	UdpAddress = config["udpAddress"]
	MistAddress = config["mistAddress"]
	LogLevel = config["logLevel"]
	DbPath = config["dbPath"]
}
