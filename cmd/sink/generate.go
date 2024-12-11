package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dwrtz/sink/internal/processor"
	"github.com/dwrtz/sink/internal/processor/markdown"
	"github.com/dwrtz/sink/internal/processor/template"
	"github.com/dwrtz/sink/internal/tokens"
	"github.com/spf13/cobra"
)

func newGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate [path]",
		Short: "Generate markdown documentation from code files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			// Validate path
			if _, err := os.Stat(path); err != nil {
				return fmt.Errorf("invalid repository path %s: %w", path, err)
			}

			// Create file processor with config
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

			var content string
			if cfg.TemplatePath != "" {
				// Use template if specified
				templateContent, err := os.ReadFile(cfg.TemplatePath)
				if err != nil {
					return fmt.Errorf("failed to read template: %w", err)
				}

				te := template.NewEngine(string(templateContent))
				content, err = te.Execute(files)
				if err != nil {
					return fmt.Errorf("failed to execute template: %w", err)
				}
			} else {
				// Use default markdown generator
				mg := markdown.NewGenerator(markdown.Config{
					NoCodeBlock:   cfg.NoCodeblock,
					LineNumbers:   cfg.LineNumbers,
					StripComments: cfg.StripComments,
				})
				content, err = mg.Generate(files)
				if err != nil {
					return fmt.Errorf("failed to generate markdown: %w", err)
				}
			}

			// Write output
			if cfg.Output != "" {
				if err := os.MkdirAll(filepath.Dir(cfg.Output), 0755); err != nil {
					return fmt.Errorf("failed to create output directory: %w", err)
				}
				if err := os.WriteFile(cfg.Output, []byte(content), 0644); err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}
				fmt.Printf("Output written to: %s\n", cfg.Output)
			} else {
				fmt.Println(content)
			}

			// Handle token counting and pricing
			if cfg.ShowTokens || cfg.ShowPrice {
				counter, err := tokens.NewCounter(cfg.TokenEncoding)
				if err != nil {
					return fmt.Errorf("failed to create token counter: %w", err)
				}

				count, err := counter.Count(content)
				if err != nil {
					return fmt.Errorf("failed to count tokens: %w", err)
				}

				if cfg.ShowTokens {
					fmt.Printf("\nToken count: %d\n", count)
				}

				if cfg.ShowPrice {
					price, err := counter.EstimatePrice(count, cfg.OutputTokens, cfg.Model)
					if err != nil {
						return fmt.Errorf("failed to estimate price: %w", err)
					}
					fmt.Printf("\nEstimated price for %s: $%.4f\n", cfg.Model, price)
				}
			}

			return nil
		},
	}

	// Add flags - these will override config file values when specified
	cmd.Flags().StringVarP(&cfg.Output, "output", "o", "", "Output file path")
	cmd.Flags().StringSliceVarP(&cfg.FilterPatterns, "filter", "f", nil, "Filter patterns to include files")
	cmd.Flags().StringSliceVarP(&cfg.ExcludePatterns, "exclude", "e", nil, "Patterns to exclude files")
	cmd.Flags().BoolVarP(&cfg.CaseSensitive, "case-sensitive", "c", false, "Use case-sensitive pattern matching")
	cmd.Flags().BoolVar(&cfg.NoCodeblock, "no-codeblock", false, "Disable wrapping code in markdown code blocks")
	cmd.Flags().BoolVarP(&cfg.LineNumbers, "line-numbers", "l", false, "Add line numbers to code blocks")
	cmd.Flags().BoolVarP(&cfg.StripComments, "strip-comments", "s", false, "Strip comments from code")
	cmd.Flags().StringVarP(&cfg.TemplatePath, "template", "t", "", "Path to template file")
	cmd.Flags().BoolVar(&cfg.ShowTokens, "tokens", false, "Show token count")
	cmd.Flags().StringVar(&cfg.TokenEncoding, "encoding", "cl100k_base", "Token encoding to use")
	cmd.Flags().BoolVar(&cfg.ShowPrice, "price", false, "Show estimated price")
	cmd.Flags().StringVar(&cfg.Provider, "provider", "openai", "Provider for price estimation")
	cmd.Flags().StringVar(&cfg.Model, "model", "gpt-3.5-turbo", "Model for price estimation")
	cmd.Flags().IntVar(&cfg.OutputTokens, "output-tokens", 1000, "Expected number of output tokens")

	return cmd
}
