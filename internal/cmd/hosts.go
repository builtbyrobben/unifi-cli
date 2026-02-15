package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/unifi-cli/internal/outfmt"
)

type HostsCmd struct {
	List HostsListCmd `cmd:"" help:"List all hosts"`
	Get  HostsGetCmd  `cmd:"" help:"Get host by ID"`
}

type HostsListCmd struct {
	PageSize int `help:"Number of results per page" name:"page-size" default:"0"`
}

func (cmd *HostsListCmd) Run(ctx context.Context) error {
	client, err := getUniFiClient()
	if err != nil {
		return err
	}

	result, err := client.Hosts().List(ctx, cmd.PageSize)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if len(result.Data) == 0 {
		fmt.Fprintln(os.Stderr, "No hosts found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d host(s)\n\n", len(result.Data))
	for _, h := range result.Data {
		fmt.Printf("ID: %s\n", h.ID)
		if h.Type != "" {
			fmt.Printf("  Type:    %s\n", h.Type)
		}
		if h.IPAddress != "" {
			fmt.Printf("  IP:      %s\n", h.IPAddress)
		}
		if h.UserData != nil && h.UserData.Name != "" {
			fmt.Printf("  Name:    %s\n", h.UserData.Name)
		}
		fmt.Printf("  Owner:   %t\n", h.Owner)
		fmt.Printf("  Blocked: %t\n", h.IsBlocked)
		if h.RegistrationTime != "" {
			fmt.Printf("  Registered: %s\n", h.RegistrationTime)
		}
		fmt.Println()
	}

	return nil
}

type HostsGetCmd struct {
	ID string `arg:"" required:"" help:"Host ID"`
}

func (cmd *HostsGetCmd) Run(ctx context.Context) error {
	client, err := getUniFiClient()
	if err != nil {
		return err
	}

	result, err := client.Hosts().Get(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	h := result.Data
	fmt.Printf("ID: %s\n", h.ID)
	if h.Type != "" {
		fmt.Printf("Type:    %s\n", h.Type)
	}
	if h.IPAddress != "" {
		fmt.Printf("IP:      %s\n", h.IPAddress)
	}
	if h.UserData != nil && h.UserData.Name != "" {
		fmt.Printf("Name:    %s\n", h.UserData.Name)
	}
	if h.HardwareID != "" {
		fmt.Printf("HW ID:   %s\n", h.HardwareID)
	}
	fmt.Printf("Owner:   %t\n", h.Owner)
	fmt.Printf("Blocked: %t\n", h.IsBlocked)
	if h.RegistrationTime != "" {
		fmt.Printf("Registered: %s\n", h.RegistrationTime)
	}
	if h.LastConnectionStateChange != "" {
		fmt.Printf("Last Connection: %s\n", h.LastConnectionStateChange)
	}
	if h.LatestBackupTime != "" {
		fmt.Printf("Last Backup: %s\n", h.LatestBackupTime)
	}

	return nil
}
