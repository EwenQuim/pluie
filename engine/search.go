package engine

import (
	"strings"

	"github.com/EwenQuim/pluie/model"
)

// HeadingMatch represents a match within a note's heading
type HeadingMatch struct {
	Note    model.Note
	Heading string // The matched heading text
	Level   int    // 1 for H1, 2 for H2, 3 for H3
	Context string // Surrounding text (~150 chars)
	LineNum int    // Line number in note for potential linking
	Score   int    // Relevance score
}

// extractHeading extracts the heading text and level from a markdown heading line
func extractHeading(line string) (string, int) {
	trimmed := strings.TrimSpace(line)
	level := 0

	// Count the # characters
	for i, char := range trimmed {
		switch char {
		case '#':
			level++
		case ' ':
			// Extract the heading text after the hashes and space
			return strings.TrimSpace(trimmed[i+1:]), level
		default:
			return "", 0
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
