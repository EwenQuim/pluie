package engine

import (
	"regexp"
	"strings"

	"github.com/EwenQuim/pluie/model"
)

// TagIndex maps tag names to notes that contain them
type TagIndex map[string][]model.Note

// BuildTagIndex creates an index of all tags found in notes
// Tags can come from:
// 1. Metadata "tags" field (array of strings)
// 2. Free text with syntax #[A-Za-z/-]+
func BuildTagIndex(notes []model.Note) TagIndex {
	tagIndex := make(TagIndex)

	for _, note := range notes {
		tags := extractAllTags(note)
		for _, tag := range tags {
			// Normalize tag (lowercase, trim spaces)
			normalizedTag := strings.ToLower(strings.TrimSpace(tag))
			if normalizedTag != "" {
				tagIndex[normalizedTag] = append(tagIndex[normalizedTag], note)
			}
		}
	}

	return tagIndex
}

// extractAllTags extracts all tags from a note (metadata + free text)
func extractAllTags(note model.Note) []string {
	var allTags []string

	// Extract tags from metadata
	metadataTags := extractMetadataTags(note.Metadata)
	allTags = append(allTags, metadataTags...)

	// Extract tags from free text content
	freeTextTags := extractFreeTextTags(note.Content)
	allTags = append(allTags, freeTextTags...)

	return allTags
}

// extractMetadataTags extracts tags from the metadata "tags" field
func extractMetadataTags(metadata map[string]any) []string {
	var tags []string

	if metadata == nil {
		return tags
	}

	// Look for "tags" field in metadata
	if tagsValue, exists := metadata["tags"]; exists {
		switch v := tagsValue.(type) {
		case []interface{}:
			// Handle array of tags
			for _, tag := range v {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
		case string:
			// Handle single tag as string
			tags = append(tags, v)
		}
	}

	return tags
}

// extractFreeTextTags extracts hashtags from free text using regex #[A-Za-z/-]+
func extractFreeTextTags(content string) []string {
	var tags []string

	// Regular expression to match hashtags: #[A-Za-z/-]+
	// This matches # followed by one or more letters, forward slashes, or hyphens
	re := regexp.MustCompile(`#([A-Za-z/-]+)`)

	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			// match[1] contains the captured group (tag without #)
			tags = append(tags, match[1])
		}
	}

	return tags
}

// GetNotesWithTag returns all notes that contain the specified tag
func (tagIndex TagIndex) GetNotesWithTag(tag string) []model.Note {
	normalizedTag := strings.ToLower(strings.TrimSpace(tag))
	if notes, exists := tagIndex[normalizedTag]; exists {
		return notes
	}
	return []model.Note{}
}

// GetAllTags returns all unique tags in the index
func (tagIndex TagIndex) GetAllTags() []string {
	var tags []string
	for tag := range tagIndex {
		tags = append(tags, tag)
	}
	return tags
}

// GetTagsContaining returns all tags that contain the specified substring
func (tagIndex TagIndex) GetTagsContaining(substring string) []string {
	var matchingTags []string
	normalizedSubstring := strings.ToLower(strings.TrimSpace(substring))

	for tag := range tagIndex {
		if strings.Contains(tag, normalizedSubstring) {
			matchingTags = append(matchingTags, tag)
		}
	}

	return matchingTags
}
