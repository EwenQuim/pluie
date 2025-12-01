package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

const (
	embeddingsTrackingFile = "embeddings_tracking.json"
	embeddingsModel        = "nomic-embed-text" // Fast, efficient embedding model
)

// EmbeddedFile tracks a file that has been embedded in the Vector Store
type EmbeddedFile struct {
	Path         string    `json:"path"`
	ContentHash  string    `json:"content_hash"`
	EmbeddedAt   time.Time `json:"embedded_at"`
	LastModified time.Time `json:"last_modified"`
}

// EmbeddingsTracker manages the tracking of embedded files
type EmbeddingsTracker struct {
	Files map[string]EmbeddedFile `json:"files"`
}

// loadEmbeddingsTracker loads the tracking file or creates a new one
func loadEmbeddingsTracker() (*EmbeddingsTracker, error) {
	tracker := &EmbeddingsTracker{
		Files: make(map[string]EmbeddedFile),
	}

	data, err := os.ReadFile(embeddingsTrackingFile)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("No existing embeddings tracking file, starting fresh")
			return tracker, nil
		}
		return nil, fmt.Errorf("reading tracking file: %w", err)
	}

	if err := json.Unmarshal(data, tracker); err != nil {
		return nil, fmt.Errorf("parsing tracking file: %w", err)
	}

	slog.Info("Loaded embeddings tracker", "tracked_files", len(tracker.Files))
	return tracker, nil
}

// saveEmbeddingsTracker saves the tracking file
func (t *EmbeddingsTracker) save() error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling tracker: %w", err)
	}

	if err := os.WriteFile(embeddingsTrackingFile, data, 0644); err != nil {
		return fmt.Errorf("writing tracking file: %w", err)
	}

	slog.Info("Saved embeddings tracker", "tracked_files", len(t.Files))
	return nil
}

// computeContentHash computes a SHA256 hash of the note content
func computeContentHash(note model.Note) string {
	// Include title and content in hash to detect changes
	data := fmt.Sprintf("%s\n%s", note.Title, note.Content)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// needsEmbedding checks if a note needs to be embedded
func (t *EmbeddingsTracker) needsEmbedding(note model.Note) bool {
	embedded, exists := t.Files[note.Path]
	if !exists {
		return true
	}

	currentHash := computeContentHash(note)
	return embedded.ContentHash != currentHash
}

// markAsEmbedded marks a note as embedded in the tracker
func (t *EmbeddingsTracker) markAsEmbedded(note model.Note, lastModified time.Time) {
	t.Files[note.Path] = EmbeddedFile{
		Path:         note.Path,
		ContentHash:  computeContentHash(note),
		EmbeddedAt:   time.Now(),
		LastModified: lastModified,
	}
}

// VectorStore defines the interface for vector storage operations
type VectorStore interface {
	// AddDocuments adds documents to the vector store
	AddDocuments(ctx context.Context, docs []schema.Document, options ...vectorstores.Option) ([]string, error)
	// SimilaritySearch performs a similarity search on the vector store
	SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]schema.Document, error)
}

// EmbeddingsManager handles all embeddings-related functionality
type EmbeddingsManager struct {
	store        VectorStore        // Vector store for semantic search
	progress     *EmbeddingProgress // Tracks embedding progress for SSE updates
	initOnce     sync.Once          // Ensures embeddings are initialized only once
	initialized  bool               // Tracks if embeddings have been initialized
	initMutex    sync.RWMutex       // Protects initialized flag
	notesService *engine.NotesService
}

// NewEmbeddingsManager creates a new EmbeddingsManager
func NewEmbeddingsManager(store VectorStore, progress *EmbeddingProgress, notesService *engine.NotesService) *EmbeddingsManager {
	return &EmbeddingsManager{
		store:        store,
		progress:     progress,
		notesService: notesService,
	}
}

// InitializeLazily initializes embeddings on first call (thread-safe)
func (em *EmbeddingsManager) InitializeLazily() {
	// Only initialize if store is available
	if em.store == nil {
		return
	}

	em.initOnce.Do(func() {
		slog.Info("Lazy-loading embeddings: triggered by search page access")

		// Mark as initialized immediately to prevent concurrent calls
		em.initMutex.Lock()
		em.initialized = true
		em.initMutex.Unlock()

		// Embed notes into vector store in background
		go func() {
			ctx := context.Background()
			allNotes := em.notesService.GetAllNotes()
			if err := embedNotesWithProgress(ctx, em.store, allNotes, em.progress); err != nil {
				slog.Error("Error embedding notes", "error", err)
				// Continue anyway - the server can still work without embeddings
			}
		}()
	})
}

// GetStore returns the vector store
func (em *EmbeddingsManager) GetStore() VectorStore {
	if em == nil {
		return nil
	}
	return em.store
}

// GetProgress returns the embedding progress tracker
func (em *EmbeddingsManager) GetProgress() *EmbeddingProgress {
	if em == nil {
		return nil
	}
	return em.progress
}
