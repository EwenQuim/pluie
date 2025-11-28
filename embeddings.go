package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/EwenQuim/pluie/model"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/weaviate"
)

const (
	embeddingsTrackingFile = "embeddings_tracking.json"
	embeddingsModel        = "nomic-embed-text" // Fast, efficient embedding model
)

// EmbeddedFile tracks a file that has been embedded in Weaviate
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

// initializeWeaviateStore creates and initializes the Weaviate store
func initializeWeaviateStore() (*weaviate.Store, error) {
	// Get Weaviate configuration from environment or use defaults
	wvHost := os.Getenv("WEAVIATE_HOST")
	if wvHost == "" {
		wvHost = "weaviate-embeddings:9035" // Default from docker-compose
	}

	wvScheme := os.Getenv("WEAVIATE_SCHEME")
	if wvScheme == "" {
		wvScheme = "http"
	}

	indexName := os.Getenv("WEAVIATE_INDEX")
	if indexName == "" {
		indexName = "Note"
	}

	slog.Info("Initializing Weaviate store",
		"host", wvHost,
		"scheme", wvScheme,
		"index", indexName,
		"embeddings_model", embeddingsModel)

	// Create Ollama client for embeddings
	embeddingsClient, err := ollama.New(
		ollama.WithServerURL("http://ollama-models:11434"),
		ollama.WithModel(embeddingsModel),
	)
	if err != nil {
		return nil, fmt.Errorf("creating ollama client: %w", err)
	}

	emb, err := embeddings.NewEmbedder(embeddingsClient)
	if err != nil {
		return nil, fmt.Errorf("creating embedder: %w", err)
	}

	// Create Weaviate store
	wvStore, err := weaviate.New(
		weaviate.WithEmbedder(emb),
		weaviate.WithScheme(wvScheme),
		weaviate.WithHost(wvHost),
		weaviate.WithIndexName(indexName),
		// Specify which metadata fields to retrieve during similarity search
		weaviate.WithQueryAttrs([]string{"text", "nameSpace", "title", "path", "slug"}),
	)
	if err != nil {
		return nil, fmt.Errorf("creating weaviate store: %w", err)
	}

	return &wvStore, nil
}

// embedNotes embeds notes into Weaviate, only processing new or changed notes
func embedNotes(ctx context.Context, wvStore *weaviate.Store, notes []model.Note) error {
	start := time.Now()

	// Load tracking file
	tracker, err := loadEmbeddingsTracker()
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

	if len(notesToEmbed) == 0 {
		slog.Info("No new notes to embed, all notes are up to date",
			"total_notes", len(notes),
			"tracked_notes", len(tracker.Files))
		return nil
	}

	slog.Info("Notes embedding status",
		"total_notes", len(notes),
		"already_embedded", len(notes)-len(notesToEmbed),
		"to_embed", len(notesToEmbed))

	// Add documents to Weaviate one at a time to show real progress
	slog.Info("Starting embedding process", "documents", len(notesToEmbed))

	for i, note := range notesToEmbed {
		docStart := time.Now()

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

		_, err = wvStore.AddDocuments(ctx, []schema.Document{doc})
		if err != nil {
			return fmt.Errorf("adding document to weaviate (title=%s, path=%s): %w", note.Title, note.Path, err)
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
	if err := tracker.save(); err != nil {
		return fmt.Errorf("saving tracker: %w", err)
	}

	slog.Info("Embedding completed",
		"embedded_notes", len(notesToEmbed),
		"duration", time.Since(start))

	return nil
}
