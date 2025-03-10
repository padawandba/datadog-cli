package tags

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

// Client provides tag-related operations
type Client struct {
	apiClient *datadog.APIClient
	ctx       context.Context
}

// NewClient creates a new tags client
func NewClient(apiClient *datadog.APIClient, ctx context.Context) *Client {
	return &Client{
		apiClient: apiClient,
		ctx:       ctx,
	}
}

// GetHostTags retrieves tags for a specific host
func (c *Client) GetHostTags(hostname string, source string) ([]string, error) {
	tagsAPI := datadogV1.NewTagsApi(c.apiClient)
	
	// Create optional parameters with proper initialization
	opts := datadogV1.NewGetHostTagsOptionalParameters()
	if source != "" {
		opts = opts.WithSource(source)
	}
	
	// Use proper error handling with context
	resp, httpResp, err := tagsAPI.GetHostTags(c.ctx, hostname, *opts)
	if err != nil {
		// Include HTTP response details in error if available
		if httpResp != nil {
			return nil, fmt.Errorf("error getting host tags (status: %d): %v", httpResp.StatusCode, err)
		}
		return nil, fmt.Errorf("error getting host tags: %v", err)
	}
	
	return resp.GetTags(), nil
}

// AddHostTags adds tags to a specific host
func (c *Client) AddHostTags(hostname string, tags []string, source string) error {
	tagsAPI := datadogV1.NewTagsApi(c.apiClient)
	
	// Create the request body with proper initialization
	body := *datadogV1.NewHostTags()
	body.SetTags(tags)
	
	// Create optional parameters with proper initialization
	opts := datadogV1.NewCreateHostTagsOptionalParameters()
	if source != "" {
		opts = opts.WithSource(source)
	}
	
	// Use proper error handling with context
	_, httpResp, err := tagsAPI.CreateHostTags(c.ctx, hostname, body, *opts)
	if err != nil {
		// Include HTTP response details in error if available
		if httpResp != nil {
			return fmt.Errorf("error adding host tags (status: %d): %v", httpResp.StatusCode, err)
		}
		return fmt.Errorf("error adding host tags: %v", err)
	}
	
	return nil
}

// RemoveHostTags removes tags from a specific host
func (c *Client) RemoveHostTags(hostname string, tags []string, source string) error {
	tagsAPI := datadogV1.NewTagsApi(c.apiClient)
	
	// Create optional parameters with proper initialization
	opts := datadogV1.NewDeleteHostTagsOptionalParameters()
	if source != "" {
		opts = opts.WithSource(source)
	}
	
	if len(tags) == 0 || (len(tags) == 1 && tags[0] == "*") {
		// Delete all tags
		httpResp, err := tagsAPI.DeleteHostTags(c.ctx, hostname, *opts)
		if err != nil {
			// Include HTTP response details in error if available
			if httpResp != nil {
				return fmt.Errorf("error removing all host tags (status: %d): %v", httpResp.StatusCode, err)
			}
			return fmt.Errorf("error removing all host tags: %v", err)
		}
		return nil
	}
	
	// For specific tags, we need to get current tags, filter them, and update
	currentTags, err := c.GetHostTags(hostname, source)
	if err != nil {
		return fmt.Errorf("error getting current host tags: %v", err)
	}
	
	// Filter out the tags to be removed
	newTags := filterTags(currentTags, tags)
	
	// If no tags left, delete all tags
	if len(newTags) == 0 {
		httpResp, err := tagsAPI.DeleteHostTags(c.ctx, hostname, *opts)
		if err != nil {
			// Include HTTP response details in error if available
			if httpResp != nil {
				return fmt.Errorf("error removing all host tags (status: %d): %v", httpResp.StatusCode, err)
			}
			return fmt.Errorf("error removing all host tags: %v", err)
		}
		return nil
	}
	
	// Update with the filtered tags
	return c.AddHostTags(hostname, newTags, source)
}

// filterTags removes the specified tags from the list of current tags
func filterTags(currentTags []string, tagsToRemove []string) []string {
	// Create a map for quick lookup of tags to remove
	removeMap := make(map[string]bool)
	for _, tag := range tagsToRemove {
		removeMap[tag] = true
	}
	
	// Filter out the tags to be removed
	var filteredTags []string
	for _, tag := range currentTags {
		if !removeMap[tag] {
			filteredTags = append(filteredTags, tag)
		}
	}
	
	return filteredTags
}
