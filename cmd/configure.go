package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"subtracker/internal/config"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Set up your AgniOps API key",
	Long:  `Interactive wizard to store your AgniOps API key securely on disk.`,
	RunE:  runConfigure,
}

func runConfigure(_ *cobra.Command, _ []string) error {
	bold := color.New(color.FgCyan, color.Bold)
	warn := color.New(color.FgYellow, color.Bold)
	ok := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)

	fmt.Println()
	bold.Println("  ┌──────────────────────────────────────────┐")
	bold.Println("  │   SubTracker — API Key Configuration     │")
	bold.Println("  └──────────────────────────────────────────┘")
	fmt.Println()

	// Show existing key status
	cfg, _ := config.Load()
	if cfg.APIKey != "" {
		masked := maskKey(cfg.APIKey)
		warn.Printf("  ⚠  Current key: %s  (will be replaced)\n\n", masked)
	} else {
		fmt.Println("  No API key configured yet.")
		fmt.Println()
	}

	fmt.Print("  Enter your AgniOps API key: ")
	reader := bufio.NewReader(os.Stdin)
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		red.Println("\n  ✗ API key cannot be empty.")
		return fmt.Errorf("API key is required")
	}

	if !strings.HasPrefix(apiKey, "at_live_") {
		fmt.Println()
		warn.Println("  ⚠  Warning: Key does not start with 'at_live_' — saving anyway.")
	}

	cfg.APIKey = apiKey
	configPath, err := config.Save(cfg)
	if err != nil {
		red.Printf("\n  ✗ Failed to save API key: %v\n", err)
		return err
	}

	fmt.Println()
	ok.Println("  ✔  API key saved successfully!")
	fmt.Printf("  📁 Config file: %s\n", configPath)
	fmt.Println()
	fmt.Println("  You can now run:")
	bold.Println("    subtracker scan --domain example.com")
	fmt.Println()

	return nil
}

// maskKey shows only the prefix and last 4 characters of the API key.
func maskKey(key string) string {
	if len(key) <= 12 {
		return strings.Repeat("*", len(key))
	}
	return key[:8] + strings.Repeat("*", len(key)-12) + key[len(key)-4:]
}
