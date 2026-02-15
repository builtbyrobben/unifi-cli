package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/unifi-cli/internal/outfmt"
)

type ISPMetricsCmd struct {
	Get ISPMetricsGetCmd `cmd:"" help:"Get ISP performance metrics"`
}

type ISPMetricsGetCmd struct {
	Type     string `arg:"" required:"" help:"Metric type: 5m or 1h"`
	Duration string `help:"Duration: 24h, 7d, or 30d" default:""`
	Begin    string `help:"Begin timestamp (epoch)" default:""`
	End      string `help:"End timestamp (epoch)" default:""`
}

func (cmd *ISPMetricsGetCmd) Run(ctx context.Context) error {
	client, err := getUniFiClient()
	if err != nil {
		return err
	}

	result, err := client.ISPMetrics().Get(ctx, cmd.Type, cmd.Duration, cmd.Begin, cmd.End)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if len(result.Data) == 0 {
		fmt.Fprintln(os.Stderr, "No ISP metrics found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d metric entry(s)\n\n", len(result.Data))

	for _, entry := range result.Data {
		fmt.Printf("Metric Type: %s\n", entry.MetricType)

		if entry.HostID != "" {
			fmt.Printf("  Host ID: %s\n", entry.HostID)
		}

		if entry.SiteID != "" {
			fmt.Printf("  Site ID: %s\n", entry.SiteID)
		}

		fmt.Printf("  Periods: %d\n", len(entry.Periods))

		for _, p := range entry.Periods {
			if p.MetricTime != "" {
				fmt.Printf("    Time: %s\n", p.MetricTime)
			}

			if p.Data != nil && p.Data.WAN != nil {
				w := p.Data.WAN
				if w.AvgLatency > 0 {
					fmt.Printf("      Avg Latency:  %.1fms\n", w.AvgLatency)
				}
				if w.MaxLatency > 0 {
					fmt.Printf("      Max Latency:  %.1fms\n", w.MaxLatency)
				}
				if w.DownloadKbps > 0 {
					fmt.Printf("      Download:     %.0f kbps\n", w.DownloadKbps)
				}
				if w.UploadKbps > 0 {
					fmt.Printf("      Upload:       %.0f kbps\n", w.UploadKbps)
				}
				if w.PacketLoss > 0 {
					fmt.Printf("      Packet Loss:  %.2f%%\n", w.PacketLoss)
				}
				if w.ISPName != "" {
					fmt.Printf("      ISP:          %s\n", w.ISPName)
				}
			}
		}

		fmt.Println()
	}

	return nil
}
