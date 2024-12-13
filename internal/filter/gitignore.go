package filter

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// GitignoreFilter handles gitignore pattern matching
type GitignoreFilter struct {
	matcher gitignore.Matcher
	fs      billy.Filesystem
}

type GitignoreConfig struct {
	RepoRoot           string `yaml:"repo-root"`
	LoadGlobalPatterns bool   `yaml:"load-global-patterns"`
	LoadSystemPatterns bool   `yaml:"load-system-patterns"`
}

func PathParts(p string) []string {
	p = filepath.Clean(p)
	if p == "." {
		// Represents the current directory, no meaningful components
		return []string{}
	}

	var parts []string
	for p != "" && p != string(filepath.Separator) {
		dir, file := filepath.Split(p)
		if file == "" {
			// This means p ended with a separator or we hit the root.
			// Trim the trailing separator from dir and continue.
			p = strings.TrimSuffix(dir, string(filepath.Separator))
			continue
		}

		parts = append(parts, file)
		// Move up one directory level by trimming the trailing separator from dir
		p = strings.TrimSuffix(dir, string(filepath.Separator))
	}

	// Reverse the collected parts to restore the original order
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return parts
}

// NewGitignoreFilter creates a new GitignoreFilter
func NewFilter(config GitignoreConfig) (*GitignoreFilter, error) {
	fs := osfs.New(config.RepoRoot)
	fmt.Println("Creating fs at", fs.Root())
	patterns, err := gitignore.ReadPatterns(fs, []string{})
	if err != nil {
		return nil, err
	}

	if config.LoadGlobalPatterns {
		globalPatterns, err := gitignore.LoadGlobalPatterns(fs)
		if err != nil {
			return nil, err
		}
		if globalPatterns != nil {
			patterns = append(patterns, globalPatterns...)
		}
	}

	if config.LoadSystemPatterns {
		systemPatterns, err := gitignore.LoadSystemPatterns(fs)
		if err != nil {
			return nil, err
		}
		if systemPatterns != nil {
			patterns = append(patterns, systemPatterns...)
		}
	}

	matcher := gitignore.NewMatcher(patterns)
	return &GitignoreFilter{matcher: matcher, fs: fs}, nil
}

func (g *GitignoreFilter) IsIgnored(path string) (bool, error) {
	info, err := g.fs.Stat(path)
	if err != nil {
		return false, err
	}

	pathParts := PathParts(path)
	ignored := g.matcher.Match(pathParts, info.IsDir())
	return ignored, nil
}
