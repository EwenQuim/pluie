package main

import (
	"flag"
	"fmt"

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
		fmt.Println(err)
		return
	}

	// Filter out private notes
	fmt.Print("Filtering private notes... ")
	publicNotes := filterPublicNotes(notes, cfg.PublicByDefault)
	fmt.Printf("%d public notes out of %d total\n", len(publicNotes), len(notes))

	// Build backreferences for public notes only
	fmt.Print("Building backreferences... ")
	publicNotes = engine.BuildBackreferences(publicNotes)
	fmt.Println("done")

	notesMap := make(map[string]model.Note)
	for _, note := range publicNotes {
		notesMap[note.Slug] = note
	}

	// Build tree structure with public notes only
	fmt.Print("Building tree structure... ")
	tree := engine.BuildTree(publicNotes)
	fmt.Println("done")

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
