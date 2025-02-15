package processor

import (
	"errors"
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

// sentinel error so we can detect when to skip a “file”
var errSkipFile = errors.New("skip this file or directory")

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

		// If it's a directory, skip .git or any directory that matches excludes
		if d.IsDir() {
			// Skip .git directory entirely
			if d.Name() == ".git" {
				return filepath.SkipDir
			}

			relPath, err := filepath.Rel(fp.fs.Root(), path)
			if err != nil {
				return err
			}

			// Check if directory is ignored by gitignore
			ignored, ignErr := fp.ignorer.IsIgnored(relPath)
			if ignErr != nil {
				fmt.Printf("Error checking if directory is ignored: %v\n", ignErr)
				return ignErr
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

		// If we got here, we have a non-dir (d.IsDir() == false), or a symlink, etc.
		if !fp.shouldProcessFile(path) {
			// Don’t abort entire walk, just skip
			return nil
		}

		fileInfo, fileErr := fp.processFile(path)
		if fileErr != nil {
			// We intentionally skip files with our sentinel error
			if errors.Is(fileErr, errSkipFile) {
				return nil
			}
			// For other errors, return up the chain
			fmt.Printf("Error processing file %s: %v\n", path, fileErr)
			return fileErr
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

	// **Double-check**: if it's a directory (or symlink to a directory), skip
	if info.IsDir() {
		// Return our sentinel, so the caller can ignore it
		return FileInfo{}, errSkipFile
	}

	// Try opening as a file
	file, err := fp.fs.Open(relPath)
	if err != nil {
		// If the OS says “is a directory”, treat as skip
		if isDirOpenError(err) {
			return FileInfo{}, errSkipFile
		}
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

// Helper to detect “is a directory” errors from the OS
func isDirOpenError(err error) bool {
	return strings.Contains(err.Error(), "is a directory")
}

// shouldProcessFile determines whether a path should be processed based on
// binary check and filter/exclude patterns.
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
			return !filter.MatchesAny(relPath, fp.config.ExcludePatterns, fp.config.CaseSensitive)
		}
		return true
	}

	// If we have filter patterns, file must match at least one
	if !filter.MatchesAny(relPath, fp.config.FilterPatterns, fp.config.CaseSensitive) {
		return false
	}

	// Finally check exclude patterns
	if len(fp.config.ExcludePatterns) > 0 {
		return !filter.MatchesAny(relPath, fp.config.ExcludePatterns, fp.config.CaseSensitive)
	}

	return true
}

func (fp *FileProcessor) detectLanguage(path string) string {
	ext := filepath.Ext(path)

	// Check syntax map first
	if lang, ok := fp.config.SyntaxMap[ext]; ok {
		return lang
	}

	// Fall back to a small set of known file types
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
	default:
		return "unknown"
	}
}
