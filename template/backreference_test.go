package template

import (
	"reflect"
	"regexp"
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
			regex := regexp.MustCompile(`\[\[([^\]]*)\]\]`)
			result := extractWikiLinks(tt.content, regex)

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
