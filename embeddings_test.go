package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/EwenQuim/pluie/model"
)

func TestComputeContentHash(t *testing.T) {
	note1 := model.Note{Title: "Test", Content: "Hello world"}
	note2 := model.Note{Title: "Test", Content: "Hello world"}
	note3 := model.Note{Title: "Test", Content: "Different content"}

	hash1 := computeContentHash(note1)
	hash2 := computeContentHash(note2)
	hash3 := computeContentHash(note3)

	if hash1 != hash2 {
		t.Error("identical notes should produce identical hashes")
	}
	if hash1 == hash3 {
		t.Error("different notes should produce different hashes")
	}
	if hash1 == "" {
		t.Error("hash should not be empty")
	}
}

func TestNeedsEmbedding(t *testing.T) {
	tracker := &EmbeddingsTracker{
		Model: "nomic-embed-text",
		Files: make(map[string]EmbeddedFile),
	}

	note := model.Note{Title: "Test", Content: "content", Path: "test.md"}

	// New note needs embedding
	if !tracker.needsEmbedding(note) {
		t.Error("new note should need embedding")
	}

	// Mark as embedded
	tracker.markAsEmbedded(note, time.Now())

	// Same content doesn't need embedding
	if tracker.needsEmbedding(note) {
		t.Error("unchanged note should not need embedding")
	}

	// Changed content needs embedding
	note.Content = "updated content"
	if !tracker.needsEmbedding(note) {
		t.Error("changed note should need embedding")
	}
}

func TestMarkAsEmbedded(t *testing.T) {
	tracker := &EmbeddingsTracker{
		Model: "nomic-embed-text",
		Files: make(map[string]EmbeddedFile),
	}

	note := model.Note{Title: "Test", Content: "content", Path: "folder/test.md"}
	modTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	tracker.markAsEmbedded(note, modTime)

	embedded, exists := tracker.Files["folder/test.md"]
	if !exists {
		t.Fatal("note should be in tracker")
	}
	if embedded.Path != "folder/test.md" {
		t.Errorf("Path = %q, want %q", embedded.Path, "folder/test.md")
	}
	if embedded.LastModified != modTime {
		t.Errorf("LastModified = %v, want %v", embedded.LastModified, modTime)
	}
	if embedded.ContentHash == "" {
		t.Error("ContentHash should not be empty")
	}
}

func TestLoadEmbeddingsTrackerNewFile(t *testing.T) {
	tmpDir := t.TempDir()
	trackingFile := filepath.Join(tmpDir, "tracking.json")

	tracker, err := loadEmbeddingsTracker(trackingFile, "nomic-embed-text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tracker.Files) != 0 {
		t.Errorf("new tracker should have 0 files, got %d", len(tracker.Files))
	}
	if tracker.Model != "nomic-embed-text" {
		t.Errorf("Model = %q, want %q", tracker.Model, "nomic-embed-text")
	}
}

func TestLoadEmbeddingsTrackerModelChange(t *testing.T) {
	tmpDir := t.TempDir()
	trackingFile := filepath.Join(tmpDir, "tracking.json")

	// Create tracker with old model
	oldTracker := &EmbeddingsTracker{
		Model: "old-model",
		Files: map[string]EmbeddedFile{
			"test.md": {Path: "test.md", ContentHash: "abc123"},
		},
	}
	data, _ := json.MarshalIndent(oldTracker, "", "  ")
	os.WriteFile(trackingFile, data, 0644)

	// Load with new model
	tracker, err := loadEmbeddingsTracker(trackingFile, "new-model")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should clear files on model change
	if len(tracker.Files) != 0 {
		t.Errorf("tracker should have 0 files after model change, got %d", len(tracker.Files))
	}
	if tracker.Model != "new-model" {
		t.Errorf("Model = %q, want %q", tracker.Model, "new-model")
	}
}

func TestLoadEmbeddingsTrackerSameModel(t *testing.T) {
	tmpDir := t.TempDir()
	trackingFile := filepath.Join(tmpDir, "tracking.json")

	// Create tracker with current model
	oldTracker := &EmbeddingsTracker{
		Model: "nomic-embed-text",
		Files: map[string]EmbeddedFile{
			"test.md": {Path: "test.md", ContentHash: "abc123"},
		},
	}
	data, _ := json.MarshalIndent(oldTracker, "", "  ")
	os.WriteFile(trackingFile, data, 0644)

	// Load with same model
	tracker, err := loadEmbeddingsTracker(trackingFile, "nomic-embed-text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should keep files when model is the same
	if len(tracker.Files) != 1 {
		t.Errorf("tracker should have 1 file, got %d", len(tracker.Files))
	}
}

func TestSaveAndLoadEmbeddingsTracker(t *testing.T) {
	tmpDir := t.TempDir()
	trackingFile := filepath.Join(tmpDir, "tracking.json")

	tracker := &EmbeddingsTracker{
		Model: "nomic-embed-text",
		Files: make(map[string]EmbeddedFile),
	}

	note := model.Note{Title: "Test", Content: "content", Path: "test.md"}
	tracker.markAsEmbedded(note, time.Now())

	if err := tracker.save(trackingFile); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := loadEmbeddingsTracker(trackingFile, "nomic-embed-text")
	if err != nil {
		t.Fatalf("load error: %v", err)
	}

	if len(loaded.Files) != 1 {
		t.Errorf("loaded tracker should have 1 file, got %d", len(loaded.Files))
	}
	if loaded.Files["test.md"].Path != "test.md" {
		t.Errorf("Path = %q, want %q", loaded.Files["test.md"].Path, "test.md")
	}
}
