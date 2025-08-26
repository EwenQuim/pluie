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
	IsPublic     bool            `json:"isPublic"`      // Whether this note is public or private
	Metadata     map[string]any  `json:"metadata"`      // YAML frontmatter metadata
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

// DetermineIsPublic sets the IsPublic field based on the hierarchy rules:
// 1. Check the note's own "public" metadata
// 2. Check parent folder metadata (if any)
// 3. Fall back to the server's default (publicByDefault parameter)
func (n *Note) DetermineIsPublic(publicByDefault bool, folderMetadata map[string]map[string]any) {
	// First, check the note's own metadata for "public" field
	if publicValue, exists := n.Metadata["public"]; exists {
		if publicBool, ok := publicValue.(bool); ok {
			n.IsPublic = publicBool
			return
		}
	}

	// Second, check parent folder metadata
	// Extract the folder path from the slug
	pathParts := strings.Split(strings.Trim(n.Slug, "/"), "/")
	if len(pathParts) > 1 {
		// Build folder path (all parts except the last one which is the file)
		folderPath := strings.Join(pathParts[:len(pathParts)-1], "/")

		// Check if folder has metadata
		if metadata, exists := folderMetadata[folderPath]; exists {
			if publicValue, exists := metadata["public"]; exists {
				if publicBool, ok := publicValue.(bool); ok {
					n.IsPublic = publicBool
					return
				}
			}
		}
	}

	// Third, fall back to server default
	n.IsPublic = publicByDefault
}
