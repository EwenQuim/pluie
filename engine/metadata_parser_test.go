package engine

import (
	"testing"

	"github.com/EwenQuim/pluie/model"
)

func TestParseWikiLinksInMetadata(t *testing.T) {
	// Create test notes
	notes := []model.Note{
		{
			Title: "Public Test Note",
			Slug:  "public-test-note",
		},
		{
			Title: "Private Test Note",
			Slug:  "private-test-note",
		},
	}

	// Build tree
	tree := BuildTree(notes)

	tests := []struct {
		name     string
		metadata map[string]any
		expected map[string]any
	}{
		{
			name: "string with wikilink",
			metadata: map[string]any{
				"description": "This references [[Public Test Note]]",
			},
			expected: map[string]any{
				"description": "This references [Public Test Note](/public-test-note)",
			},
		},
		{
			name: "string with wikilink and display name",
			metadata: map[string]any{
				"description": "This references [[Public Test Note|a public note]]",
			},
			expected: map[string]any{
				"description": "This references [a public note](/public-test-note)",
			},
		},
		{
			name: "list with wikilinks",
			metadata: map[string]any{
				"related_notes": []interface{}{
					"[[Public Test Note]]",
					"[[Private Test Note]]",
					"regular text",
				},
			},
			expected: map[string]any{
				"related_notes": []interface{}{
					"[Public Test Note](/public-test-note)",
					"[Private Test Note](/private-test-note)",
					"regular text",
				},
			},
		},
		{
			name: "nested object with wikilinks",
			metadata: map[string]any{
				"nested": map[string]interface{}{
					"reference":   "[[Public Test Note]]",
					"description": "A nested reference to [[Private Test Note|private note]]",
				},
			},
			expected: map[string]any{
				"nested": map[string]interface{}{
					"reference":   "[Public Test Note](/public-test-note)",
					"description": "A nested reference to [private note](/private-test-note)",
				},
			},
		},
		{
			name: "mixed types without wikilinks",
			metadata: map[string]any{
				"title":   "Test Title",
				"publish": true,
				"count":   42,
			},
			expected: map[string]any{
				"title":   "Test Title",
				"publish": true,
				"count":   42,
			},
		},
		{
			name: "broken wikilink",
			metadata: map[string]any{
				"description": "This references [[Nonexistent Note]]",
			},
			expected: map[string]any{
				"description": "This references Nonexistent Note",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseWikiLinksInMetadata(tt.metadata, tree)

			// Compare the results
			if !compareMetadata(result, tt.expected) {
				t.Errorf("ParseWikiLinksInMetadata() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Helper function to compare metadata maps
func compareMetadata(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}

	for key, valueA := range a {
		valueB, exists := b[key]
		if !exists {
			return false
		}

		if !compareValues(valueA, valueB) {
			return false
		}
	}

	return true
}

// Helper function to compare values recursively
func compareValues(a, b any) bool {
	switch va := a.(type) {
	case string:
		vb, ok := b.(string)
		return ok && va == vb
	case bool:
		vb, ok := b.(bool)
		return ok && va == vb
	case int:
		vb, ok := b.(int)
		return ok && va == vb
	case []interface{}:
		vb, ok := b.([]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for i, itemA := range va {
			if !compareValues(itemA, vb[i]) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		vb, ok := b.(map[string]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for key, valueA := range va {
			valueB, exists := vb[key]
			if !exists || !compareValues(valueA, valueB) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
