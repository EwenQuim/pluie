package engine

import (
	"sort"
	"strings"

	"github.com/EwenQuim/pluie/model"
)

type NoteScored struct {
	Note  model.Note
	Score int
}

// HeadingMatch represents a match within a note's heading
type HeadingMatch struct {
	Note    model.Note
	Heading string // The matched heading text
	Level   int    // 1 for H1, 2 for H2, 3 for H3
	Context string // Surrounding text (~150 chars)
	LineNum int    // Line number in note for potential linking
	Score   int    // Relevance score
}

// SearchNotesByFilename searches notes by filename (title and slug) and returns filtered results
// This function shows matches in the file name first, folder names second
func SearchNotesByFilename(notes []model.Note, searchQuery string) []model.Note {
	if searchQuery == "" {
		return notes
	}

	var filteredNotes []NoteScored
	searchLower := strings.ToLower(searchQuery)

	for _, note := range notes {
		// Check if search appears in the title (which represents the filename)
		if strings.Contains(strings.ToLower(note.Title), searchLower) {
			filteredNotes = append(filteredNotes, NoteScored{Note: note, Score: 2})
			continue
		}

		if strings.Contains(strings.ToLower(note.Slug), searchLower) {
			filteredNotes = append(filteredNotes, NoteScored{Note: note, Score: 1})
			continue
		}
	}

	// Sort by score descending
	sort.Slice(filteredNotes, func(i, j int) bool {
		return filteredNotes[i].Score > filteredNotes[j].Score
	})

	// Extract notes from scored results
	result := make([]model.Note, len(filteredNotes))
	for i, ns := range filteredNotes {
		result[i] = ns.Note
	}
	return result
}

// SearchNotesByHeadings searches for headings (H1-H3) matching the query
func SearchNotesByHeadings(notes []model.Note, searchQuery string, limit int) []HeadingMatch {
	if searchQuery == "" {
		return nil
	}

	var matches []HeadingMatch
	searchLower := strings.ToLower(searchQuery)

	for _, note := range notes {
		lines := strings.Split(note.Content, "\n")

		for i, line := range lines {
			trimmed := strings.TrimSpace(line)

			// Only check lines that start with #
			if !strings.HasPrefix(trimmed, "#") {
				continue
			}

			heading, level := extractHeading(line)

			// Only H1-H3 (skip H4-H6 and invalid headings)
			if level < 1 || level > 3 {
				continue
			}

			// Check if query matches heading
			if strings.Contains(strings.ToLower(heading), searchLower) {
				score := calculateHeadingScore(heading, searchQuery, level)
				context := extractContext(lines, i, 75)

				matches = append(matches, HeadingMatch{
					Note:    note,
					Heading: heading,
					Level:   level,
					Context: context,
					LineNum: i,
					Score:   score,
				})
			}
		}
	}

	// Sort by score descending
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// Limit to top N results
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}

	return matches
}

// extractHeading extracts the heading text and level from a markdown heading line
func extractHeading(line string) (string, int) {
	trimmed := strings.TrimSpace(line)
	level := 0

	// Count the # characters
	for i, char := range trimmed {
		if char == '#' {
			level++
		} else if char == ' ' {
			// Extract the heading text after the hashes and space
			return strings.TrimSpace(trimmed[i+1:]), level
		} else {
			// Invalid heading format (no space after #)
			break
		}
	}

	return "", 0
}

// calculateHeadingScore calculates relevance score for a heading match
func calculateHeadingScore(heading, query string, level int) int {
	score := 0
	headingLower := strings.ToLower(heading)
	queryLower := strings.ToLower(query)

	// Exact match gets highest score
	if headingLower == queryLower {
		score += 10
	} else if strings.HasPrefix(headingLower, queryLower) {
		// Starts with query
		score += 5
	} else if strings.Contains(headingLower, queryLower) {
		// Contains query
		score += 3
	}

	// Higher level headings get bonus points
	// H1=+3, H2=+2, H3=+1
	score += (4 - level)

	return score
}

// extractContext extracts surrounding text from the note content
// It gets approximately maxChars characters before and after the line
func extractContext(lines []string, lineNum int, maxChars int) string {
	if lineNum < 0 || lineNum >= len(lines) {
		return ""
	}

	var contextLines []string
	charsCollected := 0

	// Collect lines after the heading
	for i := lineNum + 1; i < len(lines) && charsCollected < maxChars; i++ {
		line := strings.TrimSpace(lines[i])

		// Skip empty lines and other headings
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		contextLines = append(contextLines, line)
		charsCollected += len(line)
	}

	// Join and truncate to max length
	context := strings.Join(contextLines, " ")
	if len(context) > maxChars {
		context = context[:maxChars] + "..."
	}

	return context
}
