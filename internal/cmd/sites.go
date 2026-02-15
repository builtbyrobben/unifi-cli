package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/unifi-cli/internal/outfmt"
)

type SitesCmd struct {
	List SitesListCmd `cmd:"" help:"List all sites"`
}

type SitesListCmd struct {
	PageSize int `help:"Number of results per page" name:"page-size" default:"0"`
}

func (cmd *SitesListCmd) Run(ctx context.Context) error {
	client, err := getUniFiClient()
	if err != nil {
		return err
	}

	result, err := client.Sites().List(ctx, cmd.PageSize)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if len(result.Data) == 0 {
		fmt.Fprintln(os.Stderr, "No sites found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d site(s)\n\n", len(result.Data))
	for _, s := range result.Data {
		fmt.Printf("Site ID: %s\n", s.SiteID)
		if s.Meta.Name != "" {
			fmt.Printf("  Name:     %s\n", s.Meta.Name)
		}
		if s.HostID != "" {
			fmt.Printf("  Host ID:  %s\n", s.HostID)
		}
		if s.Meta.Timezone != "" {
			fmt.Printf("  Timezone: %s\n", s.Meta.Timezone)
		}
		if s.Meta.Description != "" {
			fmt.Printf("  Desc:     %s\n", s.Meta.Description)
		}
		fmt.Printf("  Owner:    %t\n", s.IsOwner)
		if s.Permission != "" {
			fmt.Printf("  Perm:     %s\n", s.Permission)
		}
		if s.Statistics != nil && s.Statistics.Counts != nil {
			c := s.Statistics.Counts
			fmt.Printf("  Devices:  %d active / %d total\n", c.ActiveDevice, c.TotalDevice)
			fmt.Printf("  Clients:  %d total (%d wired, %d wireless)\n", c.TotalClient, c.WiredClient, c.WirelessClient)
		}
		fmt.Println()
	}

	return nil
}
