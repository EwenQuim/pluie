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

	// Unified search route - must be registered before the catch-all route
	fuego.Get(server, "/-/search", s.getUnifiedSearch,
		option.Query("q", "Search query for unified search (title, heading, semantic, AI)"),
	)

	// Unified search SSE stream route
	fuego.GetStd(server, "/-/search-stream", s.getUnifiedSearchStream)

	// Legacy search routes (kept for backward compatibility)
	fuego.Get(server, "/-/search-semantic", s.getSearch,
		option.Query("q", "Search query for semantic search only"),
	)

	fuego.Get(server, "/-/search-chat", s.getSearchChat,
		option.Query("q", "Question to ask about your notes"),
	)

	fuego.Get(server, "/-/search-live-chat", s.getSearchLiveChat,
		option.Query("q", "Question to ask about your notes with live streaming"),
	)

	fuego.GetStd(server, "/-/search-live-chat-stream", s.getSearchLiveChatStream)

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

// getUnifiedSearch handles the unified search page with immediate and lazy-loaded results
func (s *Server) getUnifiedSearch(ctx fuego.ContextNoBody) (fuego.Renderer, error) {
	query := ctx.QueryParam("q")

	if query == "" {
		slog.Info("Empty unified search query")
		return s.rs.UnifiedSearchResults(s.NotesService, "", nil, nil, nil)
	}

	// Get all notes for searching
	allNotes := s.NotesService.GetAllNotes()

	// Perform title search (limit to top 5)
	titleMatches := engine.SearchNotesByFilename(allNotes, query)
	if len(titleMatches) > 5 {
		titleMatches = titleMatches[:5]
	}

	// Track seen note slugs for deduplication
	seenSlugs := make(map[string]bool)
	var seenSlugsList []string
	for _, note := range titleMatches {
		seenSlugs[note.Slug] = true
		seenSlugsList = append(seenSlugsList, note.Slug)
	}

	// Perform heading search (limit to top 5, filter already-seen notes)
	allHeadingMatches := engine.SearchNotesByHeadings(allNotes, query, 0) // Get all first
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
	query := r.URL.Query().Get("q")
	seenParam := r.URL.Query().Get("seen")

	if query == "" {
		http.Error(w, "Missing query parameter 'q'", http.StatusBadRequest)
		return
	}

	// Parse seen slugs
	seenSlugs := make(map[string]bool)
	if seenParam != "" {
		for _, slug := range strings.Split(seenParam, ",") {
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
				fmt.Fprintf(w, ": keep-alive\n\n")
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
	if s.wvStore != nil {
		docs, err := s.wvStore.SimilaritySearch(r.Context(), query, 10) // Get 10, will filter to 5
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
	} else {
		slog.Warn("Vector store not available for unified search")
	}

	// Send semantic results if we have any
	if len(semanticResults) > 0 {
		html := template.RenderSemanticResultsHTML(s.rs, semanticResults)
		fmt.Fprintf(w, "event: semantic-results\ndata: %s\n\n", html)
		flusher.Flush()
		slog.Info("Sent semantic results", "query", query, "count", len(semanticResults))
	}

	// --- AI RESPONSE PHASE ---

	// Generate AI response if chat client is available
	if s.chatClient == nil {
		slog.Warn("Chat client not available for unified search")
	} else {
		// Collect all unique notes for context (title + heading + semantic)
		// Since we can't access title/heading matches here, use semantic results
		// In production, you might want to pass these via the request
		contextNotes := semanticResults
		if len(contextNotes) > 5 {
			contextNotes = contextNotes[:5]
		}

		if len(contextNotes) > 0 {
			// Build context from notes (limit to 300 chars per note)
			var contextBuilder strings.Builder
			contextBuilder.WriteString("Relevant notes:\n\n")
			for i, note := range contextNotes {
				contextBuilder.WriteString(fmt.Sprintf("%d. %s:\n", i+1, note.Title))
				content := note.Content
				if len(content) > 300 {
					content = content[:300] + "..."
				}
				contextBuilder.WriteString(content)
				contextBuilder.WriteString("\n\n")
			}

			// Build prompt
			prompt := fmt.Sprintf(`Answer this question based on the notes below.

Question: %s

%s

Answer concisely:`, query, contextBuilder.String())

			slog.Info("Generating unified search AI response", "query", query, "num_docs", len(contextNotes), "prompt", prompt)

			// Create streaming callback
			tokenCount := 0
			streamCallback := func(ctx context.Context, chunk []byte) error {
				tokenCount++
				if tokenCount == 1 {
					slog.Info("First AI token received", "query", query, "data", string(chunk))
				}
				fmt.Fprintf(w, "event: token\ndata: %s\n\n", string(chunk))
				flusher.Flush()
				return nil
			}

			// Generate response with streaming
			_, err := llms.GenerateFromSinglePrompt(
				r.Context(),
				s.chatClient,
				prompt,
				llms.WithModel(chatModel),
				llms.WithMaxTokens(512), // Shorter for unified search
				llms.WithTemperature(0.7),
				llms.WithStreamingFunc(streamCallback),
			)
			if err != nil {
				slog.Error("AI generation error", "error", err, "query", query)
				fmt.Fprintf(w, "event: error\ndata: AI generation failed\n\n")
				flusher.Flush()
				return
			}

			slog.Info("AI streaming completed", "query", query, "tokens", tokenCount)
		}
	}

	// Send completion event
	fmt.Fprintf(w, "event: done\ndata: Complete\n\n")
	flusher.Flush()
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

func (s *Server) getSearchLiveChat(ctx fuego.ContextNoBody) (fuego.Renderer, error) {
	query := ctx.QueryParam("q")
	return s.rs.SearchLiveChatResults(s.NotesService, query)
}

func (s *Server) getSearchLiveChatStream(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	if query == "" {
		http.Error(w, "Missing query parameter 'q'", http.StatusBadRequest)
		return
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
				fmt.Fprintf(w, ": keep-alive\n\n")
				flusher.Flush()
			case <-keepAliveDone:
				return
			case <-r.Context().Done():
				return
			}
		}
	}()

	// If vector store is not available, send error
	if s.wvStore == nil {
		slog.Warn("Vector store not available, cannot perform semantic search")
		fmt.Fprintf(w, "event: error\ndata: Vector store not available\n\n")
		flusher.Flush()
		return
	}

	// If chat client is not available, send error
	if s.chatClient == nil {
		slog.Warn("Chat client not available, cannot generate AI response")
		fmt.Fprintf(w, "event: error\ndata: Chat client not available\n\n")
		flusher.Flush()
		return
	}

	// Perform similarity search (limit to 3 documents to reduce context size)
	c := context.Background()
	docs, err := s.wvStore.SimilaritySearch(c, query, 3)
	if err != nil {
		slog.Error("Similarity search failed", "error", err, "query", query)
		fmt.Fprintf(w, "event: error\ndata: Search failed: %s\n\n", err.Error())
		flusher.Flush()
		return
	}

	slog.Info("Weaviate returned documents for live chat", "query", query, "doc_count", len(docs))

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

	slog.Info("Live chat search", "query", query, "weaviate_docs", len(docs), "results_found", len(searchResults))

	// Send the documents section as HTML (must be single line for SSE)
	var docsHTML string
	if len(searchResults) > 0 {
		docsHTML = fmt.Sprintf(`<div class="mb-6"><h3 class="text-lg font-semibold text-gray-700 mb-3">📚 Found %d relevant notes:</h3><div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">`, len(searchResults))

		for _, note := range searchResults {
			docsHTML += fmt.Sprintf(`<a href="/%s" class="block p-4 bg-white border border-gray-200 rounded-lg hover:shadow-md transition-shadow"><h4 class="font-semibold text-gray-900 mb-2">%s</h4><p class="text-sm text-gray-600 line-clamp-3">%s</p></a>`,
				note.Slug, note.Title, truncate(note.Content, 150))
		}

		docsHTML += `</div></div>`
	} else {
		docsHTML = `<div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6"><p class="text-yellow-800 mb-0">No relevant notes found. The AI will answer based on general knowledge.</p></div>`
	}

	// Send as SSE event
	fmt.Fprintf(w, "event: documents\ndata: %s\n\n", docsHTML)
	flusher.Flush()

	slog.Info("Sent documents HTML", "length", len(docsHTML), "num_results", len(searchResults))

	// Build context from search results (reduce content size to avoid token limits)
	var contextBuilder strings.Builder
	if len(searchResults) > 0 {
		contextBuilder.WriteString("Relevant notes:\n\n")
		for i, note := range searchResults {
			contextBuilder.WriteString(fmt.Sprintf("%d. %s:\n", i+1, note.Title))
			// Limit content to first 300 characters to avoid token limits
			content := note.Content
			if len(content) > 300 {
				content = content[:300] + "..."
			}
			contextBuilder.WriteString(content)
			contextBuilder.WriteString("\n\n")
		}
	}

	// Build a more concise prompt
	prompt := fmt.Sprintf(`Answer this question based on the notes below.

Question: %s

%s

Answer concisely:`, query, contextBuilder.String())

	slog.Info("Generating live chat response", "query", query, "num_docs", len(searchResults))

	// Create a streaming callback that sends tokens via SSE
	tokenCount := 0
	streamCallback := func(ctx context.Context, chunk []byte) error {
		tokenCount++
		if tokenCount == 1 {
			slog.Info("First token received from LLM", "query", query)
		}
		// Send each token as an SSE event
		fmt.Fprintf(w, "event: token\ndata: %s\n\n", string(chunk))
		flusher.Flush()
		return nil
	}

	// Log the prompt for debugging
	slog.Info("Sending prompt to LLM", "prompt_length", len(prompt), "query", query)

	// Generate response with streaming enabled - explicitly specify model
	_, err = llms.GenerateFromSinglePrompt(
		c,
		s.chatClient,
		prompt,
		llms.WithModel(chatModel),
		llms.WithMaxTokens(1024),
		llms.WithTemperature(0.7),
		llms.WithStreamingFunc(streamCallback),
	)
	if err != nil {
		slog.Error("streaming generation error", "error", err)
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		flusher.Flush()
		return
	}

	slog.Info("Streaming completed", "query", query, "tokens_sent", tokenCount)

	// Send completion event when done
	fmt.Fprintf(w, "event: done\ndata: Stream complete\n\n")
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
	progressChan := s.embeddingProgress.Subscribe()
	defer s.embeddingProgress.Unsubscribe(progressChan)

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

// truncate truncates a string to a maximum length, adding "..." if truncated
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
