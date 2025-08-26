package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/EwenQuim/pluie/model"
	"github.com/EwenQuim/pluie/template"
)

func main() {
	path := flag.String("path", ".", "Path to the obsidian folder")

	flag.Parse()

	explorer := Explorer{
		BasePath: *path,
	}

	notes, err := explorer.getFolderNotes("")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Build backreferences for all notes
	fmt.Print("Building backreferences... ")
	notes = template.BuildBackreferences(notes)
	fmt.Println("done")

	urls := make([]string, 0, len(notes))
	for _, note := range notes {
		urls = append(urls, strings.TrimSuffix(note.Slug, ".md"))
	}

	notesMap := make(map[string]model.Note)
	for _, note := range notes {
		notesMap[note.Slug] = note
	}

	err = Server{
		NotesMap: notesMap,
		rs: template.Resource{
			Notes: notes,
		},
	}.Start()
	if err != nil {
		panic(err)
	}

}

type Explorer struct {
	BasePath string
}

func (e Explorer) getFolderNotes(currentPath string) ([]model.Note, error) {
	if strings.HasPrefix(currentPath, "/.") || strings.Contains(currentPath, "node_modules") {
		return nil, nil
	}

	fmt.Print("Searching ", currentPath, "... ")
	dir, err := os.ReadDir(e.BasePath + "/" + currentPath)
	if err != nil {
		return nil, err
	}

	notes := make([]model.Note, 0)

	for _, entry := range dir {
		if entry.IsDir() {
			// Recursively get notes from subfolder
			subfolderNotes, err := e.getFolderNotes(currentPath + "/" + entry.Name())
			if err != nil {
				return nil, err
			}

			notes = append(notes, subfolderNotes...)
		} else {
			name := entry.Name()
			if !strings.HasSuffix(name, ".md") {
				continue
			}

			contentBytes, err := os.ReadFile(path.Join(e.BasePath, currentPath, name))
			if err != nil {
				return nil, err
			}
			content := string(contentBytes)
			note := model.Note{
				Title:   strings.TrimSuffix(name, ".md"),
				Content: content,
				Slug:    path.Join(currentPath, name),
			}
			note.BuildSlug()
			notes = append(notes, note)
		}
	}
	fmt.Println(len(notes), "notes found")

	return notes, nil
}
