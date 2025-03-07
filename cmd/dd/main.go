package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/padawandba/datadog-cli/internal/hosts"
	"github.com/padawandba/datadog-cli/internal/monitors"
	"github.com/padawandba/datadog-cli/internal/platform/config"
	ddapi "github.com/padawandba/datadog-cli/internal/platform/datadog"
	"github.com/padawandba/datadog-cli/internal/tags"
	"github.com/urfave/cli/v2"
)

// validateConfig checks if the required configuration is present
func validateConfig(cfg *config.Config) error {
	return config.Validate(cfg)
}

func main() {
	// Set up signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Check if help is requested
	isHelp := false
	for _, arg := range os.Args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			isHelp = true
			break
		}
	}

	// Set up initial basic logging
	var logLevel slog.Level
	if os.Getenv("DD_DEBUG") != "" {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}
	
	// Start with a basic stderr logger
	basicHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})
	logger := slog.New(basicHandler)
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Set up Datadog logging if not in help mode and we have API credentials
	var ddHandler *ddapi.DatadogHandler
	if !isHelp && cfg.APIKey != "" && cfg.AppKey != "" {
		// Get environment from env var or default to "dev"
		env := os.Getenv("DD_ENV")
		if env == "" {
			env = "dev"
		}
		
		// Create Datadog handler with fallback to stderr
		ddHandler = ddapi.NewDatadogHandler(cfg, &ddapi.DatadogHandlerOptions{
			MinLevel:    logLevel,
			Fallback:    basicHandler,
			Service:     "datadog-cli",
			Environment: env,
		})
		
		// Set as default logger
		logger = slog.New(ddHandler)
		slog.SetDefault(logger)
		
		// Ensure handler is closed on exit
		defer func() {
			if err := ddHandler.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing Datadog log handler: %v\n", err)
			}
		}()
		
		slog.Info("Datadog logging enabled", 
			"environment", env,
			"service", "datadog-cli",
			"level", logLevel.String())
	}

	app := &cli.App{
		Name:  "dd",
		Usage: "Datadog administration CLI tool",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "dd-api-key",
				EnvVars: []string{"DD_API_KEY"},
				Usage:   "Datadog API key",
			},
			&cli.StringFlag{
				Name:    "dd-app-key",
				EnvVars: []string{"DD_APP_KEY"},
				Usage:   "Datadog application key",
			},
			&cli.StringFlag{
				Name:    "dd-site",
				EnvVars: []string{"DD_SITE"},
				Value:   "datadoghq.com",
				Usage:   "Datadog site (e.g., datadoghq.com, datadoghq.eu)",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "table",
				Usage:   "Output format (table, json, yaml)",
			},
			&cli.BoolFlag{
				Name:    "debug",
				EnvVars: []string{"DD_DEBUG"},
				Usage:   "Enable debug logging",
			},
			&cli.StringFlag{
				Name:    "env",
				EnvVars: []string{"DD_ENV"},
				Value:   "dev",
				Usage:   "Environment for Datadog logs (e.g., dev, prod)",
			},
		},
		Before: func(c *cli.Context) error {
			// Skip validation if just showing help
			if isHelp {
				return nil
			}
			
			// Update config with command line flags (if provided)
			if apiKey := c.String("dd-api-key"); apiKey != "" {
				cfg.APIKey = apiKey
			}
			if appKey := c.String("dd-app-key"); appKey != "" {
				cfg.AppKey = appKey
			}
			if site := c.String("dd-site"); site != "" {
				cfg.Site = site
			}
			if output := c.String("output"); output != "" {
				cfg.Output = output
			}
			
			// Update log level if debug flag is set
			if c.Bool("debug") && logLevel != slog.LevelDebug {
				// Update log level for existing handlers
				newLevel := slog.LevelDebug
				
				if ddHandler != nil {
					// Create a new Datadog handler with debug level
					env := c.String("env")
					ddHandler = ddapi.NewDatadogHandler(cfg, &ddapi.DatadogHandlerOptions{
						MinLevel:    newLevel,
						Fallback:    slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: newLevel}),
						Service:     "datadog-cli",
						Environment: env,
					})
					
					logger = slog.New(ddHandler)
					slog.SetDefault(logger)
				} else {
					// Just update the stderr handler
					handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
						Level: newLevel,
					})
					logger := slog.New(handler)
					slog.SetDefault(logger)
				}
				
				slog.Debug("Debug logging enabled")
			}
			
			// Validate required configuration
			return validateConfig(cfg)
		},
		// Custom command not found handler to show help
		CommandNotFound: func(c *cli.Context, command string) {
			if isHelp {
				cli.ShowAppHelp(c)
				return
			}
			slog.Error("Command not found", "command", command)
		},
		// Custom usage error handler to show help
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			if isHelp {
				if isSubcommand {
					return cli.ShowSubcommandHelp(c)
				}
				return cli.ShowAppHelp(c)
			}
			slog.Error("Usage error", "error", err, "subcommand", isSubcommand)
			return err
		},
	}

	// Initialize the Datadog client
	var client *datadog.APIClient
	var apiCtx context.Context
	
	// Create a basic client for help mode
	configuration := datadog.NewConfiguration()
	client = datadog.NewAPIClient(configuration)
	apiCtx = ctx
	
	// Only initialize the real client if not in help mode
	if !isHelp {
		client, apiCtx = ddapi.NewClient(cfg)
		
		// Ensure client resources are cleaned up on exit
		defer ddapi.CleanupContext(apiCtx)
	}

	// Create command groups
	hostsCmd := hosts.NewCommands(client, apiCtx, cfg)
	tagsCmd := tags.NewCommands(client, apiCtx, cfg)
	monitorsCmd := monitors.NewCommands(client, apiCtx, cfg)
	
	// Add commands to the application
	app.Commands = []*cli.Command{
		hostsCmd,
		tagsCmd,
		monitorsCmd,
	}

	// Override the default help flag
	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h"},
		Usage:   "Show help",
	}

	// Log command execution
	slog.Info("Executing command", "args", os.Args)
	
	if err := app.RunContext(ctx, os.Args); err != nil {
		slog.Error("Application error", "error", err)
		os.Exit(1)
	}
	
	slog.Info("Command completed successfully")
}
