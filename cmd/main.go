package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/e3b0c442/syncip"
)

var (
	logLevel           string
	interval           string
	resolverIP         string
	cloudflareAPIToken string
	metricsAddr        string
)
var config = &syncip.Config{}

func init() {
	flag.StringVar(&logLevel, "log-level", envOrDefault("LOG_LEVEL", "info"), "log level")
	flag.StringVar(&interval, "refresh-interval", envOrDefault("REFRESH_INTERVAL", "10s"), "refresh interval")
	flag.StringVar(&config.IPServiceURL, "ip-service-url", envOrDefault("IP_SERVICE_URL", "https://api.ipify.org"), "ip service url")
	flag.StringVar(&resolverIP, "resolver-ip", envOrDefault("RESOLVER_IP", "1.1.1.1"), "resolver ip")
	flag.StringVar(&config.DNSZoneName, "dns-zone-name", envOrDefault("DNS_ZONE_NAME", ""), "zone name")
	flag.StringVar(&config.DNSRecordName, "dns-record-name", envOrDefault("DNS_RECORD_NAME", "@"), "record name")
	flag.StringVar(&cloudflareAPIToken, "cloudflare-api-token", envOrDefault("CLOUDFLARE_API_TOKEN", ""), "cloudflare api token")
	flag.StringVar(&config.HealthzAddr, "healthz-addr", envOrDefault("HEALTHZ_ADDR", ":8080"), "healthz addr")
	flag.StringVar(&metricsAddr, "metrics-addr", envOrDefault("METRICS_ADDR", ":58288"), "metrics addr")
}

func main() {
	flag.Parse()
	var err error

	var level slog.Level
	if err = level.UnmarshalText([]byte(logLevel)); err != nil {
		log.Fatal(err)
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})))

	config.Interval, err = time.ParseDuration(interval)
	if err != nil {
		log.Fatal(err)
	}

	ip := net.ParseIP(resolverIP)
	if ip == nil {
		log.Fatal("resolver-ip is invalid")
	}

	config.Resolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 1 * time.Second,
			}
			return d.DialContext(ctx, "udp", resolverIP+":53")
		},
	}

	if config.DNSZoneName == "" {
		log.Fatal("DNS zone name is required")
	}

	config.CloudflareAPI, err = cloudflare.NewWithAPIToken(cloudflareAPIToken)
	if err != nil {
		log.Fatal(err)
	}

	zoneID, err := config.CloudflareAPI.ZoneIDByName(config.DNSZoneName)
	if err != nil {
		log.Fatal(err)
	}
	config.CloudflareZoneID = cloudflare.ZoneIdentifier(zoneID)

	config.Metrics = syncip.InitMetrics(metricsAddr)

	log.Fatal(syncip.Run(config))
}

func envOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
