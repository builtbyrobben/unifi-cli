package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/unifi-cli/internal/outfmt"
)

type DevicesCmd struct {
	List    DevicesListCmd    `cmd:"" help:"List all network devices"`
	Get     DevicesGetCmd     `cmd:"" help:"Get device details by MAC address"`
	Restart DevicesRestartCmd `cmd:"" help:"Restart a device"`
}

type DevicesListCmd struct {
	Type string `help:"Filter by device type (uap|usw|ugw)" short:"t"`
}

func (cmd *DevicesListCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := getUniFiClient(flags)
	if err != nil {
		return err
	}

	devices, err := client.Devices().List(ctx, cmd.Type)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, devices)
	}

	if len(devices) == 0 {
		fmt.Fprintln(os.Stderr, "No devices found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d device(s)\n\n", len(devices))
	for _, d := range devices {
		fmt.Printf("MAC:     %s\n", d.MAC)
		if d.Name != "" {
			fmt.Printf("  Name:    %s\n", d.Name)
		}
		if d.IP != "" {
			fmt.Printf("  IP:      %s\n", d.IP)
		}
		if d.Model != "" {
			fmt.Printf("  Model:   %s\n", d.Model)
		}
		if d.Type != "" {
			fmt.Printf("  Type:    %s\n", d.Type)
		}
		if d.Version != "" {
			fmt.Printf("  Version: %s\n", d.Version)
		}
		fmt.Printf("  Adopted: %t\n", d.Adopted)
		if d.NumSta > 0 {
			fmt.Printf("  Clients: %d\n", d.NumSta)
		}
		fmt.Println()
	}

	return nil
}

type DevicesGetCmd struct {
	MAC string `arg:"" required:"" help:"Device MAC address"`
}

func (cmd *DevicesGetCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := getUniFiClient(flags)
	if err != nil {
		return err
	}

	device, err := client.Devices().Get(ctx, cmd.MAC)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, device)
	}

	fmt.Printf("MAC:     %s\n", device.MAC)
	if device.Name != "" {
		fmt.Printf("Name:    %s\n", device.Name)
	}
	if device.IP != "" {
		fmt.Printf("IP:      %s\n", device.IP)
	}
	if device.Model != "" {
		fmt.Printf("Model:   %s\n", device.Model)
	}
	if device.Type != "" {
		fmt.Printf("Type:    %s\n", device.Type)
	}
	if device.Version != "" {
		fmt.Printf("Version: %s\n", device.Version)
	}
	fmt.Printf("Adopted: %t\n", device.Adopted)
	fmt.Printf("State:   %d\n", device.State)
	if device.Uptime > 0 {
		fmt.Printf("Uptime:  %ds\n", device.Uptime)
	}
	if device.NumSta > 0 {
		fmt.Printf("Clients: %d\n", device.NumSta)
	}

	return nil
}

type DevicesRestartCmd struct {
	MAC string `arg:"" required:"" help:"Device MAC address"`
}

func (cmd *DevicesRestartCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := getUniFiClient(flags)
	if err != nil {
		return err
	}

	if err := client.Devices().Restart(ctx, cmd.MAC); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("Device %s restart initiated", cmd.MAC),
		})
	}

	fmt.Fprintf(os.Stderr, "Device %s restart initiated\n", cmd.MAC)

	return nil
}
