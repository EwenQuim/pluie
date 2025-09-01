package engine

import (
	"testing"

	"github.com/EwenQuim/pluie/model"
)

func TestBuildTagIndex(t *testing.T) {
	// Create test notes with various tag formats
	notes := []model.Note{
		{
			Title:   "Note 1",
			Slug:    "note-1",
			Content: "This is a note with #golang and #programming tags.",
			Metadata: map[string]any{
				"tags": []any{"web", "backend"},
			},
		},
		{
			Title:   "Note 2",
			Slug:    "note-2",
			Content: "Another note with #golang/web and #testing tags.",
			Metadata: map[string]any{
				"tags": "single-tag",
			},
		},
		{
			Title:   "Note 3",
			Slug:    "note-3",
			Content: "No hashtags here.",
			Metadata: map[string]any{
				"tags": []any{"golang", "tutorial"},
			},
		},
	}

	tagIndex := BuildTagIndex(notes)

	// Test that tags are properly indexed
	golangNotes := tagIndex.GetNotesWithTag("golang")
	if len(golangNotes) != 2 {
		t.Errorf("Expected 2 notes with 'golang' tag, got %d", len(golangNotes))
	}

	programmingNotes := tagIndex.GetNotesWithTag("programming")
	if len(programmingNotes) != 1 {
		t.Errorf("Expected 1 note with 'programming' tag, got %d", len(programmingNotes))
	}

	webNotes := tagIndex.GetNotesWithTag("web")
	if len(webNotes) != 1 {
		t.Errorf("Expected 1 note with 'web' tag, got %d", len(webNotes))
	}

	// Test single tag from metadata
	singleTagNotes := tagIndex.GetNotesWithTag("single-tag")
	if len(singleTagNotes) != 1 {
		t.Errorf("Expected 1 note with 'single-tag' tag, got %d", len(singleTagNotes))
	}

	// Test tag with slash
	golangWebNotes := tagIndex.GetNotesWithTag("golang/web")
	if len(golangWebNotes) != 1 {
		t.Errorf("Expected 1 note with 'golang/web' tag, got %d", len(golangWebNotes))
	}
}

func TestExtractFreeTextTags(t *testing.T) {
	testCases := []struct {
		content  string
		expected []string
	}{
		{
			content:  "This has #golang and #programming tags.",
			expected: []string{"golang", "programming"},
		},
		{
			content:  "Tags with slashes: #golang/web #testing/unit",
			expected: []string{"golang/web", "testing/unit"},
		},
		{
			content:  "Tags with hyphens: #web-dev #front-end",
			expected: []string{"web-dev", "front-end"},
		},
		{
			content:  "No tags here at all.",
			expected: []string{},
		},
		{
			content:  "Mixed case #GoLang #Programming",
			expected: []string{"GoLang", "Programming"},
		},
	}

	for _, tc := range testCases {
		result := extractFreeTextTags(tc.content)
		if len(result) != len(tc.expected) {
			t.Errorf("For content '%s', expected %d tags, got %d", tc.content, len(tc.expected), len(result))
			continue
		}

		for i, expected := range tc.expected {
			if result[i] != expected {
				t.Errorf("For content '%s', expected tag '%s', got '%s'", tc.content, expected, result[i])
			}
		}
	}
}

