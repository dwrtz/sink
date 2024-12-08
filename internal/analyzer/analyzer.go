package analyzer

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Stats represents statistics about file extensions in the codebase
type Stats struct {
	Extensions     map[string]int            // Map of extensions to count
	DirectoryCount map[string]map[string]int // Map of directories to extension counts
	TotalFiles     int                       // Total number of files
	TotalSize      int64                     // Total size in bytes
}

// Result holds the analysis results in different formats
type Result struct {
	Stats    Stats
	TreeView string
	FlatView string
}

// Analyzer performs codebase analysis
type Analyzer struct {
	mu sync.Mutex
}

// New creates a new Analyzer instance
func New() *Analyzer {
	return &Analyzer{}
}

// Analyze processes files and generates statistics
func (a *Analyzer) Analyze(files []string) (*Stats, error) {
	stats := &Stats{
		Extensions:     make(map[string]int),
		DirectoryCount: make(map[string]map[string]int),
	}

	// Use a WaitGroup for concurrent processing
	var wg sync.WaitGroup
	// Process files concurrently
	for _, file := range files {
		wg.Add(1)
		go func(filepath string) {
			defer wg.Done()
			a.processFile(filepath, stats)
		}(file)
	}
	wg.Wait()

	return stats, nil
}

// processFile analyzes a single file and updates statistics
func (a *Analyzer) processFile(path string, stats *Stats) {
	ext := filepath.Ext(path)
	dir := filepath.Dir(path)

	// Thread-safe updates to stats
	a.mu.Lock()
	defer a.mu.Unlock()

	// Update extension count
	stats.Extensions[ext]++
	stats.TotalFiles++

	// Update directory stats
	if _, exists := stats.DirectoryCount[dir]; !exists {
		stats.DirectoryCount[dir] = make(map[string]int)
	}
	stats.DirectoryCount[dir][ext]++
}

// FormatFlat returns a flat view of extension statistics
func (a *Analyzer) FormatFlat(stats *Stats) string {
	var result []string

	// Convert map to sorted slice for consistent output
	var extensions []string
	for ext := range stats.Extensions {
		extensions = append(extensions, ext)
	}
	sort.Strings(extensions)

	// Build output
	for _, ext := range extensions {
		count := stats.Extensions[ext]
		if count == 1 {
			result = append(result, fmt.Sprintf("%s: 1 file", ext))
		} else {
			result = append(result, fmt.Sprintf("%s: %d files", ext, count))
		}
	}

	return strings.Join(result, "\n")
}

// GetExtensionList returns a comma-separated list of extensions
func (a *Analyzer) GetExtensionList(stats *Stats) string {
	var extensions []string
	for ext := range stats.Extensions {
		extensions = append(extensions, ext)
	}
	sort.Strings(extensions)
	return strings.Join(extensions, ",")
}
