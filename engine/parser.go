package engine

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/EwenQuim/pluie/model"
)

// RemoveCommentBlocks removes all content between %% markers (inclusive)
func RemoveCommentBlocks(content string) string {
	// Regular expression to match content between %% markers (including the markers)
	// The (?s) flag enables dot-all mode, making . match newlines as well
	re := regexp.MustCompile(`(?s)%%.*?%%`)
	return re.ReplaceAllString(content, "")
}

// ParseWikiLinksInMetadata processes wikilinks in metadata values (strings and lists)
func ParseWikiLinksInMetadata(metadata map[string]any, tree *TreeNode) map[string]any {
	if metadata == nil {
		return metadata
	}

	result := make(map[string]any)
	for key, value := range metadata {
		result[key] = parseWikiLinksInValue(value, tree)
	}
	return result
}

// parseWikiLinksInValue recursively processes wikilinks in a metadata value
func parseWikiLinksInValue(value any, tree *TreeNode) any {
	switch v := value.(type) {
	case string:
		// Parse wikilinks in string values
		return ParseWikiLinks(v, tree)
	case []interface{}:
		// Parse wikilinks in list items
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = parseWikiLinksInValue(item, tree)
		}
		return result
	case map[string]interface{}:
		// Recursively parse wikilinks in nested objects
		result := make(map[string]interface{})
		for k, val := range v {
			result[k] = parseWikiLinksInValue(val, tree)
		}
		return result
	default:
		// Return other types unchanged (bool, int, float, etc.)
		return value
	}
}

// ProcessMarkdownLinks removes .md extensions from markdown links before rendering
func ProcessMarkdownLinks(content string) string {
	// Regular expression to match [text](link.md) patterns and remove .md extension
	// This handles cases with query parameters and anchors after .md
	re := regexp.MustCompile(`\[([^\]]*)\]\(([^)]+?)\.md(\)|[?#][^)]*\))`)

	return re.ReplaceAllStringFunc(content, func(match string) string {
		// Extract parts using the regex
		parts := re.FindStringSubmatch(match)
		if len(parts) != 4 {
			return match
		}

		linkText := parts[1]
		pathPart := parts[2]
		suffix := parts[3]

		// Return the reconstructed link without .md
		return fmt.Sprintf("[%s](%s%s", linkText, pathPart, suffix)
	})
}

// ParseWikiLinks transforms [[linktitle]] and [[linktitle|displayname]] into [title](link) format
func ParseWikiLinks(content string, tree *TreeNode) string {
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

		// Find the corresponding note by title in the tree using iterator
		var foundNote *model.Note
		tree.AllNotes(func(noteNode *TreeNode) bool {
			if noteNode.Note != nil && noteNode.Note.Title == pageTitle {
				foundNote = noteNode.Note
				return false // Stop iteration
			}
			return true // Continue iteration
		})

		if foundNote != nil {
			// Return markdown link format [displayName](link)
			return fmt.Sprintf("[%s](/%s)", displayName, foundNote.Slug)
		}

		// If no matching note found, return the display name without brackets
		// This makes it clear that the link is broken
		return displayName
	})
}
