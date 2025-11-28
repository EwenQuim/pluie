package template

import (
	"fmt"

	"github.com/EwenQuim/pluie/engine"
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
)

// searchLiveChatForm creates a reusable search live chat form component
func searchLiveChatForm(query string, autofocus bool, extraClasses string) g.Node {
	classes := "max-w-2xl"
	if extraClasses != "" {
		classes += " " + extraClasses
	}

	return Form(
		ID("live-chat-form"),
		Method("GET"),
		Action("/-/search-live-chat"),
		Class(classes),
		Div(
			Class("relative"),
			Input(
				Type("text"),
				Name("q"),
				ID("live-chat-input"),
				Placeholder("Ask a question and watch the AI respond in real-time..."),
				g.If(query != "", Value(query)),
				Class("block w-full pl-10 pr-3 py-3 border border-gray-300 rounded-lg leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-2 focus:ring-green-500 focus:border-green-500 text-base"),
				g.If(autofocus, g.Attr("autofocus", "true")),
			),
			Div(
				Class("absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none"),
				Span(
					Class("text-gray-400 text-lg"),
					g.Text("⚡"),
				),
			),
			Button(
				Type("submit"),
				ID("live-chat-submit"),
				Class("absolute inset-y-0 right-0 pr-3 flex items-center text-green-600 hover:text-green-800 font-medium"),
				g.Text("Ask"),
			),
		),
	)
}

// SearchLiveChatResults displays the live chat search page
func (rs Resource) SearchLiveChatResults(notesService *engine.NotesService, query string) (g.Node, error) {
	var title string
	var content g.Node

	if query == "" {
		title = "Live Chat Search"
		content = Div(
			Class("prose max-w-none"),
			P(
				Class("text-gray-600 mb-6"),
				g.Text("Ask a question and watch as the AI searches your notes and generates a response in real-time!"),
			),
			Div(
				Class("bg-green-50 border border-green-200 rounded-lg p-4 mb-6"),
				P(
					Class("text-sm text-green-800 mb-0"),
					g.Text("⚡ This is a live streaming version where you can see the search happening and the AI response being generated token by token."),
				),
			),
			searchLiveChatForm("", true, ""),
		)
	} else {
		title = fmt.Sprintf("Live Chat: %s", query)
		content = Div(
			Class("max-w-none"),
			// Search form
			searchLiveChatForm(query, false, "mb-6"),

			// Loading indicator (shown initially, hidden when SSE starts)
			Div(
				ID("loading-indicator"),
				Class("flex items-center gap-3 p-4 bg-blue-50 border border-blue-200 rounded-lg mb-6"),
				Div(
					Class("animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"),
				),
				Span(
					Class("text-blue-700 font-medium"),
					g.Text("Searching your notes..."),
				),
			),

			// Documents section (populated via SSE)
			Div(
				ID("documents-section"),
				Class("mb-6 hidden"),
			),

			// AI Response section (populated via SSE)
			Div(
				ID("ai-response-section"),
				Class("bg-gradient-to-r from-green-50 to-blue-50 border border-green-200 rounded-lg p-6 mb-8 shadow-sm hidden"),
				Div(
					Class("flex items-start gap-3 mb-3"),
					Span(
						Class("text-2xl"),
						g.Text("🤖"),
					),
					H2(
						Class("text-xl font-semibold text-gray-900"),
						g.Text("AI Response"),
					),
				),
				Div(
					ID("ai-response-content"),
					Class("prose max-w-none whitespace-pre-wrap"),
					// AI response will be streamed here
				),
				// Cursor indicator while streaming
				Span(
					ID("streaming-cursor"),
					Class("inline-block w-2 h-5 bg-green-600 animate-pulse ml-1"),
					g.Attr("style", "animation: pulse 1s cubic-bezier(0.4, 0, 0.6, 1) infinite;"),
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
		// SSE connection and event handlers via inline script
		g.If(query != "",
			Script(
				g.Raw(fmt.Sprintf(`
					(function() {
						const evtSource = new EventSource('/-/search-live-chat-stream?q=%s');
						const docsSection = document.getElementById('documents-section');
						const aiSection = document.getElementById('ai-response-section');
						const aiContent = document.getElementById('ai-response-content');
						const cursor = document.getElementById('streaming-cursor');
						const loading = document.getElementById('loading-indicator');

						evtSource.addEventListener('documents', function(e) {
							loading.classList.add('hidden');
							docsSection.innerHTML = e.data;
							docsSection.classList.remove('hidden');
						});

						evtSource.addEventListener('token', function(e) {
							if (aiSection.classList.contains('hidden')) {
								aiSection.classList.remove('hidden');
							}
							aiContent.insertAdjacentText('beforeend', e.data);
						});

						evtSource.addEventListener('done', function(e) {
							cursor.classList.add('hidden');
							evtSource.close();
						});

						evtSource.addEventListener('error', function(e) {
							cursor.classList.add('hidden');
							loading.classList.add('hidden');
							aiSection.classList.remove('hidden');
							aiContent.innerHTML = '<p class="text-red-600">Error: ' + e.data + '</p>';
							evtSource.close();
						});
					})();
				`, query)),
			),
		),
	)

	return rs.Layout(
		nil, // No specific note for layout
		rs.renderWithNavbar(notesService, navbarConfig{
			showSearchForm: false, // Don't show inline search form on search pages
			mainContent:    mainContent,
		}),
	), nil
}
