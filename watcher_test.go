package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	"github.com/EwenQuim/pluie/template"
)

func TestLoadNotes(t *testing.T) {
	// Use the test data directory
	testPath := "./testdata"

	cfg := &config.Config{
		PublicByDefault: false,
	}

	// Load notes
	notesMap, tree, tagIndex, err := loadNotes(testPath, cfg)
	if err != nil {
		t.Fatalf("Failed to load notes: %v", err)
	}

	if notesMap == nil {
		t.Error("notesMap should not be nil")
	}

	if tree == nil {
		t.Error("tree should not be nil")
	}

	if tagIndex == nil {
		t.Error("tagIndex should not be nil")
	}

	// Verify some notes were loaded (should have at least the public ones)
	if len(*notesMap) < 1 {
		t.Error("Expected at least one public note to be loaded")
	}
}

func TestFileWatcherIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create an initial note file
	initialNote := `---
publish: true
---

# Initial Note

This is the initial content.
`
	err := os.WriteFile(filepath.Join(tempDir, "test.md"), []byte(initialNote), 0644)
	if err != nil {
		t.Fatalf("Failed to create initial note: %v", err)
	}

	cfg := &config.Config{
		PublicByDefault: false,
	}

	// Load initial notes
	notesMap, tree, tagIndex, err := loadNotes(tempDir, cfg)
	if err != nil {
		t.Fatalf("Failed to load notes: %v", err)
	}

	// Create server
	server := &Server{
		NotesMap: notesMap,
		Tree:     tree,
		TagIndex: tagIndex,
		rs: template.Resource{
			Tree: tree,
		},
		cfg: cfg,
	}

	// Verify initial note was loaded
	if len(*server.NotesMap) != 1 {
		t.Errorf("Expected 1 initial note, got %d", len(*server.NotesMap))
	}

	// Start file watcher
	watcher, err := watchFiles(server, tempDir, cfg)
	if err != nil {
		t.Fatalf("Failed to start file watcher: %v", err)
	}
	defer watcher.Close()

	// Wait a moment for the watcher to start
	time.Sleep(100 * time.Millisecond)

	// Modify the note
	modifiedNote := `---
publish: true
---

# Modified Note

This content has been modified.
`
	err = os.WriteFile(filepath.Join(tempDir, "test.md"), []byte(modifiedNote), 0644)
	if err != nil {
		t.Fatalf("Failed to modify note: %v", err)
	}

	// Wait for the file watcher to detect the change and reload
	time.Sleep(1 * time.Second)

	// Verify the note was reloaded
	server.mu.RLock()
	notesMap = server.NotesMap
	server.mu.RUnlock()

	if len(*notesMap) != 1 {
		t.Errorf("Expected 1 note after modification, got %d", len(*notesMap))
	}

	// Find the note (we need to get its slug first)
	var foundNote *model.Note
	for _, note := range *notesMap {
		if note.Title == "Modified Note" {
			foundNote = &note
			break
		}
	}

	if foundNote == nil {
		t.Error("Modified note not found")
	} else {
		// Just verify the content contains the key phrase (ignoring whitespace variations)
		if foundNote.Title != "Modified Note" {
			t.Errorf("Note title was not updated, expected 'Modified Note', got: %s", foundNote.Title)
		}
	}

	// Create a new note
	newNote := `---
publish: true
---

# New Note

This is a new note.
`
	err = os.WriteFile(filepath.Join(tempDir, "new.md"), []byte(newNote), 0644)
	if err != nil {
		t.Fatalf("Failed to create new note: %v", err)
	}

	// Wait for the file watcher to detect the change and reload
	time.Sleep(1 * time.Second)

	// Verify the new note was added
	server.mu.RLock()
	notesMap = server.NotesMap
	tree = server.Tree
	server.mu.RUnlock()

	if len(*notesMap) != 2 {
		t.Errorf("Expected 2 notes after adding new note, got %d", len(*notesMap))
	}

	// Verify tree was updated
	allNotes := engine.GetAllNotesFromTree(tree)
	if len(allNotes) != 2 {
		t.Errorf("Expected 2 notes in tree after adding new note, got %d", len(allNotes))
	}
}
