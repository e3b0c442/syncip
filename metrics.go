package syncip

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	*http.Server
	SyncCounter          prometheus.Counter
	SyncSucceededCounter prometheus.Counter
	SyncFailedCounter    prometheus.Counter
	SyncNoChangeCounter  prometheus.Counter
	SyncUpdatedCounter   prometheus.Counter
}

func InitMetrics(addr string) *Metrics {

	metrics := &Metrics{
		SyncCounter: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "syncip_ip_sync_count",
				Help: "Number of IP checks performed",
			},
		),
		SyncSucceededCounter: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "syncip_ip_sync_succeeded",
				Help: "Number of IP checks that succeeded",
			},
		),
		SyncFailedCounter: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "syncip_ip_sync_failed",
				Help: "Number of IP checks that failed",
			},
		),
		SyncNoChangeCounter: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "syncip_ip_sync_no_change",
				Help: "Number of IP checks that resulted in no change",
			},
		),
		SyncUpdatedCounter: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "syncip_ip_sync_updated",
				Help: "Number of IP checks that resulted in an update",
			},
		),
	}

	prometheus.MustRegister(
		metrics.SyncCounter,
		metrics.SyncSucceededCounter,
		metrics.SyncFailedCounter,
		metrics.SyncNoChangeCounter,
		metrics.SyncUpdatedCounter,
	)

	metrics.Server = &http.Server{
		Addr:    addr,
		Handler: promhttp.Handler(),
	}

	return metrics
}
