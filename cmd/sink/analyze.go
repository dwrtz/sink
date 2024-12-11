package main

import (
	"fmt"
	"os"

	"github.com/dwrtz/sink/internal/analyzer"
	"github.com/dwrtz/sink/internal/processor"
	"github.com/dwrtz/sink/internal/tokens"
	"github.com/spf13/cobra"
)

type analyzeFlags struct {
	format          string
	filterPatterns  []string
	excludePatterns []string
	caseSensitive   bool
	showTokens      bool
}

func newAnalyzeCmd() *cobra.Command {
	flags := &analyzeFlags{}

	cmd := &cobra.Command{
		Use:   "analyze [path]",
		Short: "Analyze codebase structure",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Only override config values if flags were explicitly set
			if cmd.Flags().Changed("filter") {
				cfg.FilterPatterns = flags.filterPatterns
			}
			if cmd.Flags().Changed("exclude") {
				cfg.ExcludePatterns = flags.excludePatterns
			}
			if cmd.Flags().Changed("case-sensitive") {
				cfg.CaseSensitive = flags.caseSensitive
			}
			if cmd.Flags().Changed("tokens") {
				cfg.ShowTokens = flags.showTokens
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			// Validate path
			if _, err := os.Stat(path); err != nil {
				return fmt.Errorf("invalid repository path %s: %w", path, err)
			}

			// Create file processor using the global config
			fp, err := processor.NewFileProcessor(processor.Config{
				RepoRoot:        path,
				FilterPatterns:  cfg.FilterPatterns,
				ExcludePatterns: cfg.ExcludePatterns,
				CaseSensitive:   cfg.CaseSensitive,
				SyntaxMap:       cfg.SyntaxMap,
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
			if flags.format == "flat" {
				fmt.Println(a.FormatFlat(stats))
			} else if flags.format == "tree" {
				fmt.Println(a.FormatFlat(stats)) // TODO: implement a.FormatTree
			} else {
				return fmt.Errorf("invalid format: %s (must be 'flat' or 'tree')", flags.format)
			}

			// Print extension list
			fmt.Printf("\nExtensions: %s\n", a.GetExtensionList(stats))

			// Add token counting if enabled
			if cfg.ShowTokens {
				totalTokens := 0
				for _, file := range files {
					tokens, err := countFileTokens(file.Content, cfg.TokenEncoding)
					if err != nil {
						return fmt.Errorf("failed to count tokens: %w", err)
					}
					totalTokens += tokens
				}
				fmt.Printf("\nTotal tokens in codebase: %d\n", totalTokens)
			}

			return nil
		},
	}

	// Add flags bound to the local flags struct
	cmd.Flags().StringVarP(&flags.format, "format", "f", "flat", "Output format (flat or tree)")
	cmd.Flags().StringSliceVarP(&flags.filterPatterns, "filter", "i", nil, "Filter patterns to include files")
	cmd.Flags().StringSliceVarP(&flags.excludePatterns, "exclude", "e", nil, "Patterns to exclude files")
	cmd.Flags().BoolVarP(&flags.caseSensitive, "case-sensitive", "c", false, "Use case-sensitive pattern matching")
	cmd.Flags().BoolVar(&flags.showTokens, "tokens", false, "Show total token count")

	return cmd
}

// countFileTokens helper function to count tokens in a file
func countFileTokens(content, encoding string) (int, error) {
	counter, err := tokens.NewCounter(encoding)
	if err != nil {
		return 0, err
	}
	return counter.Count(content)
}
