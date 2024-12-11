package main

import (
	"fmt"
	"os"

	"github.com/dwrtz/sink/internal/analyzer"
	"github.com/dwrtz/sink/internal/processor"
	"github.com/dwrtz/sink/internal/tokens"
	"github.com/spf13/cobra"
)

func newAnalyzeCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "analyze [path]",
		Short: "Analyze codebase structure",
		Args:  cobra.ExactArgs(1),
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
			if format == "flat" {
				fmt.Println(a.FormatFlat(stats))
			} else if format == "tree" {
				fmt.Println(a.FormatFlat(stats)) // TODO: implement a.FormatTree
			} else {
				return fmt.Errorf("invalid format: %s (must be 'flat' or 'tree')", format)
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

	// Add analyze-specific flags
	cmd.Flags().StringVarP(&format, "format", "f", "flat", "Output format (flat or tree)")

	// These flags are inherited from the root command via the global config,
	// but we can add them here for command-specific help
	cmd.Flags().StringSliceVarP(&cfg.FilterPatterns, "filter", "i", nil, "Filter patterns to include files")
	cmd.Flags().StringSliceVarP(&cfg.ExcludePatterns, "exclude", "e", nil, "Patterns to exclude files")
	cmd.Flags().BoolVarP(&cfg.CaseSensitive, "case-sensitive", "c", false, "Use case-sensitive pattern matching")
	cmd.Flags().BoolVar(&cfg.ShowTokens, "tokens", false, "Show total token count")

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
