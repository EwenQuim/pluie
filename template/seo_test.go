package template

import (
	"reflect"
	"testing"

	"github.com/EwenQuim/pluie/model"
)

func TestComputeSEOData(t *testing.T) {
	tests := []struct {
		name                string
		note                *model.Note
		baseSiteTitle       string
		baseSiteDescription string
		expected            SEOData
	}{
		{
			name:                "nil note should return default values",
			note:                nil,
			baseSiteTitle:       "My Site",
			baseSiteDescription: "My Site Description",
			expected: SEOData{
				PageTitle:    "My Site",
				Description:  "My Site Description",
				CanonicalURL: "",
				OGType:       "website",
				Keywords:     nil,
				AuthorMeta:   nil,
				DateMeta:     nil,
				ModifiedMeta: nil,
			},
		},
		{
			name: "note with title should format page title correctly",
			note: &model.Note{
				Title: "Test Article",
				Slug:  "test-article",
			},
			baseSiteTitle:       "My Site",
			baseSiteDescription: "My Site Description",
			expected: SEOData{
				PageTitle:    "Test Article | My Site",
				Description:  "My Site Description",
				CanonicalURL: "/test-article",
				OGType:       "article",
				Keywords:     nil,
				AuthorMeta:   nil,
				DateMeta:     nil,
				ModifiedMeta: nil,
			},
		},
		{
			name: "note with metadata description should use it",
			note: &model.Note{
				Title: "Test Article",
				Slug:  "test-article",
				Metadata: map[string]any{
					"description": "Custom description from metadata",
					"author":      "John Doe",
					"date":        "2023-01-01",
					"modified":    "2023-01-02",
					"tags":        []interface{}{"go", "testing", "seo"},
				},
			},
			baseSiteTitle:       "My Site",
			baseSiteDescription: "My Site Description",
			expected: SEOData{
				PageTitle:    "Test Article | My Site",
				Description:  "Custom description from metadata",
				CanonicalURL: "/test-article",
				OGType:       "article",
				Keywords:     []string{"go", "testing", "seo"},
				AuthorMeta:   "John Doe",
				DateMeta:     "2023-01-01",
				ModifiedMeta: "2023-01-02",
			},
		},
		{
			name: "note with content but no metadata description should extract from content",
			note: &model.Note{
				Title:   "Test Article",
				Slug:    "test-article",
				Content: "# This is a test article\n\nThis is the content of the article with **bold** and *italic* text. It should be cleaned up for the description.",
			},
			baseSiteTitle:       "My Site",
			baseSiteDescription: "My Site Description",
			expected: SEOData{
				PageTitle:    "Test Article | My Site",
				Description:  "This is a test article\n\nThis is the content of the article with bold and italic text. It should be cleaned up for the description.",
				CanonicalURL: "/test-article",
				OGType:       "article",
				Keywords:     nil,
				AuthorMeta:   nil,
				DateMeta:     nil,
				ModifiedMeta: nil,
			},
		},
		{
			name: "note with long content should truncate description",
			note: &model.Note{
				Title:   "Test Article",
				Slug:    "test-article",
				Content: "This is a very long content that should be truncated when used as a description. It contains more than 160 characters and should be cut off at exactly 157 characters with ellipsis added at the end to indicate truncation.",
			},
			baseSiteTitle:       "My Site",
			baseSiteDescription: "My Site Description",
			expected: SEOData{
				PageTitle:    "Test Article | My Site",
				Description:  "This is a very long content that should be truncated when used as a description. It contains more than 160 characters and should be cut off at exactly 157 ch...",
				CanonicalURL: "/test-article",
				OGType:       "article",
				Keywords:     nil,
				AuthorMeta:   nil,
				DateMeta:     nil,
				ModifiedMeta: nil,
			},
		},
		{
			name: "note with empty title should use base site title",
			note: &model.Note{
				Title: "",
				Slug:  "test-article",
			},
			baseSiteTitle:       "My Site",
			baseSiteDescription: "My Site Description",
			expected: SEOData{
				PageTitle:    "My Site",
				Description:  "My Site Description",
				CanonicalURL: "/test-article",
				OGType:       "article",
				Keywords:     nil,
				AuthorMeta:   nil,
				DateMeta:     nil,
				ModifiedMeta: nil,
			},
		},
		{
			name: "note with empty description metadata should fall back to content",
			note: &model.Note{
				Title:   "Test Article",
				Slug:    "test-article",
				Content: "Content description fallback",
				Metadata: map[string]any{
					"description": "",
				},
			},
			baseSiteTitle:       "My Site",
			baseSiteDescription: "My Site Description",
			expected: SEOData{
				PageTitle:    "Test Article | My Site",
				Description:  "Content description fallback",
				CanonicalURL: "/test-article",
				OGType:       "article",
				Keywords:     nil,
				AuthorMeta:   nil,
				DateMeta:     nil,
				ModifiedMeta: nil,
			},
		},
		{
			name: "note with non-string tags should be ignored",
			note: &model.Note{
				Title: "Test Article",
				Slug:  "test-article",
				Metadata: map[string]any{
					"tags": []interface{}{"valid-tag", 123, true, "another-valid-tag"},
				},
			},
			baseSiteTitle:       "My Site",
			baseSiteDescription: "My Site Description",
			expected: SEOData{
				PageTitle:    "Test Article | My Site",
				Description:  "My Site Description",
				CanonicalURL: "/test-article",
				OGType:       "article",
				Keywords:     []string{"valid-tag", "another-valid-tag"},
				AuthorMeta:   nil,
				DateMeta:     nil,
				ModifiedMeta: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeSEOData(tt.note, tt.baseSiteTitle, tt.baseSiteDescription)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ComputeSEOData() = %+v, expected %+v", result, tt.expected)
			}
		})
	}
}

func TestComputeSEOData_EdgeCases(t *testing.T) {
	t.Run("note with nil metadata should not panic", func(t *testing.T) {
		note := &model.Note{
			Title:    "Test",
			Slug:     "test",
			Metadata: nil,
		}
		result := ComputeSEOData(note, "Site", "Description")
		if result.PageTitle != "Test | Site" {
			t.Errorf("Expected page title 'Test | Site', got '%s'", result.PageTitle)
		}
	})

	t.Run("note with empty slug should not set canonical URL", func(t *testing.T) {
		note := &model.Note{
			Title: "Test",
			Slug:  "",
		}
		result := ComputeSEOData(note, "Site", "Description")
		if result.CanonicalURL != "" {
			t.Errorf("Expected empty canonical URL, got '%s'", result.CanonicalURL)
		}
	})

	t.Run("note with non-slice tags should not panic", func(t *testing.T) {
		note := &model.Note{
			Title: "Test",
			Slug:  "test",
			Metadata: map[string]any{
				"tags": "not-a-slice",
			},
		}
		result := ComputeSEOData(note, "Site", "Description")
		if len(result.Keywords) != 0 {
			t.Errorf("Expected no keywords, got %v", result.Keywords)
		}
	})

	t.Run("note with non-string description should be ignored", func(t *testing.T) {
		note := &model.Note{
			Title: "Test",
			Slug:  "test",
			Metadata: map[string]any{
				"description": 123,
			},
		}
		result := ComputeSEOData(note, "Site", "Site Description")
		if result.Description != "Site Description" {
			t.Errorf("Expected site description fallback, got '%s'", result.Description)
		}
	})
}
