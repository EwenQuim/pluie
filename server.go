package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	"github.com/EwenQuim/pluie/static"
	"github.com/EwenQuim/pluie/template"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/vectorstores/weaviate"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
)

type Server struct {
	NotesService      *engine.NotesService
	rs                template.Resource
	cfg               *config.Config
	wvStore           *weaviate.Store    // Vector store for semantic search
	chatClient        llms.Model         // Chat client for AI responses
	embeddingProgress *EmbeddingProgress // Tracks embedding progress for SSE updates
}

// UpdateData safely updates the server's NotesMap, Tree, and TagIndex with new data
func (s *Server) UpdateData(notesMap *map[string]model.Note, tree *engine.TreeNode, tagIndex engine.TagIndex) {
	s.NotesService.UpdateData(notesMap, tree, tagIndex)
}

func (s *Server) Start() error {
	server := fuego.NewServer(
		fuego.WithAddr(":"+s.cfg.Port),
		fuego.WithEngineOptions(
			fuego.WithOpenAPIConfig(fuego.OpenAPIConfig{
				DisableLocalSave: true,
			}),
		),
	)

	// Serve static files at /static
	server.Mux.Handle("GET /static/", http.StripPrefix("/static", static.Handler()))

	// Search route - must be registered before the catch-all route
	fuego.Get(server, "/-/search", s.getSearch,
		option.Query("q", "Search query for semantic search"),
	)

	// Search chat route - must be registered before the catch-all route
	fuego.Get(server, "/-/search-chat", s.getSearchChat,
		option.Query("q", "Question to ask about your notes"),
	)

	// Embedding progress SSE route
	fuego.GetStd(server, "/-/embedding-progress", s.getEmbeddingProgress)

	// Tag route - must be registered before the catch-all route
	fuego.Get(server, "/-/tag/{tag...}", s.getTag)

	fuego.Get(server, "/{slug...}", s.getNote,
		option.Query("search", "Search query to filter notes by title"),
	)

	return server.Run()
}

func (s *Server) getNote(ctx fuego.ContextNoBody) (fuego.Renderer, error) {
	slug := ctx.PathParam("slug")
	searchQuery := ctx.QueryParam("search")

	if slug == "" {
		slug = s.NotesService.GetHomeSlug(s.cfg.HomeNoteSlug)
	}

	note, ok := s.NotesService.GetNote(slug)
	if !ok {
		slog.Info("Note not found", "slug", slug)
		return s.rs.NoteWithList(s.NotesService, nil, searchQuery)
	}

	// Additional security check: ensure note is public
	if !s.cfg.PublicByDefault && !note.IsPublic {
		slog.Info("Private note access denied", "slug", slug)
		return s.rs.NoteWithList(s.NotesService, nil, searchQuery)
	}

	return s.rs.NoteWithList(s.NotesService, &note, searchQuery)
}

func (s *Server) getTag(ctx fuego.ContextNoBody) (fuego.Renderer, error) {
	tag := ctx.PathParam("tag")

	if tag == "" {
		slog.Info("Empty tag parameter")
		return s.rs.TagList(s.NotesService, "", nil)
	}

	tagIndex := s.NotesService.GetTagIndex()

	// Get all notes that contain this tag
	notesWithTag := tagIndex.GetNotesWithTag(tag)

	// Also get all tags that contain this tag as a substring
	relatedTags := tagIndex.GetTagsContaining(tag)

	slog.Info("Tag search", "tag", tag, "notes_found", len(notesWithTag), "related_tags", len(relatedTags))

	return s.rs.TagList(s.NotesService, tag, notesWithTag)
}

