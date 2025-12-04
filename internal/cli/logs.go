package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	logsLines  int
	logsType   string
	logsFollow bool
)

var logsCmd = &cobra.Command{
	Use:   "logs <name|id>",
	Short: "View process logs",
	Long:  `View logs for a process. Shows combined stdout/stderr by default.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			exitWithError("Failed to connect to daemon", err)
		}

		logs, err := client.GetLogs(args[0], logsLines, logsType)
		if err != nil {
			exitWithError("Failed to get logs", err)
		}

		fmt.Println(strings.Join(logs, "\n"))
	},
}

func init() {
	logsCmd.Flags().IntVarP(&logsLines, "lines", "n", 100, "Number of lines to show")
	logsCmd.Flags().StringVarP(&logsType, "type", "t", "", "Log type (stdout, stderr, or empty for combined)")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output (not implemented yet)")
}
