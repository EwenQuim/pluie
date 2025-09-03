package engine

import (
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		options  SlugifyOptions
		expected string
	}{
		// Note slugification tests (with path preservation)
		{
			name:  "Note: Empty slug",
			input: "",
			options: SlugifyOptions{
				PreserveSlashes: true,
				URLEncode:       true,
				RemoveExtension: true,
				TrimSlashes:     true,
			},
			expected: "",
		},
		{
			name:  "Note: Simple title",
			input: "Hello World",
			options: SlugifyOptions{
				PreserveSlashes: true,
				URLEncode:       true,
				RemoveExtension: true,
				TrimSlashes:     true,
			},
			expected: "hello-world",
		},
		{
			name:  "Note: Title with .md extension",
			input: "hello-world.md",
			options: SlugifyOptions{
				PreserveSlashes: true,
				URLEncode:       true,
				RemoveExtension: true,
				TrimSlashes:     true,
			},
			expected: "hello-world",
		},
		{
			name:  "Note: Path with spaces",
			input: "folder/hello world.md",
			options: SlugifyOptions{
				PreserveSlashes: true,
				URLEncode:       true,
				RemoveExtension: true,
				TrimSlashes:     true,
			},
			expected: "folder/hello-world",
		},
		{
			name:  "Note: Multiple consecutive dashes",
			input: "hello---world--test",
			options: SlugifyOptions{
				PreserveSlashes: true,
				URLEncode:       true,
				RemoveExtension: true,
				TrimSlashes:     true,
			},
			expected: "hello-world-test",
		},
		{
			name:  "Note: Leading and trailing slashes/dashes",
			input: "/-hello-world-/",
			options: SlugifyOptions{
				PreserveSlashes: true,
				URLEncode:       true,
				RemoveExtension: true,
				TrimSlashes:     true,
			},
			expected: "hello-world",
		},
		{
			name:  "Note: Special characters with URL encoding",
			input: "Articles/Hello World & More",
			options: SlugifyOptions{
				PreserveSlashes: true,
				URLEncode:       true,
				RemoveExtension: true,
				TrimSlashes:     true,
			},
			expected: "articles/hello-world-&-more",
		},
		{
			name:  "Note: Only special characters",
			input: "!@#$%^&*()",
			options: SlugifyOptions{
				PreserveSlashes: true,
				URLEncode:       true,
				RemoveExtension: true,
				TrimSlashes:     true,
			},
			expected: "%21@%23$%25%5E&%2A%28%29",
		},
		{
			name:  "Note: Deep nested path",
			input: "Deep/Nested/Path Structure",
			options: SlugifyOptions{
				PreserveSlashes: true,
				URLEncode:       true,
				RemoveExtension: true,
				TrimSlashes:     true,
			},
			expected: "deep/nested/path-structure",
		},

		// Heading slugification tests (simple)
		{
			name:  "Heading: Basic text",
			input: "Hello World",
			options: SlugifyOptions{
				PreserveSlashes: false,
				URLEncode:       false,
				RemoveExtension: false,
				TrimSlashes:     false,
			},
			expected: "hello-world",
		},
		{
			name:  "Heading: Text with punctuation",
			input: "Hello, World!",
			options: SlugifyOptions{
				PreserveSlashes: false,
				URLEncode:       false,
				RemoveExtension: false,
				TrimSlashes:     false,
			},
			expected: "hello-world",
		},
		{
			name:  "Heading: Text with numbers",
			input: "Chapter 1: Introduction",
			options: SlugifyOptions{
				PreserveSlashes: false,
				URLEncode:       false,
				RemoveExtension: false,
				TrimSlashes:     false,
			},
			expected: "chapter-1-introduction",
		},
		{
			name:  "Heading: Text with special characters",
			input: "API & SDK Guide",
			options: SlugifyOptions{
				PreserveSlashes: false,
				URLEncode:       false,
				RemoveExtension: false,
				TrimSlashes:     false,
			},
			expected: "api-sdk-guide",
		},
		{
			name:  "Heading: Multiple spaces",
			input: "Multiple   Spaces   Here",
			options: SlugifyOptions{
				PreserveSlashes: false,
				URLEncode:       false,
				RemoveExtension: false,
				TrimSlashes:     false,
			},
			expected: "multiple-spaces-here",
		},
		{
			name:  "Heading: Leading/trailing spaces",
			input: "  Trimmed Text  ",
			options: SlugifyOptions{
				PreserveSlashes: false,
				URLEncode:       false,
				RemoveExtension: false,
				TrimSlashes:     false,
			},
			expected: "trimmed-text",
		},
		{
			name:  "Heading: Only special characters",
			input: "!@#$%^&*()",
			options: SlugifyOptions{
				PreserveSlashes: false,
				URLEncode:       false,
				RemoveExtension: false,
				TrimSlashes:     false,
			},
			expected: "",
		},
		{
			name:  "Heading: Mixed case with underscores",
			input: "Some_Mixed_Case_Text",
			options: SlugifyOptions{
				PreserveSlashes: false,
				URLEncode:       false,
				RemoveExtension: false,
				TrimSlashes:     false,
			},
			expected: "some-mixed-case-text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slugify(tt.input, tt.options)
			if result != tt.expected {
				t.Errorf("Slugify(%q, %+v) = %q, want %q", tt.input, tt.options, result, tt.expected)
			}
		})
	}
}

