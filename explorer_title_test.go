package main

import (
	"testing"
)

func TestExtractH1TitleFromContent(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedTitle   string
		expectedContent string
	}{
		{
			name:            "H1 at beginning",
			content:         "# My Title\n\nThis is some content.",
			expectedTitle:   "My Title",
			expectedContent: "\nThis is some content.",
		},
		{
			name:            "H1 with extra spaces",
			content:         "#   My Title with Spaces   \n\nContent here.",
			expectedTitle:   "My Title with Spaces",
			expectedContent: "\nContent here.",
		},
		{
			name:            "H1 in middle of content",
			content:         "Some intro text\n\n# Main Title\n\nMore content.",
			expectedTitle:   "Main Title",
			expectedContent: "Some intro text\n\n\nMore content.",
		},
		{
			name:            "Multiple H1s - only first is extracted",
			content:         "# First Title\n\nSome content\n\n# Second Title\n\nMore content.",
			expectedTitle:   "First Title",
			expectedContent: "\nSome content\n\n# Second Title\n\nMore content.",
		},
		{
			name:            "No H1 present",
			content:         "## H2 Title\n\nSome content without H1.",
			expectedTitle:   "",
			expectedContent: "## H2 Title\n\nSome content without H1.",
		},
		{
			name:            "H1 with special characters",
			content:         "# Title with 🚀 Emoji & Special-Characters!\n\nContent follows.",
			expectedTitle:   "Title with 🚀 Emoji & Special-Characters!",
			expectedContent: "\nContent follows.",
		},
		{
			name:            "Empty content",
			content:         "",
			expectedTitle:   "",
			expectedContent: "",
		},
		{
			name:            "Only H1",
			content:         "# Just a Title",
			expectedTitle:   "Just a Title",
			expectedContent: "",
		},
		{
			name:            "H1 with indentation (should not match)",
			content:         "    # Indented Title\n\nContent here.",
			expectedTitle:   "",
			expectedContent: "    # Indented Title\n\nContent here.",
		},
		{
			name:            "H1 without space after hash",
			content:         "#NoSpace\n\nContent here.",
			expectedTitle:   "",
			expectedContent: "#NoSpace\n\nContent here.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, modifiedContent := extractH1TitleFromContent(tt.content)

			if title != tt.expectedTitle {
				t.Errorf("extractH1TitleFromContent() title = %q, want %q", title, tt.expectedTitle)
			}

			if modifiedContent != tt.expectedContent {
				t.Errorf("extractH1TitleFromContent() modifiedContent = %q, want %q", modifiedContent, tt.expectedContent)
			}
		})
	}
}

func TestExtractTitle(t *testing.T) {
	explorer := Explorer{BasePath: "/test"}

	tests := []struct {
		name            string
		fileName        string
		metadata        map[string]any
		content         string
		expectedTitle   string
		expectedContent string
	}{
		{
			name:            "H1 title takes precedence over frontmatter",
			fileName:        "test.md",
			metadata:        map[string]any{"title": "Frontmatter Title"},
			content:         "# H1 Title\n\nContent here.",
			expectedTitle:   "H1 Title",
			expectedContent: "\nContent here.",
		},
		{
			name:            "Frontmatter title when no H1",
			fileName:        "test.md",
			metadata:        map[string]any{"title": "Frontmatter Title"},
			content:         "## H2 Title\n\nContent here.",
			expectedTitle:   "Frontmatter Title",
			expectedContent: "## H2 Title\n\nContent here.",
		},
		{
			name:            "Filename fallback when no H1 or frontmatter",
			fileName:        "my-note.md",
			metadata:        map[string]any{},
			content:         "Just some content without title.",
			expectedTitle:   "my-note",
			expectedContent: "Just some content without title.",
		},
		{
			name:            "Empty frontmatter title falls back to H1",
			fileName:        "test.md",
			metadata:        map[string]any{"title": ""},
			content:         "# H1 Title\n\nContent here.",
			expectedTitle:   "H1 Title",
			expectedContent: "\nContent here.",
		},
		{
			name:            "Non-string frontmatter title falls back to H1",
			fileName:        "test.md",
			metadata:        map[string]any{"title": 123},
			content:         "# H1 Title\n\nContent here.",
			expectedTitle:   "H1 Title",
			expectedContent: "\nContent here.",
		},
		{
			name:            "No content provided uses frontmatter",
			fileName:        "test.md",
			metadata:        map[string]any{"title": "Frontmatter Title"},
			content:         "",
			expectedTitle:   "Frontmatter Title",
			expectedContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := tt.content
			title := explorer.extractTitle(tt.fileName, tt.metadata, &content)

			if title != tt.expectedTitle {
				t.Errorf("extractTitle() title = %q, want %q", title, tt.expectedTitle)
			}

			if content != tt.expectedContent {
				t.Errorf("extractTitle() modified content = %q, want %q", content, tt.expectedContent)
			}
		})
	}
}

func TestExtractTitleWithNilContent(t *testing.T) {
	explorer := Explorer{BasePath: "/test"}

	tests := []struct {
		name          string
		fileName      string
		metadata      map[string]any
		expectedTitle string
	}{
		{
			name:          "Frontmatter title with nil content",
			fileName:      "test.md",
			metadata:      map[string]any{"title": "Frontmatter Title"},
			expectedTitle: "Frontmatter Title",
		},
		{
			name:          "Filename fallback with nil content",
			fileName:      "my-note.md",
			metadata:      map[string]any{},
			expectedTitle: "my-note",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title := explorer.extractTitle(tt.fileName, tt.metadata, nil)

			if title != tt.expectedTitle {
				t.Errorf("extractTitle() with nil content = %q, want %q", title, tt.expectedTitle)
			}
		})
	}
}
