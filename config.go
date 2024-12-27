package syncip

import (
	"net"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

type Config struct {
	Interval         time.Duration
	IPServiceURL     string
	Resolver         *net.Resolver
	DNSZoneName      string
	DNSRecordName    string
	CloudflareAPI    *cloudflare.API
	CloudflareZoneID *cloudflare.ResourceContainer
	HealthzAddr      string
	Metrics          *Metrics
}

func (c *Config) FullyQualifiedDomainName() string {
	return strings.Join([]string{c.DNSRecordName, c.DNSZoneName}, ".")
}
