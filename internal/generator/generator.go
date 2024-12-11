package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dwrtz/sink/internal/config"
	"github.com/dwrtz/sink/internal/processor"
	"github.com/dwrtz/sink/internal/processor/markdown"
	"github.com/dwrtz/sink/internal/processor/template"
	"github.com/dwrtz/sink/internal/tokens"
)

func RunGeneration(cfg *config.Config, path string) error {
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

	files, err := fp.Process()
	if err != nil {
		return fmt.Errorf("failed to process files: %w", err)
	}

	content, err := generateContent(files, cfg)
	if err != nil {
		return err
	}

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

	// Handle token counting and pricing if enabled
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
}

func generateContent(files []processor.FileInfo, cfg *config.Config) (string, error) {
	if cfg.TemplatePath != "" {
		templateContent, err := os.ReadFile(cfg.TemplatePath)
		if err != nil {
			return "", fmt.Errorf("failed to read template: %w", err)
		}
		te := template.NewEngine(string(templateContent))
		return te.Execute(files)
	}

	mg := markdown.NewGenerator(markdown.Config{
		NoCodeBlock:   cfg.NoCodeblock,
		LineNumbers:   cfg.LineNumbers,
		StripComments: cfg.StripComments,
	})
	return mg.Generate(files)
}
