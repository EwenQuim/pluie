package engine

import (
	"log/slog"
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
