package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
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
	fmt.Println(notes)

	urls := make([]string, 0, len(notes))
	for _, note := range notes {
		urls = append(urls, strings.TrimSuffix(note.Slug, ".md"))
	}
	fmt.Println(urls)

}

type Note struct {
	Title   string // May contains spaces and slashes, like "articles/Hello World"
	Slug    string // Slugified title, like "articles/hello-world"
	Content string
}

// If Slug is not defined, build it from the title
// Replace spaces with dashes
// Replace multiple dashes with a single dash
// Remove leading and trailing dashes
func (n *Note) BuildSlug() {
	n.Slug = strings.TrimSuffix(n.Slug, ".md")

	if n.Slug == "" {
		n.Slug = n.Title
	}

	n.Slug = strings.ReplaceAll(n.Slug, " ", "-")
	for strings.Contains(n.Slug, "--") {
		n.Slug = strings.ReplaceAll(n.Slug, "--", "-")
	}

	n.Slug = strings.Trim(n.Slug, "/")
	n.Slug = strings.Trim(n.Slug, "-")

	n.Slug = url.PathEscape(n.Slug)
	n.Slug = strings.ReplaceAll(n.Slug, "%2F", "/")
}

type Explorer struct {
	BasePath string
}

func (e Explorer) getFolderNotes(currentPath string) ([]Note, error) {
	if strings.HasPrefix(currentPath, "/.") || strings.Contains(currentPath, "node_modules") {
		return nil, nil
	}

	fmt.Println(e, currentPath)
	dir, err := os.ReadDir(e.BasePath + "/" + currentPath)
	if err != nil {
		return nil, err
	}

	notes := make([]Note, 0)

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
			note := Note{
				Title:   name,
				Content: content,
				Slug:    path.Join(currentPath, name),
			}
			note.BuildSlug()
			notes = append(notes, note)
		}
	}

	return notes, nil
}
