package template

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
)

// buildTestTree creates a tree structure from a list of notes for testing
func buildTestTree(notes []model.Note) *engine.TreeNode {
	root := &engine.TreeNode{
		Name:     "Notes",
		Path:     "",
		IsFolder: true,
		Children: make([]*engine.TreeNode, 0),
		IsOpen:   true,
	}

	// Sort notes by slug for consistent ordering
	sortedNotes := make([]model.Note, len(notes))
	copy(sortedNotes, notes)
	sort.Slice(sortedNotes, func(i, j int) bool {
		return sortedNotes[i].Slug < sortedNotes[j].Slug
	})

	for _, note := range sortedNotes {
		// Create a copy of the note for the pointer
		noteCopy := note
		// For simplicity in tests, put all notes at root level
		noteNode := &engine.TreeNode{
			Name:     note.Title,
			Path:     note.Slug,
			IsFolder: false,
			Note:     &noteCopy,
			Children: make([]*engine.TreeNode, 0),
		}
		root.Children = append(root.Children, noteNode)
	}

	return root
}

func TestParseWikiLinks(t *testing.T) {
	// Create sample notes for testing
	notes := []model.Note{
		{Title: "Test Note", Slug: "test-note"},
		{Title: "Another Note", Slug: "another-note"},
		{Title: "Special Characters & Symbols", Slug: "special-characters-symbols"},
		{Title: "articles/Hello World", Slug: "articles/hello-world"},
	}

	// Build tree from notes
	tree := buildTestTree(notes)

	// Create a Resource with the tree
	rs := Resource{
		Tree: tree,
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Single wiki link with existing note",
			input:    "This is a [[Test Note]] in the content.",
			expected: "This is a [Test Note](/test-note) in the content.",
		},
		{
			name:     "Multiple wiki links",
			input:    "Check out [[Test Note]] and [[Another Note]] for more info.",
			expected: "Check out [Test Note](/test-note) and [Another Note](/another-note) for more info.",
		},
		{
			name:     "Wiki link with non-existent note",
			input:    "This [[Non Existent Note]] doesn't exist.",
			expected: "This Non Existent Note doesn't exist.",
		},
		{
			name:     "Mixed existing and non-existing links",
			input:    "See [[Test Note]] but [[Missing Note]] is broken.",
			expected: "See [Test Note](/test-note) but Missing Note is broken.",
		},
		{
			name:     "Wiki link with special characters",
			input:    "Reference [[Special Characters & Symbols]] here.",
			expected: "Reference [Special Characters & Symbols](/special-characters-symbols) here.",
		},
		{
			name:     "Wiki link with path-like title",
			input:    "Check [[articles/Hello World]] for details.",
			expected: "Check [articles/Hello World](/articles/hello-world) for details.",
		},
		{
			name:     "No wiki links",
			input:    "This is just regular text with no links.",
			expected: "This is just regular text with no links.",
		},
		{
			name:     "Empty content",
			input:    "",
			expected: "",
		},
		{
			name:     "Malformed wiki links (single brackets)",
			input:    "This [Test Note] is not a wiki link.",
			expected: "This [Test Note] is not a wiki link.",
		},
		{
			name:     "Malformed wiki links (three brackets)",
			input:    "This [[[Test Note]]] has too many brackets.",
			expected: "This [[[Test Note]]] has too many brackets.",
		},
		{
			name:     "Empty wiki link",
			input:    "This [[]] is an empty wiki link.",
			expected: "This  is an empty wiki link.",
		},
		{
			name:     "Wiki link at start of content",
			input:    "[[Test Note]] is at the beginning.",
			expected: "[Test Note](/test-note) is at the beginning.",
		},
		{
			name:     "Wiki link at end of content",
			input:    "This ends with [[Test Note]]",
			expected: "This ends with [Test Note](/test-note)",
		},
		{
			name:     "Multiple wiki links of same note",
			input:    "[[Test Note]] and [[Test Note]] appear twice.",
			expected: "[Test Note](/test-note) and [Test Note](/test-note) appear twice.",
		},
		{
			name:     "Wiki link with whitespace",
			input:    "This [[ Test Note ]] has extra spaces.",
			expected: "This  Test Note  has extra spaces.",
		},
		{
			name:     "Complex markdown with wiki links",
			input:    "# Header\n\nSee [[Test Note]] for more.\n\n- Item 1\n- [[Another Note]]",
			expected: "# Header\n\nSee [Test Note](/test-note) for more.\n\n- Item 1\n- [Another Note](/another-note)",
		},
		// Custom display name tests
		{
			name:     "Wiki link with custom display name",
			input:    "Check out [[Test Note|My Custom Link]].",
			expected: "Check out [My Custom Link](/test-note).",
		},
		{
			name:     "Wiki link with custom display name for non-existent note",
			input:    "See [[Missing Note|Custom Name]] for details.",
			expected: "See Custom Name for details.",
		},
		{
			name:     "Multiple wiki links with custom display names",
			input:    "Read [[Test Note|First Link]] and [[Another Note|Second Link]].",
			expected: "Read [First Link](/test-note) and [Second Link](/another-note).",
		},
		{
			name:     "Mixed regular and custom display name links",
			input:    "See [[Test Note]] and [[Another Note|Custom Name]].",
			expected: "See [Test Note](/test-note) and [Custom Name](/another-note).",
		},
		{
			name:     "Custom display name with special characters",
			input:    "Link to [[Special Characters & Symbols|Special Chars & More!]].",
			expected: "Link to [Special Chars & More!](/special-characters-symbols).",
		},
		{
			name:     "Custom display name with whitespace",
			input:    "Link [[Test Note| Custom Name With Spaces ]].",
			expected: "Link [Custom Name With Spaces](/test-note).",
		},
		{
			name:     "Empty custom display name",
			input:    "Link [[Test Note|]].",
			expected: "Link [](/test-note).",
		},
		{
			name:     "Empty page title with custom display name",
			input:    "Link [[|Custom Display]].",
			expected: "Link Custom Display.",
		},
		{
			name:     "Multiple pipes in wiki link",
			input:    "Link [[Test Note|Custom|Extra]] here.",
			expected: "Link [Custom|Extra](/test-note) here.",
		},
		{
			name:     "Wiki link with pipe but no display name",
			input:    "Link [[Test Note|]] here.",
			expected: "Link [](/test-note) here.",
		},
		{
			name:     "Custom display name with path-like title",
			input:    "See [[articles/Hello World|My Article]].",
			expected: "See [My Article](/articles/hello-world).",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.ParseWikiLinks(tt.input, rs.Tree)
			if result != tt.expected {
				t.Errorf("parseWikiLinks() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseWikiLinksWithEmptyNotes(t *testing.T) {
	// Test with empty tree
	rs := Resource{
		Tree: &engine.TreeNode{
			Name:     "Notes",
			Path:     "",
			IsFolder: true,
			Children: []*engine.TreeNode{},
			IsOpen:   true,
		},
	}

	input := "This [[Test Note]] should not be found."
	expected := "This Test Note should not be found."
	result := engine.ParseWikiLinks(input, rs.Tree)

	if result != expected {
		t.Errorf("parseWikiLinks() with empty notes = %q, want %q", result, expected)
	}
}

func TestParseWikiLinksPerformance(t *testing.T) {
	// Create a larger set of notes for performance testing
	notes := make([]model.Note, 100)
	for i := range 100 {
		notes[i] = model.Note{
			Title: fmt.Sprintf("Note %d", i),
			Slug:  fmt.Sprintf("note-%d", i),
		}
	}

	// Build tree from notes
	tree := buildTestTree(notes)

	rs := Resource{Tree: tree}

	// Create content with multiple wiki links
	content := "Start "
	for i := range 50 {
		content += fmt.Sprintf("[[Note %d]] ", i)
	}
	content += "End"

	// This should complete reasonably quickly
	result := engine.ParseWikiLinks(content, rs.Tree)

	// Verify that some transformations occurred
	if !strings.Contains(result, "[Note 0](/note-0)") {
		t.Error("Performance test failed: expected transformations not found")
	}
}

func BenchmarkParseWikiLinks(b *testing.B) {
	notes := []model.Note{
		{Title: "Test Note", Slug: "test-note"},
		{Title: "Another Note", Slug: "another-note"},
	}

	// Build tree from notes
	tree := buildTestTree(notes)

	rs := Resource{Tree: tree}

	content := "This is a [[Test Note]] with [[Another Note]] and some [[Missing Link]] content."

	for b.Loop() {
		engine.ParseWikiLinks(content, rs.Tree)
	}
}
