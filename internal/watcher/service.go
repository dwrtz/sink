package watcher

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dwrtz/sink/internal/config"
	"github.com/dwrtz/sink/internal/filter"
	"github.com/dwrtz/sink/internal/generator"
	"github.com/dwrtz/sink/internal/utils"
	"github.com/fsnotify/fsnotify"
)

type watchedPath struct {
	path string
	dir  bool
}

type Config struct {
	RootPath        string
	RepoConfig      *config.Config
	DebounceTimeout time.Duration
}

type Service struct {
	config     Config
	watcher    *fsnotify.Watcher
	gitignorer *filter.GitignoreFilter
	debouncer  *time.Timer
	mu         sync.Mutex
	watched    map[string]*watchedPath
	configPath string
	reloading  bool
	// Add a logger for better visibility
	logger *log.Logger
}

func NewService(config Config) (*Service, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	gitignorer, err := filter.NewFilter(filter.GitignoreConfig{
		RepoRoot:           config.RootPath,
		LoadGlobalPatterns: true,
		LoadSystemPatterns: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create gitignore filter: %w", err)
	}

	// Check for config file in root directory only
	configPath := ""
	defaultConfigPath := filepath.Join(config.RootPath, "sink-config.yaml")
	if _, err := os.Stat(defaultConfigPath); err == nil {
		configPath = defaultConfigPath
	}

	// Create a logger that writes to stderr with timestamps
	logger := log.New(os.Stderr, "[watcher] ", log.LstdFlags)

	return &Service{
		config:     config,
		watcher:    watcher,
		gitignorer: gitignorer,
		debouncer:  time.NewTimer(0),
		watched:    make(map[string]*watchedPath),
		configPath: configPath,
		logger:     logger,
	}, nil
}

func (s *Service) Watch() error {
	// Create a context that's cancelled on interrupt
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Ensure cleanup
	defer s.watcher.Close()

	// Initial setup
	if err := s.reconfigureWatcher(); err != nil {
		return fmt.Errorf("failed to configure initial watches: %w", err)
	}

	// Watch config file if it exists
	if s.configPath != "" {
		if err := s.watcher.Add(s.configPath); err != nil {
			return fmt.Errorf("failed to add watch for config file: %w", err)
		}
		s.watched[s.configPath] = &watchedPath{path: s.configPath, dir: false}
		s.logger.Printf("Added watch for config file: %s", s.configPath)
	}

	// Log initial watch setup
	s.logger.Printf("Starting file watcher for root path: %s", s.config.RootPath)
	for path := range s.watched {
		s.logger.Printf("Watching: %s", path)
	}

	// Start a ticker to periodically log that the watcher is still alive
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Process events
	return s.processEvents(ctx, ticker)
}

func (s *Service) processEvents(ctx context.Context, ticker *time.Ticker) error {
	for {
		select {
		case <-ctx.Done():
			s.logger.Println("Watcher shutting down...")
			return ctx.Err()

		case <-ticker.C:
			s.logger.Println("Watcher is running...")

		case event, ok := <-s.watcher.Events:
			if !ok {
				return fmt.Errorf("watcher event channel closed")
			}
			s.logger.Printf("Received event: %s %s", event.Op.String(), event.Name)
			if err := s.handleEvent(event); err != nil {
				s.logger.Printf("Error handling event: %v", err)
			}

		case err, ok := <-s.watcher.Errors:
			if !ok {
				return fmt.Errorf("watcher error channel closed")
			}
			if e := s.handleWatchError(err); e != nil {
				return fmt.Errorf("critical watcher error: %w", e)
			}
		}
	}
}

// shouldProcessFile determines if a file should trigger a regeneration
func (s *Service) shouldProcessFile(path string) bool {
	// Skip binary files
	if utils.IsBinaryFile(path) {
		s.logger.Printf("Skipping binary file: %s", path)
		return false
	}

	// Convert to relative path for pattern matching
	relPath, err := filepath.Rel(s.config.RootPath, path)
	if err != nil {
		s.logger.Printf("Error getting relative path for %s: %v", path, err)
		return false
	}

	// Check gitignore patterns
	ignored, err := s.gitignorer.IsIgnored(relPath)
	if err != nil {
		s.logger.Printf("Error checking if %s is ignored: %v", relPath, err)
		return false
	}
	if ignored {
		s.logger.Printf("File %s is ignored by gitignore patterns", relPath)
		return false
	}

	// Check exclude patterns
	if len(s.config.RepoConfig.ExcludePatterns) > 0 {
		if filter.MatchesAny(relPath, s.config.RepoConfig.ExcludePatterns, s.config.RepoConfig.CaseSensitive) {
			s.logger.Printf("File %s matches exclude pattern", relPath)
			return false
		}
	}

	// Check filter patterns if specified
	if len(s.config.RepoConfig.FilterPatterns) > 0 {
		if !filter.MatchesAny(relPath, s.config.RepoConfig.FilterPatterns, s.config.RepoConfig.CaseSensitive) {
			s.logger.Printf("File %s does not match filter patterns", relPath)
			return false
		}
	}

	return true
}

func (s *Service) handleEvent(event fsnotify.Event) error {
	// Skip temporary files and editor backup files
	if isTemporaryFile(event.Name) {
		s.logger.Printf("Skipping temporary file: %s", event.Name)
		return nil
	}

	// Handle config file changes separately
	if event.Name == s.configPath && !s.reloading {
		if event.Op&fsnotify.Write == fsnotify.Write {
			s.logger.Println("Config file changed, reloading...")
			return s.handleConfigChange()
		} else {
			// Ignore CHMOD or other events on the config file
			return nil
		}
	}

	// Check if we should process this file
	if !s.shouldProcessFile(event.Name) {
		s.logger.Printf("Skipping event for filtered file: %s", event.Name)
		return nil
	}

	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		s.logger.Printf("File created: %s", event.Name)
		return s.handleCreate(event.Name)
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		s.logger.Printf("File removed: %s", event.Name)
		return s.handleRemove(event.Name)
	case event.Op&fsnotify.Write == fsnotify.Write:
		s.logger.Printf("File modified: %s", event.Name)
		return s.handleModify(event.Name)
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		s.logger.Printf("File renamed: %s", event.Name)
		return s.handleRename(event.Name)
	case event.Op&fsnotify.Chmod == fsnotify.Chmod:
		s.logger.Printf("File chmod: %s (ignoring)", event.Name)
		return nil
	}

	return nil
}

