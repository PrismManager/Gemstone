package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	startName        string
	startWorkDir     string
	startAutoStart   bool
	startAutoRestart bool
	startMaxRestarts int
	startUser        string
	startEnv         []string
)

var startCmd = &cobra.Command{
	Use:   "start <command> [args...]",
	Short: "Start a new process",
	Long:  `Start a new managed process with the specified command and arguments.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClient()
		if err != nil {
			exitWithError("Failed to connect to daemon", err)
		}

		command := args[0]
		cmdArgs := []string{}
		if len(args) > 1 {
			cmdArgs = args[1:]
		}

		name := startName
		if name == "" {
			name = command
		}

		// Parse environment variables
		env := make(map[string]string)
		for _, e := range startEnv {
			for i, c := range e {
				if c == '=' {
					env[e[:i]] = e[i+1:]
					break
				}
			}
		}

		req := StartRequest{
			Name:        name,
			Command:     command,
			Args:        cmdArgs,
			WorkDir:     startWorkDir,
			Env:         env,
			AutoStart:   startAutoStart,
			AutoRestart: startAutoRestart,
			MaxRestarts: startMaxRestarts,
			User:        startUser,
		}

		info, err := client.Start(&req)
		if err != nil {
			exitWithError("Failed to start process", err)
		}

		fmt.Printf("Started process '%s' (ID: %s, PID: %d)\n", info.Name, info.ID, info.PID)
	},
}

func init() {
	startCmd.Flags().StringVarP(&startName, "name", "n", "", "Process name (defaults to command name)")
	startCmd.Flags().StringVarP(&startWorkDir, "cwd", "c", "", "Working directory")
	startCmd.Flags().BoolVar(&startAutoStart, "auto-start", true, "Auto-start on daemon restart")
	startCmd.Flags().BoolVar(&startAutoRestart, "auto-restart", true, "Auto-restart on crash")
	startCmd.Flags().IntVar(&startMaxRestarts, "max-restarts", 10, "Maximum restart attempts")
	startCmd.Flags().StringVarP(&startUser, "user", "u", "", "Run as user")
	startCmd.Flags().StringArrayVarP(&startEnv, "env", "e", []string{}, "Environment variables (KEY=VALUE)")
}
