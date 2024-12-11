package config

import (
	"fmt"
	"os"
)

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate token encoding
	if !isValidEncoding(c.TokenEncoding) {
		return fmt.Errorf("invalid token encoding: %s", c.TokenEncoding)
	}

	// Validate provider
	if c.ShowPrice {
		if !isValidProvider(c.Provider) {
			return fmt.Errorf("invalid provider: %s", c.Provider)
		}
		if !isValidModel(c.Provider, c.Model) {
			return fmt.Errorf("invalid model %s for provider %s", c.Model, c.Provider)
		}
	}

	// Validate output tokens
	if c.OutputTokens < 0 {
		return fmt.Errorf("output tokens must be non-negative")
	}

	// Validate template path if specified
	if c.TemplatePath != "" {
		if _, err := os.Stat(c.TemplatePath); err != nil {
			return fmt.Errorf("invalid template path: %w", err)
		}
	}

	return nil
}

func isValidEncoding(encoding string) bool {
	validEncodings := map[string]bool{
		"cl100k_base": true,
		"p50k_base":   true,
		"p50k_edit":   true,
		"r50k_base":   true,
	}
	return validEncodings[encoding]
}

func isValidProvider(provider string) bool {
	validProviders := map[string]bool{
		"openai":    true,
		"anthropic": true,
		"google":    true,
		"mistral":   true,
		"cohere":    true,
	}
	return validProviders[provider]
}

func isValidModel(provider, model string) bool {
	validModels := map[string]map[string]bool{
		"openai": {
			"gpt-3.5-turbo": true,
			"gpt-4":         true,
			"gpt-4-32k":     true,
		},
		"anthropic": {
			"claude-2":       true,
			"claude-instant": true,
		},
		// Add more providers and their models as needed
	}

	if models, ok := validModels[provider]; ok {
		return models[model]
	}
	return false
}
