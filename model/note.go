package model

import (
	"net/url"
	"strings"
)

type NoteReference struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

type Note struct {
	Title        string          `json:"title"` // May contains spaces and slashes, like "articles/Hello World"
	Slug         string          `json:"slug"`  // Slugified title, like "articles/hello-world"
	Content      string          `json:"content"`
	ReferencedBy []NoteReference `json:"referenced_by"` // Notes that have wikilinks to this note
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
