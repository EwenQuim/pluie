package template

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	g "github.com/maragudk/gomponents"
)

// mapMap creates nodes from a map
func mapMap[T any](ts map[string]T, cb func(k string, v T) g.Node) []g.Node {
	var nodes []g.Node
	for k, v := range ts {
		nodes = append(nodes, cb(k, v))
	}
	return nodes
}

func TestRemoveObsidianCallouts(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Remove basic NOTE callout",
			input: `# Heading
> [!NOTE]
> This is a note callout
Some regular content`,
			expected: `# Heading

> This is a note callout
Some regular content`,
		},
		{
			name: "Remove WARNING callout",
			input: `> [!WARNING] This is a warning
Regular content continues`,
			expected: `
Regular content continues`,
		},
		{
			name: "Remove multiple different callouts",
			input: `> [!INFO] Information here
Some content
> [!TIP] A helpful tip
More content
> [!SUCCESS] Success message`,
			expected: `
Some content

More content
`,
		},
		{
			name: "Remove callouts with various spacing",
			input: `>[!NOTE] No space after >
> [!WARNING] One space after >
>  [!TIP] Two spaces after >
>   [!INFO] Three spaces after >`,
			expected: "\n\n\n",
		},
		{
			name: "Keep regular blockquotes",
			input: `> This is a regular blockquote
> [!NOTE] This is a callout
> Another regular blockquote line`,
			expected: `> This is a regular blockquote

> Another regular blockquote line`,
		},
		{
			name: "Remove callouts with hyphens in keywords",
			input: `> [!custom-note] Custom callout with hyphen
> [!multi-word-callout] Another custom one`,
			expected: `
`,
		},
		{
			name:     "Handle empty content",
			input:    "",
			expected: "",
		},
		{
			name: "Handle content with no callouts",
			input: `# Regular Markdown
This is normal content
> Regular blockquote
More content`,
			expected: `# Regular Markdown
This is normal content
> Regular blockquote
More content`,
		},
		{
			name: "Remove callouts mixed with other content",
			input: `## Introduction
> [!NOTE] Important note
This is regular text.
> [!WARNING] Be careful here
### Subsection
> [!TIP] Pro tip
Final paragraph.`,
			expected: `## Introduction

This is regular text.

### Subsection

Final paragraph.`,
		},
		{
			name: "Handle callouts at start and end of content",
			input: `> [!INFO] Starting callout
Middle content
> [!SUCCESS] Ending callout`,
			expected: `
Middle content
`,
		},
		{
			name: "Remove callouts with uppercase and lowercase keywords",
			input: `> [!NOTE] Uppercase
> [!note] Lowercase (should match - regex includes lowercase)
> [!Warning] Mixed case (should match - regex includes mixed case)
> [!TIP] Another uppercase`,
			expected: "\n\n\n",
		},
		{
			name: "Handle callouts with additional text after keyword",
			input: `> [!NOTE] This has additional text
> [!WARNING] So does this one
> [!TIP]- This has a dash
> [!INFO]+ This has a plus`,
			expected: "\n\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeObsidianCallouts(tt.input)
			if result != tt.expected {
				t.Errorf("removeObsidianCallouts() = %q, want %q", result, tt.expected)

				// Show line-by-line comparison for easier debugging
				resultLines := strings.Split(result, "\n")
				expectedLines := strings.Split(tt.expected, "\n")

				t.Logf("Result lines (%d):", len(resultLines))
				for i, line := range resultLines {
					t.Logf("  [%d]: %q", i, line)
				}

				t.Logf("Expected lines (%d):", len(expectedLines))
				for i, line := range expectedLines {
					t.Logf("  [%d]: %q", i, line)
				}
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic text",
			input:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "Text with punctuation",
			input:    "Hello, World!",
			expected: "hello-world",
		},
		{
			name:     "Text with numbers",
			input:    "Chapter 1: Introduction",
			expected: "chapter-1-introduction",
		},
		{
			name:     "Text with special characters",
			input:    "API & SDK Guide",
			expected: "api-sdk-guide",
		},
		{
			name:     "Text with multiple spaces",
			input:    "Multiple   Spaces   Here",
			expected: "multiple-spaces-here",
		},
		{
			name:     "Text with leading/trailing spaces",
			input:    "  Trimmed Text  ",
			expected: "trimmed-text",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only special characters",
			input:    "!@#$%^&*()",
			expected: "",
		},
		{
			name:     "Mixed case with underscores",
			input:    "Some_Mixed_Case_Text",
			expected: "some-mixed-case-text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.SlugifyHeading(tt.input)
			if result != tt.expected {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMapMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]int
		expected int // number of nodes created
	}{
		{
			name:     "Empty map",
			input:    map[string]int{},
			expected: 0,
		},
		{
			name:     "Single item",
			input:    map[string]int{"key1": 1},
			expected: 1,
		},
		{
			name:     "Multiple items",
			input:    map[string]int{"key1": 1, "key2": 2, "key3": 3},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapMap(tt.input, func(k string, v int) g.Node {
				return g.Text(k + ":" + fmt.Sprintf("%d", v))
			})
			if len(result) != tt.expected {
				t.Errorf("MapMap() returned %d nodes, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestMapMapSorted(t *testing.T) {
	input := map[string]int{
		"zebra":  1,
		"apple":  2,
		"Banana": 3, // Test case-insensitive sorting
	}

	result := MapMapSorted(input, func(k string, v int) g.Node {
		return g.Text(k)
	})

	if len(result) != 3 {
		t.Errorf("MapMapSorted() returned %d nodes, want 3", len(result))
		return
	}

	// We can't easily test the exact order without accessing the internal text,
	// but we can test that all items are present
	if len(result) != len(input) {
		t.Errorf("MapMapSorted() returned %d nodes, want %d", len(result), len(input))
	}
}

func TestSortKeysCaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "Single item",
			input:    []string{"apple"},
			expected: []string{"apple"},
		},
		{
			name:     "Already sorted",
			input:    []string{"apple", "banana", "cherry"},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:     "Reverse order",
			input:    []string{"cherry", "banana", "apple"},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:     "Mixed case",
			input:    []string{"Zebra", "apple", "Banana"},
			expected: []string{"apple", "Banana", "Zebra"},
		},
		{
			name:     "Same case insensitive",
			input:    []string{"Apple", "apple", "APPLE"},
			expected: []string{"Apple", "apple", "APPLE"}, // Stable sort maintains original order for equal elements
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying the test data
			input := make([]string, len(tt.input))
			copy(input, tt.input)

			sortKeysCaseInsensitive(input)

			if len(input) != len(tt.expected) {
				t.Errorf("sortKeysCaseInsensitive() resulted in %d items, want %d", len(input), len(tt.expected))
				return
			}

			for i, item := range input {
				if item != tt.expected[i] {
					t.Errorf("sortKeysCaseInsensitive()[%d] = %q, want %q", i, item, tt.expected[i])
				}
			}
		})
	}
}

