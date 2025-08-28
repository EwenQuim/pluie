package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	"github.com/EwenQuim/pluie/template"
)

func main() {
	path := flag.String("path", ".", "Path to the obsidian folder")

	flag.Parse()

	// Load configuration
	cfg := config.Load()

	explorer := Explorer{
		BasePath: *path,
	}

	notes, err := explorer.getFolderNotes("")
	if err != nil {
		slog.Error("Error exploring folder", "error", err)
		return
	}

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

	server := &Server{
		NotesMap: &notesMap,
		Tree:     tree,
		rs: template.Resource{
			Tree: tree,
		},
		cfg: cfg,
	}

	// Create and start file watcher
	watcher, err := NewFileWatcher(*path, server, cfg)
	if err != nil {
		slog.Error("Failed to create file watcher", "error", err)
		return
	}

	err = watcher.Start()
	if err != nil {
		slog.Error("Failed to start file watcher", "error", err)
		return
	}
	defer watcher.Stop()

	// Set up graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start()
	}()

	// Wait for either an error or a shutdown signal
	select {
	case err := <-errCh:
		if err != nil {
			slog.Error("Server error", "error", err)
		}
	case sig := <-sigCh:
		slog.Info("Shutting down", "signal", sig)
	}
}
