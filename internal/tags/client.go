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
	
	opts := datadogV1.GetHostTagsOptionalParameters{}
	if source != "" {
		opts.WithSource(source)
	}
	
	resp, _, err := tagsAPI.GetHostTags(c.ctx, hostname, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting host tags: %v", err)
	}
	
	return resp.GetTags(), nil
}

// AddHostTags adds tags to a specific host
func (c *Client) AddHostTags(hostname string, tags []string, source string) error {
	tagsAPI := datadogV1.NewTagsApi(c.apiClient)
	
	body := datadogV1.HostTags{
		Tags: tags,
	}
	
	opts := datadogV1.CreateHostTagsOptionalParameters{}
	if source != "" {
		opts.WithSource(source)
	}
	
	_, _, err := tagsAPI.CreateHostTags(c.ctx, hostname, body, opts)
	if err != nil {
		return fmt.Errorf("error adding host tags: %v", err)
	}
	
	return nil
}

// RemoveHostTags removes tags from a specific host
func (c *Client) RemoveHostTags(hostname string, tags []string, source string) error {
	tagsAPI := datadogV1.NewTagsApi(c.apiClient)
	
	if len(tags) == 0 {
		// Delete all tags
		opts := datadogV1.DeleteHostTagsOptionalParameters{}
		if source != "" {
			opts.WithSource(source)
		}
		
		_, err := tagsAPI.DeleteHostTags(c.ctx, hostname, opts)
		if err != nil {
			return fmt.Errorf("error removing all host tags: %v", err)
		}
	} else {
		// Delete specific tags
		opts := datadogV1.DeleteHostTagsOptionalParameters{}
		if source != "" {
			opts.WithSource(source)
		}
		
		_, err := tagsAPI.DeleteHostTags(c.ctx, hostname, opts)
		if err != nil {
			return fmt.Errorf("error removing host tags: %v", err)
		}
		
		// Get remaining tags
		remainingTags, err := c.GetHostTags(hostname, source)
		if err != nil {
			return fmt.Errorf("error getting remaining host tags: %v", err)
		}
		
		// Filter out the deleted tags
		newTags := []string{}
		for _, tag := range remainingTags {
			shouldKeep := true
			for _, deleteTag := range tags {
				if tag == deleteTag {
					shouldKeep = false
					break
				}
			}
			if shouldKeep {
				newTags = append(newTags, tag)
			}
		}
		
		// Add back the filtered tags
		if len(newTags) > 0 {
			if err := c.AddHostTags(hostname, newTags, source); err != nil {
				return fmt.Errorf("error updating host tags: %v", err)
			}
		}
	}
	
	return nil
}
