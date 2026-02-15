package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/unifi-cli/internal/outfmt"
	"github.com/builtbyrobben/unifi-cli/internal/unifi"
)

type TopologyCmd struct{}

type topologyOutput struct {
	Devices  []unifi.Device       `json:"devices"`
	Clients  []unifi.ClientDevice `json:"clients"`
	Networks []unifi.Network      `json:"networks"`
}

func (cmd *TopologyCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := getUniFiClient(flags)
	if err != nil {
		return err
	}

	devices, err := client.Devices().List(ctx, "")
	if err != nil {
		return fmt.Errorf("list devices: %w", err)
	}

	clients, err := client.Clients().ListActive(ctx, "")
	if err != nil {
		return fmt.Errorf("list clients: %w", err)
	}

	networks, err := client.Networks().List(ctx)
	if err != nil {
		return fmt.Errorf("list networks: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, topologyOutput{
			Devices:  devices,
			Clients:  clients,
			Networks: networks,
		})
	}

	// Human-readable topology
	fmt.Fprintf(os.Stderr, "Network Topology\n\n")

	fmt.Fprintf(os.Stderr, "=== Networks (%d) ===\n\n", len(networks))
	for _, n := range networks {
		fmt.Printf("%s", n.Name)
		if n.IPSubnet != "" {
			fmt.Printf(" (%s)", n.IPSubnet)
		}
		if n.VLANEnabled {
			fmt.Printf(" [VLAN %d]", n.VLAN)
		}
		fmt.Println()
	}

	fmt.Printf("\n")
	fmt.Fprintf(os.Stderr, "=== Devices (%d) ===\n\n", len(devices))
	for _, d := range devices {
		name := d.Name
		if name == "" {
			name = d.MAC
		}
		fmt.Printf("%s", name)
		if d.IP != "" {
			fmt.Printf(" (%s)", d.IP)
		}
		if d.Type != "" {
			fmt.Printf(" [%s]", d.Type)
		}
		if d.NumSta > 0 {
			fmt.Printf(" — %d clients", d.NumSta)
		}
		fmt.Println()
	}

	fmt.Printf("\n")
	fmt.Fprintf(os.Stderr, "=== Active Clients (%d) ===\n\n", len(clients))

	wired := 0
	wireless := 0

	for _, c := range clients {
		if c.IsWired {
			wired++
		} else {
			wireless++
		}
	}

	fmt.Printf("Wired:    %d\n", wired)
	fmt.Printf("Wireless: %d\n", wireless)
	fmt.Printf("Total:    %d\n", len(clients))

	return nil
}
