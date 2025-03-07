package monitors

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

// Client provides monitor-related operations
type Client struct {
	apiClient *datadog.APIClient
	ctx       context.Context
}

// NewClient creates a new monitors client
func NewClient(apiClient *datadog.APIClient, ctx context.Context) *Client {
	return &Client{
		apiClient: apiClient,
		ctx:       ctx,
	}
}

// Monitor represents a simplified Datadog monitor
type Monitor struct {
	ID      int64                  `json:"id"`
	Name    string                 `json:"name"`
	Status  string                 `json:"status"`
	Type    string                 `json:"type"`
	Query   string                 `json:"query"`
	Message string                 `json:"message,omitempty"`
	Tags    []string               `json:"tags,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// List retrieves a list of monitors
func (c *Client) List(query string, tags []string) ([]Monitor, error) {
	monitorsAPI := datadogV1.NewMonitorsApi(c.apiClient)
	
	opts := datadogV1.NewListMonitorsOptionalParameters()
	if query != "" {
		opts.WithName(query)
	}
	if len(tags) > 0 {
		opts.WithTags(strings.Join(tags, ","))
	}
	
	resp, _, err := monitorsAPI.ListMonitors(c.ctx, *opts)
	if err != nil {
		return nil, fmt.Errorf("error listing monitors: %v", err)
	}
	
	// Convert API response to our simplified Monitor type
	monitors := make([]Monitor, len(resp))
	for i, m := range resp {
		monitors[i] = Monitor{
			ID:      m.GetId(),
			Name:    m.GetName(),
			Status:  string(m.GetOverallState()),
			Type:    string(m.GetType()),
			Query:   m.GetQuery(),
			Message: m.GetMessage(),
			Tags:    m.GetTags(),
			Options: convertOptions(m.GetOptions()),
		}
	}
	
	return monitors, nil
}

// Mute mutes a monitor
func (c *Client) Mute(monitorID int64, scope string, endTime int64) error {
	monitorsAPI := datadogV1.NewMonitorsApi(c.apiClient)
	
	// Get the current monitor
	monitor, _, err := monitorsAPI.GetMonitor(c.ctx, monitorID)
	if err != nil {
		return fmt.Errorf("error getting monitor: %v", err)
	}
	
	// Create an update request
	updateReq := datadogV1.NewMonitorUpdateRequest()
	
	// Set up silencing
	silenced := make(map[string]int64)
	if scope == "" {
		// Mute the entire monitor
		silenced["*"] = endTime
	} else {
		// Mute a specific scope
		silenced[scope] = endTime
	}
	
	// Get or create options
	options := monitor.GetOptions()
	options.SetSilenced(silenced)
	
	updateReq.SetOptions(options)
	
	// Update the monitor
	_, _, err = monitorsAPI.UpdateMonitor(c.ctx, monitorID, *updateReq)
	if err != nil {
		return fmt.Errorf("error muting monitor: %v", err)
	}
	
	return nil
}

// Unmute unmutes a monitor
func (c *Client) Unmute(monitorID int64, scope string) error {
	monitorsAPI := datadogV1.NewMonitorsApi(c.apiClient)
	
	// Get the current monitor
	monitor, _, err := monitorsAPI.GetMonitor(c.ctx, monitorID)
	if err != nil {
		return fmt.Errorf("error getting monitor: %v", err)
	}
	
	// Create an update request
	updateReq := datadogV1.NewMonitorUpdateRequest()
	
	// Get current options and silenced settings
	options := monitor.GetOptions()
	silenced := options.GetSilenced()
	
	// Remove the silencing based on scope
	if scope == "" {
		// Clear all silencing
		options.SetSilenced(make(map[string]int64))
	} else if silenced != nil {
		// Remove specific scope
		delete(silenced, scope)
		options.SetSilenced(silenced)
	}
	
	updateReq.SetOptions(options)
	
	// Update the monitor
	_, _, err = monitorsAPI.UpdateMonitor(c.ctx, monitorID, *updateReq)
	if err != nil {
		return fmt.Errorf("error unmuting monitor: %v", err)
	}
	
	return nil
}

// Helper function to convert monitor options to a map
func convertOptions(options datadogV1.MonitorOptions) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Add relevant options fields
	if options.HasEscalationMessage() {
		result["escalation_message"] = options.GetEscalationMessage()
	}
	
	if options.HasNotifyNoData() {
		result["notify_no_data"] = options.GetNotifyNoData()
	}
	
	if options.HasNotifyAudit() {
		result["notify_audit"] = options.GetNotifyAudit()
	}
	
	if options.HasRenotifyInterval() {
		result["renotify_interval"] = options.GetRenotifyInterval()
	}
	
	if options.HasTimeoutH() {
		result["timeout_h"] = options.GetTimeoutH()
	}
	
	if options.HasSilenced() {
		result["silenced"] = options.GetSilenced()
	}
	
	return result
}

