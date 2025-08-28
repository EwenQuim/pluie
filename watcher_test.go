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

func TestFileWatcher(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "pluie-watcher-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test markdown file
	testFile := filepath.Join(tmpDir, "test.md")
	initialContent := `---
title: Test Note
publish: true
---

This is a test note.`

	err = os.WriteFile(testFile, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create initial server state
	cfg := &config.Config{PublicByDefault: true}
	explorer := Explorer{BasePath: tmpDir}
	
	notes, err := explorer.getFolderNotes("")
	if err != nil {
		t.Fatalf("Failed to get initial notes: %v", err)
	}

	publicNotes := filterPublicNotes(notes, cfg.PublicByDefault)
	publicNotes = engine.BuildBackreferences(publicNotes)
	
	notesMap := make(map[string]model.Note)
	for _, note := range publicNotes {
		notesMap[note.Slug] = note
	}
	
	tree := engine.BuildTree(publicNotes)

	server := &Server{
		NotesMap: &notesMap,
		Tree:     tree,
		rs: template.Resource{Tree: tree},
		cfg:      cfg,
	}

	// Verify initial state
	if len(*server.NotesMap) != 1 {
		t.Fatalf("Expected 1 initial note, got %d", len(*server.NotesMap))
	}

	originalNote := (*server.NotesMap)["test"]
	if originalNote.Title != "Test Note" {
		t.Fatalf("Expected title 'Test Note', got '%s'", originalNote.Title)
	}

	// Create file watcher
	watcher, err := NewFileWatcher(tmpDir, server, cfg)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	err = watcher.Start()
	if err != nil {
		t.Fatalf("Failed to start file watcher: %v", err)
	}
	defer watcher.Stop()

	// Modify the test file
	updatedContent := `---
title: Updated Test Note
publish: true
---

This is an updated test note with new content.`

	err = os.WriteFile(testFile, []byte(updatedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to update test file: %v", err)
	}

	// Wait for the file watcher to detect the change and reload
	// The debounce delay is 300ms, so we wait a bit longer
	time.Sleep(500 * time.Millisecond)

	// Verify the note was updated
	server.mutex.RLock()
	updatedNotesMap := *server.NotesMap
	server.mutex.RUnlock()

	if len(updatedNotesMap) != 1 {
		t.Fatalf("Expected 1 note after update, got %d", len(updatedNotesMap))
	}

	updatedNote := updatedNotesMap["test"]
	if updatedNote.Title != "Updated Test Note" {
		t.Fatalf("Expected updated title 'Updated Test Note', got '%s'", updatedNote.Title)
	}

	// Debug: log the note details
	t.Logf("Updated note: Title=%s, IsPublic=%t, Slug=%s", updatedNote.Title, updatedNote.IsPublic, updatedNote.Slug)

	if !updatedNote.IsPublic {
		t.Fatal("Expected updated note to be public")
	}
}

func TestFileWatcherNewFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "pluie-watcher-newfile-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create initial server state with no files
	cfg := &config.Config{PublicByDefault: true}
	
	notesMap := make(map[string]model.Note)
	tree := &engine.TreeNode{}

	server := &Server{
		NotesMap: &notesMap,
		Tree:     tree,
		rs: template.Resource{Tree: tree},
		cfg:      cfg,
	}

	// Verify initial state
	if len(*server.NotesMap) != 0 {
		t.Fatalf("Expected 0 initial notes, got %d", len(*server.NotesMap))
	}

	// Create file watcher
	watcher, err := NewFileWatcher(tmpDir, server, cfg)
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}

	err = watcher.Start()
	if err != nil {
		t.Fatalf("Failed to start file watcher: %v", err)
	}
	defer watcher.Stop()

	// Create a new test file
	testFile := filepath.Join(tmpDir, "new-note.md")
	newContent := `---
title: New Note
publish: true
---

This is a new note.`

	err = os.WriteFile(testFile, []byte(newContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create new test file: %v", err)
	}

	// Wait for the file watcher to detect the new file and reload
	time.Sleep(500 * time.Millisecond)

	// Verify the new note was added
	server.mutex.RLock()
	updatedNotesMap := *server.NotesMap
	server.mutex.RUnlock()

	if len(updatedNotesMap) != 1 {
		t.Fatalf("Expected 1 note after adding new file, got %d", len(updatedNotesMap))
	}

	newNote := updatedNotesMap["new-note"]
	if newNote.Title != "New Note" {
		t.Fatalf("Expected new note title 'New Note', got '%s'", newNote.Title)
	}
}