package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dwrtz/sink/internal/generator"
	"github.com/dwrtz/sink/internal/watcher"
	"github.com/spf13/cobra"
)

type watchFlags struct {
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
	debounceMs      int
}

func newWatchCmd() *cobra.Command {
	flags := &watchFlags{}

	cmd := &cobra.Command{
		Use:   "watch [path]",
		Short: "Watch a directory and regenerate documentation on changes",
		Long: `Watch a directory and its subdirectories for changes and automatically
regenerate documentation when files are modified. Applies the same filtering
rules as the generate command.

Examples:
  sink watch . -o output.md
  sink watch . --filter "*.go,*.md" --debounce 1000`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Convert path to absolute to ensure consistent watching
			absPath, err := filepath.Abs(args[0])
			if err != nil {
				return fmt.Errorf("failed to resolve absolute path: %w", err)
			}
			args[0] = absPath

			// Update config with CLI flags if they were explicitly set
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

			// Validate the path exists
			if _, err := os.Stat(args[0]); err != nil {
				return fmt.Errorf("invalid path %s: %w", args[0], err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := generator.RunGeneration(cfg, args[0])
			if err != nil {
				return fmt.Errorf("failed to generate file: %w", err)
			}

			watchService, err := watcher.NewService(watcher.Config{
				RootPath:        args[0],
				RepoConfig:      cfg,
				DebounceTimeout: time.Duration(flags.debounceMs) * time.Millisecond,
			})
			if err != nil {
				return fmt.Errorf("failed to create watch service: %w", err)
			}

			fmt.Printf("Watching %s for changes...\n", args[0])
			fmt.Println("Press Ctrl+C to stop")

			// Watch will block until interrupted
			if err := watchService.Watch(); err != nil {
				return fmt.Errorf("watch service error: %w", err)
			}

			return nil
		},
	}

	// Add flags
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
	cmd.Flags().IntVar(&flags.debounceMs, "debounce", 500, "Debounce timeout in milliseconds")

	return cmd
}
