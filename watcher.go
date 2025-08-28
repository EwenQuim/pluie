package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	"github.com/fsnotify/fsnotify"
)

// FileWatcher handles file system events and reloads notes when changes occur
type FileWatcher struct {
	basePath    string
	explorer    Explorer
	server      *Server
	cfg         *config.Config
	watcher     *fsnotify.Watcher
	debouncer   *time.Timer
	debounceMu  sync.Mutex
	stopCh      chan struct{}
}

// NewFileWatcher creates a new file watcher instance
func NewFileWatcher(basePath string, server *Server, cfg *config.Config) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fw := &FileWatcher{
		basePath: basePath,
		explorer: Explorer{BasePath: basePath},
		server:   server,
		cfg:      cfg,
		watcher:  watcher,
		stopCh:   make(chan struct{}),
	}

	return fw, nil
}

// Start begins watching the file system for changes
func (fw *FileWatcher) Start() error {
	// Add the base path to the watcher
	err := fw.addWatchRecursive(fw.basePath)
	if err != nil {
		return err
	}

	slog.Info("File watcher started", "path", fw.basePath)

	go fw.watchLoop()
	return nil
}

// Stop stops the file watcher
func (fw *FileWatcher) Stop() {
	close(fw.stopCh)
	if fw.watcher != nil {
		fw.watcher.Close()
	}
	
	fw.debounceMu.Lock()
	if fw.debouncer != nil {
		fw.debouncer.Stop()
	}
	fw.debounceMu.Unlock()
	
	slog.Info("File watcher stopped")
}

// addWatchRecursive adds watchers for a directory and all its subdirectories
func (fw *FileWatcher) addWatchRecursive(path string) error {
	return filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if this is a file
		if !info.IsDir() {
			return nil
		}

		// Get relative path for shouldSkipPath check
		relPath, err := filepath.Rel(fw.basePath, walkPath)
		if err != nil {
			return err
		}

		// Normalize path separators for the skip check
		relPath = filepath.ToSlash(relPath)
		if relPath == "." {
			relPath = ""
		} else {
			relPath = "/" + relPath
		}

		if fw.explorer.shouldSkipPath(relPath) {
			return filepath.SkipDir
		}

		// Add watcher for this directory
		err = fw.watcher.Add(walkPath)
		if err != nil {
			slog.Warn("Failed to add watch", "path", walkPath, "error", err)
		}

		return nil
	})
}

// watchLoop handles file system events
func (fw *FileWatcher) watchLoop() {
	for {
		select {
		case <-fw.stopCh:
			return

		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			fw.handleEvent(event)

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			slog.Error("File watcher error", "error", err)
		}
	}
}

// handleEvent processes a single file system event
func (fw *FileWatcher) handleEvent(event fsnotify.Event) {
	// Check if this is a file we care about
	if !fw.shouldProcessEvent(event) {
		return
	}

	slog.Debug("File system event", "event", event.Op.String(), "path", event.Name)

	// Handle directory creation (need to add watcher)
	if event.Op&fsnotify.Create == fsnotify.Create {
		if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
			// Add watcher for new directory
			fw.watcher.Add(event.Name)
			slog.Debug("Added watcher for new directory", "path", event.Name)
		}
	}

	// Debounce the reload to avoid excessive processing
	fw.debounceReload()
}

// shouldProcessEvent determines if we should process this file system event
func (fw *FileWatcher) shouldProcessEvent(event fsnotify.Event) bool {
	// Only process certain operations
	if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) == 0 {
		return false
	}

	// Get relative path
	relPath, err := filepath.Rel(fw.basePath, event.Name)
	if err != nil {
		return false
	}

	// Normalize path separators
	relPath = filepath.ToSlash(relPath)
	if relPath == "." {
		relPath = ""
	} else {
		relPath = "/" + relPath
	}

	// Skip paths we don't care about
	if fw.explorer.shouldSkipPath(relPath) {
		return false
	}

	// Process markdown files and directory events
	return strings.HasSuffix(event.Name, ".md") || 
		   strings.HasSuffix(event.Name, ".metadata") ||
		   event.Op&(fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0
}

// debounceReload debounces reload operations to avoid excessive processing
func (fw *FileWatcher) debounceReload() {
	fw.debounceMu.Lock()
	defer fw.debounceMu.Unlock()

	if fw.debouncer != nil {
		fw.debouncer.Stop()
	}

	fw.debouncer = time.AfterFunc(300*time.Millisecond, func() {
		fw.reloadNotes()
	})
}

// reloadNotes reloads all notes and updates the server
func (fw *FileWatcher) reloadNotes() {
	start := time.Now()
	slog.Info("Reloading notes due to file system changes")

	// Load all notes
	notes, err := fw.explorer.getFolderNotes("")
	if err != nil {
		slog.Error("Error reloading notes", "error", err)
		return
	}

	// Filter out private notes
	publicNotes := filterPublicNotes(notes, fw.cfg.PublicByDefault)

	// Build backreferences for public notes only
	publicNotes = engine.BuildBackreferences(publicNotes)

	// Create a new map of notes for quick access by slug
	notesMap := make(map[string]model.Note)
	for _, note := range publicNotes {
		notesMap[note.Slug] = note
	}

	// Build tree structure with public notes only
	tree := engine.BuildTree(publicNotes)

	// Update the server's data atomically
	fw.updateServerData(&notesMap, tree)

	slog.Info("Notes reloaded successfully", 
		"count", len(publicNotes), 
		"duration", time.Since(start))
}

// updateServerData atomically updates the server's data structures
func (fw *FileWatcher) updateServerData(notesMap *map[string]model.Note, tree *engine.TreeNode) {
	// Update the server's data using the server's update method
	fw.server.UpdateData(notesMap, tree)
}