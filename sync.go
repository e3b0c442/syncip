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

	// Get the record
	record, err := GetARecord(config, config.FullyQualifiedDomainName())
	if err != nil {
		return err
	}

	// Update the A record in Cloudflare
	slog.Info("updating IP address for FQDN", "fqdn", config.FullyQualifiedDomainName(), "oldip", ip, "newip", actual)
	if err = UpdateIP(config, record.ID, actual); err != nil {
		return err
	}
	slog.Info("ip updated successfully")

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

// Retrieve the record by name
func GetARecord(config *Config, name string) (cloudflare.DNSRecord, error) {
	records, ri, err := config.CloudflareAPI.ListDNSRecords(
		context.Background(),
		config.CloudflareZoneID,
		cloudflare.ListDNSRecordsParams{
			Name: name,
			Type: "A",
		},
	)

	if err != nil {
		return cloudflare.DNSRecord{}, err
	}

	switch ri.Count {
	case 0:
		return cloudflare.DNSRecord{}, fmt.Errorf("no A records found for %s", name)
	case 1:
	default:
		return cloudflare.DNSRecord{}, fmt.Errorf("multiple A records found for %s", name)
	}

	return records[0], nil
}

// UpdateIP updates the IP address for the DNS record.
func UpdateIP(config *Config, recordID, ip string) (err error) {
	_, err = config.CloudflareAPI.UpdateDNSRecord(
		context.Background(),
		config.CloudflareZoneID,
		cloudflare.UpdateDNSRecordParams{
			Type:    "A",
			Name:    config.DNSRecordName,
			ID:      recordID,
			Content: ip,
		},
	)

	return
}
