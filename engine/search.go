package engine

import (
	"sort"
	"strings"

	"github.com/EwenQuim/pluie/model"
)

type NoteScored struct {
	Note  model.Note
	Score int
}

// SearchNotesByFilename searches notes by filename (title and slug) and returns filtered results
// This function shows matches in the file name first, folder names second
func SearchNotesByFilename(notes []model.Note, searchQuery string) []model.Note {
	if searchQuery == "" {
		return notes
	}

	var filteredNotes []NoteScored
	searchLower := strings.ToLower(searchQuery)

	for _, note := range notes {
		// Check if search appears in the title (which represents the filename)
		if strings.Contains(strings.ToLower(note.Title), searchLower) {
			filteredNotes = append(filteredNotes, NoteScored{Note: note, Score: 2})
			continue
		}

		if strings.Contains(strings.ToLower(note.Slug), searchLower) {
			filteredNotes = append(filteredNotes, NoteScored{Note: note, Score: 1})
			continue
		}
	}

	// Sort by score descending
	sort.Slice(filteredNotes, func(i, j int) bool {
		return filteredNotes[i].Score > filteredNotes[j].Score
	})

	// Extract notes from scored results
	result := make([]model.Note, len(filteredNotes))
	for i, ns := range filteredNotes {
		result[i] = ns.Note
	}
	return result
}
