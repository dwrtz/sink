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

type generateFlags struct {
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
}

func newGenerateCmd() *cobra.Command {
	flags := &generateFlags{}

	cmd := &cobra.Command{
		Use:   "generate [path]",
		Short: "Generate markdown documentation from code files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Update config with any explicitly set flags
			if cmd.Flags().Changed("output") {
				cfg.Output = flags.output
			}
			if cmd.Flags().Changed("filter") {
				cfg.FilterPatterns = flags.filterPatterns
			}
			if cmd.Flags().Changed("exclude") {
				cfg.ExcludePatterns = flags.excludePatterns
			}
			if cmd.Flags().Changed("case-sensitive") {
				cfg.CaseSensitive = flags.caseSensitive
			}
			if cmd.Flags().Changed("no-codeblock") {
				cfg.NoCodeblock = flags.noCodeblock
			}
			if cmd.Flags().Changed("line-numbers") {
				cfg.LineNumbers = flags.lineNumbers
			}
			if cmd.Flags().Changed("strip-comments") {
				cfg.StripComments = flags.stripComments
			}
			if cmd.Flags().Changed("template") {
				cfg.TemplatePath = flags.templatePath
			}
			if cmd.Flags().Changed("tokens") {
				cfg.ShowTokens = flags.showTokens
			}
			if cmd.Flags().Changed("encoding") {
				cfg.TokenEncoding = flags.encoding
			}
			if cmd.Flags().Changed("price") {
				cfg.ShowPrice = flags.showPrice
			}
			if cmd.Flags().Changed("provider") {
				cfg.Provider = flags.provider
			}
			if cmd.Flags().Changed("model") {
				cfg.Model = flags.model
			}
			if cmd.Flags().Changed("output-tokens") {
				cfg.OutputTokens = flags.outputTokens
			}

			path := args[0]

			// Validate path
			if _, err := os.Stat(path); err != nil {
				return fmt.Errorf("invalid repository path %s: %w", path, err)
			}

			// Create file processor
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

	// Add flags bound to the local flags struct
	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file path")
	cmd.Flags().StringSliceVarP(&flags.filterPatterns, "filter", "f", nil, "Filter patterns to include files")
	cmd.Flags().StringSliceVarP(&flags.excludePatterns, "exclude", "e", nil, "Patterns to exclude files")
	cmd.Flags().BoolVarP(&flags.caseSensitive, "case-sensitive", "c", false, "Use case-sensitive pattern matching")
	cmd.Flags().BoolVar(&flags.noCodeblock, "no-codeblock", false, "Disable wrapping code in markdown code blocks")
	cmd.Flags().BoolVarP(&flags.lineNumbers, "line-numbers", "l", false, "Add line numbers to code blocks")
	cmd.Flags().BoolVarP(&flags.stripComments, "strip-comments", "s", false, "Strip comments from code")
	cmd.Flags().StringVarP(&flags.templatePath, "template", "t", "", "Path to template file")
	cmd.Flags().BoolVar(&flags.showTokens, "tokens", false, "Show token count")
	cmd.Flags().StringVar(&flags.encoding, "encoding", "cl100k_base", "Token encoding to use")
	cmd.Flags().BoolVar(&flags.showPrice, "price", false, "Show estimated price")
	cmd.Flags().StringVar(&flags.provider, "provider", "openai", "Provider for price estimation")
	cmd.Flags().StringVar(&flags.model, "model", "gpt-3.5-turbo", "Model for price estimation")
	cmd.Flags().IntVar(&flags.outputTokens, "output-tokens", 1000, "Expected number of output tokens")

	return cmd
}
