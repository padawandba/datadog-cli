package hosts

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

// Client provides host-related operations
type Client struct {
	apiClient *datadog.APIClient
	ctx       context.Context
}

// NewClient creates a new hosts client
func NewClient(apiClient *datadog.APIClient, ctx context.Context) *Client {
	return &Client{
		apiClient: apiClient,
		ctx:       ctx,
	}
}

// List retrieves a list of hosts from Datadog
func (c *Client) List(filter string) ([]datadogV1.Host, error) {
	hostsAPI := datadogV1.NewHostsApi(c.apiClient)
	
	// Create optional parameters with proper initialization
	opts := datadogV1.NewListHostsOptionalParameters()
	if filter != "" {
		opts = opts.WithFilter(filter)
	}
	
	// Use proper error handling with context
	resp, httpResp, err := hostsAPI.ListHosts(c.ctx, *opts)
	if err != nil {
		// Include HTTP response details in error if available
		if httpResp != nil {
			return nil, fmt.Errorf("error listing hosts (status: %d): %v", httpResp.StatusCode, err)
		}
		return nil, fmt.Errorf("error listing hosts: %v", err)
	}
	
	return resp.GetHostList(), nil
}

// Mute mutes a host (disables alerting)
func (c *Client) Mute(hostname string, message string, end time.Time) error {
	hostsAPI := datadogV1.NewHostsApi(c.apiClient)
	
	// Create the request body with proper initialization
	body := *datadogV1.NewHostMuteSettings()
	
	// Only set end time if it's not zero
	if !end.IsZero() {
		body.SetEnd(end.Unix())
	}
	
	// Only set message if it's not empty
	if message != "" {
		body.SetMessage(message)
	}
	
	// Use proper error handling with context
	_, httpResp, err := hostsAPI.MuteHost(c.ctx, hostname, body)
	if err != nil {
		// Include HTTP response details in error if available
		if httpResp != nil {
			return fmt.Errorf("error muting host (status: %d): %v", httpResp.StatusCode, err)
		}
		return fmt.Errorf("error muting host: %v", err)
	}
	
	return nil
}

// Unmute unmutes a host (re-enables alerting)
func (c *Client) Unmute(hostname string) error {
	hostsAPI := datadogV1.NewHostsApi(c.apiClient)
	
	// Use proper error handling with context
	_, httpResp, err := hostsAPI.UnmuteHost(c.ctx, hostname)
	if err != nil {
		// Include HTTP response details in error if available
		if httpResp != nil {
			return fmt.Errorf("error unmuting host (status: %d): %v", httpResp.StatusCode, err)
		}
		return fmt.Errorf("error unmuting host: %v", err)
	}
	
	return nil
}
