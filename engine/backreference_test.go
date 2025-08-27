package engine

import (
	"reflect"
	"sort"
	"testing"

	"github.com/EwenQuim/pluie/model"
)

func TestBuildBackreferences(t *testing.T) {
	tests := []struct {
		name     string
		notes    []model.Note
		expected []model.Note
	}{
		{
			name: "simple wikilink reference",
			notes: []model.Note{
				{
					Title:   "Note A",
					Slug:    "note-a",
					Content: "This references [[Note B]]",
				},
				{
					Title:   "Note B",
					Slug:    "note-b",
					Content: "This is note B",
				},
			},
			expected: []model.Note{
				{
					Title:        "Note A",
					Slug:         "note-a",
					Content:      "This references [[Note B]]",
					ReferencedBy: []model.NoteReference{},
				},
				{
					Title:   "Note B",
					Slug:    "note-b",
					Content: "This is note B",
					ReferencedBy: []model.NoteReference{
						{Slug: "note-a", Title: "Note A"},
					},
				},
			},
		},
		{
			name: "wikilink with custom display name",
			notes: []model.Note{
				{
					Title:   "Source Note",
					Slug:    "source-note",
					Content: "This references [[Target Note|custom display]]",
				},
				{
					Title:   "Target Note",
					Slug:    "target-note",
					Content: "This is the target",
				},
			},
			expected: []model.Note{
				{
					Title:        "Source Note",
					Slug:         "source-note",
					Content:      "This references [[Target Note|custom display]]",
					ReferencedBy: []model.NoteReference{},
				},
				{
					Title:   "Target Note",
					Slug:    "target-note",
					Content: "This is the target",
					ReferencedBy: []model.NoteReference{
						{Slug: "source-note", Title: "Source Note"},
					},
				},
			},
		},
		{
			name: "multiple references to same note",
			notes: []model.Note{
				{
					Title:   "Note A",
					Slug:    "note-a",
					Content: "References [[Target]] and [[Target|again]]",
				},
				{
					Title:   "Note B",
					Slug:    "note-b",
					Content: "Also references [[Target]]",
				},
				{
					Title:   "Target",
					Slug:    "target",
					Content: "I am referenced",
				},
			},
			expected: []model.Note{
				{
					Title:        "Note A",
					Slug:         "note-a",
					Content:      "References [[Target]] and [[Target|again]]",
					ReferencedBy: []model.NoteReference{},
				},
				{
					Title:        "Note B",
					Slug:         "note-b",
					Content:      "Also references [[Target]]",
					ReferencedBy: []model.NoteReference{},
				},
				{
					Title:   "Target",
					Slug:    "target",
					Content: "I am referenced",
					ReferencedBy: []model.NoteReference{
						{Slug: "note-a", Title: "Note A"},
						{Slug: "note-b", Title: "Note B"},
					},
				},
			},
		},
		{
			name: "circular references",
			notes: []model.Note{
				{
					Title:   "Note A",
					Slug:    "note-a",
					Content: "References [[Note B]]",
				},
				{
					Title:   "Note B",
					Slug:    "note-b",
					Content: "References [[Note A]]",
				},
			},
			expected: []model.Note{
				{
					Title:   "Note A",
					Slug:    "note-a",
					Content: "References [[Note B]]",
					ReferencedBy: []model.NoteReference{
						{Slug: "note-b", Title: "Note B"},
					},
				},
				{
					Title:   "Note B",
					Slug:    "note-b",
					Content: "References [[Note A]]",
					ReferencedBy: []model.NoteReference{
						{Slug: "note-a", Title: "Note A"},
					},
				},
			},
		},
		{
			name: "no references",
			notes: []model.Note{
				{
					Title:   "Note A",
					Slug:    "note-a",
					Content: "No wikilinks here",
				},
				{
					Title:   "Note B",
					Slug:    "note-b",
					Content: "Also no wikilinks",
				},
			},
			expected: []model.Note{
				{
					Title:        "Note A",
					Slug:         "note-a",
					Content:      "No wikilinks here",
					ReferencedBy: []model.NoteReference{},
				},
				{
					Title:        "Note B",
					Slug:         "note-b",
					Content:      "Also no wikilinks",
					ReferencedBy: []model.NoteReference{},
				},
			},
		},
		{
			name: "broken wikilinks (non-existent notes)",
			notes: []model.Note{
				{
					Title:   "Note A",
					Slug:    "note-a",
					Content: "References [[Non-existent Note]]",
				},
				{
					Title:   "Note B",
					Slug:    "note-b",
					Content: "This exists",
				},
			},
			expected: []model.Note{
				{
					Title:        "Note A",
					Slug:         "note-a",
					Content:      "References [[Non-existent Note]]",
					ReferencedBy: []model.NoteReference{},
				},
				{
					Title:        "Note B",
					Slug:         "note-b",
					Content:      "This exists",
					ReferencedBy: []model.NoteReference{},
				},
			},
		},
		{
			name: "empty wikilinks should be ignored",
			notes: []model.Note{
				{
					Title:   "Note A",
					Slug:    "note-a",
					Content: "Has empty link [[]] and valid [[Note B]]",
				},
				{
					Title:   "Note B",
					Slug:    "note-b",
					Content: "Valid note",
				},
			},
			expected: []model.Note{
				{
					Title:        "Note A",
					Slug:         "note-a",
					Content:      "Has empty link [[]] and valid [[Note B]]",
					ReferencedBy: []model.NoteReference{},
				},
				{
					Title:   "Note B",
					Slug:    "note-b",
					Content: "Valid note",
					ReferencedBy: []model.NoteReference{
						{Slug: "note-a", Title: "Note A"},
					},
				},
			},
		},
		{
			name: "wikilinks in metadata should create backlinks",
			notes: []model.Note{
				{
					Title:   "Source Note",
					Slug:    "source-note",
					Content: "This note has no content wikilinks",
					Metadata: map[string]any{
						"related": "[[Target Note]]",
						"author":  "John Doe",
					},
				},
				{
					Title:   "Target Note",
					Slug:    "target-note",
					Content: "This is the target",
				},
			},
			expected: []model.Note{
				{
					Title:   "Source Note",
					Slug:    "source-note",
					Content: "This note has no content wikilinks",
					Metadata: map[string]any{
						"related": "[[Target Note]]",
						"author":  "John Doe",
					},
					ReferencedBy: []model.NoteReference{},
				},
				{
					Title:   "Target Note",
					Slug:    "target-note",
					Content: "This is the target",
					ReferencedBy: []model.NoteReference{
						{Slug: "source-note", Title: "Source Note"},
					},
				},
			},
		},
		{
			name: "wikilinks in both content and metadata",
			notes: []model.Note{
				{
					Title:   "Source Note",
					Slug:    "source-note",
					Content: "Content references [[Note A]]",
					Metadata: map[string]any{
						"related": "[[Note B]]",
						"tags":    []any{"[[Note C]]", "regular-tag"},
					},
				},
				{
					Title:   "Note A",
					Slug:    "note-a",
					Content: "Note A content",
				},
				{
					Title:   "Note B",
					Slug:    "note-b",
					Content: "Note B content",
				},
				{
					Title:   "Note C",
					Slug:    "note-c",
					Content: "Note C content",
				},
			},
			expected: []model.Note{
				{
					Title:   "Source Note",
					Slug:    "source-note",
					Content: "Content references [[Note A]]",
					Metadata: map[string]any{
						"related": "[[Note B]]",
						"tags":    []any{"[[Note C]]", "regular-tag"},
					},
					ReferencedBy: []model.NoteReference{},
				},
				{
					Title:   "Note A",
					Slug:    "note-a",
					Content: "Note A content",
					ReferencedBy: []model.NoteReference{
						{Slug: "source-note", Title: "Source Note"},
					},
				},
				{
					Title:   "Note B",
					Slug:    "note-b",
					Content: "Note B content",
					ReferencedBy: []model.NoteReference{
						{Slug: "source-note", Title: "Source Note"},
					},
				},
				{
					Title:   "Note C",
					Slug:    "note-c",
					Content: "Note C content",
					ReferencedBy: []model.NoteReference{
						{Slug: "source-note", Title: "Source Note"},
					},
				},
			},
		},
		{
			name: "duplicate wikilinks in content and metadata should not create duplicate backlinks",
			notes: []model.Note{
				{
					Title:   "Source Note",
					Slug:    "source-note",
					Content: "Content references [[Target Note]]",
					Metadata: map[string]any{
						"related": "[[Target Note]]",
					},
				},
				{
					Title:   "Target Note",
					Slug:    "target-note",
					Content: "This is the target",
				},
			},
			expected: []model.Note{
				{
					Title:   "Source Note",
					Slug:    "source-note",
					Content: "Content references [[Target Note]]",
					Metadata: map[string]any{
						"related": "[[Target Note]]",
					},
					ReferencedBy: []model.NoteReference{},
				},
				{
					Title:   "Target Note",
					Slug:    "target-note",
					Content: "This is the target",
					ReferencedBy: []model.NoteReference{
						{Slug: "source-note", Title: "Source Note"},
					},
				},
			},
		},
		{
			name: "nested metadata objects with wikilinks",
			notes: []model.Note{
				{
					Title:   "Source Note",
					Slug:    "source-note",
					Content: "Regular content",
					Metadata: map[string]any{
						"references": map[string]any{
							"primary":   "[[Note A]]",
							"secondary": "[[Note B|Custom Display]]",
						},
						"description": "This note relates to [[Note C]]",
					},
				},
				{
					Title:   "Note A",
					Slug:    "note-a",
					Content: "Note A content",
				},
				{
					Title:   "Note B",
					Slug:    "note-b",
					Content: "Note B content",
				},
				{
					Title:   "Note C",
					Slug:    "note-c",
					Content: "Note C content",
				},
			},
			expected: []model.Note{
				{
					Title:   "Source Note",
					Slug:    "source-note",
					Content: "Regular content",
					Metadata: map[string]any{
						"references": map[string]any{
							"primary":   "[[Note A]]",
							"secondary": "[[Note B|Custom Display]]",
						},
						"description": "This note relates to [[Note C]]",
					},
					ReferencedBy: []model.NoteReference{},
				},
				{
					Title:   "Note A",
					Slug:    "note-a",
					Content: "Note A content",
					ReferencedBy: []model.NoteReference{
						{Slug: "source-note", Title: "Source Note"},
					},
				},
				{
					Title:   "Note B",
					Slug:    "note-b",
					Content: "Note B content",
					ReferencedBy: []model.NoteReference{
						{Slug: "source-note", Title: "Source Note"},
					},
				},
				{
					Title:   "Note C",
					Slug:    "note-c",
					Content: "Note C content",
					ReferencedBy: []model.NoteReference{
						{Slug: "source-note", Title: "Source Note"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildBackreferences(tt.notes)

			// Compare each note individually for better error messages
			if len(result) != len(tt.expected) {
				t.Fatalf("Expected %d notes, got %d", len(tt.expected), len(result))
			}

			for i, expectedNote := range tt.expected {
				resultNote := result[i]

				if resultNote.Title != expectedNote.Title {
					t.Errorf("Note %d: expected title %q, got %q", i, expectedNote.Title, resultNote.Title)
				}

				if resultNote.Slug != expectedNote.Slug {
					t.Errorf("Note %d: expected slug %q, got %q", i, expectedNote.Slug, resultNote.Slug)
				}

				if resultNote.Content != expectedNote.Content {
					t.Errorf("Note %d: expected content %q, got %q", i, expectedNote.Content, resultNote.Content)
				}

				if !reflect.DeepEqual(resultNote.ReferencedBy, expectedNote.ReferencedBy) {
					t.Errorf("Note %d (%s): expected ReferencedBy %+v, got %+v",
						i, expectedNote.Title, expectedNote.ReferencedBy, resultNote.ReferencedBy)
				}
			}
		})
	}
}

func TestExtractWikiLinks(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "single wikilink",
			content:  "This has [[Note A]] in it",
			expected: []string{"Note A"},
		},
		{
			name:     "multiple wikilinks",
			content:  "This has [[Note A]] and [[Note B]]",
			expected: []string{"Note A", "Note B"},
		},
		{
			name:     "wikilink with display name",
			content:  "This has [[Note A|Display Name]]",
			expected: []string{"Note A"},
		},
		{
			name:     "duplicate wikilinks",
			content:  "This has [[Note A]] and [[Note A]] again",
			expected: []string{"Note A"},
		},
		{
			name:     "empty wikilink",
			content:  "This has [[]] empty link",
			expected: nil,
		},
		{
			name:     "no wikilinks",
			content:  "This has no wikilinks",
			expected: nil,
		},
		{
			name:     "wikilink with spaces",
			content:  "This has [[ Note A ]] with spaces",
			expected: []string{"Note A"},
		},
		{
			name:     "mixed wikilinks",
			content:  "[[Note A]], [[Note B|Custom]], and [[Note A]] again",
			expected: []string{"Note A", "Note B"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractWikiLinks(tt.content)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestContainsReference(t *testing.T) {
	references := []model.NoteReference{
		{Slug: "note-a", Title: "Note A"},
		{Slug: "note-b", Title: "Note B"},
	}

	tests := []struct {
		name     string
		target   model.NoteReference
		expected bool
	}{
		{
			name:     "existing reference",
			target:   model.NoteReference{Slug: "note-a", Title: "Note A"},
			expected: true,
		},
		{
			name:     "non-existing reference",
			target:   model.NoteReference{Slug: "note-c", Title: "Note C"},
			expected: false,
		},
		{
			name:     "same slug different title",
			target:   model.NoteReference{Slug: "note-a", Title: "Different Title"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsReference(references, tt.target)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestExtractWikiLinksFromMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]any
		expected []string
	}{
		{
			name: "simple string metadata with wikilink",
			metadata: map[string]any{
				"related": "[[Note A]]",
				"author":  "John Doe",
			},
			expected: []string{"Note A"},
		},
		{
			name: "multiple wikilinks in different fields",
			metadata: map[string]any{
				"related":     "[[Note A]]",
				"description": "This relates to [[Note B]]",
				"tags":        "normal-tag",
			},
			expected: []string{"Note A", "Note B"},
		},
		{
			name: "wikilinks in array values",
			metadata: map[string]any{
				"tags": []any{"[[Note A]]", "regular-tag", "[[Note B]]"},
			},
			expected: []string{"Note A", "Note B"},
		},
		{
			name: "nested object with wikilinks",
			metadata: map[string]any{
				"references": map[string]any{
					"primary":   "[[Note A]]",
					"secondary": "[[Note B|Display]]",
				},
			},
			expected: []string{"Note A", "Note B"},
		},
		{
			name: "no wikilinks in metadata",
			metadata: map[string]any{
				"author": "John Doe",
				"tags":   []any{"tag1", "tag2"},
				"public": true,
			},
			expected: nil,
		},
		{
			name:     "empty metadata",
			metadata: map[string]any{},
			expected: nil,
		},
		{
			name: "mixed data types with wikilinks",
			metadata: map[string]any{
				"title":       "[[Note A]]",
				"count":       42,
				"enabled":     true,
				"description": "References [[Note B]] and [[Note C]]",
				"nested": map[string]any{
					"deep": "[[Note D]]",
				},
			},
			expected: []string{"Note A", "Note B", "Note C", "Note D"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractWikiLinksFromMetadata(tt.metadata)

			// Handle nil cases properly
			if tt.expected == nil {
				if result != nil && len(result) > 0 {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}

			// Sort both slices for comparison since order doesn't matter
			sort.Strings(result)
			expectedSorted := make([]string, len(tt.expected))
			copy(expectedSorted, tt.expected)
			sort.Strings(expectedSorted)

			if !reflect.DeepEqual(result, expectedSorted) {
				t.Errorf("Expected %v, got %v", expectedSorted, result)
			}
		})
	}
}

func TestRemoveDuplicateStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			input:    []string{"A", "B", "C"},
			expected: []string{"A", "B", "C"},
		},
		{
			name:     "with duplicates",
			input:    []string{"A", "B", "A", "C", "B"},
			expected: []string{"A", "B", "C"},
		},
		{
			name:     "all duplicates",
			input:    []string{"A", "A", "A"},
			expected: []string{"A"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "single item",
			input:    []string{"A"},
			expected: []string{"A"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeDuplicateStrings(tt.input)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Benchmark helper functions

// generateLargeContentWithWikilinks creates a large content string with many wikilinks
func generateLargeContentWithWikilinks(numWikilinks int) string {
	content := "# Large Note with Many Wikilinks\n\n"
	content += "This is a comprehensive note that references many other notes throughout the text. "
	content += "It demonstrates the performance characteristics of wikilink extraction on large documents.\n\n"

	for i := 0; i < numWikilinks; i++ {
		if i%10 == 0 {
			content += "\n## Section " + string(rune('A'+i/10)) + "\n\n"
		}

		// Mix different wikilink formats
		switch i % 4 {
		case 0:
			content += "This paragraph references [[Note " + string(rune('A'+i%26)) + string(rune('0'+i%10)) + "]] in the discussion. "
		case 1:
			content += "See also [[Important Topic " + string(rune('A'+i%26)) + "|custom display]] for more details. "
		case 2:
			content += "The concept of [[Research Paper " + string(rune('A'+i%26)) + string(rune('0'+i%10)) + "]] is fundamental here. "
		case 3:
			content += "For background, check [[Background Note " + string(rune('A'+i%26)) + "]] and related materials. "
		}

		if i%5 == 0 {
			content += "\n\n"
		}
	}

	content += "\n\n## Conclusion\n\nThis note demonstrates extensive cross-referencing capabilities."
	return content
}

// generateLargeMetadataWithWikilinks creates complex metadata with many wikilinks
func generateLargeMetadataWithWikilinks(numWikilinks int) map[string]any {
	metadata := map[string]any{
		"title":   "Large Note with Extensive Metadata",
		"author":  "Test Author",
		"created": "2024-01-01",
		"public":  true,
		"version": 1,
	}

	// Add wikilinks in various metadata fields
	var relatedNotes []any
	var tags []any
	var references = make(map[string]any)

	for i := 0; i < numWikilinks; i++ {
		noteName := "Meta Note " + string(rune('A'+i%26)) + string(rune('0'+i%10))

		// Add to different metadata structures
		switch i % 5 {
		case 0:
			relatedNotes = append(relatedNotes, "[["+noteName+"]]")
		case 1:
			tags = append(tags, "[["+noteName+"|tag]]")
		case 2:
			references["ref_"+string(rune('a'+i%26))] = "[[" + noteName + "]]"
		case 3:
			if i%10 == 3 {
				metadata["summary_"+string(rune('a'+i%26))] = "This section discusses [[" + noteName + "]] and its implications."
			}
		case 4:
			if i%15 == 4 {
				nestedRefs := make(map[string]any)
				nestedRefs["primary"] = "[[" + noteName + "]]"
				nestedRefs["secondary"] = "[[" + noteName + " Secondary|alt]]"
				references["nested_"+string(rune('a'+i%26))] = nestedRefs
			}
		}
	}

	metadata["related_notes"] = relatedNotes
	metadata["tags"] = tags
	metadata["references"] = references
	metadata["description"] = "A comprehensive note with [[Primary Topic]] and [[Secondary Topic|secondary]] references."

	return metadata
}

// generateLargeNoteCollection creates a collection of notes for benchmarking BuildBackreferences
func generateLargeNoteCollection(numNotes, wikilinksPerNote int) []model.Note {
	notes := make([]model.Note, numNotes)

	for i := 0; i < numNotes; i++ {
		noteTitle := "Benchmark Note " + string(rune('A'+i%26)) + string(rune('0'+(i/26)%10))
		notes[i] = model.Note{
			Title:    noteTitle,
			Slug:     "benchmark-note-" + string(rune('a'+i%26)) + string(rune('0'+(i/26)%10)),
			Content:  generateLargeContentWithWikilinks(wikilinksPerNote),
			Metadata: generateLargeMetadataWithWikilinks(wikilinksPerNote / 2),
		}
	}

	return notes
}

// Benchmarks

func BenchmarkExtractWikiLinks(b *testing.B) {
	benchmarks := []struct {
		name         string
		numWikilinks int
	}{
		{"Small_10_wikilinks", 10},
		{"Medium_100_wikilinks", 100},
		{"Large_1000_wikilinks", 1000},
		{"XLarge_5000_wikilinks", 5000},
	}

	for _, bm := range benchmarks {
		content := generateLargeContentWithWikilinks(bm.numWikilinks)
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = extractWikiLinks(content)
			}
		})
	}
}

func BenchmarkExtractWikiLinksFromMetadata(b *testing.B) {
	benchmarks := []struct {
		name         string
		numWikilinks int
	}{
		{"Small_10_wikilinks", 10},
		{"Medium_100_wikilinks", 100},
		{"Large_1000_wikilinks", 1000},
		{"XLarge_5000_wikilinks", 5000},
	}

	for _, bm := range benchmarks {
		metadata := generateLargeMetadataWithWikilinks(bm.numWikilinks)
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = extractWikiLinksFromMetadata(metadata)
			}
		})
	}
}

func BenchmarkBuildBackreferences(b *testing.B) {
	benchmarks := []struct {
		name             string
		numNotes         int
		wikilinksPerNote int
	}{
		{"Small_10_notes_10_wikilinks", 10, 10},
		{"Medium_50_notes_50_wikilinks", 50, 50},
		{"Large_100_notes_100_wikilinks", 100, 100},
		{"XLarge_200_notes_200_wikilinks", 200, 200},
	}

	for _, bm := range benchmarks {
		notes := generateLargeNoteCollection(bm.numNotes, bm.wikilinksPerNote)
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = BuildBackreferences(notes)
			}
		})
	}
}

