package hosts

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/padawandba/datadog-cli/internal/platform/config"
	"github.com/padawandba/datadog-cli/internal/platform/console"
	"github.com/urfave/cli/v2"
)

// NewCommands returns the hosts command group
func NewCommands(apiClient *datadog.APIClient, ctx context.Context, cfg *config.Config) *cli.Command {
	client := NewClient(apiClient, ctx)
	
	return &cli.Command{
		Name:  "hosts",
		Usage: "Manage Datadog hosts",
		Subcommands: []*cli.Command{
			listCommand(client, cfg),
			quietCommand(client),
			unquietCommand(client),
		},
	}
}

// listCommand returns the command to list hosts
func listCommand(client *Client, cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List hosts",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "filter",
				Usage: "Filter hosts by name",
			},
		},
		Action: func(c *cli.Context) error {
			filter := c.String("filter")
			
			hosts, err := client.List(filter)
			if err != nil {
				return fmt.Errorf("failed to list hosts: %v", err)
			}
			
			formatter := console.NewFormatter(cfg.Output)
			if c.String("output") != "" {
				formatter = console.NewFormatter(c.String("output"))
			}
			
			return formatter.Format(hosts)
		},
	}
}

// quietCommand returns the command to quiet (mute) a host
func quietCommand(client *Client) *cli.Command {
	return &cli.Command{
		Name:      "quiet",
		Usage:     "Quiet (mute) a host",
		ArgsUsage: "HOSTNAME",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "message",
				Aliases: []string{"m"},
				Usage:   "Message explaining why the host is muted",
			},
			&cli.StringFlag{
				Name:    "duration",
				Aliases: []string{"d"},
				Usage:   "Duration to mute the host (e.g., 30m, 1h, 2h30m)",
				Value:   "1h",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("hostname argument is required")
			}
			
			hostname := c.Args().First()
			message := c.String("message")
			durationStr := c.String("duration")
			
			duration, err := time.ParseDuration(durationStr)
			if err != nil {
				return fmt.Errorf("invalid duration format: %v", err)
			}
			
			endTime := time.Now().Add(duration)
			
			fmt.Printf("Quieting host %s for %s\n", hostname, durationStr)
			return client.Mute(hostname, message, endTime)
		},
	}
}

// unquietCommand returns the command to unquiet (unmute) a host
func unquietCommand(client *Client) *cli.Command {
	return &cli.Command{
		Name:      "unquiet",
		Usage:     "Unquiet (unmute) a host",
		ArgsUsage: "HOSTNAME",
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("hostname argument is required")
			}
			
			hostname := c.Args().First()
			
			fmt.Printf("Unquieting host %s\n", hostname)
			return client.Unmute(hostname)
		},
	}
}
