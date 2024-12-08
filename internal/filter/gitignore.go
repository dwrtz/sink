package filter

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// LoadGitignorePatterns loads gitignore patterns from a file
func LoadGitignorePatterns(path string) ([]string, error) {
	if path == "" {
		return nil, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}

	return patterns, scanner.Err()
}

// IsIgnored checks if a path matches any gitignore pattern
func IsIgnored(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchGitignorePattern(path, pattern) {
			return true
		}
	}
	return false
}

// matchGitignorePattern checks if a path matches a gitignore pattern
func matchGitignorePattern(path, pattern string) bool {
	// TODO: Implement proper gitignore pattern matching
	// A very rough approximation:
	match, _ := filepath.Match(pattern, filepath.Base(path))
	return match
}
