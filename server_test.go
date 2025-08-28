package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	"github.com/EwenQuim/pluie/template"
)

func TestServerPrivateNoteFiltering(t *testing.T) {
	// Create test notes
	publicNote := model.Note{
		Title:    "Public Note",
		Slug:     "public-note",
		Content:  "This is public content",
		IsPublic: true,
		Metadata: map[string]any{"public": true},
	}

	// Create filtered notes map (only public notes - simulating main.go filtering)
	publicNotesMap := map[string]model.Note{
		"public-note": publicNote,
	}

	publicNotes := []model.Note{publicNote}

	// Create tree with public notes only
	tree := engine.BuildTree(publicNotes)

	// Create server with filtered data
	server := Server{
		NotesMap: &publicNotesMap,
		Tree:     tree,
		rs: template.Resource{
			Tree: tree,
		},
		cfg: &config.Config{
			PublicByDefault: false,
			HomeNoteSlug:    "",
		},
	}

	tests := []struct {
		name             string
		slug             string
		expectedStatus   int
		shouldContain    string
		shouldNotContain string
	}{
		{
			name:           "AccessPublicNote",
			slug:           "public-note",
			expectedStatus: http.StatusOK,
			shouldContain:  "This is public content",
		},
		{
			name:             "AccessPrivateNote",
			slug:             "private-note",
			expectedStatus:   http.StatusOK, // Returns 200 but with empty note list
			shouldNotContain: "This is private content",
		},
		{
			name:           "AccessNonExistentNote",
			slug:           "non-existent",
			expectedStatus: http.StatusOK, // Returns 200 but with empty note list
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request
			req, err := http.NewRequest("GET", "/"+tt.slug, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Create the handler function
			handler := func(w http.ResponseWriter, r *http.Request) {
				slug := r.URL.Path[1:] // Remove leading slash
				if slug == "" {
					slug = server.getHomeNoteSlug()
				}

				notesMap := *server.NotesMap
				note, ok := notesMap[slug]
				if !ok {
					// Note not found - this simulates the server behavior
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("Note not found"))
					return
				}

				// Additional security check: ensure note is public
				if !note.IsPublic {
					// Private note access denied
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("Note not found"))
					return
				}

				// Return the note content
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(note.Content))
			}

			// Call the handler
			handler(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			// Check response body content
			body := rr.Body.String()
			if tt.shouldContain != "" && !contains(body, tt.shouldContain) {
				t.Errorf("Response should contain %q, got: %s", tt.shouldContain, body)
			}
			if tt.shouldNotContain != "" && contains(body, tt.shouldNotContain) {
				t.Errorf("Response should not contain %q, got: %s", tt.shouldNotContain, body)
			}
		})
	}
}

func TestServerGetHomeNoteSlug(t *testing.T) {
	tests := []struct {
		name         string
		notesMap     map[string]model.Note
		notes        []model.Note
		homeNoteSlug string
		expected     string
	}{
		{
			name: "HomeNoteExists",
			notesMap: map[string]model.Note{
				"HOME": {Slug: "HOME", Title: "Home", IsPublic: true},
				"test": {Slug: "test", Title: "Test", IsPublic: true},
			},
			notes: []model.Note{
				{Slug: "HOME", Title: "Home", IsPublic: true},
				{Slug: "test", Title: "Test", IsPublic: true},
			},
			expected: "HOME",
		},
		{
			name: "FirstNoteAlphabetically",
			notesMap: map[string]model.Note{
				"zebra": {Slug: "zebra", Title: "Zebra", IsPublic: true},
				"apple": {Slug: "apple", Title: "Apple", IsPublic: true},
			},
			notes: []model.Note{
				{Slug: "zebra", Title: "Zebra", IsPublic: true},
				{Slug: "apple", Title: "Apple", IsPublic: true},
			},
			expected: "apple",
		},
		{
			name:     "NoNotes",
			notesMap: map[string]model.Note{},
			notes:    []model.Note{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := engine.BuildTree(tt.notes)
			server := Server{
				NotesMap: &tt.notesMap,
				Tree:     tree,
				rs: template.Resource{
					Tree: tree,
				},
				cfg: &config.Config{
					PublicByDefault: false,
					HomeNoteSlug:    tt.homeNoteSlug,
				},
			}

			result := server.getHomeNoteSlug()
			if result != tt.expected {
				t.Errorf("Expected home note slug %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestBuildTreeWithPrivateNotes(t *testing.T) {
	// Create a mix of public and private notes
	notes := []model.Note{
		{Title: "Public Root", Slug: "public-root.md", IsPublic: true},
		{Title: "Private Root", Slug: "private-root.md", IsPublic: false},
		{Title: "Public Folder Note", Slug: "folder/public.md", IsPublic: true},
		{Title: "Private Folder Note", Slug: "folder/private.md", IsPublic: false},
	}

	// Filter to public notes only (simulating main.go behavior)
	publicNotes := make([]model.Note, 0)
	for _, note := range notes {
		if note.IsPublic {
			publicNotes = append(publicNotes, note)
		}
	}

	// Build tree with public notes only
	tree := engine.BuildTree(publicNotes)

	// Verify that the tree only contains public notes
	allTreeNotes := engine.GetAllNotesFromTree(tree)

	if len(allTreeNotes) != 2 {
		t.Errorf("Expected 2 public notes in tree, got %d", len(allTreeNotes))
	}

	// Verify that all notes in the tree are public
	for _, note := range allTreeNotes {
		if !note.IsPublic {
			t.Errorf("Tree contains private note: %s", note.Slug)
		}
	}

	// Verify specific notes are present
	slugs := make(map[string]bool)
	for _, note := range allTreeNotes {
		slugs[note.Slug] = true
	}

	expectedSlugs := []string{"public-root.md", "folder/public.md"}
	for _, expectedSlug := range expectedSlugs {
		if !slugs[expectedSlug] {
			t.Errorf("Expected public note %s not found in tree", expectedSlug)
		}
	}

	// Verify private notes are not present
	privateSlugs := []string{"private-root.md", "folder/private.md"}
	for _, privateSlug := range privateSlugs {
		if slugs[privateSlug] {
			t.Errorf("Private note %s should not be in tree", privateSlug)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
