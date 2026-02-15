package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/unifi-cli/internal/outfmt"
)

type StatsCmd struct {
	Overview StatsOverviewCmd `cmd:"" help:"Site health overview"`
	Site     StatsSiteCmd     `cmd:"" help:"Site system info"`
}

type StatsOverviewCmd struct{}

func (cmd *StatsOverviewCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := getUniFiClient(flags)
	if err != nil {
		return err
	}

	health, err := client.Stats().Health(ctx)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, health)
	}

	if len(health) == 0 {
		fmt.Fprintln(os.Stderr, "No health data available")
		return nil
	}

	fmt.Fprintln(os.Stderr, "Site Health Overview")
	fmt.Fprintln(os.Stderr)
	for _, h := range health {
		fmt.Printf("Subsystem: %s\n", h.Subsystem)
		fmt.Printf("  Status:  %s\n", h.Status)
		if h.NumUser > 0 {
			fmt.Printf("  Users:   %d\n", h.NumUser)
		}
		if h.NumGuest > 0 {
			fmt.Printf("  Guests:  %d\n", h.NumGuest)
		}
		if h.NumAdopted > 0 {
			fmt.Printf("  Adopted: %d\n", h.NumAdopted)
		}
		if h.ISPName != "" {
			fmt.Printf("  ISP:     %s\n", h.ISPName)
		}
		if h.Latency > 0 {
			fmt.Printf("  Latency: %dms\n", h.Latency)
		}
		fmt.Println()
	}

	return nil
}

type StatsSiteCmd struct{}

func (cmd *StatsSiteCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := getUniFiClient(flags)
	if err != nil {
		return err
	}

	sysinfo, err := client.Stats().SysInfo(ctx)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, sysinfo)
	}

	if len(sysinfo) == 0 {
		fmt.Fprintln(os.Stderr, "No system info available")
		return nil
	}

	info := sysinfo[0]
	fmt.Println("Site System Info")
	fmt.Println()
	if info.Name != "" {
		fmt.Printf("Name:     %s\n", info.Name)
	}
	if info.Hostname != "" {
		fmt.Printf("Hostname: %s\n", info.Hostname)
	}
	if info.Version != "" {
		fmt.Printf("Version:  %s\n", info.Version)
	}
	if info.Timezone != "" {
		fmt.Printf("Timezone: %s\n", info.Timezone)
	}
	if info.BuildNumber != "" {
		fmt.Printf("Build:    %s\n", info.BuildNumber)
	}
	fmt.Printf("Update:   %t\n", info.UpdateAvail)

	return nil
}
