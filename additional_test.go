package main

import (
	"net/http"
	"os"
	"testing"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	"github.com/EwenQuim/pluie/template"
)

func TestFilterPublicNotes(t *testing.T) {
	tests := []struct {
		name            string
		notes           []model.Note
		publicByDefault bool
		expectedCount   int
		expectedSlugs   []string
	}{
		{
			name: "Mixed public and private notes with publicByDefault false",
			notes: []model.Note{
				{Slug: "public1", Title: "Public 1", IsPublic: true},
				{Slug: "private1", Title: "Private 1", IsPublic: false},
				{Slug: "public2", Title: "Public 2", IsPublic: true},
				{Slug: "private2", Title: "Private 2", IsPublic: false},
			},
			publicByDefault: false,
			expectedCount:   2,
			expectedSlugs:   []string{"public1", "public2"},
		},
		{
			name: "Mixed public and private notes with publicByDefault true",
			notes: []model.Note{
				{Slug: "public1", Title: "Public 1", IsPublic: true},
				{Slug: "private1", Title: "Private 1", IsPublic: false},
				{Slug: "public2", Title: "Public 2", IsPublic: true},
				{Slug: "private2", Title: "Private 2", IsPublic: false},
			},
			publicByDefault: true,
			expectedCount:   4, // All notes returned when publicByDefault is true
			expectedSlugs:   []string{"public1", "private1", "public2", "private2"},
		},
		{
			name: "All public notes",
			notes: []model.Note{
				{Slug: "public1", Title: "Public 1", IsPublic: true},
				{Slug: "public2", Title: "Public 2", IsPublic: true},
			},
			publicByDefault: false,
			expectedCount:   2,
			expectedSlugs:   []string{"public1", "public2"},
		},
		{
			name: "All private notes with publicByDefault false",
			notes: []model.Note{
				{Slug: "private1", Title: "Private 1", IsPublic: false},
				{Slug: "private2", Title: "Private 2", IsPublic: false},
			},
			publicByDefault: false,
			expectedCount:   0,
			expectedSlugs:   []string{},
		},
		{
			name:            "Empty notes",
			notes:           []model.Note{},
			publicByDefault: false,
			expectedCount:   0,
			expectedSlugs:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterPublicNotes(tt.notes, tt.publicByDefault)

			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d notes, got %d", tt.expectedCount, len(result))
			}

			// Check specific slugs
			resultSlugs := make(map[string]bool)
			for _, note := range result {
				resultSlugs[note.Slug] = true
			}

			for _, expectedSlug := range tt.expectedSlugs {
				if !resultSlugs[expectedSlug] {
					t.Errorf("Expected slug %s not found in filtered results", expectedSlug)
				}
			}

			// When publicByDefault is false, check that all returned notes are public
			if !tt.publicByDefault {
				for _, note := range result {
					if !note.IsPublic {
						t.Errorf("filterPublicNotes returned private note: %s", note.Slug)
					}
				}
			}
		})
	}
}

func TestServerStart(t *testing.T) {
	// Create a minimal server for testing
	notes := []model.Note{
		{Slug: "test", Title: "Test", IsPublic: true, Content: "Test content"},
	}
	notesMap := make(map[string]model.Note)
	for _, note := range notes {
		notesMap[note.Slug] = note
	}

	tree := engine.BuildTree(notes)
	notesService := engine.NewNotesService(&notesMap, tree, nil)
	cfg := &config.Config{
		Port:         "0", // Use port 0 to get a random available port
		HomeNoteSlug: "test",
		SiteTitle:    "Pluie",
		SiteIcon:     "/static/pluie.webp",
	}
	server := Server{
		NotesService: notesService,
		rs:           template.NewResource(cfg),
		cfg:          cfg,
	}

	// Test that Start method exists and can be called
	// We can't easily test the actual server start without complex setup
	// so we'll just verify the method exists and doesn't panic immediately
	t.Run("Start method exists", func(t *testing.T) {
		// This test just verifies the method signature is correct
		// We can't easily test the actual server startup without port conflicts
		_ = server.Start
	})
}

func TestServerGetNote(t *testing.T) {
	// Create test notes
	notes := []model.Note{
		{Slug: "test-note", Title: "Test Note", IsPublic: true, Content: "Test content"},
		{Slug: "home", Title: "Home", IsPublic: true, Content: "Home content"},
	}
	notesMap := make(map[string]model.Note)
	for _, note := range notes {
		notesMap[note.Slug] = note
	}

	tree := engine.BuildTree(notes)
	notesService := engine.NewNotesService(&notesMap, tree, nil)
	cfg := &config.Config{
		HomeNoteSlug: "home",
		SiteTitle:    "Pluie",
		SiteIcon:     "/static/pluie.webp",
	}
	server := Server{
		NotesService: notesService,
		rs:           template.NewResource(cfg),
		cfg:          cfg,
	}

	// Test that getNote method exists and has correct signature
	// The actual testing of this method is complex due to fuego.ContextNoBody dependency
	t.Run("getNote method exists", func(t *testing.T) {
		// This test just verifies the method signature is correct
		_ = server.getNote
	})
}

func TestGetHomeNoteSlugEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		notesMap     map[string]model.Note
		notes        []model.Note
		homeNoteSlug string
		expected     string
	}{
		{
			name: "Custom home note slug exists",
			notesMap: map[string]model.Note{
				"custom-home": {Slug: "custom-home", Title: "Custom Home", IsPublic: true},
				"other":       {Slug: "other", Title: "Other", IsPublic: true},
			},
			notes: []model.Note{
				{Slug: "custom-home", Title: "Custom Home", IsPublic: true},
				{Slug: "other", Title: "Other", IsPublic: true},
			},
			homeNoteSlug: "custom-home",
			expected:     "custom-home",
		},
		{
			name: "Custom home note slug doesn't exist, fallback to first",
			notesMap: map[string]model.Note{
				"apple": {Slug: "apple", Title: "Apple", IsPublic: true},
				"zebra": {Slug: "zebra", Title: "Zebra", IsPublic: true},
			},
			notes: []model.Note{
				{Slug: "apple", Title: "Apple", IsPublic: true},
				{Slug: "zebra", Title: "Zebra", IsPublic: true},
			},
			homeNoteSlug: "non-existent",
			expected:     "apple",
		},
		{
			name:         "Empty notes map",
			notesMap:     map[string]model.Note{},
			notes:        []model.Note{},
			homeNoteSlug: "any",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := engine.BuildTree(tt.notes)
			notesService := engine.NewNotesService(&tt.notesMap, tree, nil)
			cfg := &config.Config{
				HomeNoteSlug: tt.homeNoteSlug,
				SiteTitle:    "Pluie",
				SiteIcon:     "/static/pluie.webp",
			}
			server := Server{
				NotesService: notesService,
				rs:           template.NewResource(cfg),
				cfg:          cfg,
			}

			result := server.NotesService.GetHomeSlug(server.cfg.HomeNoteSlug)
			if result != tt.expected {
				t.Errorf("Expected home note slug %q, got %q", tt.expected, result)
			}
		})
	}
}

// Mock ResponseWriter for testing
type mockResponseWriter struct {
	headers http.Header
	body    []byte
	status  int
}

func (m *mockResponseWriter) Header() http.Header {
	return m.headers
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
	m.body = append(m.body, data...)
	return len(data), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.status = statusCode
}

func TestExplorerShouldSkipPath(t *testing.T) {
	explorer := Explorer{BasePath: "/test"}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Skip hidden directory",
			path:     "/.hidden",
			expected: true,
		},
		{
			name:     "Skip node_modules",
			path:     "/project/node_modules",
			expected: true,
		},
		{
			name:     "Skip .git directory",
			path:     "/project/.git",
			expected: true,
		},
		{
			name:     "Don't skip normal directory",
			path:     "/normal/path",
			expected: false,
		},
		{
			name:     "Don't skip root",
			path:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := explorer.shouldSkipPath(tt.path)
			if result != tt.expected {
				t.Errorf("shouldSkipPath(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestExplorerExtractTitle(t *testing.T) {
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
			name:            "H1 in content takes priority",
			fileName:        "test.md",
			metadata:        map[string]any{"title": "Metadata Title"},
			content:         "# Content Title\n\nBody text",
			expectedTitle:   "Content Title",
			expectedContent: "\nBody text",
		},
		{
			name:            "Metadata title when no H1",
			fileName:        "test.md",
			metadata:        map[string]any{"title": "Metadata Title"},
			content:         "Body text without H1",
			expectedTitle:   "Metadata Title",
			expectedContent: "Body text without H1",
		},
		{
			name:            "Filename fallback",
			fileName:        "my-file.md",
			metadata:        map[string]any{},
			content:         "Body text",
			expectedTitle:   "my-file",
			expectedContent: "Body text",
		},
		{
			name:            "Empty metadata title falls back to filename",
			fileName:        "fallback.md",
			metadata:        map[string]any{"title": ""},
			content:         "Body text",
			expectedTitle:   "fallback",
			expectedContent: "Body text",
		},
		{
			name:            "Non-string metadata title falls back to filename",
			fileName:        "fallback.md",
			metadata:        map[string]any{"title": 123},
			content:         "Body text",
			expectedTitle:   "fallback",
			expectedContent: "Body text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := tt.content
			title := explorer.extractTitle(tt.fileName, tt.metadata, &content)

			if title != tt.expectedTitle {
				t.Errorf("extractTitle() = %q, want %q", title, tt.expectedTitle)
			}
			if content != tt.expectedContent {
				t.Errorf("extractTitle() modified content = %q, want %q", content, tt.expectedContent)
			}
		})
	}
}

func TestExplorerCollectFolderMetadata(t *testing.T) {
	explorer := Explorer{BasePath: "/test"}

	// Create mock directory entries
	mockEntries := []mockDirEntry{
		{name: "regular.md", isDir: false},
		{name: "folder", isDir: true},
		{name: ".pluie", isDir: false},
		{name: "config.pluie", isDir: false},
	}

	// Convert to []os.DirEntry
	entries := make([]os.DirEntry, len(mockEntries))
	for i, entry := range mockEntries {
		entries[i] = entry
	}

	result := explorer.collectFolderMetadata(entries, "/test/path")

	// Should find .pluie files but we can't test the actual parsing without file system
	// This test mainly verifies the function doesn't panic and processes entries correctly
	if result == nil {
		t.Error("collectFolderMetadata should return non-nil map")
	}
}

// Mock DirEntry for testing
type mockDirEntry struct {
	name  string
	isDir bool
}

func (m mockDirEntry) Name() string               { return m.name }
func (m mockDirEntry) IsDir() bool                { return m.isDir }
func (m mockDirEntry) Type() os.FileMode          { return 0 }
func (m mockDirEntry) Info() (os.FileInfo, error) { return nil, nil }
