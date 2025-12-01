package template

import (
	"fmt"
	"strings"

	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
)

// unifiedSearchForm creates the search form component with live search
func unifiedSearchForm(query string, autofocus bool) g.Node {
	return Form(
		Method("GET"),
		Action("/-/search"),
		Class("max-w-2xl mb-8"),
		Div(
			Class("relative"),
			Input(
				Type("text"),
				Name("q"),
				Placeholder("Search titles, headings, and content..."),
				g.If(query != "", Value(query)),
				Class("block w-full pl-10 pr-3 py-3 border border-gray-300 rounded-lg leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-base"),
				g.If(autofocus, g.Attr("autofocus", "true")),
				// HTMX live search attributes
				g.Attr("hx-get", "/-/search"),
				g.Attr("hx-trigger", "input changed delay:300ms, search"),
				g.Attr("hx-target", "#search-results-container"),
				g.Attr("hx-select", "#search-results-container"),
				g.Attr("hx-swap", "outerHTML"),
				g.Attr("hx-push-url", "true"),
				g.Attr("hx-indicator", "#search-indicator"),
			),
			Div(
				Class("absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none"),
				Span(
					Class("text-gray-400 text-lg"),
					g.Text("🔍"),
				),
			),
			// Loading indicator (shown during HTMX request)
			Div(
				ID("search-indicator"),
				Class("htmx-indicator absolute inset-y-0 right-0 pr-3 flex items-center"),
				Div(
					Class("animate-spin h-4 w-4 border-2 border-blue-500 border-t-transparent rounded-full"),
				),
			),
		),
	)
}

// UnifiedSearchResults renders the unified search page with all result types
func (rs Resource) UnifiedSearchResults(
	notesService *engine.NotesService,
	query string,
	titleMatches []model.Note,
	headingMatches []engine.HeadingMatch,
	seenSlugs []string,
) (g.Node, error) {
	var title string
	var content g.Node

	if query == "" {
		// Empty state
		title = "Search"
		content = Div(
			Class("prose max-w-none"),
			unifiedSearchForm("", true),
			Div(
				P(
					Class("text-sm italic mt-4"),
					g.Text("Start typing to search across all your notes!"),
				),
			),
		)
	} else {
		// Results page
		title = fmt.Sprintf("Search: %s", query)

		// Build seen slugs parameter for SSE
		seenParam := strings.Join(seenSlugs, ",")

		content = Div(
			Class("max-w-none"),
			// Search form at top
			unifiedSearchForm(query, false),

			// Results container (HTMX target)
			rs.renderSearchResultsContainer(query, titleMatches, headingMatches, seenParam),
		)
	}

	// Main content area
	mainContent := Div(
		Class("flex-1 container overflow-y-auto p-4 md:px-8 flex flex-col"),
		Div(
			Class("flex-1"),
			H1(
				Class("text-3xl md:text-4xl font-bold mb-4 mt-2"),
				g.Text(title),
			),
			content,
		),
		// Embedding progress indicator at the bottom
		RenderEmbeddingProgressIndicator(),
	)

	return rs.Layout(
		nil, // No specific note for layout
		rs.renderWithNavbar(notesService, navbarConfig{
			mainContent: mainContent,
		}),
	), nil
}

// renderSearchResultsContainer wraps the search results for HTMX targeting
func (rs Resource) renderSearchResultsContainer(query string, titleMatches []model.Note, headingMatches []engine.HeadingMatch, seenParam string) g.Node {
	return Div(
		ID("search-results-container"),

		// Combined results section (title + semantic, will be appended via SSE)
		g.If(len(titleMatches) > 0,
			Div(
				Class("mb-8"),
				Div(
					ID("combined-results"),
					Class("grid gap-4 md:grid-cols-2 lg:grid-cols-3"),
					g.Group(g.Map(titleMatches, func(note model.Note) g.Node {
						return rs.renderNoteCard(note)
					})),
				),
			),
		),

		// Loading indicator for semantic search (hidden when not applicable)
		g.If(len(titleMatches) > 0,
			Div(
				ID("search-loading"),
				Class("flex justify-center py-4 mb-4"),
				Div(
					Class("animate-spin h-4 w-4 border-2 border-gray-400 border-t-transparent rounded-full"),
				),
			),
		),

		// Placeholder for when no title matches but semantic will load
		g.If(len(titleMatches) == 0,
			Div(
				ID("combined-results"),
				Class("grid gap-4 md:grid-cols-2 lg:grid-cols-3 mb-8"),
			),
		),
		g.If(len(titleMatches) == 0,
			Div(
				ID("search-loading"),
				Class("flex justify-center py-8"),
				Div(
					Class("animate-spin h-5 w-5 border-2 border-gray-400 border-t-transparent rounded-full"),
				),
			),
		),

		// Heading matches section (smaller, less emphasized)
		g.If(len(headingMatches) > 0,
			Div(
				Class("mb-8 space-y-2"),
				g.Group(g.Map(headingMatches, func(match engine.HeadingMatch) g.Node {
					return rs.renderHeadingCard(match)
				})),
			),
		),

		// AI response section (populated by SSE, hidden initially)
		Div(
			ID("ai-section"),
			Class("hidden mb-8"),
			H2(
				Class("text-lg font-medium mb-2 text-gray-700"),
				g.Text("Summary"),
			),
			Div(
				Class("bg-gray-50 border border-gray-200 rounded-lg p-4"),
				Div(
					ID("ai-content"),
					Class("prose prose-sm max-w-none text-gray-700"),
				),
				// Streaming cursor
				Span(
					ID("ai-cursor"),
					Class("hidden inline-block w-1.5 h-3 bg-gray-400 ml-1 animate-pulse"),
				),
				// Disclaimer
				P(
					ID("ai-disclaimer"),
					Class("hidden text-xs text-gray-500 italic mt-3 mb-0"),
					g.Text("AI generated, might not be accurate"),
				),
			),
		),

		// SSE EventSource JavaScript
		rs.renderSSEScript(query, seenParam),
	)
}

