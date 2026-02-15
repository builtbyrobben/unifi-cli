package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/unifi-cli/internal/outfmt"
	"github.com/builtbyrobben/unifi-cli/internal/unifi"
)

type ClientsCmd struct {
	List    ClientsListCmd    `cmd:"" help:"List network clients"`
	Get     ClientsGetCmd     `cmd:"" help:"Get client details by MAC address"`
	Block   ClientsBlockCmd   `cmd:"" help:"Block a client"`
	Unblock ClientsUnblockCmd `cmd:"" help:"Unblock a client"`
}

type ClientsListCmd struct {
	Active bool   `help:"Show only currently connected clients (default: true)" default:"true"`
	Type   string `help:"Filter by client type (wired|wireless)" short:"t"`
}

func (cmd *ClientsListCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := getUniFiClient(flags)
	if err != nil {
		return err
	}

	var clients []unifi.ClientDevice

	if cmd.Active {
		clients, err = client.Clients().ListActive(ctx, cmd.Type)
	} else {
		clients, err = client.Clients().ListAll(ctx, cmd.Type)
	}

	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, clients)
	}

	if len(clients) == 0 {
		fmt.Fprintln(os.Stderr, "No clients found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d client(s)\n\n", len(clients))
	for _, c := range clients {
		printClientDevice(c)
	}

	return nil
}

type ClientsGetCmd struct {
	MAC string `arg:"" required:"" help:"Client MAC address"`
}

func (cmd *ClientsGetCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := getUniFiClient(flags)
	if err != nil {
		return err
	}

	result, err := client.Clients().Get(ctx, cmd.MAC)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	fmt.Printf("MAC:      %s\n", result.MAC)
	if result.Hostname != "" {
		fmt.Printf("Hostname: %s\n", result.Hostname)
	}
	if result.Name != "" {
		fmt.Printf("Name:     %s\n", result.Name)
	}
	if result.IP != "" {
		fmt.Printf("IP:       %s\n", result.IP)
	}
	fmt.Printf("Wired:    %t\n", result.IsWired)
	if result.Network != "" {
		fmt.Printf("Network:  %s\n", result.Network)
	}
	fmt.Printf("Blocked:  %t\n", result.Blocked)
	if result.Uptime > 0 {
		fmt.Printf("Uptime:   %ds\n", result.Uptime)
	}

	return nil
}

type ClientsBlockCmd struct {
	MAC string `arg:"" required:"" help:"Client MAC address to block"`
}

func (cmd *ClientsBlockCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := getUniFiClient(flags)
	if err != nil {
		return err
	}

	if err := client.Clients().Block(ctx, cmd.MAC); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("Client %s blocked", cmd.MAC),
		})
	}

	fmt.Fprintf(os.Stderr, "Client %s blocked\n", cmd.MAC)

	return nil
}

type ClientsUnblockCmd struct {
	MAC string `arg:"" required:"" help:"Client MAC address to unblock"`
}

func (cmd *ClientsUnblockCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := getUniFiClient(flags)
	if err != nil {
		return err
	}

	if err := client.Clients().Unblock(ctx, cmd.MAC); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("Client %s unblocked", cmd.MAC),
		})
	}

	fmt.Fprintf(os.Stderr, "Client %s unblocked\n", cmd.MAC)

	return nil
}

func printClientDevice(c unifi.ClientDevice) {
	fmt.Printf("MAC:  %s\n", c.MAC)

	name := c.Hostname
	if c.Name != "" {
		name = c.Name
	}

	if name != "" {
		fmt.Printf("  Name:    %s\n", name)
	}

	if c.IP != "" {
		fmt.Printf("  IP:      %s\n", c.IP)
	}

	connType := "wireless"
	if c.IsWired {
		connType = "wired"
	}

	fmt.Printf("  Type:    %s\n", connType)

	if c.Network != "" {
		fmt.Printf("  Network: %s\n", c.Network)
	}

	fmt.Println()
}
