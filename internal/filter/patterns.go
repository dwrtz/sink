package filter

import (
	"path/filepath"
	"strings"
)

// MatchesAny checks if a path matches any of the given patterns
func MatchesAny(path string, patterns []string, caseSensitive bool) bool {
	if !caseSensitive {
		path = strings.ToLower(path)
	}

	for _, pattern := range patterns {
		if !caseSensitive {
			pattern = strings.ToLower(pattern)
		}
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
	}
	return false
}
