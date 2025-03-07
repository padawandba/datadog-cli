package datadog

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/padawandba/datadog-cli/internal/platform/config"
)

// NewClient creates a new Datadog API client
func NewClient(cfg *config.Config) (*datadog.APIClient, context.Context) {
	ctx := context.Background()
	
	// Configure authentication
	configuration := datadog.NewConfiguration()
	
	// Create the client
	apiClient := datadog.NewAPIClient(configuration)
	
	return apiClient, ctx
}
