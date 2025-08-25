package template

import (
	"fmt"
	"regexp"
	"strings"
)

// parseWikiLinks transforms [[linktitle]] and [[linktitle|displayname]] into [title](link) format
func (rs Resource) parseWikiLinks(content string) string {
	// Regular expression to match [[linktitle]] and [[linktitle|displayname]] patterns
	// Allow empty content between brackets
	re := regexp.MustCompile(`\[\[([^\]]*)\]\]`)

	return re.ReplaceAllStringFunc(content, func(match string) string {
		// Check if this match is part of a triple bracket pattern [[[...]]]
		matchStart := strings.Index(content, match)
		if matchStart > 0 && content[matchStart-1] == '[' {
			// This is part of [[[...]], don't process it
			return match
		}
		if matchStart+len(match) < len(content) && content[matchStart+len(match)] == ']' {
			// This is part of [...]]]], don't process it
			return match
		}

		// Extract the content from [[content]]
		innerContent := strings.Trim(match, "[]")

		// Handle empty wiki links
		if innerContent == "" {
			return ""
		}

		// Check if this is a custom display name format: [[Page Title|Display Name]]
		var pageTitle, displayName string
		if strings.Contains(innerContent, "|") {
			parts := strings.SplitN(innerContent, "|", 2)
			pageTitle = strings.TrimSpace(parts[0])
			displayName = strings.TrimSpace(parts[1])
		} else {
			pageTitle = innerContent
			displayName = innerContent
		}

		// Find the corresponding note by title
		for _, note := range rs.Notes {
			if note.Title == pageTitle {
				// Return markdown link format [displayName](link)
				return fmt.Sprintf("[%s](/%s)", displayName, note.Slug)
			}
		}

		// If no matching note found, return the display name without brackets
		// This makes it clear that the link is broken
		return displayName
	})
}