// renderHeadingCard renders a single heading match as a minimal card
func (rs Resource) renderHeadingCard(match engine.HeadingMatch) g.Node {
	return A(
		Href("/"+match.Note.Slug),
		Class("block border-l-2 border-gray-300 pl-3 py-2 hover:border-gray-400 hover:bg-gray-50 transition-colors"),
		g.Attr("hx-boost", "true"),

		// Note title
		Div(
			Class("text-xs text-gray-500 mb-0.5"),
			g.Text(match.Note.Title),
		),

		// Heading text (no level badge shown, but still sorted by level)
		Div(
			Class("text-sm font-medium text-gray-700 hover:text-gray-900"),
			g.Text(match.Heading),
		),

		// Context snippet
		g.If(match.Context != "",
			P(
				Class("text-xs text-gray-600 line-clamp-1 mt-0.5 mb-0"),
				g.Text(match.Context),
			),
		),
	)
}

// renderSSEScript renders the EventSource JavaScript for SSE streaming with cleanup
func (rs Resource) renderSSEScript(query string, seenParam string) g.Node {
	return Script(
		g.Raw(fmt.Sprintf(`
(function() {
	// Close any existing SSE connection before starting a new one (for live search)
	if (window.currentSearchSSE) {
		window.currentSearchSSE.close();
		window.currentSearchSSE = null;
	}

	// Don't start SSE for empty queries
	if (!%q) {
		return;
	}

	const evtSource = new EventSource('/-/search-stream?q=%s&seen=%s');
	window.currentSearchSSE = evtSource; // Store globally for cleanup

	const loading = document.getElementById('search-loading');
	const combinedResults = document.getElementById('combined-results');
	const aiSection = document.getElementById('ai-section');
	const aiContent = document.getElementById('ai-content');
	const cursor = document.getElementById('ai-cursor');
	const disclaimer = document.getElementById('ai-disclaimer');

	evtSource.addEventListener('semantic-results', function(e) {
		if (loading) loading.classList.add('hidden');
		if (combinedResults && e.data) {
			// Append semantic results to the combined grid
			combinedResults.insertAdjacentHTML('beforeend', e.data);
		}
	});

	evtSource.addEventListener('token', function(e) {
		if (aiSection && aiSection.classList.contains('hidden')) {
			aiSection.classList.remove('hidden');
		}
		if (cursor && cursor.classList.contains('hidden')) {
			cursor.classList.remove('hidden');
		}
		if (aiContent) {
			aiContent.insertAdjacentText('beforeend', e.data);
		}
	});

	evtSource.addEventListener('done', function(e) {
		if (cursor) {
			cursor.classList.add('hidden');
			cursor.classList.remove('animate-pulse');
		}
		if (disclaimer) disclaimer.classList.remove('hidden');
		evtSource.close();
		window.currentSearchSSE = null;
	});

	evtSource.addEventListener('error', function(e) {
		console.error('SSE error:', e);
		if (loading) loading.classList.add('hidden');
		if (cursor) {
			cursor.classList.add('hidden');
			cursor.classList.remove('animate-pulse');
		}
		evtSource.close();
		window.currentSearchSSE = null;
	});
})();
		`, query, query, seenParam)),
	)
}

// RenderSemanticResultsHTML renders semantic search results as HTML for SSE streaming
// Returns individual note cards to be appended to the combined results grid
func RenderSemanticResultsHTML(rs Resource, notes []model.Note) string {
	if len(notes) == 0 {
		return ""
	}

	// Build note cards that will be appended to the existing grid
	var html strings.Builder
	for _, note := range notes {
		rs.renderNoteCard(note).Render(&html)
	}

	return html.String()
}
