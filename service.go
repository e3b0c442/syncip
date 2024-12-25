package syncip

import (
	"log/slog"
	"time"
)

func Run(config *Config) error {
	slog.Info("starting syncip service")
	slog.Debug("configuration", "interval", config.Interval, "ip-service-url", config.IPServiceURL)

	ticker := time.Tick(config.Interval)

	for {
		select {
		case <-ticker:
			if err := SyncIP(config); err != nil {
				slog.Error(err.Error())
			}
		}
	}
}
