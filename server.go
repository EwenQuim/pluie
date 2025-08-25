package main

import (
	"fmt"
	"net/http"

	"github.com/EwenQuim/pluie/model"
	"github.com/EwenQuim/pluie/static"
	"github.com/EwenQuim/pluie/template"

	"github.com/go-fuego/fuego"
)

type Server struct {
	NotesMap map[string]model.Note // Slug -> Note
	rs       template.Resource
}

func (s Server) Start() error {
	server := fuego.NewServer()

	// Serve static files at /static
	server.Mux.Handle("GET /static/", http.StripPrefix("/static", static.Handler()))

	fuego.Get(server, "/{slug...}", func(ctx fuego.ContextNoBody) (fuego.Renderer, error) {
		slug := ctx.PathParam("slug")
		if slug == "" {
			// Handle search query for home page
			searchQuery := ctx.QueryParam("search")
			return s.rs.ListWithSearch(searchQuery), nil
		}

		note, ok := s.NotesMap[slug]
		if !ok {
			return nil, fmt.Errorf("Note with slug %s not found", slug)

		}
		return s.rs.Note(note)

	})

	return server.Run()
}
