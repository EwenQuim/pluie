package template

import (
	"regexp"
	"strings"

	"github.com/EwenQuim/pluie/model"
)

// BuildBackreferences analyzes all notes and populates the ReferencedBy field
// for each note based on wikilinks found in other notes' content
func BuildBackreferences(notes []model.Note) []model.Note {
	// Create a map for quick note lookup by title
	notesByTitle := make(map[string]*model.Note)

	// Initialize all notes with empty ReferencedBy slices
	for i := range notes {
		notes[i].ReferencedBy = []model.NoteReference{}
		notesByTitle[notes[i].Title] = &notes[i]
	}

	// Regular expression to match [[linktitle]] and [[linktitle|displayname]] patterns
	wikiLinkRegex := regexp.MustCompile(`\[\[([^\]]*)\]\]`)

	// Analyze each note for wikilinks
	for _, sourceNote := range notes {
		// Find all wikilinks in the source note's content
		wikiLinks := extractWikiLinks(sourceNote.Content, wikiLinkRegex)

		// For each wikilink, add this note as a reference to the target note
		for _, targetTitle := range wikiLinks {
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
func extractWikiLinks(content string, regex *regexp.Regexp) []string {
	var links []string
	seen := make(map[string]bool)

	matches := regex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		// Check if this match is part of a triple bracket pattern [[[...]]]
		fullMatch := match[0]
		matchStart := strings.Index(content, fullMatch)
		if matchStart > 0 && content[matchStart-1] == '[' {
			// This is part of [[[...]], skip it
			continue
		}
		if matchStart+len(fullMatch) < len(content) && content[matchStart+len(fullMatch)] == ']' {
			// This is part of [...]]]], skip it
			continue
		}

		// Extract the content from [[content]]
		innerContent := strings.Trim(match[1], " ")

		// Handle empty wiki links
		if innerContent == "" {
			continue
		}

		// Extract the target title (before the | if present)
		var targetTitle string
		if strings.Contains(innerContent, "|") {
			parts := strings.SplitN(innerContent, "|", 2)
			targetTitle = strings.TrimSpace(parts[0])
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
