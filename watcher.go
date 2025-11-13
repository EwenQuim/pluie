package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	"github.com/fsnotify/fsnotify"
)

// loadNotes loads all notes from the given path, processes them, and returns the data structures
func loadNotes(basePath string, cfg *config.Config) (*map[string]model.Note, *engine.TreeNode, engine.TagIndex, error) {
	start := time.Now()

	explorer := Explorer{
		BasePath: basePath,
	}

	notes, err := explorer.getFolderNotes("")
	if err != nil {
		return nil, nil, nil, err
	}

	slog.Info("Processed files", "in", time.Since(start).String())

	// Filter out private notes
	publicNotes := filterPublicNotes(notes, cfg.PublicByDefault)

	// Build backreferences for public notes only
	publicNotes = engine.BuildBackreferences(publicNotes)

	// Create a map of notes for quick access by slug
	notesMap := make(map[string]model.Note)
	for _, note := range publicNotes {
		notesMap[note.Slug] = note
	}

	// Build tree structure with public notes only
	tree := engine.BuildTree(publicNotes)

	// Build tag index with public notes only
	tagIndex := engine.BuildTagIndex(publicNotes)

	slog.Info("Loaded notes", "total_time", time.Since(start).String(), "count", len(publicNotes))

	return &notesMap, tree, tagIndex, nil
}

// watchFiles sets up a file watcher that monitors changes in the vault directory
// and reloads the server data when changes are detected
// Returns the watcher so it can be closed by the caller
func watchFiles(server *Server, basePath string, cfg *config.Config) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Watch the base directory and all subdirectories
	err = addDirectoryRecursive(watcher, basePath)
	if err != nil {
		watcher.Close()
		return nil, err
	}

	// Start watching in a goroutine
	go func() {
		// Debounce timer to avoid reloading too frequently
		var debounceTimer *time.Timer
		debounceDuration := 500 * time.Millisecond

		defer watcher.Close()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// Only reload on write, create, remove, or rename events
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
					slog.Info("File change detected", "file", event.Name, "op", event.Op.String())

					// If a new directory was created, add it to the watcher
					if event.Op&fsnotify.Create != 0 {
						if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
							addDirectoryRecursive(watcher, event.Name)
						}
					}

					// Debounce: reset the timer on each event
					if debounceTimer != nil {
						debounceTimer.Stop()
					}

					debounceTimer = time.AfterFunc(debounceDuration, func() {
						slog.Info("Reloading notes due to file changes")
						notesMap, tree, tagIndex, err := loadNotes(basePath, cfg)
						if err != nil {
							slog.Error("Error reloading notes", "error", err)
							return
						}

						server.UpdateData(notesMap, tree, tagIndex)
						slog.Info("Notes reloaded successfully")
					})
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				slog.Error("File watcher error", "error", err)
			}
		}
	}()

	slog.Info("File watcher started", "path", basePath)
	return watcher, nil
}

// addDirectoryRecursive adds a directory and all its subdirectories to the watcher
func addDirectoryRecursive(watcher *fsnotify.Watcher, path string) error {
	watchCount := 0

	// Get absolute path of the root to compare later
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	err = filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get absolute path of current walk path
		absWalkPath, err := filepath.Abs(walkPath)
		if err != nil {
			return err
		}

		// Skip .git and other hidden directories, but don't skip the root directory itself
		base := filepath.Base(walkPath)
		if len(base) > 0 && base[0] == '.' && absWalkPath != absPath {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Add directory to watcher (we only need to watch directories on most systems)
		if info.IsDir() {
			err = watcher.Add(walkPath)
			if err != nil {
				slog.Warn("Failed to watch directory", "path", walkPath, "error", err)
			} else {
				watchCount++
				slog.Debug("Watching directory", "path", walkPath)
			}
		}

		return nil
	})

	slog.Info("Added directories to watcher", "count", watchCount, "root", path)
	return err
}
