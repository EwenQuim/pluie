package template

import (
	"fmt"
	"strings"

	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
)

// searchChatForm creates a reusable search chat form component
func searchChatForm(query string, autofocus bool, extraClasses string) g.Node {
	classes := "max-w-2xl"
	if extraClasses != "" {
		classes += " " + extraClasses
	}

	return Form(
		Method("GET"),
		Action("/-/search-chat"),
		Class(classes),
		g.Attr("hx-boost", "true"),
		Div(
			Class("relative"),
			Input(
				Type("text"),
				Name("q"),
				Placeholder("Ask a question about your notes..."),
				g.If(query != "", Value(query)),
				Class("block w-full pl-10 pr-3 py-3 border border-gray-300 rounded-lg leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-2 focus:ring-purple-500 focus:border-purple-500 text-base"),
				g.If(autofocus, g.Attr("autofocus", "true")),
			),
			Div(
				Class("absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none"),
				Span(
					Class("text-gray-400 text-lg"),
					g.Text("💬"),
				),
			),
			Button(
				Type("submit"),
				Class("absolute inset-y-0 right-0 pr-3 flex items-center text-purple-600 hover:text-purple-800 font-medium"),
				g.Text("Ask"),
			),
		),
	)
}

// SearchChatResults displays semantic search results with AI-generated summary
func (rs Resource) SearchChatResults(notesService *engine.NotesService, query string, results []model.Note, aiResponse string) (g.Node, error) {
	var title string
	var content g.Node

	if query == "" {
		title = "Search Chat"
		content = Div(
			Class("prose max-w-none"),
			P(
				Class("text-gray-600 mb-6"),
				g.Text("Ask a question and get an AI-powered answer based on your notes."),
			),
			searchChatForm("", true, ""),
		)
	} else if len(results) == 0 {
		title = fmt.Sprintf("Chat: %s", query)
		content = Div(
			Class("prose max-w-none"),
			searchChatForm(query, false, "mb-6"),
			Div(
				Class("bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6"),
				P(
					Class("text-yellow-800 mb-0"),
					g.Textf("No relevant notes found for \"%s\". Try rephrasing your question or searching for different terms.", query),
				),
			),
		)
	} else {
		title = fmt.Sprintf("Chat: %s", query)
		content = Div(
			Class("max-w-none"),
			// Search again form
			searchChatForm(query, false, "mb-6"),

			// AI Response section
			g.If(aiResponse != "",
				Div(
					Class("bg-gradient-to-r from-purple-50 to-blue-50 border border-purple-200 rounded-lg p-6 mb-8 shadow-sm"),
					Div(
						Class("flex items-start gap-3 mb-3"),
						Span(
							Class("text-2xl"),
							g.Text("🤖"),
						),
						H2(
							Class("text-xl font-semibold text-gray-900"),
							g.Text("AI Summary"),
						),
					),
					Div(
						Class("prose max-w-none"),
						g.Raw(formatAIResponse(aiResponse)),
					),
				),
			),

			// Source documents section
			Div(
				Class("mb-6"),
				H3(
					Class("text-lg font-semibold text-gray-700 mb-3"),
					g.Textf("📚 Found %d relevant notes:", len(results)),
				),
				Div(
					Class("grid gap-4 md:grid-cols-2 lg:grid-cols-3"),
					g.Group(g.Map(results, func(note model.Note) g.Node {
						return rs.renderNoteCard(note)
					})),
				),
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
			mainContent: mainContent,
		}),
	), nil
}

// formatAIResponse converts plain text AI response to HTML with basic formatting
func formatAIResponse(response string) string {
	// Split into paragraphs
	paragraphs := strings.Split(response, "\n\n")
	var formattedParagraphs []string

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// Handle bullet points
		if strings.HasPrefix(para, "- ") || strings.HasPrefix(para, "* ") {
			lines := strings.Split(para, "\n")
			var listItems []string
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if text, found := strings.CutPrefix(line, "- "); found {
					listItems = append(listItems, "<li>"+text+"</li>")
				} else if text, found := strings.CutPrefix(line, "* "); found {
					listItems = append(listItems, "<li>"+text+"</li>")
				}
			}
			if len(listItems) > 0 {
				formattedParagraphs = append(formattedParagraphs, "<ul>"+strings.Join(listItems, "")+"</ul>")
			}
		} else {
			// Regular paragraph
			formattedParagraphs = append(formattedParagraphs, "<p>"+para+"</p>")
		}
	}

	return strings.Join(formattedParagraphs, "")
}
