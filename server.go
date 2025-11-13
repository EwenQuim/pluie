package main

import (
	"log/slog"
	"net/http"
	"sort"
	"sync"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	"github.com/EwenQuim/pluie/static"
	"github.com/EwenQuim/pluie/template"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
)

type Server struct {
	mu       sync.RWMutex           // Protects NotesMap, Tree, and TagIndex
	NotesMap *map[string]model.Note // Slug -> Note
	Tree     *engine.TreeNode       // Tree structure of notes
	TagIndex engine.TagIndex        // Tag -> Notes mapping
	rs       template.Resource
	cfg      *config.Config
}

// getHomeNoteSlug determines the home note slug based on priority:
// 1. HOME_NOTE_SLUG config value
// 2. First note in alphabetical order
func (s *Server) getHomeNoteSlug() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Priority 1: Check HOME_NOTE_SLUG configuration
	if s.cfg.HomeNoteSlug != "" {
		notesMap := *s.NotesMap

		if _, exists := notesMap[s.cfg.HomeNoteSlug]; exists {
			return s.cfg.HomeNoteSlug
		}
	}

	// Priority 2: First note in alphabetical order
	if s.Tree != nil {
		// Get all notes from tree and sort by slug
		notes := engine.GetAllNotesFromTree(s.Tree)
		if len(notes) > 0 {
			sort.Slice(notes, func(i, j int) bool {
				return notes[i].Slug < notes[j].Slug
			})
			return notes[0].Slug
		}
	}

	// Fallback (should not happen if there are notes)
	return ""
}

// UpdateData safely updates the server's NotesMap, Tree, and TagIndex with new data
func (s *Server) UpdateData(notesMap *map[string]model.Note, tree *engine.TreeNode, tagIndex engine.TagIndex) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.NotesMap = notesMap
	s.Tree = tree
	s.TagIndex = tagIndex

	// Also update the resource tree for rendering
	s.rs.Tree = tree

	slog.Info("Server data updated", "notes_count", len(*notesMap))
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
		slug = s.getHomeNoteSlug()
	}

	s.mu.RLock()
	notesMap := *s.NotesMap
	s.mu.RUnlock()

	note, ok := notesMap[slug]
	if !ok {
		slog.Info("Note not found", "slug", slug)
		return s.rs.NoteWithList(nil, searchQuery)
	}

	// Additional security check: ensure note is public
	if !s.cfg.PublicByDefault && !note.IsPublic {
		slog.Info("Private note access denied", "slug", slug)
		return s.rs.NoteWithList(nil, searchQuery)
	}

	return s.rs.NoteWithList(&note, searchQuery)
}

func (s *Server) getTag(ctx fuego.ContextNoBody) (fuego.Renderer, error) {
	tag := ctx.PathParam("tag")

	if tag == "" {
		slog.Info("Empty tag parameter")
		return s.rs.TagList("", nil)
	}

	s.mu.RLock()
	tagIndex := s.TagIndex
	s.mu.RUnlock()

	// Get all notes that contain this tag
	notesWithTag := tagIndex.GetNotesWithTag(tag)

	// Also get all tags that contain this tag as a substring
	relatedTags := tagIndex.GetTagsContaining(tag)

	slog.Info("Tag search", "tag", tag, "notes_found", len(notesWithTag), "related_tags", len(relatedTags))

	return s.rs.TagList(tag, notesWithTag)
}
