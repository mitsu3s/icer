/*
Copyright (c) 2026 mitsu3s

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mitsu3s/icer/internal/config"
	"github.com/mitsu3s/icer/internal/icmp"
	"github.com/spf13/cobra"
)

var version = "0.2.0"

func main() {
	log.SetFlags(0)

	root := &cobra.Command{
		Use:   "icer",
		Short: "Send ICMP error messages for network testing",
		Long: `icer crafts and sends ICMP error messages (Destination Unreachable,
Redirect, Time Exceeded) to specified targets.

Designed for testing and debugging how systems handle ICMP error responses.

Use with caution — improper use may violate network policies or disrupt services.`,
		SilenceUsage: true,
	}

	root.AddCommand(
		newSendCmd(),
		newListTypesCmd(),
		newInitCmd(),
		newVersionCmd(),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// --- send command ---

func newSendCmd() *cobra.Command {
	var (
		icmpType uint8
		icmpCode uint8
		srcIP    string
		dstIP    string
		errSrc   string
		errDst   string
		profile  string
		cfgFile  string
		ttl      uint8
		verbose  bool
		dryRun   bool
		count    int
		interval time.Duration
		jsonOut  bool
	)

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send ICMP error packets",
		Long: `Send one or more ICMP error packets to a target.

Targets can be specified either via CLI flags or a configuration profile.

Examples:
  # CLI-only (no config file needed)
  icer send -t 3 -c 0 --src 192.168.1.1 --dst 192.168.1.2 \
    --error-src 10.0.0.1 --error-dst 10.0.0.2

  # Using a profile from config file
  icer send -t 3 -c 0 -p default

  # Dry run with verbose output
  icer send -t 5 -c 0 --src 10.0.0.1 --dst 10.0.0.2 \
    --error-src 172.16.0.1 --error-dst 172.16.0.2 --dry-run -v

  # Send 5 packets with 1 second interval
  icer send -t 3 -c 1 -p default --count 5 --interval 1s`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSend(cmd, sendFlags{
				icmpType: icmpType,
				icmpCode: icmpCode,
				srcIP:    srcIP,
				dstIP:    dstIP,
				errSrc:   errSrc,
				errDst:   errDst,
				profile:  profile,
				cfgFile:  cfgFile,
				ttl:      ttl,
				verbose:  verbose,
				dryRun:   dryRun,
				count:    count,
				interval: interval,
				jsonOut:  jsonOut,
			})
		},
	}

	f := cmd.Flags()
	f.Uint8VarP(&icmpType, "type", "t", 0, "ICMP type (3=unreachable, 5=redirect, 11=time-exceeded)")
	f.Uint8VarP(&icmpCode, "code", "c", 0, "ICMP code")
	f.StringVar(&srcIP, "src", "", "Source IP (outer header)")
	f.StringVar(&dstIP, "dst", "", "Destination IP (outer header)")
	f.StringVar(&errSrc, "error-src", "", "Source IP of the inner error packet")
	f.StringVar(&errDst, "error-dst", "", "Destination IP of the inner error packet")
	f.StringVarP(&profile, "profile", "p", "", "Configuration profile name")
	f.StringVar(&cfgFile, "config", "", "Path to config file")
	f.Uint8Var(&ttl, "ttl", 64, "IP Time To Live")
	f.BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	f.BoolVar(&dryRun, "dry-run", false, "Show what would be sent without sending")
	f.IntVar(&count, "count", 1, "Number of packets to send per target")
	f.DurationVar(&interval, "interval", time.Second, "Interval between repeated sends")
	f.BoolVar(&jsonOut, "json", false, "Output results as JSON")

	_ = cmd.MarkFlagRequired("type")

	return cmd
}

type sendFlags struct {
	icmpType uint8
	icmpCode uint8
	srcIP    string
	dstIP    string
	errSrc   string
	errDst   string
	profile  string
	cfgFile  string
	ttl      uint8
	verbose  bool
	dryRun   bool
	count    int
	interval time.Duration
	jsonOut  bool
}

// sendResult captures the outcome of a single send attempt.
type sendResult struct {
	Target   int    `json:"target"`
	Attempt  int    `json:"attempt"`
	SrcIP    string `json:"src_ip"`
	DstIP    string `json:"dst_ip"`
	TypeCode string `json:"type_code"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

func runSend(cmd *cobra.Command, f sendFlags) error {
	if err := icmp.ValidateTypeCode(f.icmpType, f.icmpCode); err != nil {
		return err
	}

	type target struct {
		srcIP, dstIP, errSrc, errDst string
	}

	var targets []target

	hasCLIIPs := f.srcIP != "" || f.dstIP != "" || f.errSrc != "" || f.errDst != ""

	if hasCLIIPs {
		if f.srcIP == "" || f.dstIP == "" || f.errSrc == "" || f.errDst == "" {
			return fmt.Errorf("when using CLI IP flags, all four are required: --src, --dst, --error-src, --error-dst")
		}
		targets = append(targets, target{f.srcIP, f.dstIP, f.errSrc, f.errDst})
	} else {
		cfgPath := config.ResolveConfigPath(f.cfgFile)
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}
		if cfg == nil {
			return fmt.Errorf("no targets specified; use --src/--dst flags or create a config file (icer init)")
		}

		if !cmd.Flags().Changed("verbose") {
			f.verbose = cfg.Options.Verbose
		}
		if !cmd.Flags().Changed("dry-run") {
			f.dryRun = cfg.Options.DryRun
		}
		if !cmd.Flags().Changed("ttl") {
			f.ttl = cfg.Options.TTL
		}

		prof, err := cfg.GetProfile(f.profile)
		if err != nil {
			return err
		}

		for _, t := range prof.Targets {
			targets = append(targets, target{t.RealIP.SrcIP, t.RealIP.DstIP, t.ErrorIP.SrcIP, t.ErrorIP.DstIP})
		}

		if f.verbose {
			fmt.Fprintf(os.Stderr, "Profile: %s", prof.Name)
			if prof.Description != "" {
				fmt.Fprintf(os.Stderr, " (%s)", prof.Description)
			}
			fmt.Fprintln(os.Stderr)
		}
	}

	typeName := icmp.TypeName(f.icmpType)
	totalTargets := len(targets)

	if f.verbose {
		fmt.Fprintf(os.Stderr, "ICMP: %s (Type=%d, Code=%d)\n", typeName, f.icmpType, f.icmpCode)
		fmt.Fprintf(os.Stderr, "Targets: %d, Count: %d, TTL: %d\n", totalTargets, f.count, f.ttl)
		if f.dryRun {
			fmt.Fprintln(os.Stderr, "Mode: DRY RUN")
		}
		fmt.Fprintln(os.Stderr)
	}

	var results []sendResult
	successCount := 0
	failureCount := 0

	for ti, t := range targets {
		for ci := range f.count {
			if ci > 0 {
				time.Sleep(f.interval)
			}

			params := &icmp.SendParams{
				Type:       f.icmpType,
				Code:       f.icmpCode,
				SrcIP:      t.srcIP,
				DstIP:      t.dstIP,
				ErrorSrcIP: t.errSrc,
				ErrorDstIP: t.errDst,
				TTL:        f.ttl,
				DryRun:     f.dryRun,
				Verbose:    f.verbose,
			}

			if !f.jsonOut && (totalTargets > 1 || f.count > 1) {
				label := fmt.Sprintf("[%d/%d]", ti+1, totalTargets)
				if f.count > 1 {
					label += fmt.Sprintf(" (attempt %d/%d)", ci+1, f.count)
				}
				fmt.Fprintf(os.Stderr, "%s %s -> %s\n", label, t.srcIP, t.dstIP)
			}

			err := icmp.Send(params)

			r := sendResult{
				Target:   ti + 1,
				Attempt:  ci + 1,
				SrcIP:    t.srcIP,
				DstIP:    t.dstIP,
				TypeCode: fmt.Sprintf("%d/%d", f.icmpType, f.icmpCode),
			}

			if err != nil {
				failureCount++
				r.Status = "failed"
				r.Error = err.Error()
				if !f.jsonOut {
					log.Printf("  [FAILED] %v", err)
				}
			} else {
				successCount++
				r.Status = "success"
			}
			results = append(results, r)
		}
	}

	if f.jsonOut {
		out := struct {
			Results []sendResult `json:"results"`
			Summary struct {
				Total   int `json:"total"`
				Success int `json:"success"`
				Failed  int `json:"failed"`
			} `json:"summary"`
		}{
			Results: results,
		}
		out.Summary.Total = len(results)
		out.Summary.Success = successCount
		out.Summary.Failed = failureCount

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	total := successCount + failureCount
	if total > 1 {
		fmt.Fprintf(os.Stderr, "\n=== Summary ===\n")
		fmt.Fprintf(os.Stderr, "Total: %d  Success: %d  Failed: %d\n", total, successCount, failureCount)
		fmt.Fprintf(os.Stderr, "===============\n")
	}

	if failureCount > 0 && successCount == 0 {
		return fmt.Errorf("all sends failed")
	}

	return nil
}

// --- list-types command ---

func newListTypesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-types",
		Short: "Show supported ICMP types and codes",
		Run: func(cmd *cobra.Command, args []string) {
			for _, t := range icmp.SupportedTypes {
				fmt.Printf("Type %d: %s\n", t.Type, t.Name)
				fmt.Printf("  %s\n", t.Description)
				for _, c := range t.Codes {
					fmt.Printf("  Code %d: %s\n", c.Code, c.Description)
				}
				fmt.Println()
			}
		},
	}
}

// --- init command ---

func newInitCmd() *cobra.Command {
	var outPath string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate a sample configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if outPath == "" {
				outPath = config.DefaultConfigPath()
			}
			if outPath == "" {
				return fmt.Errorf("cannot determine config path; use --output to specify")
			}

			if _, err := os.Stat(outPath); err == nil {
				return fmt.Errorf("config file already exists: %s (use --output for a different path)", outPath)
			}

			dir := filepath.Dir(outPath)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("creating config directory: %w", err)
			}

			if err := os.WriteFile(outPath, []byte(config.SampleConfig), 0o644); err != nil {
				return fmt.Errorf("writing config file: %w", err)
			}

			fmt.Printf("Created config file: %s\n", outPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outPath, "output", "o", "", "Output path (default: ~/.config/icer/config.yaml)")

	return cmd
}

// --- version command ---

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show icer version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("icer version %s\n", version)
		},
	}
}
