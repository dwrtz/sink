package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration structure
type Config struct {
	// Core settings
	Output          string   `yaml:"output"`
	FilterPatterns  []string `yaml:"filter-patterns"`
	ExcludePatterns []string `yaml:"exclude-patterns"`
	CaseSensitive   bool     `yaml:"case-sensitive"`

	// Processing options
	NoCodeblock   bool `yaml:"no-codeblock"`
	LineNumbers   bool `yaml:"line-numbers"`
	StripComments bool `yaml:"strip-comments"`

	// Token settings
	ShowTokens    bool   `yaml:"show-tokens"`
	TokenEncoding string `yaml:"token-encoding"`

	// Price estimation
	ShowPrice    bool   `yaml:"show-price"`
	Provider     string `yaml:"provider"`
	Model        string `yaml:"model"`
	OutputTokens int    `yaml:"output-tokens"`

	// Syntax highlighting mappings
	SyntaxMap map[string]string `yaml:"syntax-map"`

	// Template settings
	TemplatePath string `yaml:"template-path"`
}

// DefaultConfig returns a new Config with default values
func DefaultConfig() *Config {
	return &Config{
		TokenEncoding: "cl100k_base",
		Provider:      "openai",
		Model:         "gpt-3.5-turbo",
		OutputTokens:  1000,
		SyntaxMap:     make(map[string]string),
	}
}

// LoadConfig loads configuration from multiple sources with proper precedence
func LoadConfig(cmdConfigPath string) (*Config, error) {
	config := DefaultConfig()

	// 1. Load system config
	systemConfig, err := loadSystemConfig()
	if err == nil {
		config.merge(systemConfig)
	}

	// 2. Load user config
	userConfig, err := loadUserConfig()
	if err == nil {
		config.merge(userConfig)
	}

	// 3. Load local config
	localConfig, err := loadLocalConfig()
	if err == nil {
		config.merge(localConfig)
	}

	// 4. Load explicitly specified config file
	if cmdConfigPath != "" {
		explicitConfig, err := loadConfigFile(cmdConfigPath)
		if err != nil {
			return nil, fmt.Errorf("error loading specified config file: %w", err)
		}
		config.merge(explicitConfig)
	}

	return config, nil
}

// getSystemConfigPath returns the path to the system-wide config
func getSystemConfigPath() string {
	if os.Getenv("SINK_SYSTEM_CONFIG") != "" {
		return os.Getenv("SINK_SYSTEM_CONFIG")
	}
	return "/etc/sink/config.yaml"
}

// getUserConfigPath returns the path to the user's config
func getUserConfigPath() string {
	if os.Getenv("SINK_USER_CONFIG") != "" {
		return os.Getenv("SINK_USER_CONFIG")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "sink", "config.yaml")
	}

	// Fall back to ~/.config/sink/config.yaml
	return filepath.Join(homeDir, ".config", "sink", "config.yaml")
}

// getLocalConfigPath returns the path to the local config
func getLocalConfigPath() string {
	return "sink-config.yaml"
}

// loadSystemConfig loads the system-wide configuration
func loadSystemConfig() (*Config, error) {
	return loadConfigFile(getSystemConfigPath())
}

// loadUserConfig loads the user's configuration
func loadUserConfig() (*Config, error) {
	return loadConfigFile(getUserConfigPath())
}

// loadLocalConfig loads the local configuration
func loadLocalConfig() (*Config, error) {
	return loadConfigFile(getLocalConfigPath())
}

// loadConfigFile loads and parses a configuration file
func loadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("error parsing config file %s: %w", path, err)
	}

	return config, nil
}

// merge merges another config into this one
func (c *Config) merge(other *Config) {
	if other == nil {
		return
	}

	// Only override non-zero values
	if other.Output != "" {
		c.Output = other.Output
	}
	if len(other.FilterPatterns) > 0 {
		c.FilterPatterns = other.FilterPatterns
	}
	if len(other.ExcludePatterns) > 0 {
		c.ExcludePatterns = other.ExcludePatterns
	}

	// Boolean flags need special handling - they should only be overridden if explicitly set
	if other.CaseSensitive {
		c.CaseSensitive = true
	}
	if other.NoCodeblock {
		c.NoCodeblock = true
	}
	if other.LineNumbers {
		c.LineNumbers = true
	}
	if other.StripComments {
		c.StripComments = true
	}
	if other.ShowTokens {
		c.ShowTokens = true
	}
	if other.ShowPrice {
		c.ShowPrice = true
	}

	if other.TokenEncoding != "" {
		c.TokenEncoding = other.TokenEncoding
	}
	if other.Provider != "" {
		c.Provider = other.Provider
	}
	if other.Model != "" {
		c.Model = other.Model
	}
	if other.OutputTokens != 0 {
		c.OutputTokens = other.OutputTokens
	}
	if other.TemplatePath != "" {
		c.TemplatePath = other.TemplatePath
	}

	// Merge syntax map
	for k, v := range other.SyntaxMap {
		c.SyntaxMap[k] = v
	}
}

// MergeFlagSet merges cobra flag values into the config
func (c *Config) MergeFlagSet(flags *pflag.FlagSet) error {
	// Only override if flag was explicitly set
	flags.Visit(func(f *pflag.Flag) {
		switch f.Name {
		case "output":
			c.Output, _ = flags.GetString("output")
		case "filter":
			c.FilterPatterns, _ = flags.GetStringSlice("filter")
		case "exclude":
			c.ExcludePatterns, _ = flags.GetStringSlice("exclude")
		case "case-sensitive":
			c.CaseSensitive, _ = flags.GetBool("case-sensitive")
		case "no-codeblock":
			c.NoCodeblock, _ = flags.GetBool("no-codeblock")
		case "line-numbers":
			c.LineNumbers, _ = flags.GetBool("line-numbers")
		case "strip-comments":
			c.StripComments, _ = flags.GetBool("strip-comments")
		case "tokens":
			c.ShowTokens, _ = flags.GetBool("tokens")
		case "encoding":
			c.TokenEncoding, _ = flags.GetString("encoding")
		case "price":
			c.ShowPrice, _ = flags.GetBool("price")
		case "provider":
			c.Provider, _ = flags.GetString("provider")
		case "model":
			c.Model, _ = flags.GetString("model")
		case "output-tokens":
			c.OutputTokens, _ = flags.GetInt("output-tokens")
		case "template":
			c.TemplatePath, _ = flags.GetString("template")
		}
	})

	return nil
}
