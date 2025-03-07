package main

import (
	"context"
	"log/slog"
	"os"

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
	if cfg.APIKey == "" {
		return cli.Exit("API key is required", 1)
	}
	if cfg.AppKey == "" {
		return cli.Exit("Application key is required", 1)
	}
	return nil
}

// withConfigValidation wraps an action with config validation
func withConfigValidation(cfg *config.Config, action func(*cli.Context) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		// Skip validation for help
		if c.Bool("help") || c.Bool("h") {
			return cli.ShowCommandHelp(c, c.Command.Name)
		}
		
		// Validate config
		if err := validateConfig(cfg); err != nil {
			return err
		}
		
		// Run the original action
		return action(c)
	}
}

func main() {
	// Set up structured logging
	handler := slog.NewTextHandler(os.Stderr, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Check if help is requested
	isHelp := false
	for _, arg := range os.Args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			isHelp = true
			break
		}
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
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
			return err
		},
	}

	// Initialize the Datadog client
	var client *datadog.APIClient
	var ctx context.Context
	
	// Create a basic client for help mode
	configuration := datadog.NewConfiguration()
	client = datadog.NewAPIClient(configuration)
	ctx = context.Background()
	
	// Only initialize the real client if not in help mode
	if !isHelp {
		client, ctx = ddapi.NewClient(cfg)
	}

	// Create command groups
	hostsCmd := hosts.NewCommands(client, ctx, cfg)
	tagsCmd := tags.NewCommands(client, ctx, cfg)
	monitorsCmd := monitors.NewCommands(client, ctx, cfg)
	
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

	if err := app.Run(os.Args); err != nil {
		slog.Error("Application error", "error", err)
		os.Exit(1)
	}
}
