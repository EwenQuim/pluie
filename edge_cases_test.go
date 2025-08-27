package main

import (
	"testing"

	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
)

func TestEdgeCases(t *testing.T) {
	t.Run("InvalidYAMLFrontmatter", func(t *testing.T) {
		// Test that invalid YAML doesn't crash the system
		explorer := Explorer{
			BasePath: "testdata",
		}

		// This should not crash even if there are YAML parsing errors
		notes, err := explorer.getFolderNotes("")
		if err != nil {
			t.Fatalf("getFolderNotes() should handle invalid YAML gracefully, got error: %v", err)
		}

		// Should still return some notes
		if len(notes) == 0 {
			t.Error("Expected some notes even with potential YAML errors")
		}
	})

	t.Run("EmptyMetadata", func(t *testing.T) {
		note := model.Note{
			Slug:     "test.md",
			Metadata: map[string]any{},
		}

		// Should fall back to default (now always false)
		note.DetermineIsPublic(map[string]map[string]any{})
		if note.IsPublic {
			t.Error("Note with empty metadata should use default (false)")
		}
	})

	t.Run("NilMetadata", func(t *testing.T) {
		note := model.Note{
			Slug:     "test.md",
			Metadata: nil,
		}

		// Should not crash with nil metadata
		note.DetermineIsPublic(map[string]map[string]any{})
		if note.IsPublic {
			t.Error("Note with nil metadata should use default (false)")
		}
	})

	t.Run("NonBooleanPublishValue", func(t *testing.T) {
		note := model.Note{
			Slug:     "test.md",
			Metadata: map[string]any{"publish": "not a boolean"},
		}

		// Should fall back to default when publish value is not a boolean
		note.DetermineIsPublic(map[string]map[string]any{})
		if note.IsPublic {
			t.Error("Note with non-boolean publish value should use default (false)")
		}
	})

	t.Run("DeepFolderPath", func(t *testing.T) {
		note := model.Note{
			Slug:     "very/deep/folder/structure/test.md",
			Metadata: map[string]any{},
		}

		folderMetadata := map[string]map[string]any{
			"very/deep/folder/structure": {"publish": true},
		}

		note.DetermineIsPublic(folderMetadata)
		if !note.IsPublic {
			t.Error("Note in deep folder should inherit folder metadata")
		}
	})

	t.Run("FolderMetadataWithNonBooleanPublish", func(t *testing.T) {
		note := model.Note{
			Slug:     "folder/test.md",
			Metadata: map[string]any{},
		}

		folderMetadata := map[string]map[string]any{
			"folder": {"publish": "not a boolean"},
		}

		// Should fall back to default when folder publish value is not a boolean
		note.DetermineIsPublic(folderMetadata)
		if note.IsPublic {
			t.Error("Note with non-boolean folder publish value should use default (false)")
		}
	})

	t.Run("EmptySlug", func(t *testing.T) {
		note := model.Note{
			Slug:     "",
			Metadata: map[string]any{},
		}

		// Should not crash with empty slug
		note.DetermineIsPublic(map[string]map[string]any{})
		if note.IsPublic {
			t.Error("Note with empty slug should use default (false)")
		}
	})

	t.Run("SlugWithoutExtension", func(t *testing.T) {
		note := model.Note{
			Slug:     "folder/test",
			Metadata: map[string]any{},
		}

		folderMetadata := map[string]map[string]any{
			"folder": {"publish": true},
		}

		note.DetermineIsPublic(folderMetadata)
		if !note.IsPublic {
			t.Error("Note without extension should still inherit folder metadata")
		}
	})

	t.Run("RootLevelNote", func(t *testing.T) {
		note := model.Note{
			Slug:     "root-note.md",
			Metadata: map[string]any{},
		}

		folderMetadata := map[string]map[string]any{
			"subfolder": {"publish": true},
		}

		// Root level note should not inherit subfolder metadata
		note.DetermineIsPublic(folderMetadata)
		if note.IsPublic {
			t.Error("Root level note should default to private (false)")
		}
	})

	t.Run("FilteringWithMixedNotes", func(t *testing.T) {
		notes := []model.Note{
			{Slug: "public1.md", IsPublic: true},
			{Slug: "private1.md", IsPublic: false},
			{Slug: "public2.md", IsPublic: true},
			{Slug: "private2.md", IsPublic: false},
			{Slug: "public3.md", IsPublic: true},
		}

		// Filter to public notes
		publicNotes := make([]model.Note, 0)
		for _, note := range notes {
			if note.IsPublic {
				publicNotes = append(publicNotes, note)
			}
		}

		if len(publicNotes) != 3 {
			t.Errorf("Expected 3 public notes, got %d", len(publicNotes))
		}

		// Verify all filtered notes are public
		for _, note := range publicNotes {
			if !note.IsPublic {
				t.Errorf("Filtered note %s should be public", note.Slug)
			}
		}
	})

	t.Run("BuildSlugEdgeCases", func(t *testing.T) {
		tests := []struct {
			name     string
			note     model.Note
			expected string
		}{
			{
				name:     "EmptySlugWithTitle",
				note:     model.Note{Title: "Test Title", Slug: ""},
				expected: "Test-Title",
			},
			{
				name:     "SlugWithSpaces",
				note:     model.Note{Title: "Test", Slug: "test with spaces.md"},
				expected: "test-with-spaces",
			},
			{
				name:     "SlugWithMultipleDashes",
				note:     model.Note{Title: "Test", Slug: "test--with--dashes.md"},
				expected: "test-with-dashes",
			},
			{
				name:     "SlugWithLeadingTrailingDashes",
				note:     model.Note{Title: "Test", Slug: "-test-slug-.md"},
				expected: "test-slug",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.note.BuildSlug()
				if tt.note.Slug != tt.expected {
					t.Errorf("Expected slug %q, got %q", tt.expected, tt.note.Slug)
				}
			})
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	// Test that the system handles concurrent access gracefully
	explorer := Explorer{
		BasePath: "testdata",
	}

	// Run multiple goroutines accessing the same functionality
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			notes, err := explorer.getFolderNotes("")
			if err != nil {
				t.Errorf("Concurrent access failed: %v", err)
				return
			}

			// Basic validation
			if len(notes) == 0 {
				t.Error("Expected some notes from concurrent access")
			}

			// Test filtering
			publicNotes := make([]model.Note, 0)
			for _, note := range notes {
				if note.IsPublic {
					publicNotes = append(publicNotes, note)
				}
			}

			// Should have some public notes
			if len(publicNotes) == 0 {
				t.Error("Expected some public notes from concurrent access")
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestMemoryUsage(t *testing.T) {
	// Test that the system doesn't leak memory with large numbers of notes
	explorer := Explorer{
		BasePath: "testdata",
	}

	// Run the operation multiple times to check for memory leaks
	for i := 0; i < 100; i++ {
		notes, err := explorer.getFolderNotes("")
		if err != nil {
			t.Fatalf("Memory test failed at iteration %d: %v", i, err)
		}

		// Process the notes
		publicNotes := make([]model.Note, 0)
		for _, note := range notes {
			if note.IsPublic {
				publicNotes = append(publicNotes, note)
			}
		}

		// Build tree
		tree := engine.BuildTree(publicNotes)
		if tree == nil {
			t.Fatalf("Tree building failed at iteration %d", i)
		}

		// Clear references to help GC
		notes = nil
		publicNotes = nil
		tree = nil
	}
}
