package collector

import (
	"net/http"

	"github.com/nanopack/logvac/config"
	"github.com/nanopack/logvac/drain"
)

var (
	CollectHandler  http.HandlerFunc
	RetreiveHandler http.HandlerFunc
)

func Init() error {
	if config.ListenTcp != "" {
		err := SyslogTCPStart("app", config.ListenTcp)
		if err != nil {
			return err
		}
		config.Log.Info("Collector listening on tcp://%v...", config.ListenTcp)
	}

	if config.ListenUdp != "" {
		err := SyslogUDPStart("app", config.ListenUdp)
		if err != nil {
			return err
		}
		config.Log.Info("Collector listening on udp://%v...", config.ListenUdp)
	}

	if config.ListenUdp != "" {
		CollectHandler = GenerateHttpCollector("deploy")
		RetreiveHandler = GenerateArchiveEndpoint(drain.Archiver)
		config.Log.Debug("Collector listening on https://%v...", config.ListenHttp)
	}

	return nil
}
