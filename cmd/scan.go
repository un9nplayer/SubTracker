package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/un9nplayer/subtracker/internal/api"
	"github.com/un9nplayer/subtracker/internal/config"
	"github.com/un9nplayer/subtracker/internal/output"
)

var (
	flagDomain    string
	flagOutputFmt string
	flagOutFile   string
	flagTimeout   int
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Discover subdomains for a target domain",
	Long: `Query the AgniOps Intelligence engine to enumerate subdomains for any domain.

Examples:
  subtracker scan --domain example.com
  subtracker scan -d example.com -o json
  subtracker scan -d example.com -o csv --out-file results.csv
  subtracker scan -d example.com -o plain | sort`,
	RunE: runScan,
}

func init() {
	scanCmd.Flags().StringVarP(&flagDomain, "domain", "d", "", "Target domain to scan (required)")
	scanCmd.Flags().StringVarP(&flagOutputFmt, "output", "o", "table", "Output format: table | json | plain | csv")
	scanCmd.Flags().StringVarP(&flagOutFile, "out-file", "f", "", "Save results to a file (path)")
	scanCmd.Flags().IntVarP(&flagTimeout, "timeout", "t", 30, "HTTP timeout in seconds")
	_ = scanCmd.MarkFlagRequired("domain")
}

func runScan(_ *cobra.Command, _ []string) error {
	red := color.New(color.FgRed, color.Bold)
	green := color.New(color.FgGreen, color.Bold)

	// ── 1. Load API key ────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil || cfg.APIKey == "" {
		fmt.Println()
		red.Println("  ✗ No API key configured.")
		fmt.Println("  Run:  subtracker configure")
		fmt.Println()
		return fmt.Errorf("API key not set")
	}

	// ── 2. Validate output format ──────────────────────────────────────────────
	switch flagOutputFmt {
	case "table", "json", "plain", "txt", "csv":
		// valid
	default:
		return fmt.Errorf("unknown output format %q — use: table, json, plain, csv", flagOutputFmt)
	}

	// ── 3. Run scan with spinner ───────────────────────────────────────────────
	client := api.NewClient(cfg.APIKey, time.Duration(flagTimeout)*time.Second)

	done := make(chan struct{}, 1)
	go output.Spinner(fmt.Sprintf("Scanning subdomains for %s", flagDomain), done)

	start := time.Now()
	result, err := client.Scan(flagDomain)
	done <- struct{}{}
	elapsed := time.Since(start)

	if err != nil {
		fmt.Println()
		red.Printf("\n  ✗ Scan failed: %v\n\n", err)
		return err
	}

	// ── 4. Print to stdout ─────────────────────────────────────────────────────
	switch flagOutputFmt {
	case "json":
		output.PrintJSON(result)
	case "plain", "txt":
		output.PrintPlain(result)
	case "csv":
		output.PrintCSV(result)
	default:
		output.PrintTable(result, elapsed)
	}

	// ── 5. Save to file ────────────────────────────────────────────────────────
	if flagOutFile != "" {
		if saveErr := saveResultToFile(result, flagOutFile, flagOutputFmt); saveErr != nil {
			red.Printf("  ✗ Failed to save file: %v\n", saveErr)
		} else {
			green.Printf("  📄 Saved to: %s\n", flagOutFile)
		}
	}

	// ── 6. Footer (table mode only) ────────────────────────────────────────────
	if flagOutputFmt == "table" {
		fmt.Printf("\n  📊 Quota remaining today: %d / %d\n",
			result.Meta.QuotaRemaining, result.Meta.DailyQuota)
		green.Printf("  ✔  Scan complete in %.2fs\n\n", elapsed.Seconds())
	}

	return nil
}

// saveResultToFile writes the scan result to a file in the requested format.
func saveResultToFile(result *api.ScanResult, path, format string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot create file: %w", err)
	}
	defer f.Close()

	switch format {
	case "json":
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		return enc.Encode(result)

	case "csv":
		w := csv.NewWriter(f)
		if err := w.Write([]string{"Subdomain", "IP Address", "Cloudflare"}); err != nil {
			return err
		}
		for _, s := range result.Subdomains {
			cf := "No"
			if s.Cloudflare {
				cf = "Yes"
			}
			if err := w.Write([]string{s.Subdomain, s.IP, cf}); err != nil {
				return err
			}
		}
		w.Flush()
		return w.Error()

	default: // plain / txt / table — one subdomain per line
		for _, s := range result.Subdomains {
			if _, err := fmt.Fprintln(f, s.Subdomain); err != nil {
				return err
			}
		}
		return nil
	}
}
