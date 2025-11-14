package engine

import (
	"log/slog"
	"strings"
	"time"

	"github.com/EwenQuim/pluie/model"
)

// BuildBackreferences analyzes all notes and populates the ReferencedBy field
// for each note based on wikilinks found in other notes' content
func BuildBackreferences(notes []model.Note) []model.Note {
	start := time.Now()
	defer func() {
		slog.Info("Backreferences built", "in", time.Since(start).String())
	}()

	// Create a map for quick note lookup by title
	notesByTitle := make(map[string]*model.Note)

	// Initialize all notes with empty ReferencedBy slices
	for i := range notes {
		notes[i].ReferencedBy = []model.NoteReference{}
		notesByTitle[notes[i].Title] = &notes[i]
	}

	// Analyze each note for wikilinks
	for _, sourceNote := range notes {
		// Find all wikilinks in the source note's content
		contentWikiLinks := extractWikiLinks(sourceNote.Content)

		// Find all wikilinks in the source note's metadata
		metadataWikiLinks := extractWikiLinksFromMetadata(sourceNote.Metadata)

		// Combine all wikilinks from both content and metadata
		allWikiLinks := append(contentWikiLinks, metadataWikiLinks...)

		// Remove duplicates from the combined list
		uniqueWikiLinks := removeDuplicateStrings(allWikiLinks)

		// For each wikilink, add this note as a reference to the target note
		for _, targetTitle := range uniqueWikiLinks {
			if targetNote, exists := notesByTitle[targetTitle]; exists {
				// Add the source note as a reference to the target note
				reference := model.NoteReference{
					Slug:  sourceNote.Slug,
					Title: sourceNote.Title,
				}

				// Check if this reference already exists to avoid duplicates
				if !containsReference(targetNote.ReferencedBy, reference) {
					targetNote.ReferencedBy = append(targetNote.ReferencedBy, reference)
				}
			}
		}
	}

	return notes
}

// extractWikiLinks extracts all unique target titles from wikilinks in the content
func extractWikiLinks(content string) []string {
	var links []string
	seen := make(map[string]bool)

	// Use FindAllStringSubmatchIndex for better performance - returns positions instead of creating substrings
	matchIndices := wikiLinkRegex.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matchIndices {
		if len(match) < 4 {
			continue
		}

		matchStart := match[0]
		matchEnd := match[1]
		innerStart := match[2]
		innerEnd := match[3]

		// Check if this match is part of a triple bracket pattern [[[...]]]
		if matchStart > 0 && content[matchStart-1] == '[' {
			// This is part of [[[...]], skip it
			continue
		}
		if matchEnd < len(content) && content[matchEnd] == ']' {
			// This is part of [...]]]], skip it
			continue
		}

		// Extract the content from [[content]]
		innerContent := strings.TrimSpace(content[innerStart:innerEnd])

		// Handle empty wiki links
		if innerContent == "" {
			continue
		}

		// Extract the target title (before the | if present)
		var targetTitle string
		if pipeIdx := strings.IndexByte(innerContent, '|'); pipeIdx != -1 {
			targetTitle = strings.TrimSpace(innerContent[:pipeIdx])
		} else {
			targetTitle = innerContent
		}

		// Add to links if not already seen
		if targetTitle != "" && !seen[targetTitle] {
			links = append(links, targetTitle)
			seen[targetTitle] = true
		}
	}

	return links
}

// containsReference checks if a reference already exists in the slice
func containsReference(references []model.NoteReference, target model.NoteReference) bool {
	for _, ref := range references {
		if ref.Slug == target.Slug && ref.Title == target.Title {
			return true
		}
	}
	return false
}

// extractWikiLinksFromMetadata extracts wikilinks from all string values in metadata
func extractWikiLinksFromMetadata(metadata map[string]any) []string {
	var allLinks []string

	for _, value := range metadata {
		links := extractWikiLinksFromValue(value)
		allLinks = append(allLinks, links...)
	}

	return allLinks
}

// extractWikiLinksFromValue recursively extracts wikilinks from any value type
func extractWikiLinksFromValue(value any) []string {
	var links []string

	switch v := value.(type) {
	case string:
		// Extract wikilinks from string values
		links = append(links, extractWikiLinks(v)...)
	case []any:
		// Handle arrays/slices
		for _, item := range v {
			links = append(links, extractWikiLinksFromValue(item)...)
		}
	case map[string]any:
		// Handle nested objects
		for _, nestedValue := range v {
			links = append(links, extractWikiLinksFromValue(nestedValue)...)
		}
		// For other types (int, bool, etc.), we don't extract wikilinks
	}

	return links
}

// removeDuplicateStrings removes duplicate strings from a slice while preserving order
func removeDuplicateStrings(slice []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
