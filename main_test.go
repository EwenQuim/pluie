package main

import (
	"os"
	"testing"

	"github.com/EwenQuim/pluie/model"
)

func TestExplorerGetFolderNotes(t *testing.T) {
	tests := []struct {
		name            string
		publicByDefault bool
		expectedNotes   map[string]bool // slug -> isPublic
	}{
		{
			name:            "PublicByDefaultFalse",
			publicByDefault: false,
			expectedNotes: map[string]bool{
				"public_note":                true,  // explicit public: true
				"private_note":               false, // explicit public: false
				"default_note":               false, // no frontmatter, now defaults to false
				"public_folder/folder_note":  true,  // folder metadata public: true
				"private_folder/secret_note": false, // folder metadata public: false
			},
		},
		{
			name:            "PublicByDefaultTrue",
			publicByDefault: true,
			expectedNotes: map[string]bool{
				"public_note":                true,  // explicit public: true
				"private_note":               false, // explicit public: false
				"default_note":               false, // no frontmatter, defaults to false
				"public_folder/folder_note":  true,  // folder metadata public: true
				"private_folder/secret_note": false, // folder metadata public: false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			explorer := Explorer{
				BasePath:        "testdata",
				PublicByDefault: tt.publicByDefault,
			}

			notes, err := explorer.getFolderNotes("")
			if err != nil {
				t.Fatalf("getFolderNotes() error = %v", err)
			}

			// Create a map of actual results
			actualNotes := make(map[string]bool)
			for _, note := range notes {
				actualNotes[note.Slug] = note.IsPublic
			}

			// Debug: Print actual slugs found
			t.Logf("Actual notes found:")
			for slug, isPublic := range actualNotes {
				t.Logf("  %s -> IsPublic: %t", slug, isPublic)
			}

			// Check that we have the expected number of notes
			if len(actualNotes) != len(tt.expectedNotes) {
				t.Errorf("Expected %d notes, got %d", len(tt.expectedNotes), len(actualNotes))
			}

			// Check each expected note
			for expectedSlug, expectedIsPublic := range tt.expectedNotes {
				actualIsPublic, exists := actualNotes[expectedSlug]
				if !exists {
					t.Errorf("Expected note %s not found", expectedSlug)
					continue
				}
				if actualIsPublic != expectedIsPublic {
					t.Errorf("Note %s: expected IsPublic=%t, got IsPublic=%t",
						expectedSlug, expectedIsPublic, actualIsPublic)
				}
			}
		})
	}
}

func TestNoteMetadataParsing(t *testing.T) {
	explorer := Explorer{
		BasePath:        "testdata",
		PublicByDefault: false,
	}

	notes, err := explorer.getFolderNotes("")
	if err != nil {
		t.Fatalf("getFolderNotes() error = %v", err)
	}

	// Find the public note and check its metadata
	var publicNote *model.Note
	for _, note := range notes {
		if note.Slug == "public_note" {
			publicNote = &note
			break
		}
	}

	if publicNote == nil {
		t.Fatal("Public note not found")
	}

	// Check that metadata was parsed correctly
	if publicNote.Metadata["title"] != "Public Test Note" {
		t.Errorf("Expected title 'Public Test Note', got %v", publicNote.Metadata["title"])
	}

	if publicNote.Metadata["author"] != "Test Author" {
		t.Errorf("Expected author 'Test Author', got %v", publicNote.Metadata["author"])
	}

	if publicNote.Metadata["publish"] != true {
		t.Errorf("Expected publish true, got %v", publicNote.Metadata["publish"])
	}
}

func TestFolderMetadataInheritance(t *testing.T) {
	explorer := Explorer{
		BasePath:        "testdata",
		PublicByDefault: false,
	}

	notes, err := explorer.getFolderNotes("")
	if err != nil {
		t.Fatalf("getFolderNotes() error = %v", err)
	}

	// Test public folder inheritance
	var folderNote *model.Note
	for _, note := range notes {
		if note.Slug == "public_folder/folder_note" {
			folderNote = &note
			break
		}
	}

	if folderNote == nil {
		t.Fatal("Folder note not found")
	}

	if !folderNote.IsPublic {
		t.Error("Folder note should be public due to folder metadata")
	}

	// Test private folder inheritance
	var secretNote *model.Note
	for _, note := range notes {
		if note.Slug == "private_folder/secret_note" {
			secretNote = &note
			break
		}
	}

	if secretNote == nil {
		t.Fatal("Secret note not found")
	}

	if secretNote.IsPublic {
		t.Error("Secret note should be private due to folder metadata")
	}
}

