package syncip

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run(config *Config) error {
	slog.Info("starting syncip service")
	slog.Debug("configuration", "interval", config.Interval, "ip-service-url", config.IPServiceURL)

	ticker := time.Tick(config.Interval)

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	go config.Metrics.ListenAndServe()
	go http.ListenAndServe(":8080", nil)

	for {
		select {
		case <-ticker:
			config.Metrics.SyncCounter.Inc()
			if err := SyncIP(config); err != nil {
				config.Metrics.SyncFailedCounter.Inc()
				slog.Error(err.Error())
				continue
			}
			config.Metrics.SyncSucceededCounter.Inc()
		case <-term:
			slog.Info("received sigterm, exiting")
			return nil
		}
	}
}
