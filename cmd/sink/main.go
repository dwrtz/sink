package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
}

func init() {
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
