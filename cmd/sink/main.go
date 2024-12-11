package main

import (
	"fmt"
	"os"

	"github.com/dwrtz/sink/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg     *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "sink [path]",
	Short: "Sink - A tool for generating AI prompts from codebases",
	Long: `Sink analyzes codebases and generates well-structured prompts for AI models.
It supports multiple languages, comment stripping, and customizable output formats.

Example usage:
  sink generate . -o output.md
  sink analyze . --format flat
  sink generate . --tokens --price --model gpt-4`,
	Version: "1.0.0",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error

		// Load configuration
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		// Merge command line flags
		if err := cfg.MergeFlagSet(cmd.Flags()); err != nil {
			return fmt.Errorf("error merging command line flags: %w", err)
		}

		// Validate the final configuration
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(newGenerateCmd())
	rootCmd.AddCommand(newAnalyzeCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
