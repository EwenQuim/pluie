package template

import (
	"fmt"
	"os"

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
	// Get site title, icon, and description from environment variables
	siteTitle := os.Getenv("SITE_TITLE")
	if siteTitle == "" {
		siteTitle = "Pluie"
	}
	siteIcon := os.Getenv("SITE_ICON")
	siteDescription := os.Getenv("SITE_DESCRIPTION")

	var title string
	var content g.Node

	if query == "" {
		title = "Semantic Search"
		content = Div(
			Class("prose max-w-none"),
			P(
				Class("text-gray-600 mb-6"),
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

	return rs.Layout(
		nil, // No specific note for layout
		Div(
			Class("flex flex-col md:flex-row md:gap-2 h-screen w-screen justify-between"),
			// Mobile top bar (hidden on desktop)
			Div(
				Class("md:hidden bg-white border-b border-gray-200 p-4 flex items-center justify-between z-50"),
				// Site title and icon
				Div(
					Class("flex items-center gap-3"),
					g.If(siteIcon != "",
						Img(
							Src(siteIcon),
							Alt("Site Icon"),
							Class("w-6 h-6 object-contain rounded-md"),
						),
					),
					H1(
						Class("text-lg font-bold text-gray-900"),
						g.Text(siteTitle),
					),
				),
				// Burger menu button
				Button(
					Class("p-2 rounded-md hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-gray-200"),
					ID("burger-menu"),
					g.Attr("onclick", "toggleMobileSidebar()"),
					g.Attr("aria-label", "Toggle navigation menu"),
					Div(
						Class("w-6 h-6 flex flex-col justify-center items-center space-y-1"),
						Div(Class("w-5 h-0.5 bg-gray-600 transition-all duration-300 ease-in-out"), ID("burger-line-1")),
						Div(Class("w-5 h-0.5 bg-gray-600 transition-all duration-300 ease-in-out"), ID("burger-line-2")),
						Div(Class("w-5 h-0.5 bg-gray-600 transition-all duration-300 ease-in-out"), ID("burger-line-3")),
					),
				),
			),
			// Mobile sidebar overlay
			Div(
				Class("fixed inset-0 bg-black bg-opacity-50 z-40 md:hidden opacity-0 invisible transition-all duration-300"),
				ID("mobile-sidebar-overlay"),
				g.Attr("onclick", "closeMobileSidebar()"),
			),
			// Left sidebar with notes list
			Div(
				Class("w-3/4 md:w-1/4 max-w-md bg-white border-r border-gray-200 p-4 flex flex-col h-full md:relative fixed top-0 left-0 z-50 md:z-auto -translate-x-full md:translate-x-0 transition-transform duration-300 ease-in-out"),
				ID("mobile-sidebar"),
				// Site header with title and icon
				Div(
					Class("mb-6"),
					Div(
						Class("flex items-center gap-3 mb-2"),
						// Site icon
						g.If(siteIcon != "",
							Img(
								Src(siteIcon),
								Alt("Site Icon"),
								Class("w-8 h-8 object-contain rounded-md"),
							),
						),
						// Site title
						H1(
							Class("text-xl font-bold text-gray-900"),
							g.Text(siteTitle),
						),
					),
					// Site description
					g.If(siteDescription != "",
						P(
							Class("text-sm text-gray-500 italic"),
							g.Text(siteDescription),
						),
					),
				),
				// Back to home link
				Div(
					Class("mb-6"),
					A(
						Href("/"),
						Class("inline-flex items-center gap-2 text-sm text-blue-600 hover:text-blue-800 hover:underline"),
						g.Text("← Back to notes"),
					),
				),
				// Scrollable container for notes tree
				Div(
					Class("flex-1 overflow-y-auto"),
					// Notes tree
					Div(
						ID("notes-list"),
						Class(""),
						g.If(notesService.GetTree() != nil && len(notesService.GetTree().Children) > 0,
							Ul(
								Class(""),
								g.Group(g.Map(notesService.GetTree().Children, func(child *engine.TreeNode) g.Node {
									return rs.renderTreeNode(child, "")
								})),
							),
						),
					),
				),
			),
			// Main content area with the search results
			Div(
				Class("flex-1 container overflow-y-auto p-4 md:px-8"),
				H1(
					Class("text-3xl md:text-4xl font-bold mb-4 mt-2"),
					g.Text(title),
				),
				content,
			),
		),
	), nil
}
