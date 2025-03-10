package tags

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/padawandba/datadog-cli/internal/platform/config"
	"github.com/padawandba/datadog-cli/internal/platform/console"
	"github.com/urfave/cli/v2"
)

// NewCommands returns the tags command group
func NewCommands(apiClient *datadog.APIClient, ctx context.Context, cfg *config.Config) *cli.Command {
	client := NewClient(apiClient, ctx)
	
	return &cli.Command{
		Name:  "tags",
		Usage: "Manage Datadog tags",
		Subcommands: []*cli.Command{
			listCommand(client, cfg),
			addCommand(client),
			removeCommand(client),
		},
	}
}

// listCommand returns the command to list tags
func listCommand(client *Client, cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "List tags for a host",
		ArgsUsage: "HOSTNAME",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "source",
				Usage: "Source of the tags (e.g., user, chef, puppet)",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("hostname argument is required")
			}
			
			hostname := c.Args().First()
			source := c.String("source")
			
			tags, err := client.GetHostTags(hostname, source)
			if err != nil {
				return fmt.Errorf("failed to get host tags: %v", err)
			}
			
			formatter := console.NewFormatter(cfg.Output)
			if c.String("output") != "" {
				formatter = console.NewFormatter(c.String("output"))
			}
			
			// Use our custom formatter for host tags
			return FormatHostTags(formatter, hostname, tags)
		},
	}
}

// addCommand returns the command to add tags
func addCommand(client *Client) *cli.Command {
	return &cli.Command{
		Name:      "add",
		Usage:     "Add tags to a host",
		ArgsUsage: "HOSTNAME TAG [TAG...]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "source",
				Usage: "Source of the tags (e.g., user, chef, puppet)",
				Value: "user",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() < 2 {
				return fmt.Errorf("hostname and at least one tag argument are required")
			}
			
			hostname := c.Args().First()
			tags := c.Args().Slice()[1:]
			source := c.String("source")
			
			fmt.Printf("Adding tags to host %s: %s\n", hostname, strings.Join(tags, ", "))
			return client.AddHostTags(hostname, tags, source)
		},
	}
}

// removeCommand returns the command to remove tags
func removeCommand(client *Client) *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Usage:     "Remove tags from a host",
		ArgsUsage: "HOSTNAME [TAG...]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "source",
				Usage: "Source of the tags (e.g., user, chef, puppet)",
				Value: "user",
			},
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Remove all tags",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("hostname argument is required")
			}
			
			hostname := c.Args().First()
			source := c.String("source")
			
			var tags []string
			if !c.Bool("all") && c.NArg() < 2 {
				return fmt.Errorf("at least one tag argument is required unless --all is specified")
			}
			
			if !c.Bool("all") {
				tags = c.Args().Slice()[1:]
				fmt.Printf("Removing tags from host %s: %s\n", hostname, strings.Join(tags, ", "))
			} else {
				fmt.Printf("Removing all tags from host %s\n", hostname)
			}
			
			return client.RemoveHostTags(hostname, tags, source)
		},
	}
}
