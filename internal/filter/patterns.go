package filter

import (
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// MatchesAny checks if a path matches any of the given glob patterns
func MatchesAny(path string, patterns []string, caseSensitive bool) bool {
	if len(patterns) == 0 {
		return true // No patterns means match everything
	}

	// Normalize path separators and handle case sensitivity
	path = filepath.ToSlash(path)
	if !caseSensitive {
		path = strings.ToLower(path)
	}

	// Get basename for simple patterns
	basename := filepath.Base(path)

	for _, pattern := range patterns {
		if !caseSensitive {
			pattern = strings.ToLower(pattern)
		}
		pattern = filepath.ToSlash(pattern)

		// If pattern has no slashes, match against basename
		matchPath := path
		if !strings.Contains(pattern, "/") {
			matchPath = basename
		}

		matched, err := doublestar.Match(pattern, matchPath)
		if err == nil && matched {
			return true
		}
	}

	return false
}
