package template

import (
	"testing"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/model"
)

func TestLayout(t *testing.T) {
	tests := []struct {
		name string
		note *model.Note
		cfg  *config.Config
	}{
		{
			name: "Layout with nil note",
			note: nil,
			cfg: &config.Config{
				SiteTitle:       "Test Site",
				SiteDescription: "Test Description",
				SiteIcon:        "/test-icon.png",
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
			cfg: &config.Config{
				SiteTitle:       "Test Site",
				SiteDescription: "Test Description",
				SiteIcon:        "/test-icon.png",
			},
		},
		{
			name: "Layout with default values",
			note: &model.Note{
				Title: "Test Note",
				Slug:  "test-note",
			},
			cfg: &config.Config{
				SiteTitle:       "Pluie",
				SiteIcon:        "/static/pluie.webp",
				SiteDescription: "",
			},
		},
		{
			name: "Layout with empty site icon",
			note: &model.Note{
				Title: "Test Note",
				Slug:  "test-note",
			},
			cfg: &config.Config{
				SiteTitle:       "Test Site",
				SiteDescription: "Test Description",
				SiteIcon:        "", // Empty icon
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := NewResource(tt.cfg)
			result := rs.Layout(tt.note)
			if result == nil {
				t.Errorf("Layout() returned nil for note %v", tt.note)
			}
		})
	}
}
