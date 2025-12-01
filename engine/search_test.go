package engine

import (
	"fmt"
	"reflect"
	"strings"
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

func TestSearchNotesByHeadings(t *testing.T) {
	testNotes := []model.Note{
		{
			Title: "Go Programming",
			Slug:  "go-programming",
			Content: `# Introduction to Go
This is an introduction to the Go programming language.

## Getting Started with Go
Learn how to install and set up Go on your system.

### Go Installation
Step by step installation guide.

## Advanced Go Topics
Deep dive into advanced concepts.

#### This is H4 and should be ignored
This should not be found.`,
		},
		{
			Title: "Python Guide",
			Slug:  "python-guide",
			Content: `# Python Basics
Learn Python programming from scratch.

## Data Structures in Python
Arrays, lists, and dictionaries.

### Python Lists
Working with lists in Python.`,
		},
		{
			Title: "JavaScript Tutorial",
			Slug:  "javascript-tutorial",
			Content: `# JavaScript Introduction
Modern JavaScript development.

## Getting Started
Set up your development environment.`,
		},
	}

	tests := []struct {
		name          string
		notes         []model.Note
		searchQuery   string
		limit         int
		expectedCount int
		expectedFirst string // Expected first heading match
		expectedLevel int    // Expected level of first match
	}{
		{
			name:          "empty search query returns nil",
			notes:         testNotes,
			searchQuery:   "",
			limit:         5,
			expectedCount: 0,
		},
		{
			name:          "search for 'Go' matches multiple headings",
			notes:         testNotes,
			searchQuery:   "Go",
			limit:         10,
			expectedCount: 4, // "Introduction to Go", "Getting Started with Go", "Go Installation", "Advanced Go Topics"
			expectedFirst: "Introduction to Go",
			expectedLevel: 1,
		},
		{
			name:          "search for 'Python' matches headings",
			notes:         testNotes,
			searchQuery:   "Python",
			limit:         10,
			expectedCount: 3,
			expectedFirst: "Python Basics",
			expectedLevel: 1,
		},
		{
			name:          "case insensitive search",
			notes:         testNotes,
			searchQuery:   "javascript",
			limit:         10,
			expectedCount: 1,
			expectedFirst: "JavaScript Introduction",
			expectedLevel: 1,
		},
		{
			name:          "limit results to top 2",
			notes:         testNotes,
			searchQuery:   "Go",
			limit:         2,
			expectedCount: 2,
			expectedFirst: "Introduction to Go", // H1 exact should rank highest
			expectedLevel: 1,
		},
		{
			name:          "search for 'Getting Started'",
			notes:         testNotes,
			searchQuery:   "Getting Started",
			limit:         10,
			expectedCount: 2, // "Getting Started", "Getting Started with Go"
			expectedFirst: "Getting Started", // Exact match should rank higher
			expectedLevel: 2,
		},
		{
			name:          "H4 headings are ignored",
			notes:         testNotes,
			searchQuery:   "H4",
			limit:         10,
			expectedCount: 0,
		},
		{
			name:          "no matches returns empty",
			notes:         testNotes,
			searchQuery:   "Rust",
			limit:         10,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SearchNotesByHeadings(tt.notes, tt.searchQuery, tt.limit)

			if len(result) != tt.expectedCount {
				t.Errorf("SearchNotesByHeadings() returned %d results, expected %d", len(result), tt.expectedCount)
				for i, match := range result {
					t.Logf("  [%d] %s (level %d, score %d)", i, match.Heading, match.Level, match.Score)
				}
			}

			if tt.expectedCount > 0 && len(result) > 0 {
				if result[0].Heading != tt.expectedFirst {
					t.Errorf("First heading = %q, expected %q", result[0].Heading, tt.expectedFirst)
				}
				if tt.expectedLevel > 0 && result[0].Level != tt.expectedLevel {
					t.Errorf("First heading level = %d, expected %d", result[0].Level, tt.expectedLevel)
				}
			}
		})
	}
}

func TestExtractHeading(t *testing.T) {
	tests := []struct {
		name          string
		line          string
		expectedText  string
		expectedLevel int
	}{
		{
			name:          "H1 heading",
			line:          "# Introduction",
			expectedText:  "Introduction",
			expectedLevel: 1,
		},
		{
			name:          "H2 heading",
			line:          "## Getting Started",
			expectedText:  "Getting Started",
			expectedLevel: 2,
		},
		{
			name:          "H3 heading",
			line:          "### Installation",
			expectedText:  "Installation",
			expectedLevel: 3,
		},
		{
			name:          "H4 heading",
			line:          "#### Details",
			expectedText:  "Details",
			expectedLevel: 4,
		},
		{
			name:          "heading with extra spaces",
			line:          "##   Multiple   Spaces",
			expectedText:  "Multiple   Spaces",
			expectedLevel: 2,
		},
		{
			name:          "heading with leading spaces",
			line:          "   ## Indented Heading",
			expectedText:  "Indented Heading",
			expectedLevel: 2,
		},
		{
			name:          "invalid heading no space",
			line:          "##NoSpace",
			expectedText:  "",
			expectedLevel: 0,
		},
		{
			name:          "not a heading",
			line:          "Just regular text",
			expectedText:  "",
			expectedLevel: 0,
		},
		{
			name:          "empty line",
			line:          "",
			expectedText:  "",
			expectedLevel: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, level := extractHeading(tt.line)

			if text != tt.expectedText {
				t.Errorf("extractHeading() text = %q, expected %q", text, tt.expectedText)
			}
			if level != tt.expectedLevel {
				t.Errorf("extractHeading() level = %d, expected %d", level, tt.expectedLevel)
			}
		})
	}
}

