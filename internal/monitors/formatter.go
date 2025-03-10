package monitors

import (
	"fmt"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/padawandba/datadog-cli/internal/platform/console"
)

// FormatMonitors formats a slice of Monitor structs for display
func FormatMonitors(formatter *console.Formatter, monitors []Monitor) error {
	// Use the formatter to display the monitors
	// (Our Monitor struct is already simplified)
	return formatter.Format(monitors)
}

// FormatAPIMonitors formats a slice of Datadog API Monitor structs for display
func FormatAPIMonitors(formatter *console.Formatter, monitors []datadogV1.Monitor) error {
	// Convert the monitors to our simplified format for display
	simplifiedMonitors := make([]Monitor, 0, len(monitors))
	for _, monitor := range monitors {
		simplifiedMonitors = append(simplifiedMonitors, simplifyMonitor(monitor))
	}

	// Use the formatter to display the simplified monitors
	return formatter.Format(simplifiedMonitors)
}

// simplifyMonitor converts a Datadog Monitor to our simplified Monitor
func simplifyMonitor(monitor datadogV1.Monitor) Monitor {
	simplified := Monitor{
		ID:      monitor.GetId(),
		Name:    monitor.GetName(),
		Status:  string(monitor.GetOverallState()),
		Type:    string(monitor.GetType()),
		Query:   monitor.GetQuery(),
		Message: monitor.GetMessage(),
		Tags:    monitor.GetTags(),
	}

	// Handle options
	if monitor.HasOptions() {
		options := monitor.GetOptions()
		optionsMap := make(map[string]interface{})

		// Add relevant options fields
		if options.HasEscalationMessage() {
			optionsMap["escalation_message"] = options.GetEscalationMessage()
		}

		if options.HasNotifyNoData() {
			optionsMap["notify_no_data"] = options.GetNotifyNoData()
		}

		if options.HasNotifyAudit() {
			optionsMap["notify_audit"] = options.GetNotifyAudit()
		}

		if options.HasRenotifyInterval() {
			optionsMap["renotify_interval"] = options.GetRenotifyInterval()
		}

		if options.HasTimeoutH() {
			optionsMap["timeout_h"] = options.GetTimeoutH()
		}

		if options.HasSilenced() {
			// Format silenced entries for better readability
			silenced := options.GetSilenced()
			if len(silenced) > 0 {
				silencedStr := make([]string, 0, len(silenced))
				for scope, until := range silenced {
					var untilStr string
					if until > 0 {
						untilTime := time.Unix(until, 0)
						untilStr = fmt.Sprintf("until %s", untilTime.Format(time.RFC3339))
					} else {
						untilStr = "indefinitely"
					}
					
					silencedStr = append(silencedStr, fmt.Sprintf("%s: %s", scope, untilStr))
				}
				optionsMap["silenced"] = strings.Join(silencedStr, ", ")
			}
		}

		simplified.Options = optionsMap
	}

	return simplified
} 