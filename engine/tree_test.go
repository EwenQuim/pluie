package engine

import (
	"testing"

	"github.com/EwenQuim/pluie/model"
)

func TestFindNoteInTree(t *testing.T) {
	// Create test notes
	notes := []model.Note{
		{Slug: "root-note", Title: "Root Note", IsPublic: true},
		{Slug: "folder/child-note", Title: "Child Note", IsPublic: true},
		{Slug: "folder/subfolder/deep-note", Title: "Deep Note", IsPublic: true},
	}

	tree := BuildTree(notes)

	tests := []struct {
		name     string
		slug     string
		expected bool
	}{
		{
			name:     "Find root note",
			slug:     "root-note",
			expected: true,
		},
		{
			name:     "Find child note",
			slug:     "folder/child-note",
			expected: true,
		},
		{
			name:     "Find deep note",
			slug:     "folder/subfolder/deep-note",
			expected: true,
		},
		{
			name:     "Note not found",
			slug:     "non-existent",
			expected: false,
		},
		{
			name:     "Empty slug",
			slug:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindNoteInTree(tree, tt.slug)
			found := result != nil
			if found != tt.expected {
				t.Errorf("FindNoteInTree(%q) found=%v, want found=%v", tt.slug, found, tt.expected)
			}
			if found && result.Note.Slug != tt.slug {
				t.Errorf("FindNoteInTree(%q) returned note with slug %q", tt.slug, result.Note.Slug)
			}
		})
	}
}

func TestGetAllNotesFromTree(t *testing.T) {
	tests := []struct {
		name          string
		notes         []model.Note
		expectedCount int
		expectedSlugs []string
	}{
		{
			name: "Multiple notes in tree",
			notes: []model.Note{
				{Slug: "root-note", Title: "Root Note", IsPublic: true},
				{Slug: "folder/child-note", Title: "Child Note", IsPublic: true},
				{Slug: "folder/subfolder/deep-note", Title: "Deep Note", IsPublic: true},
			},
			expectedCount: 3,
			expectedSlugs: []string{"root-note", "folder/child-note", "folder/subfolder/deep-note"},
		},
		{
			name:          "Empty tree",
			notes:         []model.Note{},
			expectedCount: 0,
			expectedSlugs: []string{},
		},
		{
			name: "Single note",
			notes: []model.Note{
				{Slug: "single-note", Title: "Single Note", IsPublic: true},
			},
			expectedCount: 1,
			expectedSlugs: []string{"single-note"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := BuildTree(tt.notes)
			result := GetAllNotesFromTree(tree)

			if len(result) != tt.expectedCount {
				t.Errorf("GetAllNotesFromTree() returned %d notes, want %d", len(result), tt.expectedCount)
			}

			// Check that all expected slugs are present
			resultSlugs := make(map[string]bool)
			for _, note := range result {
				resultSlugs[note.Slug] = true
			}

			for _, expectedSlug := range tt.expectedSlugs {
				if !resultSlugs[expectedSlug] {
					t.Errorf("Expected slug %q not found in results", expectedSlug)
				}
			}
		})
	}
}

func TestFilterTreeBySearch(t *testing.T) {
	// Create test notes
	notes := []model.Note{
		{Slug: "apple-note", Title: "Apple Note", IsPublic: true, Content: "This is about apples"},
		{Slug: "banana-note", Title: "Banana Note", IsPublic: true, Content: "This is about bananas"},
		{Slug: "fruit/orange-note", Title: "Orange Note", IsPublic: true, Content: "This is about oranges"},
		{Slug: "vegetable/carrot-note", Title: "Carrot Note", IsPublic: true, Content: "This is about carrots"},
	}

	tree := BuildTree(notes)

	tests := []struct {
		name          string
		searchQuery   string
		expectedCount int
		expectedSlugs []string
	}{
		{
			name:          "Search by title - apple",
			searchQuery:   "apple",
			expectedCount: 1,
			expectedSlugs: []string{"apple-note"},
		},
		{
			name:          "Search by title - note",
			searchQuery:   "note",
			expectedCount: 4, // All notes have "Note" in title
			expectedSlugs: []string{"apple-note", "banana-note", "fruit/orange-note", "vegetable/carrot-note"},
		},
		{
			name:          "Search by title - orange",
			searchQuery:   "orange",
			expectedCount: 1,
			expectedSlugs: []string{"fruit/orange-note"},
		},
		{
			name:          "Case insensitive search",
			searchQuery:   "BANANA",
			expectedCount: 1,
			expectedSlugs: []string{"banana-note"},
		},
		{
			name:          "No matches",
			searchQuery:   "xyz",
			expectedCount: 0,
			expectedSlugs: []string{},
		},
		{
			name:          "Empty search query",
			searchQuery:   "",
			expectedCount: 4, // Should return all notes
			expectedSlugs: []string{"apple-note", "banana-note", "fruit/orange-note", "vegetable/carrot-note"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterTreeBySearch(tree, tt.searchQuery)
			allNotes := GetAllNotesFromTree(result)

			if len(allNotes) != tt.expectedCount {
				t.Errorf("FilterTreeBySearch(%q) returned %d notes, want %d", tt.searchQuery, len(allNotes), tt.expectedCount)
			}

			// Check that all expected slugs are present
			resultSlugs := make(map[string]bool)
			for _, note := range allNotes {
				resultSlugs[note.Slug] = true
			}

			for _, expectedSlug := range tt.expectedSlugs {
				if !resultSlugs[expectedSlug] {
					t.Errorf("Expected slug %q not found in filtered results for query %q", expectedSlug, tt.searchQuery)
				}
			}
		})
	}
}

