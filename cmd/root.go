package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "subtracker",
	Version: "1.0.1",
	Short:   "SubTracker — Professional Subdomain Discovery Tool",
	Long: `
  ███████╗██╗   ██╗██████╗ ████████╗██████╗  █████╗  ██████╗██╗  ██╗███████╗██████╗
  ██╔════╝██║   ██║██╔══██╗╚══██╔══╝██╔══██╗██╔══██╗██╔════╝██║ ██╔╝██╔════╝██╔══██╗
  ███████╗██║   ██║██████╔╝   ██║   ██████╔╝███████║██║     █████╔╝ █████╗  ██████╔╝
  ╚════██║██║   ██║██╔══██╗   ██║   ██╔══██╗██╔══██║██║     ██╔═██╗ ██╔══╝  ██╔══██╗
  ███████║╚██████╔╝██████╔╝   ██║   ██║  ██║██║  ██║╚██████╗██║  ██╗███████╗██║  ██║
  ╚══════╝ ╚═════╝ ╚═════╝    ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝

  Powered by AgniOps Intelligence Node | v1.0.0
  Run 'subtracker --help' for usage.
`,
	// Silence default usage on error — we print our own messages
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		color.Red("\n  ✗ Error: %v\n\n", err)
		os.Exit(1)
	}
}

func init() {
	// Disable the default completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Register sub-commands
	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(scanCmd)

	// Custom version template
	rootCmd.SetVersionTemplate(fmt.Sprintf(
		"\n  SubTracker v%s — Subdomain Discovery Tool\n  Powered by AgniOps Intelligence Node\n\n",
		"1.0.1",
	))
}
