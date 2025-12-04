package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "ps"},
	Short:   "List all processes",
	Long:    `List all managed processes with their status.`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			exitWithError("Failed to connect to daemon", err)
		}

		processes, err := client.List()
		if err != nil {
			exitWithError("Failed to list processes", err)
		}

		if len(processes) == 0 {
			fmt.Println("No processes running")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSTATUS\tPID\tCPU\tMEMORY\tUPTIME")

		for _, p := range processes {
			uptime := "-"
			if p.Uptime > 0 {
				uptime = formatDuration(time.Duration(p.Uptime) * time.Second)
			}

			cpu := "-"
			if p.Status == "running" {
				cpu = fmt.Sprintf("%.1f%%", p.CPU)
			}

			memory := "-"
			if p.Status == "running" && p.Memory > 0 {
				memory = formatBytes(p.Memory)
			}

			pid := "-"
			if p.PID > 0 {
				pid = fmt.Sprintf("%d", p.PID)
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				p.ID, p.Name, p.Status, pid, cpu, memory, uptime)
		}

		w.Flush()
	},
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fG", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1fM", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1fK", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
