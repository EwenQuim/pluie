package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	"github.com/EwenQuim/pluie/static"
	"github.com/EwenQuim/pluie/template"
	"github.com/tmc/langchaingo/vectorstores/weaviate"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
)

type Server struct {
	NotesService *engine.NotesService
	rs           template.Resource
	cfg          *config.Config
	wvStore      *weaviate.Store // Vector store for semantic search
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

	slog.Info("Semantic search", "query", query, "results_found", len(searchResults))

	return s.rs.SearchResults(s.NotesService, query, searchResults)
}
