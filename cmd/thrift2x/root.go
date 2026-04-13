package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/81120/thrift2x/internal/converter"
	"github.com/81120/thrift2x/internal/targets"
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "thrift2x",
		Short: "Convert thrift files to typed definitions",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	cmd.PersistentFlags().Bool("help", false, "help for thrift2x")

	cmd.AddCommand(newGenerateCmd())
	cmd.AddCommand(newTargetsCmd())
	return cmd
}

func newGenerateCmd() *cobra.Command {
	var inDir string
	var outDir string
	var exclude string
	var i64Type string
	var jobs string
	var target string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate typed definitions from thrift files",
		RunE: func(_ *cobra.Command, args []string) error {
			if strings.TrimSpace(inDir) == "" || strings.TrimSpace(outDir) == "" {
				return fmt.Errorf("both -in and -out are required")
			}

			targetName := strings.TrimSpace(target)

			targetOptions := map[string]string{}
			if targetName == "typescript" && strings.TrimSpace(i64Type) != "" {
				targetOptions["i64-type"] = i64Type
			}

			stats, err := converter.Run(converter.Config{
				InDir:   inDir,
				OutDir:  outDir,
				Exclude: splitExclude(exclude),
				Target:  targetName,
				TargetOptions: targetOptions,
				Jobs: jobs,
			})

			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				return err
			}

			fmt.Printf("done: total=%d success=%d failed=%d jobs=%d\n", stats.Total, stats.Success, stats.Failed, stats.Jobs)
			fmt.Printf("timing: scan=%s convert=%s total=%s\n", converter.FormatDuration(stats.ScanDuration), converter.FormatDuration(stats.ConvertDuration), converter.FormatDuration(stats.TotalDuration))

			if stats.Failed > 0 {
				return fmt.Errorf("conversion finished with failures")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&inDir, "in", "", "input directory containing .thrift files")
	cmd.Flags().StringVar(&outDir, "out", "", "output directory for generated files")
	cmd.Flags().StringVar(&exclude, "exclude", "", "comma-separated path substrings to exclude")
	cmd.Flags().StringVar(&target, "target", "", "output target language (e.g. typescript)")
	cmd.Flags().StringVar(&i64Type, "i64-type", "string", "typescript target option: Type for thrift i64 (string|number|bigint)")
	cmd.Flags().StringVar(&jobs, "jobs", "auto", "parallel workers: integer or 'auto'")
	_ = cmd.MarkFlagRequired("target")
	return cmd
}

func newTargetsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "targets",
		Short: "List supported output targets",
		Run: func(cmd *cobra.Command, args []string) {
			for _, t := range targets.List() {
				fmt.Fprintln(os.Stdout, t)
			}
		},
	}
}

func splitExclude(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
