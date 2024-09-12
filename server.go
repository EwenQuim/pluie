package main

import (
	"fmt"

	"github.com/EwenQuim/pluie/model"
	"github.com/EwenQuim/pluie/template"

	"github.com/go-fuego/fuego"
)

type Server struct {
	NotesMap map[string]model.Note // Slug -> Note
	rs       template.Resource
}

func (s Server) Start() error {
	server := fuego.NewServer()

	fuego.Get(server, "/{slug...}", func(ctx *fuego.ContextNoBody) (fuego.Renderer, error) {
		slug := ctx.PathParam("slug")
		if slug == "" {
			return s.rs.Home(), nil
		}

		note, ok := s.NotesMap[slug]
		if !ok {
			return nil, fmt.Errorf("Note with slug %s not found", slug)

		}
		return s.rs.Note(note), nil

	})

	return server.Run()
}