func BenchmarkRemoveDuplicateStrings(b *testing.B) {
	benchmarks := []struct {
		name       string
		size       int
		duplicates int // percentage of duplicates
	}{
		{"Small_100_no_duplicates", 100, 0},
		{"Small_100_50pct_duplicates", 100, 50},
		{"Medium_1000_no_duplicates", 1000, 0},
		{"Medium_1000_50pct_duplicates", 1000, 50},
		{"Large_10000_no_duplicates", 10000, 0},
		{"Large_10000_50pct_duplicates", 10000, 50},
	}

	for _, bm := range benchmarks {
		// Generate test data
		input := make([]string, bm.size)
		uniqueCount := bm.size * (100 - bm.duplicates) / 100

		for i := 0; i < bm.size; i++ {
			if i < uniqueCount {
				input[i] = "Item" + string(rune('A'+i%26)) + string(rune('0'+(i/26)%10))
			} else {
				// Add duplicates
				input[i] = input[i%uniqueCount]
			}
		}

		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = removeDuplicateStrings(input)
			}
		})
	}
}

// Comprehensive benchmark that tests the entire pipeline
func BenchmarkFullWikilinkProcessing(b *testing.B) {
	benchmarks := []struct {
		name             string
		numNotes         int
		wikilinksPerNote int
	}{
		{"Realistic_20_notes_25_wikilinks", 20, 25},
		{"Heavy_50_notes_100_wikilinks", 50, 100},
		{"Extreme_100_notes_500_wikilinks", 100, 500},
	}

	for _, bm := range benchmarks {
		notes := generateLargeNoteCollection(bm.numNotes, bm.wikilinksPerNote)
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result := BuildBackreferences(notes)
				// Ensure the result is used to prevent optimization
				_ = len(result)
			}
		})
	}
}