func TestCalculateHeadingScore(t *testing.T) {
	tests := []struct {
		name     string
		heading  string
		query    string
		level    int
		expected int
	}{
		{
			name:     "exact match H1",
			heading:  "Go Programming",
			query:    "Go Programming",
			level:    1,
			expected: 13, // 10 (exact) + 3 (H1 bonus)
		},
		{
			name:     "exact match H2",
			heading:  "Getting Started",
			query:    "Getting Started",
			level:    2,
			expected: 12, // 10 (exact) + 2 (H2 bonus)
		},
		{
			name:     "exact match H3",
			heading:  "Installation",
			query:    "Installation",
			level:    3,
			expected: 11, // 10 (exact) + 1 (H3 bonus)
		},
		{
			name:     "starts with H1",
			heading:  "Go Programming Basics",
			query:    "Go",
			level:    1,
			expected: 8, // 5 (starts with) + 3 (H1 bonus)
		},
		{
			name:     "starts with H2",
			heading:  "Python Tutorial",
			query:    "Python",
			level:    2,
			expected: 7, // 5 (starts with) + 2 (H2 bonus)
		},
		{
			name:     "contains H1",
			heading:  "Advanced Go Topics",
			query:    "Go",
			level:    1,
			expected: 6, // 3 (contains) + 3 (H1 bonus)
		},
		{
			name:     "contains H3",
			heading:  "Working with Lists",
			query:    "Lists",
			level:    3,
			expected: 4, // 3 (contains) + 1 (H3 bonus)
		},
		{
			name:     "case insensitive exact match",
			heading:  "JavaScript",
			query:    "javascript",
			level:    1,
			expected: 13, // 10 (exact) + 3 (H1 bonus)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateHeadingScore(tt.heading, tt.query, tt.level)

			if score != tt.expected {
				t.Errorf("calculateHeadingScore(%q, %q, %d) = %d, expected %d",
					tt.heading, tt.query, tt.level, score, tt.expected)
			}
		})
	}
}

func TestExtractContext(t *testing.T) {
	lines := []string{
		"# Main Heading",
		"This is the first line after the heading.",
		"This is the second line with more content.",
		"",
		"This is after a blank line.",
		"## Another Heading",
		"This should not be included.",
	}

	tests := []struct {
		name      string
		lines     []string
		lineNum   int
		maxChars  int
		shouldContain string
	}{
		{
			name:     "extract context from H1",
			lines:    lines,
			lineNum:  0,
			maxChars: 75,
			shouldContain: "first line",
		},
		{
			name:     "extract limited context",
			lines:    lines,
			lineNum:  0,
			maxChars: 30,
			shouldContain: "first line",
		},
		{
			name:     "context stops at next heading",
			lines:    lines,
			lineNum:  0,
			maxChars: 200,
			shouldContain: "blank line",
		},
		{
			name:     "invalid line number returns empty",
			lines:    lines,
			lineNum:  -1,
			maxChars: 75,
			shouldContain: "",
		},
		{
			name:     "out of bounds returns empty",
			lines:    lines,
			lineNum:  100,
			maxChars: 75,
			shouldContain: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := extractContext(tt.lines, tt.lineNum, tt.maxChars)

			if tt.shouldContain != "" {
				if !strings.Contains(context, tt.shouldContain) {
					t.Errorf("extractContext() = %q, should contain %q", context, tt.shouldContain)
				}
			} else {
				if context != "" {
					t.Errorf("extractContext() = %q, expected empty string", context)
				}
			}

			// Check length constraint (allowing for "...")
			if len(context) > tt.maxChars+3 {
				t.Errorf("extractContext() length = %d, should be <= %d", len(context), tt.maxChars+3)
			}
		})
	}
}

func BenchmarkSearchNotesByHeadings(b *testing.B) {
	// Create test notes with headings
	notes := make([]model.Note, 100)
	for i := 0; i < 100; i++ {
		notes[i] = model.Note{
			Title: fmt.Sprintf("note-%d", i),
			Slug:  fmt.Sprintf("notes/note-%d", i),
			Content: fmt.Sprintf(`# Heading %d
Content for heading %d.

## Subheading %d
More content here.

### Details %d
Even more content.`, i, i, i, i),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SearchNotesByHeadings(notes, "Heading", 5)
	}
}