func (s *Server) getSearch(ctx fuego.ContextNoBody) (fuego.Renderer, error) {
	query := ctx.QueryParam("q")

	if query == "" {
		slog.Info("Empty search query")
		return s.rs.SearchResults(s.NotesService, "", nil)
	}

	// If vector store is not available, return empty results
	if s.wvStore == nil {
		slog.Warn("Vector store not available, cannot perform semantic search")
		return s.rs.SearchResults(s.NotesService, query, nil)
	}

	// Perform similarity search
	c := context.Background()
	docs, err := s.wvStore.SimilaritySearch(c, query, 10)
	if err != nil {
		slog.Error("Similarity search failed", "error", err, "query", query)
		return s.rs.SearchResults(s.NotesService, query, nil)
	}

	slog.Info("Weaviate returned documents", "query", query, "doc_count", len(docs))

	// Convert documents to notes using metadata
	var searchResults []model.Note
	notesMap := s.NotesService.GetNotesMap()

	for _, doc := range docs {
		// Try to get the slug from metadata
		if slug, ok := doc.Metadata["slug"].(string); ok {
			if note, exists := notesMap[slug]; exists {
				searchResults = append(searchResults, note)
			}
		}
	}

	slog.Info("Semantic search", "query", query, "weaviate_docs", len(docs), "results_found", len(searchResults))

	return s.rs.SearchResults(s.NotesService, query, searchResults)
}

func (s *Server) getSearchChat(ctx fuego.ContextNoBody) (fuego.Renderer, error) {
	query := ctx.QueryParam("q")

	if query == "" {
		slog.Info("Empty search chat query")
		return s.rs.SearchChatResults(s.NotesService, "", nil, "")
	}

	// If vector store is not available, return empty results
	if s.wvStore == nil {
		slog.Warn("Vector store not available, cannot perform semantic search")
		return s.rs.SearchChatResults(s.NotesService, query, nil, "")
	}

	// Perform similarity search
	c := context.Background()
	docs, err := s.wvStore.SimilaritySearch(c, query, 10)
	if err != nil {
		slog.Error("Similarity search failed", "error", err, "query", query)
		return s.rs.SearchChatResults(s.NotesService, query, nil, "")
	}

	slog.Info("Weaviate returned documents for chat", "query", query, "doc_count", len(docs))

	// Convert documents to notes using metadata
	var searchResults []model.Note
	notesMap := s.NotesService.GetNotesMap()

	for _, doc := range docs {
		// Try to get the slug from metadata
		if slug, ok := doc.Metadata["slug"].(string); ok {
			if note, exists := notesMap[slug]; exists {
				searchResults = append(searchResults, note)
			}
		}
	}

	slog.Info("Chat search", "query", query, "weaviate_docs", len(docs), "results_found", len(searchResults))

	// Generate AI response if we have results and chat client is available
	var aiResponse string
	if len(searchResults) > 0 && s.chatClient != nil {
		aiResponse, err = generateChatResponse(c, s.chatClient, query, searchResults)
		if err != nil {
			slog.Error("Failed to generate chat response", "error", err, "query", query)
			// Continue without AI response rather than failing completely
		}
	} else if s.chatClient == nil {
		slog.Warn("Chat client not available, cannot generate AI response")
	}

	return s.rs.SearchChatResults(s.NotesService, query, searchResults, aiResponse)
}

func (s *Server) getEmbeddingProgress(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Get flusher for SSE
	flusher, ok := w.(http.Flusher)
	if !ok {
		slog.Error("Streaming not supported")
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Subscribe to embedding progress updates
	progressChan := s.embeddingProgress.Subscribe()
	defer s.embeddingProgress.Unsubscribe(progressChan)

	// Create a ticker for periodic updates (every 3 seconds)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	// Helper function to send HTML update using gomponent
	sendUpdate := func(status EmbeddingStatus) {
		// Create progress data
		data := template.EmbeddingProgressData{
			Embedded:    status.EmbeddedNotes,
			Total:       status.TotalNotes,
			IsEmbedding: status.IsEmbedding,
		}

		// Render the progress content using the SAME gomponent as in navbar
		progressNode := template.RenderEmbeddingProgressContent(data)

		// Write SSE message
		fmt.Fprint(w, "data: ")
		progressNode.Render(w)
		fmt.Fprint(w, "\n\n")
		flusher.Flush()
	}

	// Send initial status immediately
	sendUpdate(s.embeddingProgress.GetStatus())

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			// Send periodic update
			sendUpdate(s.embeddingProgress.GetStatus())
		case status := <-progressChan:
			// Send update when progress changes
			sendUpdate(status)
		}
	}
}
