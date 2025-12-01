package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/EwenQuim/pluie/model"
	"github.com/tmc/langchaingo/schema"
)

// EmbeddingProgress tracks the current state of embedding operations
type EmbeddingProgress struct {
	mu            sync.RWMutex
	TotalNotes    int
	EmbeddedNotes int
	IsEmbedding   bool
	CurrentNote   string
	LastUpdated   time.Time
	subscribers   []chan EmbeddingStatus
	subscribersMu sync.Mutex
}

// EmbeddingStatus represents a point-in-time snapshot of embedding progress
type EmbeddingStatus struct {
	TotalNotes    int       `json:"total_notes"`
	EmbeddedNotes int       `json:"embedded_notes"`
	IsEmbedding   bool      `json:"is_embedding"`
	CurrentNote   string    `json:"current_note,omitempty"`
	LastUpdated   time.Time `json:"last_updated"`
}

// NewEmbeddingProgress creates a new embedding progress tracker
func NewEmbeddingProgress() *EmbeddingProgress {
	return &EmbeddingProgress{
		subscribers: make([]chan EmbeddingStatus, 0),
		LastUpdated: time.Now(),
	}
}

// GetStatus returns the current embedding status
func (ep *EmbeddingProgress) GetStatus() EmbeddingStatus {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	return EmbeddingStatus{
		TotalNotes:    ep.TotalNotes,
		EmbeddedNotes: ep.EmbeddedNotes,
		IsEmbedding:   ep.IsEmbedding,
		CurrentNote:   ep.CurrentNote,
		LastUpdated:   ep.LastUpdated,
	}
}

// UpdateProgress updates the embedding progress and notifies subscribers
func (ep *EmbeddingProgress) UpdateProgress(embedded, total int, currentNote string, isEmbedding bool) {
	ep.mu.Lock()
	ep.TotalNotes = total
	ep.EmbeddedNotes = embedded
	ep.CurrentNote = currentNote
	ep.IsEmbedding = isEmbedding
	ep.LastUpdated = time.Now()
	ep.mu.Unlock()

	// Notify all subscribers
	status := ep.GetStatus()
	ep.subscribersMu.Lock()
	for _, ch := range ep.subscribers {
		select {
		case ch <- status:
		default:
			// Skip if channel is full
		}
	}
	ep.subscribersMu.Unlock()
}

// Subscribe adds a new subscriber that will receive progress updates
func (ep *EmbeddingProgress) Subscribe() chan EmbeddingStatus {
	ch := make(chan EmbeddingStatus, 10)
	ep.subscribersMu.Lock()
	ep.subscribers = append(ep.subscribers, ch)
	ep.subscribersMu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber
func (ep *EmbeddingProgress) Unsubscribe(ch chan EmbeddingStatus) {
	ep.subscribersMu.Lock()
	defer ep.subscribersMu.Unlock()

	for i, subscriber := range ep.subscribers {
		if subscriber == ch {
			ep.subscribers = append(ep.subscribers[:i], ep.subscribers[i+1:]...)
			close(ch)
			break
		}
	}
}

// embedNotesWithProgress embeds notes into a vector store with progress tracking
func (em *EmbeddingsManager) embedNotesWithProgress(ctx context.Context, store VectorStore, notes []model.Note, progress *EmbeddingProgress) error {
	start := time.Now()

	// Load tracking file
	tracker, err := loadEmbeddingsTracker(em.embeddingsTrackingFile)
	if err != nil {
		return fmt.Errorf("loading embeddings tracker: %w", err)
	}

	// Filter notes that need embedding
	var notesToEmbed []model.Note
	for _, note := range notes {
		if tracker.needsEmbedding(note) {
			notesToEmbed = append(notesToEmbed, note)
		}
	}

	totalNotes := len(notes)
	alreadyEmbedded := len(tracker.Files)

	if len(notesToEmbed) == 0 {
		slog.Info("No new notes to embed, all notes are up to date",
			"total_notes", totalNotes,
			"tracked_notes", alreadyEmbedded)
		progress.UpdateProgress(alreadyEmbedded, totalNotes, "", false)
		return nil
	}

	slog.Info("Notes embedding status",
		"total_notes", totalNotes,
		"already_embedded", alreadyEmbedded,
		"to_embed", len(notesToEmbed))

	// Mark as embedding in progress
	progress.UpdateProgress(alreadyEmbedded, totalNotes, "", true)

	// Add documents to Weaviate one at a time to show real progress
	slog.Info("Starting embedding process", "documents", len(notesToEmbed))

	for i, note := range notesToEmbed {
		docStart := time.Now()

		// Update progress with current note
		progress.UpdateProgress(alreadyEmbedded+i, totalNotes, note.Title, true)

		// Combine title and content for better semantic search
		content := fmt.Sprintf("# %s\n\n%s", note.Title, note.Content)
		doc := schema.Document{
			PageContent: content,
			Metadata: map[string]any{
				"title": note.Title,
				"path":  note.Path,
				"slug":  note.Slug,
			},
		}

		_, err = store.AddDocuments(ctx, []schema.Document{doc})
		if err != nil {
			progress.UpdateProgress(alreadyEmbedded+i, totalNotes, "", false)
			return fmt.Errorf("adding document to vector store (title=%s, path=%s): %w", note.Title, note.Path, err)
		}

		slog.Info("Document embedded successfully",
			"document", i+1,
			"total", len(notesToEmbed),
			"title", note.Title,
			"duration", time.Since(docStart))
	}

	// Update tracking file
	for _, note := range notesToEmbed {
		// Get file modification time
		info, err := os.Stat(filepath.Join(".", note.Path))
		var modTime time.Time
		if err == nil {
			modTime = info.ModTime()
		} else {
			modTime = time.Now()
		}
		tracker.markAsEmbedded(note, modTime)
	}

	// Save tracker
	if err := tracker.save(em.embeddingsTrackingFile); err != nil {
		progress.UpdateProgress(alreadyEmbedded+len(notesToEmbed), totalNotes, "", false)
		return fmt.Errorf("saving tracker: %w", err)
	}

	slog.Info("Embedding completed",
		"embedded_notes", len(notesToEmbed),
		"duration", time.Since(start))

	// Mark as complete
	progress.UpdateProgress(alreadyEmbedded+len(notesToEmbed), totalNotes, "", false)

	return nil
}
