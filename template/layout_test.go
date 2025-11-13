package template

import (
	"os"
	"testing"

	"github.com/EwenQuim/pluie/model"
)

func TestLayout(t *testing.T) {
	rs := Resource{} // Layout doesn't need NotesService

	tests := []struct {
		name string
		note *model.Note
		env  map[string]string
	}{
		{
			name: "Layout with nil note",
			note: nil,
			env: map[string]string{
				"SITE_TITLE":       "Test Site",
				"SITE_DESCRIPTION": "Test Description",
				"SITE_ICON":        "/test-icon.png",
			},
		},
		{
			name: "Layout with note",
			note: &model.Note{
				Title: "Test Note",
				Slug:  "test-note",
				Metadata: map[string]any{
					"description": "Test note description",
					"author":      "Test Author",
					"tags":        []interface{}{"tag1", "tag2"},
				},
			},
			env: map[string]string{
				"SITE_TITLE":       "Test Site",
				"SITE_DESCRIPTION": "Test Description",
				"SITE_ICON":        "/test-icon.png",
			},
		},
		{
			name: "Layout with default values",
			note: &model.Note{
				Title: "Test Note",
				Slug:  "test-note",
			},
			env: map[string]string{}, // No environment variables set
		},
		{
			name: "Layout with empty site icon",
			note: &model.Note{
				Title: "Test Note",
				Slug:  "test-note",
			},
			env: map[string]string{
				"SITE_TITLE":       "Test Site",
				"SITE_DESCRIPTION": "Test Description",
				"SITE_ICON":        "", // Empty icon
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.env {
				os.Setenv(key, value)
			}
			defer func() {
				// Clean up environment variables
				for key := range tt.env {
					os.Unsetenv(key)
				}
			}()

			result := rs.Layout(tt.note)
			if result == nil {
				t.Errorf("Layout() returned nil for note %v", tt.note)
			}
		})
	}
}
