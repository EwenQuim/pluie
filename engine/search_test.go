package engine

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/EwenQuim/pluie/model"
)

func TestSearchNotesByFilename(t *testing.T) {
	// Create test notes with various filename patterns
	testNotes := []model.Note{
		{
			Title: "hello-world",
			Slug:  "hello-world",
		},
		{
			Title: "Getting Started",
			Slug:  "getting-started",
		},
		{
			Title: "API Documentation",
			Slug:  "api-documentation",
		},
		{
			Title: "nested-file",
			Slug:  "folder/nested-file",
		},
		{
			Title: "another-nested",
			Slug:  "deep/folder/structure/another-nested",
		},
		{
			Title: "README",
			Slug:  "projects/myproject/README",
		},
		{
			Title: "config",
			Slug:  "config/app/config",
		},
		{
			Title: "World",
			Slug:  "hello",
		},
	}

	tests := []struct {
		name        string
		notes       []model.Note
		searchQuery string
		expected    []model.Note
	}{
		{
			name:        "empty search query returns all notes",
			notes:       testNotes,
			searchQuery: "",
			expected:    testNotes,
		},
		{
			name:        "search by exact title match",
			notes:       testNotes,
			searchQuery: "hello-world",
			expected: []model.Note{
				{Title: "hello-world", Slug: "hello-world"},
			},
		},
		{
			name:        "search by partial title match",
			notes:       testNotes,
			searchQuery: "hello",
			expected: []model.Note{
				{Title: "hello-world", Slug: "hello-world"},
				{Title: "World", Slug: "hello"},
			},
		},
		{
			name:        "case insensitive search",
			notes:       testNotes,
			searchQuery: "HELLO",
			expected: []model.Note{
				{Title: "hello-world", Slug: "hello-world"},
				{Title: "World", Slug: "hello"},
			},
		},
		{
			name:        "search by title with spaces",
			notes:       testNotes,
			searchQuery: "Getting",
			expected: []model.Note{
				{Title: "Getting Started", Slug: "getting-started"},
			},
		},
		{
			name:        "search by filename in nested path (should match filename only)",
			notes:       testNotes,
			searchQuery: "nested-file",
			expected: []model.Note{
				{Title: "nested-file", Slug: "folder/nested-file"},
			},
		},
		{
			name:        "search matches folder names in slug",
			notes:       testNotes,
			searchQuery: "folder",
			expected: []model.Note{
				{Title: "nested-file", Slug: "folder/nested-file"},
				{Title: "another-nested", Slug: "deep/folder/structure/another-nested"},
			},
		},
		{
			name:        "search matches deep folder names in slug",
			notes:       testNotes,
			searchQuery: "deep",
			expected: []model.Note{
				{Title: "another-nested", Slug: "deep/folder/structure/another-nested"},
			},
		},
		{
			name:        "search by filename in deeply nested structure",
			notes:       testNotes,
			searchQuery: "another-nested",
			expected: []model.Note{
				{Title: "another-nested", Slug: "deep/folder/structure/another-nested"},
			},
		},
		{
			name:        "search by common filename in different paths",
			notes:       testNotes,
			searchQuery: "config",
			expected: []model.Note{
				{Title: "config", Slug: "config/app/config"},
			},
		},
		{
			name:        "search with partial match",
			notes:       testNotes,
			searchQuery: "API",
			expected: []model.Note{
				{Title: "API Documentation", Slug: "api-documentation"},
			},
		},
		{
			name:        "search with no matches",
			notes:       testNotes,
			searchQuery: "nonexistent",
			expected:    []model.Note{},
		},
		{
			name:        "search with multiple matches",
			notes:       testNotes,
			searchQuery: "nested",
			expected: []model.Note{
				{Title: "nested-file", Slug: "folder/nested-file"},
				{Title: "another-nested", Slug: "deep/folder/structure/another-nested"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SearchNotesByFilename(tt.notes, tt.searchQuery)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("SearchNotesByFilename() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestSearchNotesByFilename_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		notes       []model.Note
		searchQuery string
		expected    []model.Note
	}{
		{
			name:        "empty notes slice",
			notes:       []model.Note{},
			searchQuery: "test",
			expected:    []model.Note{},
		},
		{
			name: "notes with empty titles and slugs",
			notes: []model.Note{
				{Title: "", Slug: ""},
				{Title: "valid", Slug: "valid"},
			},
			searchQuery: "valid",
			expected: []model.Note{
				{Title: "valid", Slug: "valid"},
			},
		},
		{
			name: "search query with special characters",
			notes: []model.Note{
				{Title: "file-with-dashes", Slug: "file-with-dashes"},
				{Title: "file_with_underscores", Slug: "file_with_underscores"},
				{Title: "file.with.dots", Slug: "file.with.dots"},
			},
			searchQuery: "with-dashes",
			expected: []model.Note{
				{Title: "file-with-dashes", Slug: "file-with-dashes"},
			},
		},
		{
			name: "slug without path separator",
			notes: []model.Note{
				{Title: "simple", Slug: "simple"},
			},
			searchQuery: "simple",
			expected: []model.Note{
				{Title: "simple", Slug: "simple"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SearchNotesByFilename(tt.notes, tt.searchQuery)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("SearchNotesByFilename() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Benchmark test to ensure the search function performs well
func BenchmarkSearchNotesByFilename(b *testing.B) {
	// Create a large set of test notes
	notes := make([]model.Note, 1000)
	for i := 0; i < 1000; i++ {
		notes[i] = model.Note{
			Title: fmt.Sprintf("note-%d", i),
			Slug:  fmt.Sprintf("folder/subfolder/note-%d", i),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SearchNotesByFilename(notes, "note-500")
	}
}
