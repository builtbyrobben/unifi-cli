package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/unifi-cli/internal/outfmt"
)

type DevicesCmd struct {
	List DevicesListCmd `cmd:"" help:"List all devices"`
}

type DevicesListCmd struct {
	Host     string `help:"Filter by host ID" name:"host" default:""`
	PageSize int    `help:"Number of results per page" name:"page-size" default:"0"`
}

func (cmd *DevicesListCmd) Run(ctx context.Context) error {
	client, err := getUniFiClient()
	if err != nil {
		return err
	}

	result, err := client.Devices().List(ctx, cmd.Host, cmd.PageSize)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"HOST_ID", "DEVICE_ID", "NAME", "MODEL", "IP", "MAC", "STATUS"}
		var rows [][]string
		for _, dh := range result.Data {
			for _, d := range dh.Devices {
				rows = append(rows, []string{dh.HostID, d.ID, d.Name, d.Model, d.IP, d.MAC, d.Status})
			}
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Data) == 0 {
		fmt.Fprintln(os.Stderr, "No devices found")
		return nil
	}

	totalDevices := 0
	for _, dh := range result.Data {
		totalDevices += len(dh.Devices)
	}

	fmt.Fprintf(os.Stderr, "Found %d device(s) across %d host(s)\n\n", totalDevices, len(result.Data))

	for _, dh := range result.Data {
		hostLabel := dh.HostID
		if dh.HostName != "" {
			hostLabel = dh.HostName + " (" + dh.HostID + ")"
		}

		fmt.Printf("Host: %s\n", hostLabel)

		for _, d := range dh.Devices {
			fmt.Printf("  Device: %s\n", d.ID)
			if d.Name != "" {
				fmt.Printf("    Name:       %s\n", d.Name)
			}
			if d.Model != "" {
				fmt.Printf("    Model:      %s\n", d.Model)
			}
			if d.Shortname != "" {
				fmt.Printf("    Shortname:  %s\n", d.Shortname)
			}
			if d.IP != "" {
				fmt.Printf("    IP:         %s\n", d.IP)
			}
			if d.MAC != "" {
				fmt.Printf("    MAC:        %s\n", d.MAC)
			}
			if d.Status != "" {
				fmt.Printf("    Status:     %s\n", d.Status)
			}
			if d.Version != "" {
				fmt.Printf("    Version:    %s\n", d.Version)
			}
			if d.ProductLine != "" {
				fmt.Printf("    Product:    %s\n", d.ProductLine)
			}
			if d.FirmwareStatus != "" {
				fmt.Printf("    Firmware:   %s\n", d.FirmwareStatus)
			}
		}

		fmt.Println()
	}

	return nil
}
