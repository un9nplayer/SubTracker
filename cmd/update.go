package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const repoReleasesURL = "https://api.github.com/repos/un9nplayer/SubTracker/releases/latest"

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for new SubTracker updates",
	Long:  `Check GitHub Releases for the latest version of SubTracker and get download instructions.`,
	RunE:  runUpdate,
}

type ghRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

func runUpdate(_ *cobra.Command, _ []string) error {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	red := color.New(color.FgRed, color.Bold)

	fmt.Println()
	cyan.Println("  🔍 Checking for SubTracker updates...")

	resp, err := http.Get(repoReleasesURL)
	if err != nil {
		red.Printf("  ✗ Failed to check for updates: %v\n\n", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		red.Printf("  ✗ Could not fetch release info (HTTP %d)\n\n", resp.StatusCode)
		return fmt.Errorf("HTTP error %d", resp.StatusCode)
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		red.Printf("  ✗ Failed to parse release info: %v\n\n", err)
		return err
	}

	currentVersion := "v" + rootCmd.Version

	fmt.Println()
	fmt.Printf("  Current version : %s\n", currentVersion)
	fmt.Printf("  Latest version  : %s\n", rel.TagName)
	fmt.Println()

	if rel.TagName == currentVersion {
		green.Println("  ✔ You are running the latest version of SubTracker!")
		fmt.Println()
		return nil
	}

	yellow.Printf("  🎉 A new version of SubTracker is available (%s)!\n\n", rel.TagName)
	fmt.Println("  To update:")
	cyan.Println("    go install github.com/un9nplayer/subtracker@latest")
	fmt.Println()
	fmt.Printf("  Or download pre-built binaries from:\n  %s\n\n", rel.HTMLURL)

	return nil
}
