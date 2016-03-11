package collector

import (
	"net/http"

	"github.com/nanopack/logvac/config"
)

var (
	CollectHandler http.HandlerFunc
)

func Init() error {
	if config.ListenTcp != "" {
		err := SyslogTCPStart(config.ListenTcp)
		if err != nil {
			return err
		}
		config.Log.Info("Collector listening on tcp://%v...", config.ListenTcp)
	}

	if config.ListenUdp != "" {
		err := SyslogUDPStart(config.ListenUdp)
		if err != nil {
			return err
		}
		config.Log.Info("Collector listening on udp://%v...", config.ListenUdp)
	}

	if config.ListenHttp != "" {
		CollectHandler = GenerateHttpCollector()
		config.Log.Info("Collector listening on http://%v...", config.ListenHttp)
	}

	return nil
}
