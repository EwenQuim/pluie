package main

import (
	"sync"
	"testing"
	"time"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	"github.com/EwenQuim/pluie/template"
)

func TestServerReactiveDataUpdate(t *testing.T) {
	// Create initial notes
	initialNote1 := model.Note{
		Title:    "Initial Note 1",
		Slug:     "initial-1",
		Content:  "This is initial content 1",
		IsPublic: true,
	}
	initialNote2 := model.Note{
		Title:    "Initial Note 2",
		Slug:     "initial-2",
		Content:  "This is initial content 2",
		IsPublic: true,
	}

	// Create initial notes map
	initialNotesMap := map[string]model.Note{
		"initial-1": initialNote1,
		"initial-2": initialNote2,
	}

	initialNotes := []model.Note{initialNote1, initialNote2}
	initialTree := engine.BuildTree(initialNotes)
	initialTagIndex := engine.BuildTagIndex(initialNotes)

	// Create server with initial data
	server := &Server{
		NotesMap: &initialNotesMap,
		Tree:     initialTree,
		TagIndex: initialTagIndex,
		rs: template.Resource{
			Tree: initialTree,
		},
		cfg: &config.Config{
			PublicByDefault: true,
			HomeNoteSlug:    "",
		},
	}

	// Test 1: Verify initial data is accessible
	server.mu.RLock()
	notesMap1 := *server.NotesMap
	server.mu.RUnlock()

	if len(notesMap1) != 2 {
		t.Errorf("Expected 2 initial notes, got %d", len(notesMap1))
	}

	if note, ok := notesMap1["initial-1"]; !ok {
		t.Error("Initial note 1 not found")
	} else if note.Content != "This is initial content 1" {
		t.Errorf("Expected initial content 1, got %s", note.Content)
	}

	// Test 2: Update data concurrently with reads
	var wg sync.WaitGroup
	numReaders := 10
	readErrors := make(chan error, numReaders)

	// Start multiple readers
	for i := range numReaders {
		wg.Add(1)
		go func(readerId int) {
			defer wg.Done()
			for range 100 {
				server.mu.RLock()
				notesMap := *server.NotesMap
				server.mu.RUnlock()

				// Verify we can always read valid data
				if len(notesMap) == 0 {
					readErrors <- nil
					return
				}

				// Small delay to increase chance of concurrent access
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	// Update data while readers are running
	time.Sleep(10 * time.Millisecond)

	updatedNote1 := model.Note{
		Title:    "Updated Note 1",
		Slug:     "updated-1",
		Content:  "This is updated content 1",
		IsPublic: true,
	}
	updatedNote2 := model.Note{
		Title:    "Updated Note 2",
		Slug:     "updated-2",
		Content:  "This is updated content 2",
		IsPublic: true,
	}
	updatedNote3 := model.Note{
		Title:    "New Note 3",
		Slug:     "new-3",
		Content:  "This is new content 3",
		IsPublic: true,
	}

	updatedNotesMap := map[string]model.Note{
		"updated-1": updatedNote1,
		"updated-2": updatedNote2,
		"new-3":     updatedNote3,
	}

	updatedNotes := []model.Note{updatedNote1, updatedNote2, updatedNote3}
	updatedTree := engine.BuildTree(updatedNotes)
	updatedTagIndex := engine.BuildTagIndex(updatedNotes)

	// Update server data
	server.UpdateData(&updatedNotesMap, updatedTree, updatedTagIndex)

	// Wait for all readers to finish
	wg.Wait()
	close(readErrors)

	// Check if any reader encountered an error
	for err := range readErrors {
		if err != nil {
			t.Error(err)
		}
	}

	// Test 3: Verify updated data is now accessible
	server.mu.RLock()
	notesMap2 := *server.NotesMap
	tree2 := server.Tree
	server.mu.RUnlock()

	if len(notesMap2) != 3 {
		t.Errorf("Expected 3 updated notes, got %d", len(notesMap2))
	}

	if note, ok := notesMap2["updated-1"]; !ok {
		t.Error("Updated note 1 not found")
	} else if note.Content != "This is updated content 1" {
		t.Errorf("Expected updated content 1, got %s", note.Content)
	}

	if note, ok := notesMap2["new-3"]; !ok {
		t.Error("New note 3 not found")
	} else if note.Content != "This is new content 3" {
		t.Errorf("Expected new content 3, got %s", note.Content)
	}

	// Verify old notes are gone
	if _, ok := notesMap2["initial-1"]; ok {
		t.Error("Initial note 1 should not exist after update")
	}

	// Verify tree was also updated
	allNotes := engine.GetAllNotesFromTree(tree2)
	if len(allNotes) != 3 {
		t.Errorf("Expected 3 notes in updated tree, got %d", len(allNotes))
	}

	// Test 4: Verify resource tree was also updated
	if server.rs.Tree != tree2 {
		t.Error("Resource tree was not updated")
	}
}

func TestServerReactiveConcurrentUpdates(t *testing.T) {
	// Create initial server
	initialNotesMap := map[string]model.Note{
		"note-1": {
			Title:    "Note 1",
			Slug:     "note-1",
			Content:  "Content 1",
			IsPublic: true,
		},
	}

	initialNotes := []model.Note{initialNotesMap["note-1"]}
	initialTree := engine.BuildTree(initialNotes)
	initialTagIndex := engine.BuildTagIndex(initialNotes)

	server := &Server{
		NotesMap: &initialNotesMap,
		Tree:     initialTree,
		TagIndex: initialTagIndex,
		rs: template.Resource{
			Tree: initialTree,
		},
		cfg: &config.Config{
			PublicByDefault: true,
		},
	}

	// Perform multiple concurrent updates
	var wg sync.WaitGroup
	numUpdates := 5

	for i := range numUpdates {
		wg.Add(1)
		go func(updateId int) {
			defer wg.Done()

			// Create unique notes for this update
			notes := []model.Note{
				{
					Title:    "Updated Note",
					Slug:     "updated-note",
					Content:  "Updated content",
					IsPublic: true,
				},
			}

			notesMap := map[string]model.Note{
				"updated-note": notes[0],
			}

			tree := engine.BuildTree(notes)
			tagIndex := engine.BuildTagIndex(notes)

			// Update server data
			server.UpdateData(&notesMap, tree, tagIndex)

			// Small delay
			time.Sleep(time.Millisecond)
		}(i)
	}

	wg.Wait()

	// Verify server is still in a consistent state
	server.mu.RLock()
	notesMap := *server.NotesMap
	tree := server.Tree
	server.mu.RUnlock()

	if len(notesMap) != 1 {
		t.Errorf("Expected 1 note after updates, got %d", len(notesMap))
	}

	if tree == nil {
		t.Error("Tree should not be nil after updates")
	}

	allNotes := engine.GetAllNotesFromTree(tree)
	if len(allNotes) != 1 {
		t.Errorf("Expected 1 note in tree after updates, got %d", len(allNotes))
	}
}
