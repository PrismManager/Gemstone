package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [name|id]",
	Short: "Show process status",
	Long:  `Show detailed status for a specific process or all processes.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			exitWithError("Failed to connect to daemon", err)
		}

		if len(args) == 0 {
			// Show all processes with stats
			showAllStats(client)
			return
		}

		// Show specific process
		info, err := client.Get(args[0])
		if err != nil {
			exitWithError("Failed to get process", err)
		}

		if info == nil {
			exitWithError("Process not found", nil)
		}

		fmt.Printf("Process: %s\n", info.Name)
		fmt.Printf("  ID:           %s\n", info.ID)
		fmt.Printf("  Status:       %s\n", info.Status)
		fmt.Printf("  PID:          %d\n", info.PID)
		fmt.Printf("  Command:      %s\n", info.Command)
		if len(info.Args) > 0 {
			fmt.Printf("  Args:         %v\n", info.Args)
		}
		if info.WorkDir != "" {
			fmt.Printf("  Working Dir:  %s\n", info.WorkDir)
		}
		fmt.Printf("  Auto-start:   %v\n", info.AutoStart)
		fmt.Printf("  Auto-restart: %v\n", info.AutoRestart)
		fmt.Printf("  Max restarts: %d\n", info.MaxRestarts)
		fmt.Printf("  Restart count:%d\n", info.RestartCount)
		fmt.Printf("  Created at:   %s\n", info.CreatedAt.Format("2006-01-02 15:04:05"))
		if info.StartedAt != nil {
			fmt.Printf("  Started at:   %s\n", info.StartedAt.Format("2006-01-02 15:04:05"))
		}
		if info.Status == "running" {
			fmt.Printf("  CPU:          %.1f%%\n", info.CPU)
			fmt.Printf("  Memory:       %s (%.1f%%)\n", formatBytes(info.Memory), info.MemoryPercent)
		}
	},
}

func showAllStats(client *Client) {
	stats, err := client.GetAllStats()
	if err != nil {
		exitWithError("Failed to get stats", err)
	}

	if len(stats) == 0 {
		fmt.Println("No running processes")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tPID\tCPU\tMEMORY\tTHREADS\tFDs\tREAD\tWRITE")

	for _, s := range stats {
		fmt.Fprintf(w, "%s\t%d\t%.1f%%\t%s\t%d\t%d\t%s\t%s\n",
			s.ID, s.PID, s.CPU, formatBytes(s.Memory),
			s.NumThreads, s.NumFDs,
			formatBytes(s.ReadBytes), formatBytes(s.WriteBytes))
	}

	w.Flush()
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show system and daemon information",
	Long:  `Show information about the system and the gemstone daemon.`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			exitWithError("Failed to connect to daemon", err)
		}

		info, err := client.GetSystemInfo()
		if err != nil {
			exitWithError("Failed to get system info", err)
		}

		fmt.Printf("Gemstone Daemon\n")
		fmt.Printf("  Version:        %s\n", info.Version)
		fmt.Printf("  Process count:  %d\n", info.ProcessCount)
		fmt.Println()
		fmt.Printf("System Stats\n")
		fmt.Printf("  CPU:            %.1f%%\n", info.SystemStats.CPUPercent)
		fmt.Printf("  Memory:         %s / %s (%.1f%%)\n",
			formatBytes(info.SystemStats.MemoryUsed),
			formatBytes(info.SystemStats.MemoryTotal),
			info.SystemStats.MemoryPercent)
		fmt.Printf("  Disk:           %s / %s (%.1f%%)\n",
			formatBytes(info.SystemStats.DiskUsed),
			formatBytes(info.SystemStats.DiskTotal),
			info.SystemStats.DiskPercent)
		if len(info.SystemStats.LoadAverage) >= 3 {
			fmt.Printf("  Load Average:   %.2f %.2f %.2f\n",
				info.SystemStats.LoadAverage[0],
				info.SystemStats.LoadAverage[1],
				info.SystemStats.LoadAverage[2])
		}
	},
}
