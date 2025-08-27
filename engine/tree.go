package engine

import (
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/EwenQuim/pluie/model"
)

// TreeNode represents a node in the file tree structure
type TreeNode struct {
	Name     string      `json:"name"`     // Display name (folder name or note title)
	Path     string      `json:"path"`     // Full path from root
	IsFolder bool        `json:"isFolder"` // True if this is a folder, false if it's a note
	Note     *model.Note `json:"note"`     // Reference to the note if this is a note node
	Children []*TreeNode `json:"children"` // Child nodes (subfolders and notes)
	IsOpen   bool        `json:"isOpen"`   // Whether the folder is expanded in the UI
}

// AllNotes yields all notes in the tree using Go 1.23 iterator pattern
func (t *TreeNode) AllNotes(yield func(*TreeNode) bool) {
	if t == nil {
		return
	}

	if !t.IsFolder && t.Note != nil {
		if !yield(t) {
			return
		}
	}

	for _, child := range t.Children {
		child.AllNotes(yield)
	}
}

// BuildTree creates a tree structure from a list of notes
func BuildTree(notes []model.Note) *TreeNode {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		slog.Info("Tree built", "in", elapsed.String())
	}()

	root := &TreeNode{
		Name:     "Notes",
		Path:     "",
		IsFolder: true,
		Children: make([]*TreeNode, 0),
		IsOpen:   true, // Root is always open
	}

	// Create a map to store folder nodes for quick lookup
	folderMap := make(map[string]*TreeNode)
	folderMap[""] = root

	// Sort notes by path to ensure consistent ordering
	sortedNotes := make([]model.Note, len(notes))
	copy(sortedNotes, notes)
	sort.Slice(sortedNotes, func(i, j int) bool {
		return sortedNotes[i].Path < sortedNotes[j].Path
	})

	for _, note := range sortedNotes {
		// Clean the path and split into path components
		cleanPath := strings.TrimPrefix(note.Path, "/")
		// Remove the .md extension for path processing
		cleanPath = strings.TrimSuffix(cleanPath, ".md")
		pathParts := strings.Split(cleanPath, "/")

		// If it's just a filename, put it in root
		if len(pathParts) == 1 {
			noteNode := &TreeNode{
				Name:     note.Title,
				Path:     note.Slug,
				IsFolder: false,
				Note:     &note,
				Children: make([]*TreeNode, 0),
			}
			root.Children = append(root.Children, noteNode)
			continue
		}

		// Create folder structure if it doesn't exist
		currentPath := ""
		currentParent := root

		// Process all path parts except the last one (which is the file)
		for i, part := range pathParts[:len(pathParts)-1] {
			if currentPath == "" {
				currentPath = part
			} else {
				currentPath = currentPath + "/" + part
			}

			// Check if folder already exists
			folderNode, exists := folderMap[currentPath]
			if !exists {
				// Create new folder node
				folderNode = &TreeNode{
					Name:     part,
					Path:     currentPath,
					IsFolder: true,
					Children: make([]*TreeNode, 0),
					IsOpen:   i == 0, // Only open first level by default
				}
				folderMap[currentPath] = folderNode
				currentParent.Children = append(currentParent.Children, folderNode)
			}
			currentParent = folderNode
		}

		// Add the note to the appropriate folder
		noteNode := &TreeNode{
			Name:     note.Title,
			Path:     note.Slug,
			IsFolder: false,
			Note:     &note,
			Children: make([]*TreeNode, 0),
		}
		currentParent.Children = append(currentParent.Children, noteNode)
	}

	// Sort children at each level (folders first, then notes, both alphabetically)
	sortTreeChildren(root)

	return root
}

// sortTreeChildren recursively sorts children in each node
func sortTreeChildren(node *TreeNode) {
	if len(node.Children) == 0 {
		return
	}

	// Sort children: folders first, then notes, both alphabetically
	sort.Slice(node.Children, func(i, j int) bool {
		a, b := node.Children[i], node.Children[j]

		// Folders come before notes
		if a.IsFolder && !b.IsFolder {
			return true
		}
		if !a.IsFolder && b.IsFolder {
			return false
		}

		// Both are same type, sort alphabetically
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	})

	// Recursively sort children
	for _, child := range node.Children {
		sortTreeChildren(child)
	}
}

// FindNoteInTree searches for a note by slug in the tree
func FindNoteInTree(root *TreeNode, slug string) *TreeNode {
	if !root.IsFolder && root.Note != nil && root.Note.Slug == slug {
		return root
	}

	for _, child := range root.Children {
		if result := FindNoteInTree(child, slug); result != nil {
			return result
		}
	}

	return nil
}

// GetAllNotesFromTree extracts all notes from the tree structure
func GetAllNotesFromTree(root *TreeNode) []model.Note {
	var notes []model.Note

	if !root.IsFolder && root.Note != nil {
		notes = append(notes, *root.Note)
	}

	for _, child := range root.Children {
		notes = append(notes, GetAllNotesFromTree(child)...)
	}

	return notes
}

// FilterTreeBySearch filters the tree to only show nodes that match the search query
func FilterTreeBySearch(root *TreeNode, query string) *TreeNode {
	if query == "" {
		return root
	}

	query = strings.ToLower(query)

	// Create a new root for filtered results
	filteredRoot := &TreeNode{
		Name:     root.Name,
		Path:     root.Path,
		IsFolder: true,
		Children: make([]*TreeNode, 0),
		IsOpen:   true,
	}

	// Recursively filter children
	filterChildren(root, filteredRoot, query)

	return filteredRoot
}

// filterChildren recursively filters children based on search query
func filterChildren(source *TreeNode, target *TreeNode, query string) {
	for _, child := range source.Children {
		if child.IsFolder {
			// Check if folder name matches the search query
			folderMatches := strings.Contains(strings.ToLower(child.Name), query)

			// Create temp folder to check for matching descendants
			tempFolder := &TreeNode{
				Name:     child.Name,
				Path:     child.Path,
				IsFolder: true,
				Children: make([]*TreeNode, 0),
				IsOpen:   true, // Open folders in search results
			}

			// Recursively filter children
			filterChildren(child, tempFolder, query)

			// Include folder if:
			// 1. The folder name itself matches, OR
			// 2. The folder has matching descendants
			if folderMatches || len(tempFolder.Children) > 0 {
				// If folder name matches, include ALL its contents
				if folderMatches {
					tempFolder = copyEntireSubtree(child)
					tempFolder.IsOpen = true // Ensure matched folders are open
				}
				target.Children = append(target.Children, tempFolder)
			}
		} else {
			// For notes, check if title matches
			if strings.Contains(strings.ToLower(child.Name), query) {
				noteNode := &TreeNode{
					Name:     child.Name,
					Path:     child.Path,
					IsFolder: false,
					Note:     child.Note,
					Children: make([]*TreeNode, 0),
				}
				target.Children = append(target.Children, noteNode)
			}
		}
	}
}

// copyEntireSubtree creates a complete copy of a subtree with all its contents
func copyEntireSubtree(source *TreeNode) *TreeNode {
	if source == nil {
		return nil
	}

	copy := &TreeNode{
		Name:     source.Name,
		Path:     source.Path,
		IsFolder: source.IsFolder,
		Note:     source.Note,
		IsOpen:   true, // Open all folders in search results
		Children: make([]*TreeNode, len(source.Children)),
	}

	for i, child := range source.Children {
		copy.Children[i] = copyEntireSubtree(child)
	}

	return copy
}
