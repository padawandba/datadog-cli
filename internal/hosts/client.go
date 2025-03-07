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
	
	opts := datadogV1.ListHostsOptionalParameters{}
	if filter != "" {
		opts.WithFilter(filter)
	}
	
	resp, _, err := hostsAPI.ListHosts(c.ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("error listing hosts: %v", err)
	}
	
	return resp.HostList, nil
}

// Mute mutes a host (disables alerting)
func (c *Client) Mute(hostname string, message string, end time.Time) error {
	hostsAPI := datadogV1.NewHostsApi(c.apiClient)
	
	body := datadogV1.HostMuteSettings{
		End:     datadog.PtrInt64(end.Unix()),
		Message: datadog.PtrString(message),
	}
	
	_, _, err := hostsAPI.MuteHost(c.ctx, hostname, body)
	if err != nil {
		return fmt.Errorf("error muting host: %v", err)
	}
	
	return nil
}

// Unmute unmutes a host (re-enables alerting)
func (c *Client) Unmute(hostname string) error {
	hostsAPI := datadogV1.NewHostsApi(c.apiClient)
	
	_, _, err := hostsAPI.UnmuteHost(c.ctx, hostname)
	if err != nil {
		return fmt.Errorf("error unmuting host: %v", err)
	}
	
	return nil
}
