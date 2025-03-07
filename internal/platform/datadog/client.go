package datadog

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/padawandba/datadog-cli/internal/platform/config"
)

// NewClient creates a new Datadog API client
func NewClient(cfg *config.Config) (*datadog.APIClient, context.Context) {
	// Create a context with timeout for API operations
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	
	// Store the cancel function in the context for cleanup
	ctx = context.WithValue(ctx, "cancelFunc", cancel)
	
	// Configure authentication and settings
	configuration := datadog.NewConfiguration()
	configuration.AddDefaultHeader("DD-API-KEY", cfg.APIKey)
	configuration.AddDefaultHeader("DD-APPLICATION-KEY", cfg.AppKey)
	configuration.Host = cfg.Site
	
	// Configure HTTP client with reasonable timeouts
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &loggingTransport{
			transport: http.DefaultTransport,
		},
	}
	configuration.HTTPClient = httpClient
	
	// Create the client
	apiClient := datadog.NewAPIClient(configuration)
	
	slog.Debug("Initialized Datadog API client", 
		"site", cfg.Site,
		"timeout", "30s")
	
	return apiClient, ctx
}

// loggingTransport is an http.RoundTripper that logs requests and responses
type loggingTransport struct {
	transport http.RoundTripper
}

// RoundTrip implements the http.RoundTripper interface
func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Log the request
	slog.Debug("Datadog API request",
		"method", req.Method,
		"url", req.URL.String(),
	)
	
	// Perform the request
	start := time.Now()
	resp, err := t.transport.RoundTrip(req)
	duration := time.Since(start)
	
	// Handle transport-level errors
	if err != nil {
		slog.Error("Datadog API transport error",
			"method", req.Method,
			"url", req.URL.String(),
			"error", err,
			"duration_ms", duration.Milliseconds(),
		)
		return nil, err
	}
	
	// Log the response
	level := slog.LevelDebug
	if resp.StatusCode >= 400 {
		level = slog.LevelError
	} else if resp.StatusCode >= 300 {
		level = slog.LevelWarn
	}
	
	slog.Log(context.Background(), level, "Datadog API response",
		"method", req.Method,
		"url", req.URL.String(),
		"status", resp.Status,
		"status_code", resp.StatusCode,
		"duration_ms", duration.Milliseconds(),
	)
	
	return resp, nil
}

// CleanupContext ensures any resources associated with the context are released
func CleanupContext(ctx context.Context) {
	if cancel, ok := ctx.Value("cancelFunc").(context.CancelFunc); ok {
		cancel()
	}
}
