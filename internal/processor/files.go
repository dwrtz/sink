package processor

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/dwrtz/sink/internal/filter"
	"github.com/dwrtz/sink/internal/utils"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
)

type FileInfo struct {
	Path     string
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

			// Check if directory matches gitignore patterns
			ignored, err := fp.ignorer.IsIgnored(path)
			if err != nil {
				fmt.Printf("Error checking if directory is ignored: %v\n", err)
				return err
			}
			if ignored {
				return filepath.SkipDir
			}

			// Check directory against exclude patterns
			if len(fp.config.ExcludePatterns) > 0 &&
				filter.MatchesAny(path, fp.config.ExcludePatterns, fp.config.CaseSensitive) {
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
	info, err := fp.fs.Stat(path)
	if err != nil {
		return FileInfo{}, err
	}

	file, err := fp.fs.Open(path)
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
		Content:  string(content),
		Language: detectLanguage(path),
		Size:     info.Size(),
		Created:  info.ModTime(),
		Modified: info.ModTime(),
	}, nil
}

func (fp *FileProcessor) shouldProcessFile(path string) bool {
	// Check if file is ignored by gitignore patterns
	ignored, err := fp.ignorer.IsIgnored(path)
	if err != nil || ignored {
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
