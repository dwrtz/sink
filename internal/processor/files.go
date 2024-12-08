package processor

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/dwrtz/sink/internal/filter" // for gitignore and pattern matching
	"github.com/dwrtz/sink/internal/utils"  // for binary file detection
)

// FileInfo represents information about a processed file
type FileInfo struct {
	Path     string
	Content  string
	Language string
	Size     int64
	Created  time.Time
	Modified time.Time
}

// Config holds the file processor configuration
type Config struct {
	Paths           []string
	GitignorePath   string
	FilterPatterns  []string
	ExcludePatterns []string
	CaseSensitive   bool
}

// FileProcessor handles the processing of files
type FileProcessor struct {
	config            Config
	gitignorePatterns []string
}

// NewFileProcessor creates a new file processor with the given configuration
func NewFileProcessor(config Config) (*FileProcessor, error) {
	patterns, err := filter.LoadGitignorePatterns(config.GitignorePath)
	if err != nil {
		return nil, err
	}

	return &FileProcessor{
		config:            config,
		gitignorePatterns: patterns,
	}, nil
}

// Process processes files based on the configured paths and returns file information
func (fp *FileProcessor) Process() ([]FileInfo, error) {
	var files []FileInfo

	for _, path := range fp.config.Paths {
		err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			if !fp.shouldProcessFile(path) {
				return nil
			}

			info, err := fp.processFile(path)
			if err != nil {
				return err
			}

			files = append(files, info)
			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

// processFile processes a single file and returns its information
func (fp *FileProcessor) processFile(path string) (FileInfo, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return FileInfo{}, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		Path:     path,
		Content:  string(content),
		Language: detectLanguage(path),
		Size:     stat.Size(),
		Created:  stat.ModTime(), // Note: Creation time not available in all OS
		Modified: stat.ModTime(),
	}, nil
}

// shouldProcessFile determines if a file should be processed based on patterns
func (fp *FileProcessor) shouldProcessFile(path string) bool {
	// Check gitignore patterns
	if filter.IsIgnored(path, fp.gitignorePatterns) {
		return false
	}

	// Check if file matches filter patterns
	if len(fp.config.FilterPatterns) > 0 && !filter.MatchesAny(path, fp.config.FilterPatterns, fp.config.CaseSensitive) {
		return false
	}

	// Check if file matches exclude patterns
	if len(fp.config.ExcludePatterns) > 0 && filter.MatchesAny(path, fp.config.ExcludePatterns, fp.config.CaseSensitive) {
		return false
	}

	return !utils.IsBinaryFile(path)
}
