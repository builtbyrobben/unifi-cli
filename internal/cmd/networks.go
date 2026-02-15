package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/unifi-cli/internal/outfmt"
)

type NetworksCmd struct {
	List NetworksListCmd `cmd:"" help:"List all network/VLAN configurations"`
	Get  NetworksGetCmd  `cmd:"" help:"Get specific network config"`
}

type NetworksListCmd struct{}

func (cmd *NetworksListCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := getUniFiClient(flags)
	if err != nil {
		return err
	}

	networks, err := client.Networks().List(ctx)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, networks)
	}

	if len(networks) == 0 {
		fmt.Fprintln(os.Stderr, "No networks found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d network(s)\n\n", len(networks))
	for _, n := range networks {
		fmt.Printf("ID:      %s\n", n.ID)
		fmt.Printf("  Name:    %s\n", n.Name)
		if n.Purpose != "" {
			fmt.Printf("  Purpose: %s\n", n.Purpose)
		}
		if n.IPSubnet != "" {
			fmt.Printf("  Subnet:  %s\n", n.IPSubnet)
		}
		if n.VLANEnabled {
			fmt.Printf("  VLAN:    %d\n", n.VLAN)
		}
		fmt.Println()
	}

	return nil
}

type NetworksGetCmd struct {
	ID string `arg:"" required:"" help:"Network ID"`
}

func (cmd *NetworksGetCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := getUniFiClient(flags)
	if err != nil {
		return err
	}

	network, err := client.Networks().Get(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, network)
	}

	fmt.Printf("ID:       %s\n", network.ID)
	fmt.Printf("Name:     %s\n", network.Name)
	if network.Purpose != "" {
		fmt.Printf("Purpose:  %s\n", network.Purpose)
	}
	if network.IPSubnet != "" {
		fmt.Printf("Subnet:   %s\n", network.IPSubnet)
	}
	if network.VLANEnabled {
		fmt.Printf("VLAN:     %d\n", network.VLAN)
	}
	if network.DomainName != "" {
		fmt.Printf("Domain:   %s\n", network.DomainName)
	}
	fmt.Printf("DHCP:     %t\n", network.DHCPDEnabled)
	if network.DHCPDStart != "" {
		fmt.Printf("  Start:  %s\n", network.DHCPDStart)
		fmt.Printf("  Stop:   %s\n", network.DHCPDStop)
	}

	return nil
}
