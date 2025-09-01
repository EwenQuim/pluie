package model

import (
	"testing"
)

func TestNote_BuildSlug(t *testing.T) {
	tests := []struct {
		name     string
		note     Note
		expected string
	}{
		{
			name: "Empty slug uses title",
			note: Note{
				Title: "Hello World",
				Slug:  "",
			},
			expected: "Hello-World",
		},
		{
			name: "Existing slug is used",
			note: Note{
				Title: "Hello World",
				Slug:  "custom-slug",
			},
			expected: "custom-slug",
		},
		{
			name: "Slug with .md extension is trimmed",
			note: Note{
				Title: "Hello World",
				Slug:  "hello-world.md",
			},
			expected: "hello-world",
		},
		{
			name: "Spaces replaced with dashes",
			note: Note{
				Title: "Multiple Spaces Here",
				Slug:  "",
			},
			expected: "Multiple-Spaces-Here",
		},
		{
			name: "Multiple consecutive dashes cleaned up",
			note: Note{
				Title: "Hello World",
				Slug:  "hello---world--test",
			},
			expected: "hello-world-test",
		},
		{
			name: "Leading and trailing slashes/dashes removed",
			note: Note{
				Title: "Hello World",
				Slug:  "/-hello-world-/",
			},
			expected: "hello-world",
		},
		{
			name: "URL encoding with preserved forward slashes",
			note: Note{
				Title: "Hello World",
				Slug:  "folder/hello world",
			},
			expected: "folder/hello-world",
		},
		{
			name: "Complex path with spaces and special chars",
			note: Note{
				Title: "Articles/Hello World & More",
				Slug:  "",
			},
			expected: "Articles/Hello-World-&-More",
		},
		{
			name: "Empty title and slug",
			note: Note{
				Title: "",
				Slug:  "",
			},
			expected: "",
		},
		{
			name: "Only special characters",
			note: Note{
				Title: "!@#$%^&*()",
				Slug:  "",
			},
			expected: "%21@%23$%25%5E&%2A%28%29",
		},
		{
			name: "Path with multiple levels",
			note: Note{
				Title: "Deep/Nested/Path Structure",
				Slug:  "",
			},
			expected: "Deep/Nested/Path-Structure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.note.BuildSlug()
			if tt.note.Slug != tt.expected {
				t.Errorf("BuildSlug() = %q, want %q", tt.note.Slug, tt.expected)
			}
		})
	}
}

func TestCleanMultipleDashes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No dashes",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "Single dash",
			input:    "hello-world",
			expected: "hello-world",
		},
		{
			name:     "Multiple consecutive dashes",
			input:    "hello---world",
			expected: "hello-world",
		},
		{
			name:     "Mixed multiple dashes",
			input:    "hello--world---test----end",
			expected: "hello-world-test-end",
		},
		{
			name:     "Dashes at start and end",
			input:    "---hello-world---",
			expected: "-hello-world-",
		},
		{
			name:     "Only dashes",
			input:    "-----",
			expected: "-",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Single character",
			input:    "a",
			expected: "a",
		},
		{
			name:     "Single dash",
			input:    "-",
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanMultipleDashes(tt.input)
			if result != tt.expected {
				t.Errorf("cleanMultipleDashes(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNote_DetermineIsPublic(t *testing.T) {
	tests := []struct {
		name           string
		note           Note
		folderMetadata map[string]map[string]any
		expected       bool
	}{
		{
			name: "Note metadata publish true",
			note: Note{
				Slug: "test-note",
				Metadata: map[string]any{
					"publish": true,
				},
			},
			folderMetadata: map[string]map[string]any{},
			expected:       true,
		},
		{
			name: "Note metadata publish false",
			note: Note{
				Slug: "test-note",
				Metadata: map[string]any{
					"publish": false,
				},
			},
			folderMetadata: map[string]map[string]any{},
			expected:       false,
		},
		{
			name: "Note metadata publish non-boolean ignored",
			note: Note{
				Slug: "test-note",
				Metadata: map[string]any{
					"publish": "true", // string, not boolean
				},
			},
			folderMetadata: map[string]map[string]any{},
			expected:       false, // falls back to default
		},
		{
			name: "No note metadata, folder metadata publish true",
			note: Note{
				Slug:     "folder/test-note",
				Metadata: map[string]any{},
			},
			folderMetadata: map[string]map[string]any{
				"folder": {
					"publish": true,
				},
			},
			expected: true,
		},
		{
			name: "No note metadata, folder metadata publish false",
			note: Note{
				Slug:     "folder/test-note",
				Metadata: map[string]any{},
			},
			folderMetadata: map[string]map[string]any{
				"folder": {
					"publish": false,
				},
			},
			expected: false,
		},
		{
			name: "Note metadata overrides folder metadata",
			note: Note{
				Slug: "folder/test-note",
				Metadata: map[string]any{
					"publish": false,
				},
			},
			folderMetadata: map[string]map[string]any{
				"folder": {
					"publish": true,
				},
			},
			expected: false, // note metadata takes precedence
		},
		{
			name: "Deep nested folder path",
			note: Note{
				Slug:     "deep/nested/folder/test-note",
				Metadata: map[string]any{},
			},
			folderMetadata: map[string]map[string]any{
				"deep/nested/folder": {
					"publish": true,
				},
			},
			expected: true,
		},
		{
			name: "No metadata anywhere, defaults to false",
			note: Note{
				Slug:     "test-note",
				Metadata: map[string]any{},
			},
			folderMetadata: map[string]map[string]any{},
			expected:       false,
		},
		{
			name: "Root level note with no metadata",
			note: Note{
				Slug:     "test-note",
				Metadata: map[string]any{},
			},
			folderMetadata: map[string]map[string]any{},
			expected:       false,
		},
		{
			name: "Folder metadata non-boolean ignored",
			note: Note{
				Slug:     "folder/test-note",
				Metadata: map[string]any{},
			},
			folderMetadata: map[string]map[string]any{
				"folder": {
					"publish": "yes", // string, not boolean
				},
			},
			expected: false, // falls back to default
		},
		{
			name: "Folder exists but no publish metadata",
			note: Note{
				Slug:     "folder/test-note",
				Metadata: map[string]any{},
			},
			folderMetadata: map[string]map[string]any{
				"folder": {
					"other": "value",
				},
			},
			expected: false,
		},
		{
			name: "Slug with leading/trailing slashes",
			note: Note{
				Slug:     "/folder/test-note/",
				Metadata: map[string]any{},
			},
			folderMetadata: map[string]map[string]any{
				"folder": {
					"publish": true,
				},
			},
			expected: true,
		},
		{
			name: "Empty slug",
			note: Note{
				Slug:     "",
				Metadata: map[string]any{},
			},
			folderMetadata: map[string]map[string]any{},
			expected:       false,
		},
		{
			name: "Single level path (no folder)",
			note: Note{
				Slug:     "test-note",
				Metadata: map[string]any{},
			},
			folderMetadata: map[string]map[string]any{
				"some-folder": {
					"publish": true,
				},
			},
			expected: false, // no folder path to check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.note.DetermineIsPublic(tt.folderMetadata)
			if tt.note.IsPublic != tt.expected {
				t.Errorf("DetermineIsPublic() set IsPublic = %v, want %v", tt.note.IsPublic, tt.expected)
			}
		})
	}
}
