package syncip

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/cloudflare/cloudflare-go"
)

func SyncIP(config *Config) error {
	// retrieve current IP
	actual, err := GetCurrentIP(config)
	if err != nil {
		return err
	}
	slog.Debug("retrieved current IP", "ip", actual)

	// DNS lookup for FQDN
	ip, err := ResolveHost(config)
	if err != nil {
		return err
	}
	slog.Debug("retrieved IP for FQDN", "fqdn", config.FullyQualifiedDomainName(), "ip", ip)

	// if IPs match, nothing to do
	if ip == actual {
		slog.Info("No IP change")
		return nil
	}

	return nil
}

// GetCurrentIP retrieves the current external IP address from this service.
func GetCurrentIP(config *Config) (string, error) {
	r, err := http.Get(config.IPServiceURL)
	if err != nil {
		return "", fmt.Errorf("failed to get current IP: %w", err)
	}

	if r.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get current IP: %s", r.Status)
	}

	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read current IP: %w", err)
	}

	return string(b), nil
}

// GetDNSRecordIP retrieves the current IP address for the DNS record.
func ResolveHost(config *Config) (string, error) {
	ips, err := config.Resolver.LookupHost(context.Background(), config.FullyQualifiedDomainName())
	if err == nil {
		switch len(ips) {
		case 0:
			err = errors.New("no IPs returned for hostname")
		case 1:
		default:
			err = fmt.Errorf("multiple IPs returned for hostname")
		}
	}
	if err != nil {
		return "", fmt.Errorf("failed to lookup IP for %s: %w", config.FullyQualifiedDomainName(), err)
	}
	return ips[0], nil
}

// UpdateIP updates the IP address for the DNS record.
func UpdateIP(config *Config, ip string) (err error) {
	_, err = config.CloudflareAPI.UpdateDNSRecord(
		context.Background(),
		config.CloudflareZoneID,
		cloudflare.UpdateDNSRecordParams{
			Type:    "A",
			Name:    config.DNSRecordName,
			Content: ip,
		},
	)

	return
}
