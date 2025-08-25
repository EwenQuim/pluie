package template

import (
	"strings"

	"github.com/EwenQuim/pluie/model"
	"github.com/maragudk/gomponents"
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
)

type Resource struct {
	Notes []model.Note
}

// Displays the list of notes with optional search filtering
func (rs Resource) List() gomponents.Node {
	return rs.ListWithSearch("")
}

// Displays the list of notes with search functionality
func (rs Resource) ListWithSearch(searchQuery string) gomponents.Node {
	filteredNotes := rs.Notes

	// Filter notes by title if search query is provided
	if searchQuery != "" {
		filteredNotes = make([]model.Note, 0)
		searchLower := strings.ToLower(searchQuery)
		for _, note := range rs.Notes {
			if strings.Contains(strings.ToLower(note.Title), searchLower) {
				filteredNotes = append(filteredNotes, note)
			}
		}
	}

	return rs.Layout(
		H1(g.Text("Obsidian")),
		// Search form
		Form(
			Method("GET"),
			Action("/"),
			Class("mb-4"),
			Div(
				Class("flex gap-2"),
				Input(
					Type("text"),
					Name("search"),
					Placeholder("Search notes by title..."),
					Value(searchQuery),
					Class("flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"),
				),
				Button(
					Type("submit"),
					Class("px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500"),
					g.Text("Search"),
				),
			),
		),
		// Results info
		g.If(searchQuery != "",
			P(
				Class("text-gray-600 mb-2"),
				g.Textf("Found %d notes matching \"%s\"", len(filteredNotes), searchQuery),
			),
		),
		// Notes list
		Ul(
			g.Group(g.Map(filteredNotes, func(note model.Note) gomponents.Node {
				return Li(
					Class("mb-1"),
					A(
						Href("/"+note.Slug),
						Class("text-blue-600 hover:text-blue-800 hover:underline"),
						g.Text(note.Title),
					),
				)
			}),
			),
		),
		// Show message if no results found
		g.If(searchQuery != "" && len(filteredNotes) == 0,
			P(
				Class("text-gray-500 mt-4"),
				g.Text("No notes found matching your search."),
			),
		),
	)
}
