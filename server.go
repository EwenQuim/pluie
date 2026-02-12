package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	"github.com/EwenQuim/pluie/static"
	"github.com/EwenQuim/pluie/template"
	"github.com/tmc/langchaingo/llms"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
)

type HealthResponse struct {
	Status string `json:"status"`
}

type Server struct {
	NotesService      *engine.NotesService
	rs                template.Resource
	cfg               *config.Config
	chatClient        llms.Model         // Chat client for AI responses
	embeddingsManager *EmbeddingsManager // Manages all embeddings functionality
}

// UpdateData safely updates the server's NotesMap, Tree, and TagIndex with new data
func (s *Server) UpdateData(notesMap *map[string]model.Note, tree *engine.TreeNode, tagIndex engine.TagIndex) {
	s.NotesService.UpdateData(notesMap, tree, tagIndex)
}

func (s *Server) registerRoutes(server *fuego.Server) {
	// Serve static files at /static
	server.Mux.Handle("GET /static/", http.StripPrefix("/static", static.Handler()))

	// Health check endpoint for Docker/K8s probes
	fuego.Get(server, "/-/health", func(c fuego.ContextNoBody) (HealthResponse, error) {
		return HealthResponse{Status: "ok"}, nil
	}, option.Summary("health"), option.Tags("Health"))

	// Unified search route - must be registered before the catch-all route
	fuego.Get(server, "/-/search", s.getUnifiedSearch,
		option.Query("q", "Search query for unified search (title, heading, semantic, AI)"),
	)

	// Unified search SSE stream route
	fuego.GetStd(server, "/-/search-stream", s.getUnifiedSearchStream)

	// Embedding progress SSE route
	fuego.GetStd(server, "/-/embedding-progress", s.getEmbeddingProgress)

	// Tag route - must be registered before the catch-all route
	fuego.Get(server, "/-/tag/{tag...}", s.getTag)

	fuego.Get(server, "/{slug...}", s.getNote,
		option.Query("search", "Search query to filter notes by title"),
	)
}

func (s *Server) Start(ctx context.Context) error {
	server := fuego.NewServer(
		fuego.WithAddr(":"+s.cfg.Port),
		fuego.WithEngineOptions(
			fuego.WithOpenAPIConfig(fuego.OpenAPIConfig{
				DisableLocalSave: true,
			}),
		),
	)

	s.registerRoutes(server)

	// Start server in a goroutine so we can listen for shutdown signal
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Run()
	}()

	// Wait for shutdown signal or server error
	select {
	case <-ctx.Done():
		slog.Info("Shutdown signal received, draining connections...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("Server shutdown error", "error", err)
			return err
		}
		slog.Info("Server shut down gracefully")
		return nil
	case err := <-errCh:
		return err
	}
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

// getUnifiedSearch handles the unified search page with immediate and lazy-loaded results
func (s *Server) getUnifiedSearch(ctx fuego.ContextNoBody) (fuego.Renderer, error) {
	// Trigger lazy initialization of embeddings on first search access
	if s.embeddingsManager != nil {
		s.embeddingsManager.InitializeLazily()
	}

	query := ctx.QueryParam("q")

	if query == "" {
		slog.Info("Empty unified search query")
		return s.rs.UnifiedSearchResults(s.NotesService, "", nil, nil, nil)
	}

	// Perform title search (limit to top 5)
	titleMatches := s.NotesService.SearchNotesByFilename(query, 5)

	// Track seen note slugs for deduplication
	seenSlugs := make(map[string]bool)
	var seenSlugsList []string
	for _, note := range titleMatches {
		seenSlugs[note.Slug] = true
		seenSlugsList = append(seenSlugsList, note.Slug)
	}

	// Perform heading search (limit to top 5, filter already-seen notes)
	allHeadingMatches := s.NotesService.SearchNotesByHeadings(query, 0) // Get all first
	var headingMatches []engine.HeadingMatch
	for _, match := range allHeadingMatches {
		if !seenSlugs[match.Note.Slug] {
			headingMatches = append(headingMatches, match)
			seenSlugs[match.Note.Slug] = true
			seenSlugsList = append(seenSlugsList, match.Note.Slug)

			if len(headingMatches) >= 5 {
				break
			}
		}
	}

	slog.Info("Unified search",
		"query", query,
		"title_matches", len(titleMatches),
		"heading_matches", len(headingMatches),
		"seen_slugs", len(seenSlugsList))

	return s.rs.UnifiedSearchResults(s.NotesService, query, titleMatches, headingMatches, seenSlugsList)
}

