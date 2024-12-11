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

// rootCmd represents the base command
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
}

func initConfig() error {
	var err error
	cfg, err = config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}
	return nil
}

func initialize() {
	// Add persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")

	// Disable default completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Initialize config before adding subcommands
	cobra.OnInitialize(func() {
		if err := initConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
			os.Exit(1)
		}
	})

	// Add subcommands after config initialization
	rootCmd.AddCommand(newGenerateCmd())
	rootCmd.AddCommand(newAnalyzeCmd())
}

func main() {
	initialize()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
