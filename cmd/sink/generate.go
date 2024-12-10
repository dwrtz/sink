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
	var (
		output          string
		filterPatterns  []string
		excludePatterns []string
		caseSensitive   bool
		noCodeblock     bool
		lineNumbers     bool
		stripComments   bool
		templatePath    string
		showTokens      bool
		encoding        string
		showPrice       bool
		provider        string
		model           string
		outputTokens    int
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate markdown documentation from code files",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("at least one input path is required")
			}

			// Validate paths
			for _, path := range args {
				if _, err := os.Stat(path); err != nil {
					return fmt.Errorf("invalid path %s: %w", path, err)
				}
			}

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

			var content string
			if templatePath != "" {
				// Use template if specified
				templateContent, err := os.ReadFile(templatePath)
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
					NoCodeBlock:   noCodeblock,
					LineNumbers:   lineNumbers,
					StripComments: stripComments,
				})
				content, err = mg.Generate(files)
				if err != nil {
					return fmt.Errorf("failed to generate markdown: %w", err)
				}
			}

			// Write output
			if output != "" {
				if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
					return fmt.Errorf("failed to create output directory: %w", err)
				}
				if err := os.WriteFile(output, []byte(content), 0644); err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}
				fmt.Printf("Output written to: %s\n", output)
			} else {
				fmt.Println(content)
			}

			// Handle token counting and pricing
			if showTokens || showPrice {
				counter, err := tokens.NewCounter(encoding)
				if err != nil {
					return fmt.Errorf("failed to create token counter: %w", err)
				}

				count, err := counter.Count(content)
				if err != nil {
					return fmt.Errorf("failed to count tokens: %w", err)
				}

				if showTokens {
					fmt.Printf("\nToken count: %d\n", count)
				}

				if showPrice {
					if provider != "openai" {
						fmt.Printf("Warning: Provider '%s' not currently implemented. Using openai pricing.\n", provider)
					}
					price, err := counter.EstimatePrice(count, outputTokens, model)
					if err != nil {
						return fmt.Errorf("failed to estimate price: %w", err)
					}
					fmt.Printf("\nEstimated price for %s: $%.4f\n", model, price)
				}
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")
	cmd.Flags().StringSliceVarP(&filterPatterns, "filter", "f", nil, "Filter patterns to include files")
	cmd.Flags().StringSliceVarP(&excludePatterns, "exclude", "e", nil, "Patterns to exclude files")
	cmd.Flags().BoolVarP(&caseSensitive, "case-sensitive", "c", false, "Use case-sensitive pattern matching")
	cmd.Flags().BoolVar(&noCodeblock, "no-codeblock", false, "Disable wrapping code in markdown code blocks")
	cmd.Flags().BoolVarP(&lineNumbers, "line-numbers", "l", false, "Add line numbers to code blocks")
	cmd.Flags().BoolVarP(&stripComments, "strip-comments", "s", false, "Strip comments from code")
	cmd.Flags().StringVarP(&templatePath, "template", "t", "", "Path to template file")
	cmd.Flags().BoolVar(&showTokens, "tokens", false, "Show token count")
	cmd.Flags().StringVar(&encoding, "encoding", "cl100k_base", "Token encoding to use")
	cmd.Flags().BoolVar(&showPrice, "price", false, "Show estimated price")
	cmd.Flags().StringVar(&provider, "provider", "openai", "Provider for price estimation")
	cmd.Flags().StringVar(&model, "model", "gpt-3.5-turbo", "Model for price estimation")
	cmd.Flags().IntVar(&outputTokens, "output-tokens", 1000, "Expected number of output tokens")

	return cmd
}
