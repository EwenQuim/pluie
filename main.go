package main

import (
	"flag"
	"log/slog"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/template"
)

func main() {
	path := flag.String("path", ".", "Path to the obsidian folder")
	watch := flag.Bool("watch", true, "Enable file watching to auto-reload on changes")

	flag.Parse()

	// Load configuration
	cfg := config.Load()

	// Load initial notes
	notesMap, tree, tagIndex, err := loadNotes(*path, cfg)
	if err != nil {
		slog.Error("Error loading notes", "error", err)
		return
	}

	server := &Server{
		NotesMap: notesMap,
		Tree:     tree,
		TagIndex: tagIndex,
		rs: template.Resource{
			Tree: tree,
		},
		cfg: cfg,
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
