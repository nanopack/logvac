// Package collector initializes tcp, udp, and http servers for collecting logs.
package collector

import (
	"net/http"

	"github.com/nanopack/logvac/config"
)

var (
	// CollectHandler handles the posting of logs via http. It is passed to
	// the api on start.
	CollectHandler http.HandlerFunc
)

// Init initializes the tcp, udp, and http servers, if configured
func Init() error {
	// todo: handle similar to mist listeners
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
