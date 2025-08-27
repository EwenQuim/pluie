package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	"github.com/EwenQuim/pluie/template"
	"github.com/adrg/frontmatter"
)

func main() {
	path := flag.String("path", ".", "Path to the obsidian folder")

	flag.Parse()

	// Get PUBLIC_BY_DEFAULT environment variable
	publicByDefault := false
	if envValue := os.Getenv("PUBLIC_BY_DEFAULT"); envValue != "" {
		if parsed, err := strconv.ParseBool(envValue); err == nil {
			publicByDefault = parsed
			fmt.Println("PUBLIC_BY_DEFAULT set to", publicByDefault)
		}
	}

	explorer := Explorer{
		BasePath:        *path,
		PublicByDefault: publicByDefault,
	}

	notes, err := explorer.getFolderNotes("")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Filter out private notes
	fmt.Print("Filtering private notes... ")
	publicNotes := make([]model.Note, 0)
	if publicByDefault {
		publicNotes = notes
	} else {
		for _, note := range notes {
			if note.IsPublic {
				publicNotes = append(publicNotes, note)
			}
		}
	}
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
	}.Start()
	if err != nil {
		panic(err)
	}

}

type Explorer struct {
	BasePath        string
	PublicByDefault bool
}

func (e Explorer) getFolderNotes(currentPath string) ([]model.Note, error) {
	if strings.HasPrefix(currentPath, "/.") || strings.Contains(currentPath, "node_modules") || strings.Contains(currentPath, ".git") {
		return nil, nil
	}

	fmt.Print("Searching ", currentPath, "... ")
	dir, err := os.ReadDir(e.BasePath + "/" + currentPath)
	if err != nil {
		return nil, err
	}

	notes := make([]model.Note, 0)
	folderMetadata := make(map[string]map[string]any)

	// First pass: collect folder metadata from .pluie files
	for _, entry := range dir {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".pluie") {
			metadataBytes, err := os.ReadFile(path.Join(e.BasePath, currentPath, entry.Name()))
			if err != nil {
				continue // Skip if can't read metadata file
			}

			var metadata map[string]any
			_, err = frontmatter.Parse(strings.NewReader(string(metadataBytes)), &metadata)
			if err != nil {
				continue // Skip if can't parse metadata
			}

			// Store metadata for the current folder
			folderPath := strings.Trim(currentPath, "/")
			folderMetadata[folderPath] = metadata
		}
	}

	// Second pass: process files and folders
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

			// Parse frontmatter
			var metadata map[string]any
			parsedContent, err := frontmatter.Parse(strings.NewReader(string(contentBytes)), &metadata)
			var finalContent string
			if err != nil {
				// If frontmatter parsing fails, use the original content
				finalContent = string(contentBytes)
				metadata = make(map[string]any)
			} else {
				finalContent = string(parsedContent)
			}

			// Set title from frontmatter if available, otherwise use filename
			title := strings.TrimSuffix(name, ".md")
			if frontmatterTitle, exists := metadata["title"]; exists {
				if titleStr, ok := frontmatterTitle.(string); ok && titleStr != "" {
					title = titleStr
				}
			}

			note := model.Note{
				Title:    title,
				Content:  finalContent,
				Slug:     path.Join(currentPath, name),
				Path:     path.Join(currentPath, name),
				Metadata: metadata,
			}
			note.BuildSlug()

			// Determine if the note is public based on hierarchy
			note.DetermineIsPublic(folderMetadata)

			notes = append(notes, note)
		}
	}
	fmt.Println(len(notes), "notes found")

	return notes, nil
}
