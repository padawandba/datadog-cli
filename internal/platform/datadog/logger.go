package datadog

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/padawandba/datadog-cli/internal/platform/config"
)

const (
	// DatadogLogsEndpoint is the endpoint for sending logs to Datadog
	DatadogLogsEndpoint = "https://http-intake.logs.%s/api/v2/logs"
	
	// BatchSize is the number of logs to batch before sending
	BatchSize = 20
	
	// FlushInterval is the maximum time to wait before sending logs
	FlushInterval = 5 * time.Second
)

// DatadogLogEntry represents a log entry in Datadog format
type DatadogLogEntry struct {
	Message     string                 `json:"message"`
	Status      string                 `json:"status"`
	Service     string                 `json:"service"`
	DDSource    string                 `json:"ddsource"`
	Hostname    string                 `json:"hostname"`
	Timestamp   int64                  `json:"timestamp"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
	Environment string                 `json:"env,omitempty"`
}

// DatadogHandler is a slog.Handler that sends logs to Datadog
type DatadogHandler struct {
	cfg          *config.Config
	minLevel     slog.Level
	fallback     slog.Handler
	service      string
	environment  string
	hostname     string
	logChan      chan DatadogLogEntry
	wg           sync.WaitGroup
	stopChan     chan struct{}
	client       *http.Client
	flushTicker  *time.Ticker
	logBuffer    []DatadogLogEntry
	bufferMutex  sync.Mutex
}

// DatadogHandlerOptions contains options for creating a DatadogHandler
type DatadogHandlerOptions struct {
	// MinLevel is the minimum level to log
	MinLevel slog.Level
	
	// Fallback is a handler to use if sending to Datadog fails
	Fallback slog.Handler
	
	// Service is the name of the service
	Service string
	
	// Environment is the environment (e.g., prod, staging)
	Environment string
}

// NewDatadogHandler creates a new DatadogHandler
func NewDatadogHandler(cfg *config.Config, opts *DatadogHandlerOptions) *DatadogHandler {
	if opts == nil {
		opts = &DatadogHandlerOptions{
			MinLevel: slog.LevelInfo,
		}
	}
	
	// Use os.Stderr as fallback if not provided
	if opts.Fallback == nil {
		opts.Fallback = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: opts.MinLevel,
		})
	}
	
	// Default service name if not provided
	if opts.Service == "" {
		opts.Service = "datadog-cli"
	}
	
	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	
	h := &DatadogHandler{
		cfg:         cfg,
		minLevel:    opts.MinLevel,
		fallback:    opts.Fallback,
		service:     opts.Service,
		environment: opts.Environment,
		hostname:    hostname,
		logChan:     make(chan DatadogLogEntry, 100),
		stopChan:    make(chan struct{}),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		flushTicker: time.NewTicker(FlushInterval),
		logBuffer:   make([]DatadogLogEntry, 0, BatchSize),
	}
	
	// Start the background worker
	h.wg.Add(1)
	go h.processLogs()
	
	return h
}

// Enabled implements slog.Handler.
func (h *DatadogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

// Handle implements slog.Handler.
func (h *DatadogHandler) Handle(ctx context.Context, record slog.Record) error {
	// Always log to fallback handler first
	if err := h.fallback.Handle(ctx, record); err != nil {
		return err
	}
	
	// Skip if below minimum level
	if !h.Enabled(ctx, record.Level) {
		return nil
	}
	
	// Convert slog.Record to DatadogLogEntry
	entry := DatadogLogEntry{
		Message:     record.Message,
		Status:      levelToStatus(record.Level),
		Service:     h.service,
		DDSource:    "datadog-cli",
		Hostname:    h.hostname,
		Timestamp:   record.Time.UnixNano() / int64(time.Millisecond),
		Environment: h.environment,
		Attributes:  make(map[string]interface{}),
	}
	
	// Add file and line information
	if pc, file, line, ok := runtime.Caller(4); ok {
		entry.Attributes["file"] = file
		entry.Attributes["line"] = line
		if fn := runtime.FuncForPC(pc); fn != nil {
			entry.Attributes["function"] = fn.Name()
		}
	}
	
	// Add attributes from record
	record.Attrs(func(attr slog.Attr) bool {
		addAttrToMap(entry.Attributes, "", attr)
		return true
	})
	
	// Send to channel for async processing
	select {
	case h.logChan <- entry:
		// Successfully queued
	default:
		// Channel full, log to fallback
		return fmt.Errorf("datadog log channel full, dropping log")
	}
	
	return nil
}

// WithAttrs implements slog.Handler.
func (h *DatadogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Create a new handler with the same configuration
	newHandler := &DatadogHandler{
		cfg:         h.cfg,
		minLevel:    h.minLevel,
		fallback:    h.fallback.WithAttrs(attrs),
		service:     h.service,
		environment: h.environment,
		hostname:    h.hostname,
		logChan:     h.logChan,
		stopChan:    h.stopChan,
		client:      h.client,
		flushTicker: h.flushTicker,
		logBuffer:   h.logBuffer,
		bufferMutex: h.bufferMutex,
	}
	
	return newHandler
}

// WithGroup implements slog.Handler.
func (h *DatadogHandler) WithGroup(name string) slog.Handler {
	// Create a new handler with the same configuration
	newHandler := &DatadogHandler{
		cfg:         h.cfg,
		minLevel:    h.minLevel,
		fallback:    h.fallback.WithGroup(name),
		service:     h.service,
		environment: h.environment,
		hostname:    h.hostname,
		logChan:     h.logChan,
		stopChan:    h.stopChan,
		client:      h.client,
		flushTicker: h.flushTicker,
		logBuffer:   h.logBuffer,
		bufferMutex: h.bufferMutex,
	}
	
	return newHandler
}

// Close stops the background worker and flushes any remaining logs
func (h *DatadogHandler) Close() error {
	// Signal the worker to stop
	close(h.stopChan)
	
	// Wait for the worker to finish
	h.wg.Wait()
	
	return nil
}

// processLogs processes logs in the background
func (h *DatadogHandler) processLogs() {
	defer h.wg.Done()
	
	for {
		select {
		case <-h.stopChan:
			// Flush any remaining logs
			h.flush()
			return
			
		case entry := <-h.logChan:
			// Add to buffer
			h.bufferMutex.Lock()
			h.logBuffer = append(h.logBuffer, entry)
			
			// Flush if buffer is full
			if len(h.logBuffer) >= BatchSize {
				h.flushLocked()
			}
			h.bufferMutex.Unlock()
			
		case <-h.flushTicker.C:
			// Flush on timer
			h.flush()
		}
	}
}

// flush flushes the log buffer
func (h *DatadogHandler) flush() {
	h.bufferMutex.Lock()
	defer h.bufferMutex.Unlock()
	
	h.flushLocked()
}

// flushLocked flushes the log buffer (must be called with bufferMutex held)
func (h *DatadogHandler) flushLocked() {
	if len(h.logBuffer) == 0 {
		return
	}
	
	// Copy buffer and reset
	logs := make([]DatadogLogEntry, len(h.logBuffer))
	copy(logs, h.logBuffer)
	h.logBuffer = h.logBuffer[:0]
	
	// Send logs in a separate goroutine
	go h.sendLogs(logs)
}

// sendLogs sends logs to Datadog
func (h *DatadogHandler) sendLogs(logs []DatadogLogEntry) {
	// Skip if no API key
	if h.cfg.APIKey == "" {
		return
	}
	
	// Construct URL
	site := h.cfg.Site
	if site == "" {
		site = "datadoghq.com"
	}
	url := fmt.Sprintf(DatadogLogsEndpoint, site)
	
	// Marshal logs to JSON
	data, err := json.Marshal(logs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling logs: %v\n", err)
		return
	}
	
	// Create request
	req, err := http.NewRequest("POST", url, strings.NewReader(string(data)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		return
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("DD-API-KEY", h.cfg.APIKey)
	
	// Send request
	resp, err := h.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending logs to Datadog: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode >= 300 {
		fmt.Fprintf(os.Stderr, "Error from Datadog logs API: %s\n", resp.Status)
	}
}

// levelToStatus converts a slog.Level to a Datadog status
func levelToStatus(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return "error"
	case level >= slog.LevelWarn:
		return "warning"
	case level >= slog.LevelInfo:
		return "info"
	default:
		return "debug"
	}
}

// addAttrToMap adds a slog.Attr to a map
func addAttrToMap(m map[string]interface{}, prefix string, attr slog.Attr) {
	key := attr.Key
	if prefix != "" {
		key = prefix + "." + key
	}
	
	switch attr.Value.Kind() {
	case slog.KindGroup:
		// For groups, recursively add attributes with prefixed keys
		for _, a := range attr.Value.Group() {
			addAttrToMap(m, key, a)
		}
	case slog.KindAny:
		// For any values, use the value directly
		m[key] = attr.Value.Any()
	case slog.KindLogValuer:
		// For LogValuer, resolve and add
		m[key] = attr.Value.Resolve().Any()
	default:
		// For other kinds, use the value directly
		m[key] = attr.Value.Any()
	}
} 