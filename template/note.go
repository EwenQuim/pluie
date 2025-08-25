package template

import (
	"fmt"
	"strings"

	"github.com/EwenQuim/pluie/model"
	"github.com/adrg/frontmatter"
	"github.com/maragudk/gomponents"
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"

	"github.com/go-fuego/fuego/extra/markdown"
)

type Resource struct {
	Notes []model.Note
}

func MapMap[T any](ts map[string]T, cb func(k string, v T) g.Node) []g.Node {
	var nodes []g.Node
	for k, v := range ts {
		nodes = append(nodes, cb(k, v))
	}
	return nodes
}

// NoteWithList displays a note with the list of all notes on the left side
func (rs Resource) NoteWithList(note model.Note, searchQuery string) (gomponents.Node, error) {
	matter := map[string]any{}
	content, err := frontmatter.Parse(strings.NewReader(note.Content), &matter)
	if err != nil {
		content = []byte(note.Content)
		fmt.Println("Error parsing frontmatter:", err)
	}

	// Parse wiki-style links before markdown processing
	parsedContent := rs.parseWikiLinks(string(content))

	filteredNotes := rs.Notes

	// Filter notes by title if search query is provided
	if searchQuery != "" {
		filteredNotes = make([]model.Note, 0)
		searchLower := strings.ToLower(searchQuery)
		for _, n := range rs.Notes {
			if strings.Contains(strings.ToLower(n.Title), searchLower) {
				filteredNotes = append(filteredNotes, n)
			} else if strings.Contains(strings.ToLower(n.Slug), searchLower) {
				filteredNotes = append(filteredNotes, n)
			}
		}
	}

	return rs.Layout(
		Div(
			Class("flex gap-6"),
			// Left sidebar with notes list
			Div(
				Class("w-1/4 bg-gray-50 p-4 rounded-lg"),
				H2(
					Class("text-xl font-bold mb-4"),
					g.Text("Notes"),
				),
				// Search form
				Form(
					Method("GET"),
					Action("/"+note.Slug),
					Class("mb-4"),
					g.Attr("hx-boost", "true"),
					g.Attr("hx-push-url", "true"),
					g.Attr("hx-target", "#notes-list"),
					g.Attr("hx-select", "#notes-list"),
					g.Attr("hx-swap", "outerHTML"),
					Div(
						Class("flex gap-2"),
						Input(
							Type("text"),
							Name("search"),
							Placeholder("Search notes..."),
							Value(searchQuery),
							Class("flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"),
							g.Attr("hx-get", "/"+note.Slug),
							g.Attr("hx-trigger", "input changed delay:100ms, search"),
							g.Attr("hx-target", "#notes-list"),
							g.Attr("hx-select", "#notes-list"),
							g.Attr("hx-swap", "outerHTML"),
							g.Attr("hx-push-url", "true"),
						),
						NoScript(
							Button(
								Type("submit"),
								Class("px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500"),
								g.Text("Search"),
							),
						),
					),
				),
				// Results info
				g.If(searchQuery != "",
					P(
						Class("text-gray-600 mb-2 text-sm"),
						g.Textf("Found %d notes matching \"%s\"", len(filteredNotes), searchQuery),
					),
				),
				// Notes list
				Ul(
					ID("notes-list"),
					Class("space-y-1"),
					g.Group(g.Map(filteredNotes, func(n model.Note) gomponents.Node {
						isActive := n.Slug == note.Slug
						var linkClass string
						if isActive {
							linkClass = "block px-3 py-2 text-blue-800 bg-blue-100 rounded-md font-medium"
						} else {
							linkClass = "block px-3 py-2 text-blue-600 hover:text-blue-800 hover:bg-blue-50 rounded-md"
						}
						return Li(
							A(
								Href("/"+n.Slug),
								Class(linkClass),
								g.Text(n.Title),
							),
						)
					})),
				),
				// Show message if no results found
				g.If(searchQuery != "" && len(filteredNotes) == 0,
					P(
						Class("text-gray-500 mt-4 text-sm"),
						g.Text("No notes found matching your search."),
					),
				),
			),
			// Right content area with the note
			Div(
				Class("flex-1"),
				H1(
					Class("text-3xl font-bold mb-4"),
					g.Text(note.Title),
				),
				g.If(len(matter) > 0,
					Ul(
						Class("bg-gray-100 p-4 rounded-lg mb-6"),
						g.Group(MapMap(matter, func(key string, value any) gomponents.Node {
							return Li(
								g.Text(fmt.Sprintf("%s: %v", key, value)),
							)
						})),
					),
				),
				Div(
					Class("prose lg:prose-xl max-w-none"),
					g.Raw(string(markdown.Markdown(parsedContent))),
				),
			),
		),
	), nil
}
