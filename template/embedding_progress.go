package template

import (
	"fmt"

	g "github.com/maragudk/gomponents"
	h "github.com/maragudk/gomponents/html"
)

// EmbeddingProgressData holds the data for rendering embedding progress
type EmbeddingProgressData struct {
	Embedded    int
	Total       int
	IsEmbedding bool
}

// RenderEmbeddingProgressContent renders the inner content that gets swapped by SSE
// This is used both for initial render and SSE updates
func RenderEmbeddingProgressContent(data EmbeddingProgressData) g.Node {
	// Calculate percentage
	percentage := 0
	if data.Total > 0 {
		percentage = (data.Embedded * 100) / data.Total
	}

	// Determine bar color based on status
	barColor := "bg-purple-600"
	if !data.IsEmbedding && data.Embedded == data.Total && data.Total > 0 {
		barColor = "bg-green-600"
	}

	return h.Div(
		h.Class("text-xs text-gray-500 px-2"),
		h.Div(
			h.Class("flex items-center justify-between mb-1"),
			h.Span(g.Text("Embeddings:")),
			h.Span(
				h.ID("embedding-progress-text"),
				h.Class("font-mono"),
				g.Textf("%d/%d", data.Embedded, data.Total),
			),
		),
		h.Div(
			h.Class("w-full bg-gray-200 rounded-full h-1.5"),
			h.Div(
				h.ID("embedding-progress-bar"),
				h.Class(fmt.Sprintf("%s h-1.5 rounded-full transition-all duration-300", barColor)),
				g.Attr("style", fmt.Sprintf("width: %d%%", percentage)),
			),
		),
	)
}

// RenderEmbeddingProgressIndicator renders the complete embedding progress indicator
// for initial page load
func RenderEmbeddingProgressIndicator() g.Node {
	// Initial state: 0/0
	initialData := EmbeddingProgressData{
		Embedded:    0,
		Total:       0,
		IsEmbedding: false,
	}

	return h.Div(
		h.Class("mt-auto pt-4 border-t border-gray-200"),
		g.Attr("hx-ext", "sse"),
		g.Attr("sse-connect", "/-/embedding-progress"),
		g.Attr("sse-swap", "message"),
		RenderEmbeddingProgressContent(initialData),
	)
}
