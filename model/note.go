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
	Slug         string          `json:"slug"`  // Slugified title, like "my-articles/hello-world"
	Path         string          `json:"path"`  // Full path relative to the base directory, like "My articles/Hello World.md"
	Content      string          `json:"content"`
	ReferencedBy []NoteReference `json:"referenced_by"` // Notes that have wikilinks to this note
	IsPublic     bool            `json:"isPublic"`      // Whether this note is public or private
	Metadata     map[string]any  `json:"metadata"`      // YAML frontmatter metadata
}

// BuildSlug creates a URL-friendly slug from the note's title or existing slug
// This uses the unified slugification approach for notes (matches engine.SlugifyNoteWithCaseLogic)
func (n *Note) BuildSlug() {
	text := n.Slug
	usingTitle := false
	if text == "" {
		text = n.Title
		usingTitle = true
	}

	// Remove file extension
	slug := strings.TrimSuffix(text, ".md")

	// Only convert to lowercase if we're using an existing slug (not creating from title)
	if !usingTitle {
		slug = strings.ToLower(slug)
	}

	// For paths: replace spaces with dashes but preserve forward slashes
	slug = strings.ReplaceAll(slug, " ", "-")

	// Clean up multiple consecutive dashes
	slug = cleanMultipleDashes(slug)

	// Clean leading/trailing slashes and dashes
	slug = strings.Trim(slug, "/-")

	// URL encode while preserving forward slashes
	slug = url.PathEscape(slug)
	slug = strings.ReplaceAll(slug, "%2F", "/")

	n.Slug = slug
}

// cleanMultipleDashes removes consecutive dashes using a single pass
func cleanMultipleDashes(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	prevDash := false
	for _, r := range s {
		if r == '-' {
			if !prevDash {
				result.WriteRune(r)
				prevDash = true
			}
		} else {
			result.WriteRune(r)
			prevDash = false
		}
	}

	return result.String()
}

// DetermineIsPublic sets the IsPublic field based on the hierarchy rules:
// 1. Check the note's own "publish" metadata
// 2. Check parent folder metadata (if any)
// 3. Fall back to private by default
func (n *Note) DetermineIsPublic(folderMetadata map[string]map[string]any) {
	// First, check the note's own metadata for "publish" field
	if publishValue, exists := n.Metadata["publish"]; exists {
		if publishBool, ok := publishValue.(bool); ok {
			n.IsPublic = publishBool
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
			if publishValue, exists := metadata["publish"]; exists {
				if publishBool, ok := publishValue.(bool); ok {
					n.IsPublic = publishBool
					return
				}
			}
		}
	}

	// Third, fall back to private by default
	n.IsPublic = false
}