func (s *Service) handleCreate(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("error stating new path %s: %w", path, err)
	}

	if info.IsDir() {
		if s.shouldWatchDirectory(path) {
			if err := s.addWatchRecursive(path); err != nil {
				return fmt.Errorf("error adding watch to new directory %s: %w", path, err)
			}
		}
	}

	return s.triggerRegeneration()
}

func (s *Service) handleRemove(path string) error {
	s.mu.Lock()

	watched, exists := s.watched[path]
	if !exists {
		// Path wasn't being watched
		s.mu.Unlock()
		return s.triggerRegeneration()
	}

	// Remove the watch for this path
	if err := s.watcher.Remove(path); err != nil {
		// Log but don't fail - the path might already be gone
		log.Printf("Error removing watch for %s: %v", path, err)
	}
	delete(s.watched, path)

	// If it was a directory, remove watches for all subdirectories
	if watched.dir {
		prefix := path + string(filepath.Separator)
		for watchedPath := range s.watched {
			if strings.HasPrefix(watchedPath, prefix) {
				if err := s.watcher.Remove(watchedPath); err != nil {
					log.Printf("Error removing watch for %s: %v", watchedPath, err)
				}
				delete(s.watched, watchedPath)
			}
		}
	}

	s.mu.Unlock()
	return s.triggerRegeneration()
}

func (s *Service) handleModify(_ string) error {
	return s.triggerRegeneration()
}

func (s *Service) handleRename(path string) error {
	// Handle rename similar to remove
	return s.handleRemove(path)
}

