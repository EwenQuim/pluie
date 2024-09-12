package model

import (
	"net/url"
	"strings"
)

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
