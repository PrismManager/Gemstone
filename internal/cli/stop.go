package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop <name|id>",
	Short: "Stop a running process",
	Long:  `Stop a running process by name or ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			exitWithError("Failed to connect to daemon", err)
		}

		if err := client.Stop(args[0]); err != nil {
			exitWithError("Failed to stop process", err)
		}

		fmt.Printf("Stopped process '%s'\n", args[0])
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart <name|id>",
	Short: "Restart a process",
	Long:  `Restart a process by name or ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			exitWithError("Failed to connect to daemon", err)
		}

		if err := client.Restart(args[0]); err != nil {
			exitWithError("Failed to restart process", err)
		}

		fmt.Printf("Restarted process '%s'\n", args[0])
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete <name|id>",
	Short: "Delete a process",
	Long:  `Delete a process by name or ID. The process will be stopped if running.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			exitWithError("Failed to connect to daemon", err)
		}

		if err := client.Delete(args[0]); err != nil {
			exitWithError("Failed to delete process", err)
		}

		fmt.Printf("Deleted process '%s'\n", args[0])
	},
}