func TestFilterChildren(t *testing.T) {
	// Create test notes
	notes := []model.Note{
		{Slug: "folder/apple-note", Path: "folder/apple-note.md", Title: "Apple Note", IsPublic: true, Content: "This is about apples"},
		{Slug: "folder/banana-note", Path: "folder/banana-note.md", Title: "Banana Note", IsPublic: true, Content: "This is about bananas"},
		{Slug: "folder/carrot-note", Path: "folder/carrot-note.md", Title: "Carrot Note", IsPublic: true, Content: "This is about carrots"},
	}

	tree := BuildTree(notes)

	// Find the folder node
	var folderNode *TreeNode
	for _, child := range tree.Children {
		if child.Name == "folder" && child.IsFolder {
			folderNode = child
			break
		}
	}

	if folderNode == nil {
		// Debug: print tree structure
		t.Logf("Tree structure:")
		for i, child := range tree.Children {
			t.Logf("  Child %d: Name=%q, IsFolder=%v, Path=%q", i, child.Name, child.IsFolder, child.Path)
		}
		t.Fatal("Could not find folder node in tree")
	}

	tests := []struct {
		name          string
		searchQuery   string
		expectedCount int
	}{
		{
			name:          "Filter by apple",
			searchQuery:   "apple",
			expectedCount: 1,
		},
		{
			name:          "Filter by note",
			searchQuery:   "note",
			expectedCount: 3, // All notes have "Note" in title
		},
		{
			name:          "No matches",
			searchQuery:   "xyz",
			expectedCount: 0,
		},
		{
			name:          "Empty query",
			searchQuery:   "",
			expectedCount: 3, // Should return all children
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a target node to receive filtered results
			target := &TreeNode{
				Name:     "filtered",
				Path:     "",
				IsFolder: true,
				Children: make([]*TreeNode, 0),
				IsOpen:   true,
			}

			// Call filterChildren with correct signature
			filterChildren(folderNode, target, tt.searchQuery)

			if len(target.Children) != tt.expectedCount {
				t.Errorf("filterChildren(%q) returned %d children, want %d", tt.searchQuery, len(target.Children), tt.expectedCount)
			}
		})
	}
}

func TestCopyEntireSubtree(t *testing.T) {
	// Create test notes
	notes := []model.Note{
		{Slug: "folder/note1", Path: "folder/note1.md", Title: "Note 1", IsPublic: true, Content: "Content 1"},
		{Slug: "folder/subfolder/note2", Path: "folder/subfolder/note2.md", Title: "Note 2", IsPublic: true, Content: "Content 2"},
	}

	tree := BuildTree(notes)

	// Find the folder node
	var folderNode *TreeNode
	for _, child := range tree.Children {
		if child.Name == "folder" && child.IsFolder {
			folderNode = child
			break
		}
	}

	if folderNode == nil {
		t.Fatal("Could not find folder node in tree")
	}

	t.Run("Copy subtree", func(t *testing.T) {
		copied := copyEntireSubtree(folderNode)

		if copied == nil {
			t.Fatal("copyEntireSubtree returned nil")
		}

		if copied.Name != folderNode.Name {
			t.Errorf("Copied node name = %q, want %q", copied.Name, folderNode.Name)
		}

		if copied.IsFolder != folderNode.IsFolder {
			t.Errorf("Copied node IsFolder = %v, want %v", copied.IsFolder, folderNode.IsFolder)
		}

		if len(copied.Children) != len(folderNode.Children) {
			t.Errorf("Copied node has %d children, want %d", len(copied.Children), len(folderNode.Children))
		}

		// Verify it's a deep copy (different memory addresses)
		if copied == folderNode {
			t.Error("copyEntireSubtree returned the same object, not a copy")
		}
	})

	t.Run("Copy nil node", func(t *testing.T) {
		copied := copyEntireSubtree(nil)
		if copied != nil {
			t.Error("copyEntireSubtree(nil) should return nil")
		}
	})
}

func TestBuildTreeEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		notes []model.Note
	}{
		{
			name: "Notes with special characters in path",
			notes: []model.Note{
				{Slug: "folder with spaces/note", Title: "Note", IsPublic: true},
				{Slug: "folder-with-dashes/note", Title: "Note", IsPublic: true},
				{Slug: "folder_with_underscores/note", Title: "Note", IsPublic: true},
			},
		},
		{
			name: "Very deep nesting",
			notes: []model.Note{
				{Slug: "a/b/c/d/e/f/deep-note", Title: "Deep Note", IsPublic: true},
			},
		},
		{
			name: "Notes with same name in different folders",
			notes: []model.Note{
				{Slug: "folder1/note", Title: "Note 1", IsPublic: true},
				{Slug: "folder2/note", Title: "Note 2", IsPublic: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := BuildTree(tt.notes)
			if tree == nil {
				t.Error("BuildTree returned nil")
			}

			// Verify all notes are accessible
			allNotes := GetAllNotesFromTree(tree)
			if len(allNotes) != len(tt.notes) {
				t.Errorf("Expected %d notes in tree, got %d", len(tt.notes), len(allNotes))
			}
		})
	}
}