func TestCountNotesInTree(t *testing.T) {
	tests := []struct {
		name     string
		tree     *engine.TreeNode
		expected int
	}{
		{
			name:     "Nil tree",
			tree:     nil,
			expected: 0,
		},
		{
			name: "Empty folder",
			tree: &engine.TreeNode{
				Name:     "root",
				IsFolder: true,
				Children: []*engine.TreeNode{},
			},
			expected: 0,
		},
		{
			name: "Single note",
			tree: &engine.TreeNode{
				Name:     "root",
				IsFolder: false,
				Note:     &model.Note{Title: "Test Note"},
			},
			expected: 1,
		},
		{
			name: "Folder with notes",
			tree: &engine.TreeNode{
				Name:     "root",
				IsFolder: true,
				Children: []*engine.TreeNode{
					{
						Name:     "note1",
						IsFolder: false,
						Note:     &model.Note{Title: "Note 1"},
					},
					{
						Name:     "note2",
						IsFolder: false,
						Note:     &model.Note{Title: "Note 2"},
					},
				},
			},
			expected: 2,
		},
		{
			name: "Nested folders with notes",
			tree: &engine.TreeNode{
				Name:     "root",
				IsFolder: true,
				Children: []*engine.TreeNode{
					{
						Name:     "note1",
						IsFolder: false,
						Note:     &model.Note{Title: "Note 1"},
					},
					{
						Name:     "subfolder",
						IsFolder: true,
						Children: []*engine.TreeNode{
							{
								Name:     "note2",
								IsFolder: false,
								Note:     &model.Note{Title: "Note 2"},
							},
						},
					},
				},
			},
			expected: 2,
		},
		{
			name: "Folder node without note should not count",
			tree: &engine.TreeNode{
				Name:     "root",
				IsFolder: false,
				Note:     nil, // No note attached
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countNotesInTree(tt.tree)
			if result != tt.expected {
				t.Errorf("countNotesInTree() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestRenderTOC(t *testing.T) {
	tests := []struct {
		name     string
		input    []TOCItem
		expected int // number of nodes expected
	}{
		{
			name:     "Empty TOC",
			input:    []TOCItem{},
			expected: 1, // Should return "No headings found" message
		},
		{
			name: "Single heading",
			input: []TOCItem{
				{ID: "heading-1", Text: "Heading 1", Level: 1},
			},
			expected: 1,
		},
		{
			name: "Multiple headings",
			input: []TOCItem{
				{ID: "heading-1", Text: "Heading 1", Level: 1},
				{ID: "heading-2", Text: "Heading 2", Level: 2},
				{ID: "heading-3", Text: "Heading 3", Level: 3},
			},
			expected: 3,
		},
		{
			name: "Deep headings",
			input: []TOCItem{
				{ID: "h1", Text: "H1", Level: 1},
				{ID: "h2", Text: "H2", Level: 2},
				{ID: "h3", Text: "H3", Level: 3},
				{ID: "h4", Text: "H4", Level: 4},
				{ID: "h5", Text: "H5", Level: 5},
				{ID: "h6", Text: "H6", Level: 6},
			},
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderTOC(tt.input)
			if len(result) != tt.expected {
				t.Errorf("renderTOC() returned %d nodes, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestRenderYamlValue(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{
			name:  "Boolean true",
			input: true,
		},
		{
			name:  "Boolean false",
			input: false,
		},
		{
			name:  "String value",
			input: "test string",
		},
		{
			name:  "Empty string",
			input: "",
		},
		{
			name:  "URL string",
			input: "https://example.com",
		},
		{
			name:  "Email string",
			input: "test@example.com",
		},
		{
			name:  "Date string",
			input: "2023-01-01",
		},
		{
			name:  "String with markdown links",
			input: "[Link](https://example.com)",
		},
		{
			name:  "Integer",
			input: 42,
		},
		{
			name:  "Float",
			input: 3.14,
		},
		{
			name:  "Empty array",
			input: []any{},
		},
		{
			name:  "Array with strings",
			input: []any{"tag1", "tag2", "tag3"},
		},
		{
			name:  "Array with mixed types",
			input: []any{"string", 123, true},
		},
		{
			name:  "Array with markdown links",
			input: []any{"[Link](https://example.com)", "regular tag"},
		},
		{
			name:  "Empty map",
			input: map[string]any{},
		},
		{
			name:  "Map with values",
			input: map[string]any{"key1": "value1", "key2": 42},
		},
		{
			name:  "Unknown type",
			input: struct{ Name string }{Name: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderYamlValue(tt.input)
			if result == nil {
				t.Errorf("renderYamlValue() returned nil for input %v", tt.input)
			}
		})
	}
}

func TestRenderYamlProperty(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value any
	}{
		{
			name:  "String property",
			key:   "title",
			value: "Test Title",
		},
		{
			name:  "Boolean property",
			key:   "published",
			value: true,
		},
		{
			name:  "Array property",
			key:   "tags",
			value: []any{"tag1", "tag2"},
		},
		{
			name:  "Map property",
			key:   "metadata",
			value: map[string]any{"author": "John Doe"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderYamlProperty(tt.key, tt.value)
			if result == nil {
				t.Errorf("renderYamlProperty() returned nil for key %q, value %v", tt.key, tt.value)
			}
		})
	}
}

func TestRenderTreeNode(t *testing.T) {
	rs := Resource{}

	tests := []struct {
		name        string
		node        *engine.TreeNode
		currentSlug string
	}{
		{
			name:        "Nil node",
			node:        nil,
			currentSlug: "",
		},
		{
			name: "Folder node",
			node: &engine.TreeNode{
				Name:     "folder",
				Path:     "folder",
				IsFolder: true,
				Children: []*engine.TreeNode{},
			},
			currentSlug: "",
		},
		{
			name: "Note node",
			node: &engine.TreeNode{
				Name:     "note",
				Path:     "note",
				IsFolder: false,
				Note:     &model.Note{Title: "Test Note", Slug: "test-note"},
			},
			currentSlug: "test-note",
		},
		{
			name: "Note node not current",
			node: &engine.TreeNode{
				Name:     "note",
				Path:     "note",
				IsFolder: false,
				Note:     &model.Note{Title: "Test Note", Slug: "test-note"},
			},
			currentSlug: "other-note",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rs.renderTreeNode(tt.node, tt.currentSlug)
			if result == nil {
				t.Errorf("renderTreeNode() returned nil for node %v", tt.node)
			}
		})
	}
}

func TestRenderFolderNode(t *testing.T) {
	rs := Resource{}

	tests := []struct {
		name        string
		node        *engine.TreeNode
		currentSlug string
	}{
		{
			name: "Empty folder",
			node: &engine.TreeNode{
				Name:     "folder",
				Path:     "folder",
				IsFolder: true,
				Children: []*engine.TreeNode{},
			},
			currentSlug: "",
		},
		{
			name: "Folder with children",
			node: &engine.TreeNode{
				Name:     "folder",
				Path:     "folder",
				IsFolder: true,
				IsOpen:   true,
				Children: []*engine.TreeNode{
					{
						Name:     "child",
						Path:     "folder/child",
						IsFolder: false,
						Note:     &model.Note{Title: "Child Note", Slug: "child"},
					},
				},
			},
			currentSlug: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rs.renderFolderNode(tt.node, tt.currentSlug)
			if result == nil {
				t.Errorf("renderFolderNode() returned nil for node %v", tt.node)
			}
		})
	}
}

func TestRenderNoteNode(t *testing.T) {
	rs := Resource{}

	tests := []struct {
		name        string
		node        *engine.TreeNode
		currentSlug string
	}{
		{
			name: "Active note",
			node: &engine.TreeNode{
				Name:     "note",
				Path:     "note",
				IsFolder: false,
				Note:     &model.Note{Title: "Test Note", Slug: "test-note"},
			},
			currentSlug: "test-note",
		},
		{
			name: "Inactive note",
			node: &engine.TreeNode{
				Name:     "note",
				Path:     "note",
				IsFolder: false,
				Note:     &model.Note{Title: "Test Note", Slug: "test-note"},
			},
			currentSlug: "other-note",
		},
		{
			name: "Note without note object",
			node: &engine.TreeNode{
				Name:     "note",
				Path:     "note",
				IsFolder: false,
				Note:     nil,
			},
			currentSlug: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rs.renderNoteNode(tt.node, tt.currentSlug)
			if result == nil {
				t.Errorf("renderNoteNode() returned nil for node %v", tt.node)
			}
		})
	}
}

func TestRenderChevronIcon(t *testing.T) {
	rs := Resource{}

	tests := []struct {
		name string
		node *engine.TreeNode
	}{
		{
			name: "Open folder",
			node: &engine.TreeNode{
				Name:     "folder",
				Path:     "folder",
				IsFolder: true,
				IsOpen:   true,
			},
		},
		{
			name: "Closed folder",
			node: &engine.TreeNode{
				Name:     "folder",
				Path:     "folder",
				IsFolder: true,
				IsOpen:   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rs.renderChevronIcon(tt.node)
			if result == nil {
				t.Errorf("renderChevronIcon() returned nil for node %v", tt.node)
			}
		})
	}
}

func TestRenderFolderChildren(t *testing.T) {
	rs := Resource{}

	tests := []struct {
		name        string
		node        *engine.TreeNode
		currentSlug string
	}{
		{
			name: "No children",
			node: &engine.TreeNode{
				Name:     "folder",
				Path:     "folder",
				IsFolder: true,
				Children: []*engine.TreeNode{},
			},
			currentSlug: "",
		},
		{
			name: "Open folder with children",
			node: &engine.TreeNode{
				Name:     "folder",
				Path:     "folder",
				IsFolder: true,
				IsOpen:   true,
				Children: []*engine.TreeNode{
					{
						Name:     "child",
						Path:     "folder/child",
						IsFolder: false,
						Note:     &model.Note{Title: "Child Note", Slug: "child"},
					},
				},
			},
			currentSlug: "",
		},
		{
			name: "Closed folder with children",
			node: &engine.TreeNode{
				Name:     "folder",
				Path:     "folder",
				IsFolder: true,
				IsOpen:   false,
				Children: []*engine.TreeNode{
					{
						Name:     "child",
						Path:     "folder/child",
						IsFolder: false,
						Note:     &model.Note{Title: "Child Note", Slug: "child"},
					},
				},
			},
			currentSlug: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rs.renderFolderChildren(tt.node, tt.currentSlug)
			if result == nil {
				t.Errorf("renderFolderChildren() returned nil for node %v", tt.node)
			}
		})
	}
}

func TestNoteWithList(t *testing.T) {
	// Create a simple tree for testing
	tree := &engine.TreeNode{
		Name:     "root",
		IsFolder: true,
		Children: []*engine.TreeNode{
			{
				Name:     "test-note",
				Path:     "test-note",
				IsFolder: false,
				Note:     &model.Note{Title: "Test Note", Slug: "test-note"},
			},
		},
	}

	notesMap := make(map[string]model.Note)
	notesService := engine.NewNotesService(&notesMap, tree, nil)
	rs := Resource{}

	tests := []struct {
		name        string
		note        *model.Note
		searchQuery string
		env         map[string]string
	}{
		{
			name:        "Nil note",
			note:        nil,
			searchQuery: "",
			env: map[string]string{
				"SITE_TITLE":       "Test Site",
				"SITE_DESCRIPTION": "Test Description",
			},
		},
		{
			name: "Valid note",
			note: &model.Note{
				Title:   "Test Note",
				Slug:    "test-note",
				Content: "# Test Content\n\nThis is test content.",
				Metadata: map[string]any{
					"description": "Test description",
					"tags":        []any{"tag1", "tag2"},
				},
				ReferencedBy: []model.NoteReference{
					{Title: "Reference Note", Slug: "reference-note"},
				},
			},
			searchQuery: "",
			env: map[string]string{
				"SITE_TITLE":            "Test Site",
				"SITE_DESCRIPTION":      "Test Description",
				"SITE_ICON":             "/test-icon.png",
				"HIDE_YAML_FRONTMATTER": "false",
			},
		},
		{
			name: "Note with search query",
			note: &model.Note{
				Title:   "Test Note",
				Slug:    "test-note",
				Content: "# Test Content\n\nThis is test content.",
			},
			searchQuery: "test",
			env: map[string]string{
				"SITE_TITLE":       "Test Site",
				"SITE_DESCRIPTION": "Test Description",
			},
		},
		{
			name: "Note with hidden YAML frontmatter",
			note: &model.Note{
				Title:   "Test Note",
				Slug:    "test-note",
				Content: "# Test Content\n\nThis is test content.",
				Metadata: map[string]any{
					"description": "Test description",
				},
			},
			searchQuery: "",
			env: map[string]string{
				"SITE_TITLE":            "Test Site",
				"SITE_DESCRIPTION":      "Test Description",
				"HIDE_YAML_FRONTMATTER": "true",
			},
		},
		{
			name: "Note with default environment",
			note: &model.Note{
				Title:   "Test Note",
				Slug:    "test-note",
				Content: "# Test Content\n\nThis is test content.",
			},
			searchQuery: "",
			env:         map[string]string{}, // No environment variables
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

			result, err := rs.NoteWithList(notesService, tt.note, tt.searchQuery)
			if err != nil {
				t.Errorf("NoteWithList() returned error: %v", err)
			}
			if result == nil {
				t.Errorf("NoteWithList() returned nil result")
			}
		})
	}
}

func TestExtractHeadings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TOCItem
	}{
		{
			name: "Basic headings",
			input: `# Heading 1
## Heading 2
### Heading 3`,
			expected: []TOCItem{
				{ID: "heading-1", Text: "Heading 1", Level: 1},
				{ID: "heading-2", Text: "Heading 2", Level: 2},
				{ID: "heading-3", Text: "Heading 3", Level: 3},
			},
		},
		{
			name: "Headings with duplicate text",
			input: `# Introduction
## Introduction
### Introduction`,
			expected: []TOCItem{
				{ID: "introduction", Text: "Introduction", Level: 1},
				{ID: "introduction-2", Text: "Introduction", Level: 2},
				{ID: "introduction-3", Text: "Introduction", Level: 3},
			},
		},
		{
			name: "Mixed heading levels",
			input: `# Main Title
### Subsection
## Section
#### Deep Section`,
			expected: []TOCItem{
				{ID: "main-title", Text: "Main Title", Level: 1},
				{ID: "subsection", Text: "Subsection", Level: 3},
				{ID: "section", Text: "Section", Level: 2},
				{ID: "deep-section", Text: "Deep Section", Level: 4},
			},
		},
		{
			name:     "No headings",
			input:    "Just regular text with no headings",
			expected: []TOCItem{},
		},
		{
			name:     "Empty content",
			input:    "",
			expected: []TOCItem{},
		},
		{
			name: "Headings with special characters",
			input: `# API & SDK Guide
## Getting Started!
### FAQ?`,
			expected: []TOCItem{
				{ID: "api-sdk-guide", Text: "API & SDK Guide", Level: 1},
				{ID: "getting-started", Text: "Getting Started!", Level: 2},
				{ID: "faq", Text: "FAQ?", Level: 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractHeadings(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("extractHeadings() returned %d items, want %d", len(result), len(tt.expected))
				return
			}

			for i, item := range result {
				expected := tt.expected[i]
				if item.ID != expected.ID || item.Text != expected.Text || item.Level != expected.Level {
					t.Errorf("extractHeadings()[%d] = {ID: %q, Text: %q, Level: %d}, want {ID: %q, Text: %q, Level: %d}",
						i, item.ID, item.Text, item.Level, expected.ID, expected.Text, expected.Level)
				}
			}
		})
	}
}
