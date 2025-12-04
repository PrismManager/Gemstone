package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Gemstone version %s\n", version)
	},
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Daemon management commands",
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon",
	Long:  `Start the gemstone daemon. Usually managed by systemd.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use 'systemctl start gemstone' to start the daemon")
	},
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the daemon",
	Long:  `Stop the gemstone daemon. Usually managed by systemd.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use 'systemctl stop gemstone' to stop the daemon")
	},
}

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check daemon status",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			fmt.Println("Daemon is not running")
			return
		}

		info, err := client.GetSystemInfo()
		if err != nil {
			fmt.Println("Daemon is not responding")
			return
		}

		fmt.Printf("Daemon is running (version %s)\n", info.Version)
		fmt.Printf("Managing %d processes\n", info.ProcessCount)
	},
}

func init() {
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
}