func (s *Service) handleConfigChange() error {
	s.mu.Lock()
	s.reloading = true

	newConfig, err := config.LoadConfig("")
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("error reloading config: %w", err)
	}
	s.config.RepoConfig = newConfig

	if err := s.reconfigureWatcher(); err != nil {
		s.mu.Unlock()
		return fmt.Errorf("error reconfiguring watcher: %w", err)
	}

	if s.configPath != "" {
		if err := s.watcher.Add(s.configPath); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("error re-adding watch for config file: %w", err)
		}
		s.watched[s.configPath] = &watchedPath{path: s.configPath, dir: false}
	}

	s.reloading = false
	s.mu.Unlock()

	return s.triggerRegeneration()
}

func (s *Service) handleWatchError(err error) error {
	// Determine if the error is critical
	if isCriticalError(err) {
		return err
	}
	// Log non-critical errors
	log.Printf("Watch error: %v", err)
	return nil
}

func (s *Service) reconfigureWatcher() error {
	for path := range s.watched {
		s.watcher.Remove(path)
	}
	s.watched = make(map[string]*watchedPath)

	if err := s.addWatchRecursive(s.config.RootPath); err != nil {
		return fmt.Errorf("failed to add watches: %w", err)
	}

	return nil
}

func (s *Service) addWatchRecursive(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}

			if !s.shouldWatchDirectory(path) {
				return filepath.SkipDir
			}

			if err := s.watcher.Add(path); err != nil {
				return fmt.Errorf("failed to add watch for %s: %w", path, err)
			}
			s.watched[path] = &watchedPath{path: path, dir: true}
		}

		return nil
	})
}

func (s *Service) triggerRegeneration() error {
	s.mu.Lock()
	s.logger.Println("Triggering regeneration...")

	// Stop the timer first
	if !s.debouncer.Stop() {
		// Timer already fired, drain the channel
		select {
		case <-s.debouncer.C:
		default:
		}
	}

	// Now reset the timer to the configured debounce duration
	s.debouncer.Reset(s.config.DebounceTimeout)
	s.mu.Unlock()

	// Spawn a goroutine to wait for the debounce to expire and then regenerate
	go func() {
		<-s.debouncer.C
		s.logger.Println("Debounce timeout reached, regenerating...")
		if err := s.Generate(); err != nil {
			s.logger.Printf("Failed to regenerate: %v", err)
		}
	}()
	return nil
}

func (s *Service) Generate() error {
	fmt.Println("Generating...")
	return generator.RunGeneration(s.config.RepoConfig, s.config.RootPath)
}

func (s *Service) shouldWatchDirectory(path string) bool {
	// Convert absolute path to relative path from repo root
	relPath, err := filepath.Rel(s.config.RootPath, path)
	if err != nil {
		s.logger.Printf("Error getting relative path for %s: %v", path, err)
		return false
	}

	// Skip the root directory itself
	if relPath == "." {
		return true
	}

	ignored, err := s.gitignorer.IsIgnored(relPath)
	if err != nil {
		s.logger.Printf("Error checking if %s is ignored: %v", relPath, err)
		return false
	}
	if ignored {
		s.logger.Printf("Directory %s is ignored by gitignore", relPath)
		return false
	}

	if len(s.config.RepoConfig.ExcludePatterns) > 0 {
		if filter.MatchesAny(relPath, s.config.RepoConfig.ExcludePatterns, s.config.RepoConfig.CaseSensitive) {
			s.logger.Printf("Directory %s matches exclude pattern", relPath)
			return false
		}
	}

	return true
}

// Helper functions

func isCriticalError(err error) bool {
	// TODO: Add logic to determine if an error is critical
	// For example, permission errors or watcher resource exhaustion
	fmt.Println("isCriticalError", err)
	return false // Placeholder implementation
}

func isTemporaryFile(path string) bool {
	base := filepath.Base(path)
	return base == ".DS_Store" || // macOS
		base == "Thumbs.db" || // Windows
		base[0] == '.' || // Hidden files
		base[len(base)-1] == '~' || // Vim/Emacs backup
		base[0] == '#' && base[len(base)-1] == '#' // Emacs auto-save
}