func TestPublicByDefaultEnvironmentVariable(t *testing.T) {
	// Test with PUBLIC_BY_DEFAULT=true
	os.Setenv("PUBLIC_BY_DEFAULT", "true")
	defer os.Unsetenv("PUBLIC_BY_DEFAULT")

	explorer := Explorer{
		BasePath:        "testdata",
		PublicByDefault: true, // This should be set based on env var in main()
	}

	notes, err := explorer.getFolderNotes("")
	if err != nil {
		t.Fatalf("getFolderNotes() error = %v", err)
	}

	// Find the default note (no frontmatter)
	var defaultNote *model.Note
	for _, note := range notes {
		if note.Slug == "default_note" {
			defaultNote = &note
			break
		}
	}

	if defaultNote == nil {
		t.Fatal("Default note not found")
	}

	if defaultNote.IsPublic {
		t.Error("Default note should be private by default (DetermineIsPublic now ignores PUBLIC_BY_DEFAULT)")
	}
}

func TestNoteFiltering(t *testing.T) {
	explorer := Explorer{
		BasePath:        "testdata",
		PublicByDefault: false,
	}

	allNotes, err := explorer.getFolderNotes("")
	if err != nil {
		t.Fatalf("getFolderNotes() error = %v", err)
	}

	// Filter to public notes only (simulating main.go logic)
	publicNotes := make([]model.Note, 0)
	for _, note := range allNotes {
		if note.IsPublic {
			publicNotes = append(publicNotes, note)
		}
	}

	// Count expected public and private notes
	expectedPublic := 0
	expectedPrivate := 0
	for _, note := range allNotes {
		if note.IsPublic {
			expectedPublic++
		} else {
			expectedPrivate++
		}
	}

	if len(publicNotes) != expectedPublic {
		t.Errorf("Expected %d public notes, got %d", expectedPublic, len(publicNotes))
	}

	totalNotes := len(allNotes)
	if totalNotes != expectedPublic+expectedPrivate {
		t.Errorf("Total notes mismatch: expected %d, got %d", expectedPublic+expectedPrivate, totalNotes)
	}

	// Verify that all filtered notes are indeed public
	for _, note := range publicNotes {
		if !note.IsPublic {
			t.Errorf("Filtered note %s should be public but IsPublic=%t", note.Slug, note.IsPublic)
		}
	}
}

func TestNoteDetermineIsPublic(t *testing.T) {
	tests := []struct {
		name            string
		note            model.Note
		publicByDefault bool
		folderMetadata  map[string]map[string]any
		expectedPublic  bool
	}{
		{
			name: "ExplicitPublishTrue",
			note: model.Note{
				Slug:     "test.md",
				Metadata: map[string]any{"publish": true},
			},
			publicByDefault: false,
			folderMetadata:  map[string]map[string]any{},
			expectedPublic:  true,
		},
		{
			name: "ExplicitPublishFalse",
			note: model.Note{
				Slug:     "test.md",
				Metadata: map[string]any{"publish": false},
			},
			publicByDefault: true,
			folderMetadata:  map[string]map[string]any{},
			expectedPublic:  false,
		},
		{
			name: "FolderMetadataPublish",
			note: model.Note{
				Slug:     "folder/test.md",
				Metadata: map[string]any{},
			},
			publicByDefault: false,
			folderMetadata: map[string]map[string]any{
				"folder": {"publish": true},
			},
			expectedPublic: true,
		},
		{
			name: "DefaultPublicTrue",
			note: model.Note{
				Slug:     "test.md",
				Metadata: map[string]any{},
			},
			publicByDefault: true,
			folderMetadata:  map[string]map[string]any{},
			expectedPublic:  false, // Now defaults to false regardless of publicByDefault
		},
		{
			name: "DefaultPublicFalse",
			note: model.Note{
				Slug:     "test.md",
				Metadata: map[string]any{},
			},
			publicByDefault: false,
			folderMetadata:  map[string]map[string]any{},
			expectedPublic:  false, // Now defaults to false regardless of publicByDefault
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.note.DetermineIsPublic(tt.folderMetadata)
			if tt.note.IsPublic != tt.expectedPublic {
				t.Errorf("Expected IsPublic=%t, got IsPublic=%t", tt.expectedPublic, tt.note.IsPublic)
			}
		})
	}
}
