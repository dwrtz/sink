package tokens

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"
)

// Counter handles token counting operations
type Counter struct {
	encoding string
}

// NewCounter creates a new token counter with the specified encoding
func NewCounter(encoding string) (*Counter, error) {
	// Validate encoding
	if !isValidEncoding(encoding) {
		return nil, fmt.Errorf("invalid encoding: %s", encoding)
	}

	return &Counter{
		encoding: encoding,
	}, nil
}

// Count returns the number of tokens in the given text
func (c *Counter) Count(text string) (int, error) {
	tkm, err := tiktoken.GetEncoding(c.encoding)
	if err != nil {
		return 0, fmt.Errorf("failed to get encoding: %w", err)
	}

	tokens := tkm.Encode(text, nil, nil)
	return len(tokens), nil
}

// CountFiles counts tokens in multiple files and returns the total
func (c *Counter) CountFiles(files []string) (int, error) {
	total := 0
	for _, file := range files {
		count, err := c.Count(file)
		if err != nil {
			return 0, fmt.Errorf("failed to count tokens in file %s: %w", file, err)
		}
		total += count
	}
	return total, nil
}

// isValidEncoding checks if the encoding is supported
func isValidEncoding(encoding string) bool {
	validEncodings := map[string]bool{
		"cl100k_base": true,
		"p50k_base":   true,
		"r50k_base":   true,
	}
	return validEncodings[encoding]
}

// EstimatePrice calculates the estimated price for the given number of tokens
func (c *Counter) EstimatePrice(inputTokens, outputTokens int, model string) (float64, error) {
	prices := map[string]struct {
		input  float64
		output float64
	}{
		"gpt-3.5-turbo": {input: 0.0015, output: 0.002},
		"gpt-4":         {input: 0.03, output: 0.06},
		"gpt-4-32k":     {input: 0.06, output: 0.12},
	}

	modelPrices, ok := prices[model]
	if !ok {
		return 0, fmt.Errorf("unsupported model: %s", model)
	}

	inputCost := float64(inputTokens) * modelPrices.input / 1000
	outputCost := float64(outputTokens) * modelPrices.output / 1000

	return inputCost + outputCost, nil
}