func TestExtractMetadataTags(t *testing.T) {
	testCases := []struct {
		metadata map[string]any
		expected []string
	}{
		{
			metadata: map[string]any{
				"tags": []any{"golang", "web", "backend"},
			},
			expected: []string{"golang", "web", "backend"},
		},
		{
			metadata: map[string]any{
				"tags": "single-tag",
			},
			expected: []string{"single-tag"},
		},
		{
			metadata: map[string]any{
				"other": "value",
			},
			expected: []string{},
		},
		{
			metadata: nil,
			expected: []string{},
		},
		{
			metadata: map[string]any{
				"tags": []any{},
			},
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		result := extractMetadataTags(tc.metadata)
		if len(result) != len(tc.expected) {
			t.Errorf("Expected %d tags, got %d", len(tc.expected), len(result))
			continue
		}

		for i, expected := range tc.expected {
			if result[i] != expected {
				t.Errorf("Expected tag '%s', got '%s'", expected, result[i])
			}
		}
	}
}

func TestGetTagsContaining(t *testing.T) {
	notes := []model.Note{
		{
			Title:   "Note 1",
			Content: "Content with #golang and #golang/web tags.",
		},
		{
			Title:   "Note 2",
			Content: "Content with #programming and #golang/backend tags.",
		},
	}

	tagIndex := BuildTagIndex(notes)

	// Test getting tags containing "golang"
	golangTags := tagIndex.GetTagsContaining("golang")
	expectedCount := 3 // golang, golang/web, golang/backend
	if len(golangTags) != expectedCount {
		t.Errorf("Expected %d tags containing 'golang', got %d: %v", expectedCount, len(golangTags), golangTags)
	}

	// Test getting tags containing "web"
	webTags := tagIndex.GetTagsContaining("web")
	expectedWebCount := 1 // golang/web
	if len(webTags) != expectedWebCount {
		t.Errorf("Expected %d tags containing 'web', got %d: %v", expectedWebCount, len(webTags), webTags)
	}
}

func TestParseHashtagLinks(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "Single hashtag",
			content:  "This is about #golang programming.",
			expected: "This is about [#golang](/-/tag/golang) programming.",
		},
		{
			name:     "Multiple hashtags",
			content:  "Learn #golang and #programming together.",
			expected: "Learn [#golang](/-/tag/golang) and [#programming](/-/tag/programming) together.",
		},
		{
			name:     "Hashtag with slash",
			content:  "Using #golang/web for development.",
			expected: "Using [#golang/web](/-/tag/golang/web) for development.",
		},
		{
			name:     "Hashtag with hyphen",
			content:  "Modern #web-development practices.",
			expected: "Modern [#web-development](/-/tag/web-development) practices.",
		},
		{
			name:     "No hashtags",
			content:  "This content has no tags.",
			expected: "This content has no tags.",
		},
		{
			name:     "Mixed content",
			content:  "# Header\n\nThis is about #golang and #web-dev.\n\n## Another section with #programming",
			expected: "# Header\n\nThis is about [#golang](/-/tag/golang) and [#web-dev](/-/tag/web-dev).\n\n## Another section with [#programming](/-/tag/programming)",
		},
		{
			name:     "False positive: hashtag in code block",
			content:  "Other than that, we use the `#MOC` tag for Maps of Content (MOC)",
			expected: "Other than that, we use the `#MOC` tag for Maps of Content (MOC)",
		},
		{
			name:     "False positive: hashtag in inline code",
			content:  "Use `#include <stdio.h>` in C programming.",
			expected: "Use `#include <stdio.h>` in C programming.",
		},
		{
			name:     "False positive: hashtag with character before",
			content:  "README#What is the Obsidian Hub",
			expected: "README#What is the Obsidian Hub",
		},
		{
			name:     "False positive: URL fragment",
			content:  "Visit https://example.com#section for more info.",
			expected: "Visit https://example.com#section for more info.",
		},
		{
			name:     "Mixed valid and invalid hashtags",
			content:  "This is a valid #tag but `#code` and URL#fragment should not be parsed.",
			expected: "This is a valid [#tag](/-/tag/tag) but `#code` and URL#fragment should not be parsed.",
		},
		{
			name:     "Code block with multiple lines",
			content:  "```\n#define MAX 100\n#include <stdio.h>\n```\nBut this #tag should work.",
			expected: "```\n#define MAX 100\n#include <stdio.h>\n```\nBut this [#tag](/-/tag/tag) should work.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseHashtagLinks(tc.content)
			if result != tc.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tc.expected, result)
			}
		})
	}
}

func TestParseTagLinksInMetadata(t *testing.T) {
	testCases := []struct {
		name     string
		metadata map[string]any
		expected map[string]any
	}{
		{
			name: "Single tag string",
			metadata: map[string]any{
				"tags":   "golang",
				"author": "test",
			},
			expected: map[string]any{
				"tags":   "[#golang](/-/tag/golang)",
				"author": "test",
			},
		},
		{
			name: "Array of tags",
			metadata: map[string]any{
				"tags": []any{"golang", "programming", "web-dev"},
			},
			expected: map[string]any{
				"tags": []any{
					"[#golang](/-/tag/golang)",
					"[#programming](/-/tag/programming)",
					"[#web-dev](/-/tag/web-dev)",
				},
			},
		},
		{
			name: "No tags field",
			metadata: map[string]any{
				"author": "test",
				"date":   "2023-01-01",
			},
			expected: map[string]any{
				"author": "test",
				"date":   "2023-01-01",
			},
		},
		{
			name:     "Nil metadata",
			metadata: nil,
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseTagLinksInMetadata(tc.metadata)

			// Compare the results
			if tc.expected == nil && result != nil {
				t.Errorf("Expected nil, got %v", result)
				return
			}
			if tc.expected != nil && result == nil {
				t.Errorf("Expected %v, got nil", tc.expected)
				return
			}
			if tc.expected == nil && result == nil {
				return
			}

			// Check each key-value pair
			for key, expectedValue := range tc.expected {
				actualValue, exists := result[key]
				if !exists {
					t.Errorf("Expected key '%s' not found in result", key)
					continue
				}

				// Handle array comparison
				if expectedArray, ok := expectedValue.([]any); ok {
					actualArray, ok := actualValue.([]any)
					if !ok {
						t.Errorf("Expected array for key '%s', got %T", key, actualValue)
						continue
					}
					if len(expectedArray) != len(actualArray) {
						t.Errorf("Array length mismatch for key '%s': expected %d, got %d", key, len(expectedArray), len(actualArray))
						continue
					}
					for i, expectedItem := range expectedArray {
						if expectedItem != actualArray[i] {
							t.Errorf("Array item mismatch for key '%s' at index %d: expected %v, got %v", key, i, expectedItem, actualArray[i])
						}
					}
				} else {
					// Handle simple value comparison
					if expectedValue != actualValue {
						t.Errorf("Value mismatch for key '%s': expected %v, got %v", key, expectedValue, actualValue)
					}
				}
			}
		})
	}
}
