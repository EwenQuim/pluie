package main

import (
	"flag"
	"log/slog"

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

	err = Server{
		NotesMap: notesMap,
		Tree:     tree,
		rs: template.Resource{
			Tree: tree,
		},
		cfg: cfg,
	}.Start()
	if err != nil {
		panic(err)
	}

}
