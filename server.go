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
	NotesMap *map[string]model.Note // Slug -> Note
	Tree     *engine.TreeNode       // Tree structure of notes
	rs       template.Resource
	cfg      *config.Config
	mutex    sync.RWMutex // Protects data updates
}

// getHomeNoteSlug determines the home note slug based on priority:
// 1. HOME_NOTE_SLUG config value
// 2. First note in alphabetical order
func (s Server) getHomeNoteSlug() string {
	// Priority 1: Check HOME_NOTE_SLUG configuration
	if s.cfg.HomeNoteSlug != "" {
		s.mutex.RLock()
		notesMap := *s.NotesMap
		s.mutex.RUnlock()

		if _, exists := notesMap[s.cfg.HomeNoteSlug]; exists {
			return s.cfg.HomeNoteSlug
		}
	}

	// Priority 2: First note in alphabetical order
	s.mutex.RLock()
	tree := s.Tree
	s.mutex.RUnlock()
	
	if tree != nil {
		// Get all notes from tree and sort by slug
		notes := engine.GetAllNotesFromTree(tree)
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

func (s Server) Start() error {
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

	fuego.Get(server, "/{slug...}", s.getNote,
		option.Query("search", "Search query to filter notes by title"),
	)

	return server.Run()
}

func (s Server) getNote(ctx fuego.ContextNoBody) (fuego.Renderer, error) {
	slug := ctx.PathParam("slug")
	searchQuery := ctx.QueryParam("search")

	if slug == "" {
		slug = s.getHomeNoteSlug()
	}

	// Use read lock for concurrent access
	s.mutex.RLock()
	notesMap := *s.NotesMap
	s.mutex.RUnlock()

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

// UpdateData atomically updates the server's data structures
func (s *Server) UpdateData(notesMap *map[string]model.Note, tree *engine.TreeNode) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.NotesMap = notesMap
	s.Tree = tree
	s.rs = template.Resource{
		Tree: tree,
	}
}
