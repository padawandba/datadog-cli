package tags

import (
	"github.com/padawandba/datadog-cli/internal/platform/console"
)

// SimplifiedHostTags represents a simplified view of host tags
type SimplifiedHostTags struct {
	Host string   `json:"host"`
	Tags []string `json:"tags"`
}

// FormatHostTags formats host tags for display
func FormatHostTags(formatter *console.Formatter, host string, tags []string) error {
	// Create a simplified representation
	simplifiedTags := SimplifiedHostTags{
		Host: host,
		Tags: tags,
	}

	// Use the formatter to display the simplified tags
	return formatter.Format(simplifiedTags)
}

// SimplifiedTagsBySource represents tags grouped by source
type SimplifiedTagsBySource struct {
	Source string   `json:"source"`
	Tags   []string `json:"tags"`
}

// FormatTagsBySource formats tags by source for display
func FormatTagsBySource(formatter *console.Formatter, tagsBySource map[string][]string) error {
	// Convert the map to a slice of SimplifiedTagsBySource for better display
	simplifiedTags := make([]SimplifiedTagsBySource, 0, len(tagsBySource))
	
	for source, tags := range tagsBySource {
		simplifiedTags = append(simplifiedTags, SimplifiedTagsBySource{
			Source: source,
			Tags:   tags,
		})
	}

	// Use the formatter to display the simplified tags
	return formatter.Format(simplifiedTags)
}

// FormatAllTags formats a map of tags to hosts for display
func FormatAllTags(formatter *console.Formatter, tagsToHosts map[string][]string) error {
	// Convert the map to a more readable format
	simplifiedTags := make([]SimplifiedTagsBySource, 0, len(tagsToHosts))
	
	for tag, hosts := range tagsToHosts {
		simplifiedTags = append(simplifiedTags, SimplifiedTagsBySource{
			Source: tag,
			Tags:   hosts,
		})
	}
	
	// Use the formatter to display the simplified tags
	return formatter.Format(simplifiedTags)
} 