// getUnifiedSearchStream handles SSE streaming for semantic search and AI response
func (s *Server) getUnifiedSearchStream(w http.ResponseWriter, r *http.Request) {
	// Trigger lazy initialization of embeddings on first search access
	if s.embeddingsManager != nil {
		s.embeddingsManager.InitializeLazily()
	}

	query := r.URL.Query().Get("q")
	seenParam := r.URL.Query().Get("seen")

	if query == "" {
		http.Error(w, "Missing query parameter 'q'", http.StatusBadRequest)
		return
	}

	// Parse seen slugs
	seenSlugs := make(map[string]bool)
	if seenParam != "" {
		for slug := range strings.SplitSeq(seenParam, ",") {
			if slug != "" {
				seenSlugs[slug] = true
			}
		}
	}

	// Set headers for Server-Sent Events
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Set write deadline to 5 minutes for long-running SSE connections
	rc := http.NewResponseController(w)
	if err := rc.SetWriteDeadline(time.Now().Add(5 * time.Minute)); err != nil {
		slog.Warn("Failed to set write deadline", "error", err)
	}

	// Start keep-alive ticker to prevent timeout
	keepAliveTicker := time.NewTicker(15 * time.Second)
	defer keepAliveTicker.Stop()

	keepAliveDone := make(chan bool)
	defer close(keepAliveDone)

	go func() {
		for {
			select {
			case <-keepAliveTicker.C:
				if _, err := fmt.Fprintf(w, ": keep-alive\n\n"); err != nil {
					slog.Debug("SSE keep-alive write failed (client likely disconnected)", "error", err)
					return
				}
				flusher.Flush()
			case <-keepAliveDone:
				return
			case <-r.Context().Done():
				return
			}
		}
	}()

	// --- SEMANTIC SEARCH PHASE ---

	// If vector store is available, perform semantic search
	var semanticResults []model.Note
	vectorStore := s.embeddingsManager.GetStore()
	if vectorStore == nil {
		slog.Warn("Vector store not available for unified search")
	} else {
		docs, err := vectorStore.SimilaritySearch(r.Context(), query, 10) // Get 10, will filter to 5
		if err != nil {
			slog.Error("Similarity search failed", "error", err, "query", query)
		} else {
			slog.Info("Weaviate returned documents for unified search", "query", query, "doc_count", len(docs))

			// Convert documents to notes using metadata
			notesMap := s.NotesService.GetNotesMap()
			for _, doc := range docs {
				if slug, ok := doc.Metadata["slug"].(string); ok {
					if note, exists := notesMap[slug]; exists {
						// Only add if not already seen
						if !seenSlugs[note.Slug] {
							semanticResults = append(semanticResults, note)
							seenSlugs[note.Slug] = true

							// Stop at 5 results
							if len(semanticResults) >= 5 {
								break
							}
						}
					}
				}
			}
		}
	}

	// Send semantic results if we have any
	if len(semanticResults) > 0 {
		html := template.RenderSemanticResultsHTML(s.rs, semanticResults)
		if _, err := fmt.Fprintf(w, "event: semantic-results\ndata: %s\n\n", html); err != nil {
			slog.Debug("SSE semantic results write failed", "error", err, "query", query)
			return
		}
		flusher.Flush()
		slog.Info("Sent semantic results", "query", query, "count", len(semanticResults))
	}

	// --- AI RESPONSE PHASE ---

	// Generate AI response if chat client is available
	if s.chatClient == nil {
		slog.Warn("Chat client not available for unified search")
	} else {
		// Collect all unique notes for context (title + heading + semantic)
		// Re-perform title and heading searches to get all relevant notes

		// Get title matches (no limit - get all)
		titleMatches := s.NotesService.SearchNotesByFilename(query, 10)

		// Get heading matches (no limit - get all)
		headingMatches := s.NotesService.SearchNotesByHeadings(query, 10)

		// Combine all results: title, heading, then semantic
		contextNotes := make([]model.Note, 0, 10)
		contextSlugs := make(map[string]bool)

		// Add title matches first
		for _, note := range titleMatches {
			if !contextSlugs[note.Slug] {
				contextNotes = append(contextNotes, note)
				contextSlugs[note.Slug] = true
			}
		}

		// Add heading matches
		for _, match := range headingMatches {
			if !contextSlugs[match.Note.Slug] {
				contextNotes = append(contextNotes, match.Note)
				contextSlugs[match.Note.Slug] = true
			}
		}

		// Add semantic matches
		for _, note := range semanticResults {
			if !contextSlugs[note.Slug] {
				contextNotes = append(contextNotes, note)
				contextSlugs[note.Slug] = true
			}
		}

		// Limit to 15 notes to maximize 2K token context
		if len(contextNotes) > 10 {
			contextNotes = contextNotes[:10]
		}

		slog.Info("Combined context notes for AI response",
			"query", query,
			"title_matches", len(titleMatches),
			"heading_matches", len(headingMatches),
			"semantic_matches", len(semanticResults),
			"total_context_notes", len(contextNotes))

		if len(contextNotes) > 0 {
			// Build context from notes (limit to ~600 chars per note to maximize 2K token usage)
			var contextBuilder strings.Builder
			contextBuilder.WriteString("Relevant notes:\n\n")
			for i, note := range contextNotes {
				contextBuilder.WriteString(fmt.Sprintf("%d. %s:\n", i+1, note.Title))
				content := note.Content
				if len(content) > 600 {
					content = content[:600] + "..."
				}
				contextBuilder.WriteString(content)
				contextBuilder.WriteString("\n\n")
			}

			userPrompt := contextBuilder.String()

			// Build prompt
			prompt := fmt.Sprintf(`Answer this question based on the notes below.

Question: %s

%s

Answer concisely:`, query, userPrompt)

			slog.Info("Generating unified search AI response", "query", query, "context_size", len(userPrompt), "user_prompt", userPrompt)

			// Create streaming callback
			tokenCount := 0
			streamCallback := func(ctx context.Context, chunk []byte) error {
				tokenCount++
				if tokenCount == 1 {
					slog.Info("First AI token received", "query", query, "data", string(chunk))
				}
				if _, err := fmt.Fprintf(w, "event: token\ndata: %s\n\n", string(chunk)); err != nil {
					slog.Debug("SSE token write failed", "error", err, "query", query)
					return err
				}
				flusher.Flush()
				return nil
			}

			// Generate response with streaming
			_, err := llms.GenerateFromSinglePrompt(
				r.Context(),
				s.chatClient,
				prompt,
				llms.WithMaxTokens(512), // Shorter for unified search
				llms.WithTemperature(0.7),
				llms.WithStreamingFunc(streamCallback),
			)
			if err != nil {
				slog.Error("AI generation error", "error", err, "query", query)
				if _, writeErr := fmt.Fprintf(w, "event: error\ndata: AI generation failed\n\n"); writeErr != nil {
					slog.Debug("SSE error write failed", "error", writeErr, "query", query)
				}
				flusher.Flush()
				return
			}

			slog.Info("AI streaming completed", "query", query, "tokens", tokenCount)
		}
	}

	// Send completion event
	if _, err := fmt.Fprintf(w, "event: done\ndata: Complete\n\n"); err != nil {
		slog.Debug("SSE done write failed", "error", err, "query", query)
		return
	}
	flusher.Flush()
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
	embeddingProgress := s.embeddingsManager.GetProgress()
	if embeddingProgress == nil {
		http.Error(w, "Embedding progress not available", http.StatusServiceUnavailable)
		return
	}

	progressChan := embeddingProgress.Subscribe()
	defer embeddingProgress.Unsubscribe(progressChan)

	// Create a ticker for periodic updates (every 10 seconds)
	ticker := time.NewTicker(10 * time.Second)
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
		if _, err := fmt.Fprint(w, "data: "); err != nil {
			slog.Debug("SSE embedding progress write failed", "error", err)
			return
		}
		if err := progressNode.Render(w); err != nil {
			slog.Debug("SSE embedding progress render failed", "error", err)
			return
		}
		if _, err := fmt.Fprint(w, "\n\n"); err != nil {
			slog.Debug("SSE embedding progress write failed", "error", err)
			return
		}
		flusher.Flush()
	}

	// Send initial status immediately
	sendUpdate(embeddingProgress.GetStatus())

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			// Send periodic update
			sendUpdate(embeddingProgress.GetStatus())
		case status := <-progressChan:
			// Send update when progress changes
			sendUpdate(status)
		}
	}
}
