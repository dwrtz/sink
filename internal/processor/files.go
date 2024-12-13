package processor

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/dwrtz/sink/internal/filter"
	"github.com/dwrtz/sink/internal/utils"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
)

type FileInfo struct {
	Path     string
	Ext      string
	Content  string
	Language string
	Size     int64
	Created  time.Time
	Modified time.Time
}

type Config struct {
	RepoRoot        string
	FilterPatterns  []string
	ExcludePatterns []string
	CaseSensitive   bool
	SyntaxMap       map[string]string
}

type FileProcessor struct {
	config  Config
	fs      billy.Filesystem
	ignorer *filter.GitignoreFilter
}

func NewFileProcessor(config Config) (*FileProcessor, error) {
	// Create filesystem relative to repo root
	fs := osfs.New(config.RepoRoot)

	// Create GitignoreFilter using repo root
	ignorer, err := filter.NewFilter(filter.GitignoreConfig{
		RepoRoot:           config.RepoRoot,
		LoadGlobalPatterns: true,
		LoadSystemPatterns: true,
	})
	if err != nil {
		return nil, err
	}

	return &FileProcessor{
		config:  config,
		fs:      fs,
		ignorer: ignorer,
	}, nil
}

func (fp *FileProcessor) Process() ([]FileInfo, error) {
	var files []FileInfo

	// Walk the entire repository from root
	err := filepath.WalkDir(fp.config.RepoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Skip .git directory entirely
			if d.Name() == ".git" {
				return filepath.SkipDir
			}

			relPath, err := filepath.Rel(fp.fs.Root(), path)
			if err != nil {
				return err
			}

			// Check if directory matches gitignore patterns
			ignored, err := fp.ignorer.IsIgnored(relPath)
			if err != nil {
				fmt.Printf("Error checking if directory is ignored: %v\n", err)
				return err
			}
			if ignored {
				return filepath.SkipDir
			}

			// Check directory against exclude patterns
			if len(fp.config.ExcludePatterns) > 0 &&
				filter.MatchesAny(relPath, fp.config.ExcludePatterns, fp.config.CaseSensitive) {
				return filepath.SkipDir
			}

			return nil
		}

		if !fp.shouldProcessFile(path) {
			return nil
		}

		fileInfo, err := fp.processFile(path)
		if err != nil {
			fmt.Printf("Error processing file %s: %v\n", path, err)
			return err
		}

		files = append(files, fileInfo)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func (fp *FileProcessor) processFile(path string) (FileInfo, error) {
	relPath, err := filepath.Rel(fp.fs.Root(), path)
	if err != nil {
		return FileInfo{}, err
	}

	info, err := fp.fs.Stat(relPath)
	if err != nil {
		return FileInfo{}, err
	}

	file, err := fp.fs.Open(relPath)
	if err != nil {
		return FileInfo{}, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		Path:     path,
		Ext:      filepath.Ext(path),
		Content:  string(content),
		Language: fp.detectLanguage(path),
		Size:     info.Size(),
		Created:  info.ModTime(),
		Modified: info.ModTime(),
	}, nil
}

// shouldProcessFile determines whether a file should be processed based on
// various filtering criteria.
func (fp *FileProcessor) shouldProcessFile(path string) bool {
	// Check if file is binary
	if utils.IsBinaryFile(path) {
		return false
	}

	relPath, err := filepath.Rel(fp.fs.Root(), path)
	if err != nil {
		return false
	}

	// Check if file is ignored by gitignore patterns
	ignored, err := fp.ignorer.IsIgnored(relPath)
	if err != nil || ignored {
		return false
	}

	// If no filter patterns specified, only exclude patterns matter
	if len(fp.config.FilterPatterns) == 0 {
		// Check exclude patterns if any
		if len(fp.config.ExcludePatterns) > 0 {
			return !matchesAnyPattern(relPath, fp.config.ExcludePatterns, fp.config.CaseSensitive)
		}
		return true
	}

	// If we have filter patterns, file must match at least one
	if !matchesAnyPattern(relPath, fp.config.FilterPatterns, fp.config.CaseSensitive) {
		return false
	}

	// Finally check exclude patterns
	if len(fp.config.ExcludePatterns) > 0 {
		return !matchesAnyPattern(relPath, fp.config.ExcludePatterns, fp.config.CaseSensitive)
	}

	return true
}

// matchesAnyPattern checks if a path matches any of the given glob patterns
func matchesAnyPattern(path string, patterns []string, caseSensitive bool) bool {
	if !caseSensitive {
		path = strings.ToLower(path)
	}

	for _, pattern := range patterns {
		if !caseSensitive {
			pattern = strings.ToLower(pattern)
		}

		// Handle both relative and absolute paths
		_, filename := filepath.Split(path)

		// Try matching against full path
		matched, err := filepath.Match(pattern, path)
		if err == nil && matched {
			return true
		}

		// Try matching against just the filename
		matched, err = filepath.Match(pattern, filename)
		if err == nil && matched {
			return true
		}
	}

	return false
}

func (fp *FileProcessor) detectLanguage(path string) string {
	ext := filepath.Ext(path)

	// Check syntax map first
	if lang, ok := fp.config.SyntaxMap[ext]; ok {
		return lang
	}

	// Fall back to default language detection
	switch ext {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".java":
		return "java"
	case ".cpp", ".hpp", ".cc", ".hh":
		return "cpp"
	case ".c", ".h":
		return "c"
	// Add more language mappings as needed
	default:
		return "unknown"
	}
}
