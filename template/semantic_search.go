package template

import (
	"fmt"

	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
)

// searchForm creates a reusable search form component
func searchForm(query string, autofocus bool, extraClasses string) g.Node {
	classes := "max-w-2xl"
	if extraClasses != "" {
		classes += " " + extraClasses
	}

	return Form(
		Method("GET"),
		Action("/-/search"),
		Class(classes),
		g.Attr("hx-boost", "true"),
		Div(
			Class("relative"),
			Input(
				Type("text"),
				Name("q"),
				Placeholder("What are you looking for?"),
				g.If(query != "", Value(query)),
				Class("block w-full pl-10 pr-3 py-3 border border-gray-300 rounded-lg leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-2 focus:ring-purple-500 focus:border-purple-500 text-base"),
				g.If(autofocus, g.Attr("autofocus", "true")),
			),
			Div(
				Class("absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none"),
				Span(
					Class("text-gray-400 text-lg"),
					g.Text("🔍"),
				),
			),
			Button(
				Type("submit"),
				Class("absolute inset-y-0 right-0 pr-3 flex items-center text-purple-600 hover:text-purple-800 font-medium"),
				g.Text("Search"),
			),
		),
	)
}

// SearchResults displays semantic search results
func (rs Resource) SearchResults(notesService *engine.NotesService, query string, results []model.Note) (g.Node, error) {
	var title string
	var content g.Node

	if query == "" {
		title = "Semantic Search"
		content = Div(
			Class("prose max-w-none"),
			P(
				Class("text-gray-600 mb-4"),
				g.Text("Enter a search query to find related notes using semantic search."),
			),
			searchForm("", true, ""),
		)
	} else if len(results) == 0 {
		title = fmt.Sprintf("Search: %s", query)
		content = Div(
			Class("prose max-w-none"),
			P(
				Class("text-gray-600 mb-6"),
				g.Textf("No results found for \"%s\".", query),
			),
			// Search again form
			searchForm(query, false, ""),
		)
	} else {
		title = fmt.Sprintf("Search: %s (%d results)", query, len(results))
		content = Div(
			Class("prose max-w-none"),
			// Search again form
			searchForm(query, false, "mb-8"),
			P(
				Class("text-gray-600 mb-6"),
				g.Textf("Found %d semantically related notes:", len(results)),
			),
			Div(
				Class("grid gap-4 md:grid-cols-2 lg:grid-cols-3"),
				g.Group(g.Map(results, func(note model.Note) g.Node {
					return rs.renderNoteCard(note)
				})),
			),
		)
	}

	// Main content area
	mainContent := Div(
		Class("flex-1 container overflow-y-auto p-4 md:px-8"),
		H1(
			Class("text-3xl md:text-4xl font-bold mb-4 mt-2"),
			g.Text(title),
		),
		content,
	)

	return rs.Layout(
		nil, // No specific note for layout
		rs.renderWithNavbar(notesService, navbarConfig{
			showSearchForm: false, // Don't show inline search form on search pages
			mainContent:    mainContent,
		}),
	), nil
}
