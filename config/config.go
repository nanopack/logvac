package config

import (
	"io/ioutil"
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/jcelliott/lumber"
)

var (
	HttpAddress = ":1234"
	UdpAddress = ":1234"
	MistAddress = ""
	LogLevel = "info"
	DbPath  = "/tmp/logvac.bolt"
	AuthType = "none"
	AuthConfig = ""
	Token = "secret"
	Log = lumber.NewConsoleLogger(lumber.INFO)
)

func AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&HttpAddress, "http-address", "", HttpAddress, "[server] HttpListenAddress")
	cmd.Flags().StringVarP(&UdpAddress, "udp-address", "", UdpAddress, "[server] UDPListenAddress")
	cmd.Flags().StringVarP(&MistAddress, "mist-address", "", MistAddress, "[server] MistAddress")
	cmd.Flags().StringVarP(&LogLevel, "log-level", "", LogLevel, "[server] LogLevel")
	cmd.Flags().StringVarP(&DbPath, "db-path", "", DbPath, "[server] DbPath")
	cmd.Flags().StringVarP(&AuthType, "auth-type", "", AuthType, "[server] AuthType")
	cmd.Flags().StringVarP(&AuthConfig, "auth-config", "", AuthConfig, "[server] AuthConfig")
	cmd.PersistentFlags().StringVarP(&Token, "token", "", Token, "Token security")
}


func Setup(configFile string) {
	Log.Prefix("[logtap]")
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

