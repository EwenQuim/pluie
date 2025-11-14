package main

import (
	"flag"
	"log/slog"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/template"
)

func main() {
	path := flag.String("path", ".", "Path to the obsidian folder")
	watch := flag.Bool("watch", true, "Enable file watching to auto-reload on changes")
	mode := flag.String("mode", "server", "Mode to run in: server (default) or static")

	flag.Parse()

	// Load configuration
	cfg := config.Load()

	// Load initial notes
	notesMap, tree, tagIndex, err := loadNotes(*path, cfg)
	if err != nil {
		slog.Error("Error loading notes", "error", err)
		return
	}

	notesService := engine.NewNotesService(notesMap, tree, tagIndex)

	// Run in static mode if requested
	if *mode == "static" {
		err := generateStaticSite(notesService, cfg)
		if err != nil {
			slog.Error("Error generating static site", "error", err)
			return
		}
		slog.Info("Static site generated successfully in /dist folder")
		return
	}

	// Otherwise run in server mode
	server := &Server{
		NotesService: notesService,
		rs:           template.Resource{},
		cfg:          cfg,
	}

	// Start file watcher if enabled
	if *watch {
		_, err = watchFiles(server, *path, cfg)
		if err != nil {
			slog.Error("Error starting file watcher", "error", err)
			// Continue anyway - the server can still work without file watching
		}
	}

	err = server.Start()
	if err != nil {
		panic(err)
	}

}
