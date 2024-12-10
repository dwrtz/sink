package main

import (
	"fmt"

	"github.com/dwrtz/sink/internal/analyzer"
	"github.com/dwrtz/sink/internal/processor"
	"github.com/spf13/cobra"
)

func newAnalyzeCmd() *cobra.Command {
	var (
		format          string
		filterPatterns  []string
		excludePatterns []string
		caseSensitive   bool
	)

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze codebase structure",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create file processor
			fp, err := processor.NewFileProcessor(processor.Config{
				Paths:           args,
				FilterPatterns:  filterPatterns,
				ExcludePatterns: excludePatterns,
				CaseSensitive:   caseSensitive,
			})
			if err != nil {
				return fmt.Errorf("failed to create file processor: %w", err)
			}

			// Process files
			files, err := fp.Process()
			if err != nil {
				return fmt.Errorf("failed to process files: %w", err)
			}

			// Convert FileInfo to paths for analyzer
			var paths []string
			for _, f := range files {
				paths = append(paths, f.Path)
			}

			// Create and run analyzer
			a := analyzer.New()
			stats, err := a.Analyze(paths)
			if err != nil {
				return fmt.Errorf("failed to analyze codebase: %w", err)
			}

			// Output results based on format
			if format == "flat" {
				fmt.Println(a.FormatFlat(stats))
			} else {
				// Tree format not implemented yet
				return fmt.Errorf("tree format not implemented")
			}

			// Print extension list
			fmt.Printf("\nExtensions: %s\n", a.GetExtensionList(stats))

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&format, "format", "f", "flat", "Output format (flat or tree)")
	cmd.Flags().StringSliceVarP(&filterPatterns, "filter", "i", nil, "Filter patterns to include files")
	cmd.Flags().StringSliceVarP(&excludePatterns, "exclude", "e", nil, "Patterns to exclude files")
	cmd.Flags().BoolVarP(&caseSensitive, "case-sensitive", "c", false, "Use case-sensitive pattern matching")

	return cmd
}
