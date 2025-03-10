package hosts

import (
	"fmt"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/padawandba/datadog-cli/internal/platform/console"
)

// FormatHosts formats a slice of Host structs for display
func FormatHosts(formatter *console.Formatter, hosts []datadogV1.Host) error {
	// Convert the hosts to a simplified format for display
	simplifiedHosts := make([]SimplifiedHost, 0, len(hosts))
	for _, host := range hosts {
		simplifiedHosts = append(simplifiedHosts, simplifyHost(host))
	}

	// Use the formatter to display the simplified hosts
	return formatter.Format(simplifiedHosts)
}

// SimplifiedHost is a simplified representation of a Datadog Host
type SimplifiedHost struct {
	Name            string   `json:"name"`
	Aliases         []string `json:"aliases"`
	Apps            []string `json:"apps"`
	HostName        string   `json:"host_name"`
	LastReportedAt  string   `json:"last_reported_at"`
	IsMuted         bool     `json:"is_muted"`
	Up              bool     `json:"up"`
	Sources         []string `json:"sources"`
	TagsBySource    string   `json:"tags_by_source"`
}

// simplifyHost converts a Datadog Host to a SimplifiedHost
func simplifyHost(host datadogV1.Host) SimplifiedHost {
	simplified := SimplifiedHost{}
	
	// Handle Name
	if host.HasName() {
		simplified.Name = host.GetName()
	}
	
	// Handle Aliases
	if host.HasAliases() {
		simplified.Aliases = host.GetAliases()
	}
	
	// Handle Apps
	if host.HasApps() {
		simplified.Apps = host.GetApps()
	}
	
	// Handle HostName
	if host.HasHostName() {
		simplified.HostName = host.GetHostName()
	}
	
	// Handle LastReportedAt
	if host.HasLastReportedTime() {
		timestamp := host.GetLastReportedTime()
		// Convert Unix timestamp to readable time
		t := time.Unix(timestamp, 0)
		simplified.LastReportedAt = t.Format(time.RFC3339)
	}
	
	// Handle IsMuted
	if host.HasIsMuted() {
		simplified.IsMuted = host.GetIsMuted()
	}
	
	// Handle Up
	if host.HasUp() {
		simplified.Up = host.GetUp()
	}
	
	// Handle Sources
	if host.HasSources() {
		simplified.Sources = host.GetSources()
	}
	
	// Handle TagsBySource
	if host.HasTagsBySource() {
		tagsBySource := host.GetTagsBySource()
		// Format tags by source as a string
		parts := make([]string, 0, len(tagsBySource))
		for source, tags := range tagsBySource {
			parts = append(parts, fmt.Sprintf("%s:[%s]", source, strings.Join(tags, ", ")))
		}
		simplified.TagsBySource = strings.Join(parts, ", ")
	}
	
	return simplified
} 