func TestSlugifyNote(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Simple title",
			input:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "Title with .md extension",
			input:    "hello-world.md",
			expected: "hello-world",
		},
		{
			name:     "Path with spaces",
			input:    "folder/hello world.md",
			expected: "folder/hello-world",
		},
		{
			name:     "Multiple consecutive dashes",
			input:    "hello---world--test",
			expected: "hello-world-test",
		},
		{
			name:     "Leading and trailing slashes/dashes",
			input:    "/-hello-world-/",
			expected: "hello-world",
		},
		{
			name:     "Special characters",
			input:    "Articles/Hello World & More",
			expected: "articles/hello-world-&-more",
		},
		{
			name:     "Deep nested path",
			input:    "Deep/Nested/Path Structure",
			expected: "deep/nested/path-structure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SlugifyNote(tt.input)
			if result != tt.expected {
				t.Errorf("SlugifyNote(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSlugifyHeading(t *testing.T) {
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
			name:     "Multiple spaces",
			input:    "Multiple   Spaces   Here",
			expected: "multiple-spaces-here",
		},
		{
			name:     "Leading/trailing spaces",
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
			result := SlugifyHeading(tt.input)
			if result != tt.expected {
				t.Errorf("SlugifyHeading(%q) = %q, want %q", tt.input, result, tt.expected)
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

func TestSlugifyNoteWithCaseLogic(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		existingSlug string
		expected     string
	}{
		{
			name:         "Empty slug uses title with case preserved",
			text:         "Hello World",
			existingSlug: "",
			expected:     "Hello-World",
		},
		{
			name:         "Existing slug is converted to lowercase",
			text:         "custom-slug",
			existingSlug: "custom-slug",
			expected:     "custom-slug",
		},
		{
			name:         "Existing slug with mixed case is lowercased",
			text:         "Custom-Slug",
			existingSlug: "Custom-Slug",
			expected:     "custom-slug",
		},
		{
			name:         "Title with path preserves case",
			text:         "Articles/Hello World & More",
			existingSlug: "",
			expected:     "Articles/Hello-World-&-More",
		},
		{
			name:         "Existing slug with path is lowercased",
			text:         "Articles/Hello-World",
			existingSlug: "Articles/Hello-World",
			expected:     "articles/hello-world",
		},
		{
			name:         "Title with .md extension",
			text:         "Hello World.md",
			existingSlug: "",
			expected:     "Hello-World",
		},
		{
			name:         "Existing slug with .md extension",
			text:         "hello-world.md",
			existingSlug: "hello-world.md",
			expected:     "hello-world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SlugifyNoteWithCaseLogic(tt.text, tt.existingSlug)
			if result != tt.expected {
				t.Errorf("SlugifyNoteWithCaseLogic(%q, %q) = %q, want %q", tt.text, tt.existingSlug, result, tt.expected)
			}
		})
	}
}

func TestDefaultOptions(t *testing.T) {
	t.Run("DefaultNoteSlugOptions", func(t *testing.T) {
		options := DefaultNoteSlugOptions()
		expected := SlugifyOptions{
			PreserveSlashes: true,
			URLEncode:       true,
			RemoveExtension: true,
			TrimSlashes:     true,
		}
		if options != expected {
			t.Errorf("DefaultNoteSlugOptions() = %+v, want %+v", options, expected)
		}
	})

	t.Run("DefaultHeadingSlugOptions", func(t *testing.T) {
		options := DefaultHeadingSlugOptions()
		expected := SlugifyOptions{
			PreserveSlashes: false,
			URLEncode:       false,
			RemoveExtension: false,
			TrimSlashes:     false,
		}
		if options != expected {
			t.Errorf("DefaultHeadingSlugOptions() = %+v, want %+v", options, expected)
		}
	})
}
