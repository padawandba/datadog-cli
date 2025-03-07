package monitors

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/padawandba/datadog-cli/internal/platform/config"
	"github.com/padawandba/datadog-cli/internal/platform/console"
	"github.com/urfave/cli/v2"
)

// NewCommands returns the monitors command group
func NewCommands(apiClient *datadog.APIClient, ctx context.Context, cfg *config.Config) *cli.Command {
	client := NewClient(apiClient, ctx)
	
	return &cli.Command{
		Name:  "monitors",
		Usage: "Manage Datadog monitors",
		Subcommands: []*cli.Command{
			listCommand(client, cfg),
			muteCommand(client),
			unmuteCommand(client),
		},
	}
}

// listCommand returns the command to list monitors
func listCommand(client *Client, cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List monitors",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "query",
				Aliases: []string{"q"},
				Usage:   "Filter monitors by query",
			},
			&cli.StringSliceFlag{
				Name:    "tags",
				Aliases: []string{"t"},
				Usage:   "Filter monitors by tags",
			},
		},
		Action: func(c *cli.Context) error {
			query := c.String("query")
			tags := c.StringSlice("tags")
			
			monitors, err := client.List(query, tags)
			if err != nil {
				return fmt.Errorf("failed to list monitors: %v", err)
			}
			
			formatter := console.NewFormatter(cfg.Output)
			if c.String("output") != "" {
				formatter = console.NewFormatter(c.String("output"))
			}
			
			return formatter.Format(monitors)
		},
	}
}

// muteCommand returns the command to mute a monitor
func muteCommand(client *Client) *cli.Command {
	return &cli.Command{
		Name:      "mute",
		Usage:     "Mute a monitor",
		ArgsUsage: "MONITOR_ID",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "scope",
				Usage: "Scope to mute the monitor for (e.g., 'host:myhost')",
			},
			&cli.StringFlag{
				Name:    "duration",
				Aliases: []string{"d"},
				Usage:   "Duration to mute the monitor (e.g., 30m, 1h, 2h30m)",
				Value:   "1h",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("monitor ID argument is required")
			}
			
			monitorIDStr := c.Args().First()
			monitorID, err := strconv.ParseInt(monitorIDStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid monitor ID: %v", err)
			}
			
			scope := c.String("scope")
			durationStr := c.String("duration")
			
			duration, err := time.ParseDuration(durationStr)
			if err != nil {
				return fmt.Errorf("invalid duration format: %v", err)
			}
			
			endTime := time.Now().Add(duration).Unix()
			
			fmt.Printf("Muting monitor %d", monitorID)
			if scope != "" {
				fmt.Printf(" with scope '%s'", scope)
			}
			fmt.Printf(" for %s\n", durationStr)
			
			return client.Mute(monitorID, scope, endTime)
		},
	}
}

// unmuteCommand returns the command to unmute a monitor
func unmuteCommand(client *Client) *cli.Command {
	return &cli.Command{
		Name:      "unmute",
		Usage:     "Unmute a monitor",
		ArgsUsage: "MONITOR_ID",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "scope",
				Usage: "Scope to unmute the monitor for (e.g., 'host:myhost')",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("monitor ID argument is required")
			}
			
			monitorIDStr := c.Args().First()
			monitorID, err := strconv.ParseInt(monitorIDStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid monitor ID: %v", err)
			}
			
			scope := c.String("scope")
			
			fmt.Printf("Unmuting monitor %d", monitorID)
			if scope != "" {
				fmt.Printf(" with scope '%s'", scope)
			}
			fmt.Println()
			
			return client.Unmute(monitorID, scope)
		},
	}
}