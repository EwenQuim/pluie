package engine

import (
	"log/slog"
	"sort"
	"strings"
	"sync"

	"github.com/EwenQuim/pluie/model"
)

// NotesService manages the notes data with thread-safe access
type NotesService struct {
	mu       sync.RWMutex           // Protects notesMap, tree, and tagIndex
	notesMap *map[string]model.Note // Slug -> Note
	tree     *TreeNode              // Tree structure of notes
	tagIndex TagIndex               // Tag -> Notes mapping
}

// NewNotesService creates a new NotesService with the given data
func NewNotesService(notesMap *map[string]model.Note, tree *TreeNode, tagIndex TagIndex) *NotesService {
	return &NotesService{
		notesMap: notesMap,
		tree:     tree,
		tagIndex: tagIndex,
	}
}

// UpdateData safely updates the service's notesMap, tree, and tagIndex with new data
func (ns *NotesService) UpdateData(notesMap *map[string]model.Note, tree *TreeNode, tagIndex TagIndex) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	ns.notesMap = notesMap
	ns.tree = tree
	ns.tagIndex = tagIndex

	slog.Info("Notes data updated", "notes_count", len(*notesMap))
}

// GetNotesMap returns a thread-safe copy of the notesMap
func (ns *NotesService) GetNotesMap() map[string]model.Note {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	if ns.notesMap == nil {
		return make(map[string]model.Note)
	}
	return *ns.notesMap
}

// GetTree returns the tree with a read lock
func (ns *NotesService) GetTree() *TreeNode {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	return ns.tree
}

// GetTagIndex returns the tag index with a read lock
func (ns *NotesService) GetTagIndex() TagIndex {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	return ns.tagIndex
}

// GetNote safely retrieves a note by slug
func (ns *NotesService) GetNote(slug string) (model.Note, bool) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	if ns.notesMap == nil {
		return model.Note{}, false
	}

	note, ok := (*ns.notesMap)[slug]
	return note, ok
}

// ParseWikiLinksInMetadata processes wikilinks in metadata values
// This is a convenience method that wraps engine.ParseWikiLinksInMetadata
func (ns *NotesService) ParseWikiLinksInMetadata(metadata map[string]any) map[string]any {
	return ParseWikiLinksInMetadata(metadata, ns.GetTree())
}

// ParseWikiLinks processes wiki-style links in content
// This is a convenience method that wraps engine.ParseWikiLinks
func (ns *NotesService) ParseWikiLinks(content string) string {
	return ParseWikiLinks(content, ns.GetTree())
}

// FilterTreeBySearch filters the tree to only show nodes matching the search query
// This is a convenience method that wraps engine.FilterTreeBySearch
func (ns *NotesService) FilterTreeBySearch(query string) *TreeNode {
	return FilterTreeBySearch(ns.GetTree(), query)
}

// GetAllNotes extracts all notes from the tree structure
// This is a convenience method that wraps engine.GetAllNotesFromTree
func (ns *NotesService) GetAllNotes() []model.Note {
	tree := ns.GetTree()
	if tree == nil {
		return []model.Note{}
	}
	return GetAllNotesFromTree(tree)
}

// GetHomeSlug determines the home note slug based on priority:
// 1. The provided homeNoteSlug config value (if it exists in notes)
// 2. First note in alphabetical order
// 3. Empty string if no notes exist
func (ns *NotesService) GetHomeSlug(homeNoteSlug string) string {
	// Priority 1: Check provided homeNoteSlug configuration
	if homeNoteSlug != "" {
		if _, exists := ns.GetNote(homeNoteSlug); exists {
			return homeNoteSlug
		}
	}

	// Priority 2: First note in alphabetical order
	notes := ns.GetAllNotes()
	if len(notes) > 0 {
		sort.Slice(notes, func(i, j int) bool {
			return notes[i].Slug < notes[j].Slug
		})
		return notes[0].Slug
	}

	// Fallback: no notes available
	return ""
}

// SearchNotesByFilename searches notes by filename (title and slug) with a maximum result limit
// This function shows matches in the file name first (score 2), folder names second (score 1)
// Returns early if maxResults is reached to optimize performance (0 means no limit)
func (ns *NotesService) SearchNotesByFilename(searchQuery string, maxResults int) []model.Note {
	notes := ns.GetAllNotes()

	if searchQuery == "" {
		// Sort by slug for consistent ordering
		sort.Slice(notes, func(i, j int) bool {
			return notes[i].Slug < notes[j].Slug
		})
		if maxResults > 0 && len(notes) > maxResults {
			return notes[:maxResults]
		}
		return notes
	}

	searchLower := strings.ToLower(searchQuery)

	// Separate high-score (title) and low-score (slug) matches
	var titleMatches []model.Note
	var slugMatches []model.Note

	for _, note := range notes {
		// Check title first (higher priority)
		if strings.Contains(strings.ToLower(note.Title), searchLower) {
			titleMatches = append(titleMatches, note)
			// Early exit if we have enough title matches (they're highest priority)
			if maxResults > 0 && len(titleMatches) >= maxResults {
				// Sort before returning for consistent order
				sort.Slice(titleMatches, func(i, j int) bool {
					return titleMatches[i].Slug < titleMatches[j].Slug
				})
				return titleMatches
			}
			continue
		}

		// Check slug (lower priority)
		if strings.Contains(strings.ToLower(note.Slug), searchLower) {
			slugMatches = append(slugMatches, note)
		}
	}

	// Sort each group by slug for consistent ordering
	sort.Slice(titleMatches, func(i, j int) bool {
		return titleMatches[i].Slug < titleMatches[j].Slug
	})
	sort.Slice(slugMatches, func(i, j int) bool {
		return slugMatches[i].Slug < slugMatches[j].Slug
	})

	// Combine results: title matches first, then slug matches
	result := make([]model.Note, 0, len(titleMatches)+len(slugMatches))
	result = append(result, titleMatches...)
	result = append(result, slugMatches...)

	// Apply limit if specified
	if maxResults > 0 && len(result) > maxResults {
		result = result[:maxResults]
	}

	return result
}

// SearchNotesByHeadings searches for headings (H1-H3) matching the query
// Returns early if maxResults is reached after sorting (0 means no limit)
func (ns *NotesService) SearchNotesByHeadings(searchQuery string, maxResults int) []HeadingMatch {
	if searchQuery == "" {
		return nil
	}

	notes := ns.GetAllNotes()
	searchLower := strings.ToLower(searchQuery)

	var matches []HeadingMatch

	for _, note := range notes {
		lines := strings.Split(note.Content, "\n")

		for i, line := range lines {
			trimmed := strings.TrimSpace(line)

			// Only check lines that start with #
			if !strings.HasPrefix(trimmed, "#") {
				continue
			}

			heading, level := extractHeading(line)

			// Only H1-H3 (skip H4-H6 and invalid headings)
			if level < 1 || level > 3 {
				continue
			}

			// Check if query matches heading
			if strings.Contains(strings.ToLower(heading), searchLower) {
				score := calculateHeadingScore(heading, searchQuery, level)
				context := extractContext(lines, i, 75)

				matches = append(matches, HeadingMatch{
					Note:    note,
					Heading: heading,
					Level:   level,
					Context: context,
					LineNum: i,
					Score:   score,
				})
			}
		}
	}

	// Sort by score descending
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// Apply limit if specified
	if maxResults > 0 && len(matches) > maxResults {
		matches = matches[:maxResults]
	}

	return matches
}
