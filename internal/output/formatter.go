// Package output handles all terminal and file output formatting for SubTracker.
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"subtracker/internal/api"
)

// ─── Spinner ───────────────────────────────────────────────────────────────────

// Spinner renders an animated braille spinner on stderr until a value is sent
// on the done channel. Writing to stderr keeps stdout clean for piped output.
// Call from a goroutine; send struct{}{} to stop.
func Spinner(msg string, done <-chan struct{}) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	cyan := color.New(color.FgCyan)
	// Force colour output on stderr (colour is disabled for non-TTY by default)
	cyan.EnableColor()
	i := 0
	for {
		select {
		case <-done:
			// Clear the spinner line on stderr
			fmt.Fprintf(os.Stderr, "\r%s\r", repeat(" ", len(msg)+12))
			return
		default:
			fmt.Fprintf(os.Stderr, "\r  %s  %s... ", frames[i%len(frames)], msg)
			time.Sleep(80 * time.Millisecond)
			i++
		}
	}
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

// ─── Table (default) ──────────────────────────────────────────────────────────

// PrintTable renders a rich, coloured ASCII table of scan results.
func PrintTable(result *api.ScanResult, elapsed time.Duration) {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	yellow := color.New(color.FgYellow)
	white := color.New(color.FgWhite)

	fmt.Println()
	cyan.Println("  ╔══════════════════════════════════════════════╗")
	cyan.Println("  ║        SubTracker — Scan Results             ║")
	cyan.Println("  ╚══════════════════════════════════════════════╝")
	fmt.Println()

	white.Printf("  %-14s", "🌐 Domain")
	fmt.Printf(": %s\n", result.Domain)

	white.Printf("  %-14s", "🔍 Engine")
	fmt.Printf(": %s\n", result.Engine)

	white.Printf("  %-14s", "🗺  Country")
	fmt.Printf(": %s\n", result.Country)

	white.Printf("  %-14s", "🔝 Top IP")
	fmt.Printf(": %s\n", result.MostUsedIP)

	white.Printf("  %-14s", "📅 Scan Date")
	fmt.Printf(": %s\n", result.ScanDate)

	green.Printf("  %-14s", "✔  Found")
	green.Printf(": %d subdomain(s)\n", result.SubdomainsCount)

	fmt.Println()

	// Build table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"#", "Subdomain", "IP Address", "Cloudflare"})
	table.SetBorder(true)
	table.SetCenterSeparator("┼")
	table.SetColumnSeparator("│")
	table.SetRowSeparator("─")
	table.SetHeaderLine(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
	)
	table.SetColumnColor(
		tablewriter.Colors{tablewriter.FgHiBlackColor},
		tablewriter.Colors{tablewriter.FgHiWhiteColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgCyanColor},
		tablewriter.Colors{},
	)

	for i, s := range result.Subdomains {
		cfLabel := "  ✗ No"
		cfColor := tablewriter.Colors{tablewriter.FgRedColor}
		if s.Cloudflare {
			cfLabel = "  ✔ Yes"
			cfColor = tablewriter.Colors{tablewriter.FgGreenColor, tablewriter.Bold}
		}

		table.Rich(
			[]string{
				fmt.Sprintf(" %d", i+1),
				" " + s.Subdomain,
				" " + s.IP,
				cfLabel,
			},
			[]tablewriter.Colors{
				{tablewriter.FgHiBlackColor},
				{tablewriter.FgHiWhiteColor, tablewriter.Bold},
				{tablewriter.FgCyanColor},
				cfColor,
			},
		)
		_ = yellow // suppress unused warning; used in future
	}

	table.Render()
}

// ─── JSON ─────────────────────────────────────────────────────────────────────

// PrintJSON outputs the full scan result as pretty-printed JSON.
func PrintJSON(result *api.ScanResult) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(result)
}

// ─── Plain ────────────────────────────────────────────────────────────────────

// PrintPlain outputs one subdomain per line — ideal for piping to other tools.
func PrintPlain(result *api.ScanResult) {
	for _, s := range result.Subdomains {
		fmt.Println(s.Subdomain)
	}
}

// ─── CSV ──────────────────────────────────────────────────────────────────────

// PrintCSV outputs results in CSV format to stdout.
func PrintCSV(result *api.ScanResult) {
	fmt.Println("Subdomain,IP Address,Cloudflare")
	for _, s := range result.Subdomains {
		cf := "No"
		if s.Cloudflare {
			cf = "Yes"
		}
		fmt.Printf("%s,%s,%s\n", s.Subdomain, s.IP, cf)
	}
}
