package template

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	"github.com/maragudk/gomponents"
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"

	"github.com/go-fuego/fuego/extra/markdown"
)

type Resource struct {
	Tree *engine.TreeNode
}

// MapMap creates nodes from a map
func MapMap[T any](ts map[string]T, cb func(k string, v T) g.Node) []g.Node {
	var nodes []g.Node
	for k, v := range ts {
		nodes = append(nodes, cb(k, v))
	}
	return nodes
}

// MapMapSorted creates nodes from a map with keys sorted alphabetically
func MapMapSorted[T any](ts map[string]T, cb func(k string, v T) g.Node) []g.Node {
	// Get all keys and sort them
	keys := make([]string, 0, len(ts))
	for k := range ts {
		keys = append(keys, k)
	}

	// Sort keys alphabetically (case-insensitive)
	sortKeys(keys)

	// Create nodes in sorted order
	var nodes []g.Node
	for _, k := range keys {
		nodes = append(nodes, cb(k, ts[k]))
	}
	return nodes
}

// sortKeys sorts a slice of strings alphabetically (case-insensitive)
func sortKeys(keys []string) {
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if strings.ToLower(keys[i]) > strings.ToLower(keys[j]) {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
}

// renderTreeNode renders a single tree node with its children
func (rs Resource) renderTreeNode(node *engine.TreeNode, currentSlug string) gomponents.Node {
	if node == nil {
		return g.Text("")
	}

	if node.IsFolder {
		// Render folder
		return Li(
			Class(""),
			Div(
				Class("flex items-center py-1"),
				Button(
					Class("flex items-center text-left w-full px-2 py-1 text-gray-900 hover:text-black hover:bg-gray-50"),
					g.Attr("onclick", fmt.Sprintf("toggleFolder('%s')", node.Path)),
					// Folder icon (chevron)
					Span(
						Class("mr-2 transition-transform duration-200 text-gray-400 text-xs"),
						ID("chevron-"+node.Path),
						g.If(node.IsOpen,
							g.Text("▼"),
						),
						g.If(!node.IsOpen,
							g.Text("▶"),
						),
					),
					Span(g.Text(node.Name)),
				),
			),
			// Children container
			g.If(len(node.Children) > 0,
				Ul(
					Class("ml-4"),
					ID("folder-"+node.Path),
					g.If(node.IsOpen,
						g.Attr("style", "display: block;"),
					),
					g.If(!node.IsOpen,
						g.Attr("style", "display: none;"),
					),
					g.Group(g.Map(node.Children, func(child *engine.TreeNode) gomponents.Node {
						return rs.renderTreeNode(child, currentSlug)
					})),
				),
			),
		)
	} else {
		// Render note
		isActive := node.Note != nil && node.Note.Slug == currentSlug
		var linkClass string
		if isActive {
			linkClass = "flex items-center px-2 py-1 text-purple-600 bg-purple-50 border-l border-purple-600 font-medium"
		} else {
			linkClass = "flex items-center px-2 py-1 text-gray-600 hover:text-gray-900 hover:bg-gray-50 border-l border-gray-300 hover:border-gray-800"
		}

		return Li(
			A(
				Href("/"+node.Path),
				Class(linkClass),
				g.Attr("hx-boost", "true"),
				g.Text(node.Name),
			),
		)
	}
}

// TOCItem represents a table of contents item
type TOCItem struct {
	ID    string
	Text  string
	Level int
}

// slugify converts a string to a URL-friendly slug
func slugify(text string) string {
	// Convert to lowercase
	slug := strings.ToLower(text)

	// Replace spaces and common punctuation with hyphens
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	return slug
}

// extractHeadings extracts headings from markdown content and returns TOC items
func extractHeadings(content string) []TOCItem {
	var tocItems []TOCItem
	usedSlugs := make(map[string]int) // Track used slugs to handle duplicates

	// Regular expression to match markdown headings
	headingRegex := regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if matches := headingRegex.FindStringSubmatch(line); matches != nil {
			level := len(matches[1]) // Number of # characters
			text := strings.TrimSpace(matches[2])

			// Generate a slug from the heading text
			baseSlug := slugify(text)
			id := baseSlug

			// Handle duplicate slugs by appending a number
			if count, exists := usedSlugs[baseSlug]; exists {
				usedSlugs[baseSlug] = count + 1
				id = fmt.Sprintf("%s-%d", baseSlug, count+1)
			} else {
				usedSlugs[baseSlug] = 0
			}

			tocItems = append(tocItems, TOCItem{
				ID:    id,
				Text:  text,
				Level: level,
			})
		}
	}

	return tocItems
}

// renderTOC renders the table of contents as HTML nodes
func renderTOC(tocItems []TOCItem) []gomponents.Node {
	if len(tocItems) == 0 {
		return []gomponents.Node{
			P(
				Class("text-sm text-gray-500 italic"),
				g.Text("No headings found"),
			),
		}
	}

	var nodes []gomponents.Node

	for _, item := range tocItems {
		// Calculate indentation based on heading level
		var indentClass string
		if item.Level > 2 {
			indentClass = fmt.Sprintf("ml-%d", (item.Level-2)*3)
		}

		node := A(
			Href("#"+item.ID),
			Class(fmt.Sprintf("block py-1 px-2 text-sm text-gray-600 hover:text-gray-900 hover:bg-gray-50 rounded-md transition-colors %s", indentClass)),
			g.Attr("onclick", "handleTOCClick(event, this)"),
			g.Text(item.Text),
		)

		nodes = append(nodes, node)
	}

	return nodes
}

// NoteWithList displays a note with the list of all notes on the left side
func (rs Resource) NoteWithList(note *model.Note, searchQuery string) (gomponents.Node, error) {

	matter := map[string]any{}
	var content []byte
	var slug string
	var title string
	var referencedBy []model.NoteReference

	if note != nil {
		matter = note.Metadata
		slug = note.Slug
		title = note.Title
		referencedBy = note.ReferencedBy
		content = []byte(note.Content)
	} else {
		title = "404 : Not found"
		content = []byte("This note does not exist or is private.")
	}

	// Parse wiki-style links before markdown processing
	parsedContent := engine.ParseWikiLinks(string(content), rs.Tree)

	// Extract headings for table of contents
	tocItems := extractHeadings(parsedContent)

	// Filter tree based on search query
	displayTree := rs.Tree
	if searchQuery != "" {
		displayTree = engine.FilterTreeBySearch(rs.Tree, searchQuery)
	}

	// Get site title, icon, and description from environment variables
	siteTitle := os.Getenv("SITE_TITLE")
	if siteTitle == "" {
		siteTitle = "Pluie"
	}
	siteIcon := os.Getenv("SITE_ICON")
	siteDescription := os.Getenv("SITE_DESCRIPTION")

	return rs.Layout(
		Div(
			Class("flex gap-2 h-screen"),
			// Left sidebar with notes list
			Div(
				Class("w-1/4 max-w-md bg-white border-r border-gray-200 p-4 flex flex-col h-full"),
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
				// Search form
				Form(
					Method("GET"),
					Action("/"+slug),
					Class("mb-6"),
					g.Attr("hx-boost", "true"),
					g.Attr("hx-push-url", "true"),
					g.Attr("hx-target", "#notes-list"),
					g.Attr("hx-select", "#notes-list"),
					g.Attr("hx-swap", "outerHTML"),
					Div(
						Class("relative"),
						Input(
							Type("text"),
							Name("search"),
							Placeholder("Search page or heading..."),
							Value(searchQuery),
							Class("block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 text-sm"),
							g.Attr("hx-get", "/"+slug),
							g.Attr("hx-trigger", "input changed delay:200ms, search"),
							g.Attr("hx-target", "#notes-list"),
							g.Attr("hx-select", "#notes-list"),
							g.Attr("hx-swap", "outerHTML"),
							g.Attr("hx-push-url", "true"),
						),
						Div(
							Class("absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none"),
							Span(
								Class("text-gray-400 text-sm"),
								g.Text("🔍"),
							),
						),
						NoScript(
							Button(
								Type("submit"),
								Class("absolute inset-y-0 right-0 pr-3 flex items-center"),
								g.Text("Search"),
							),
						),
					),
				),
				// Fold/Unfold all buttons
				Div(
					Class("mb-4 flex gap-2"),
					Button(
						Class("px-3 py-1 text-sm bg-gray-200 hover:bg-gray-300 text-gray-700 rounded-md transition-colors cursor-pointer"),
						g.Attr("onclick", "expandAllFolders()"),
						g.Text("Expand All"),
					),
					Button(
						Class("px-3 py-1 text-sm bg-gray-200 hover:bg-gray-300 text-gray-700 rounded-md transition-colors cursor-pointer"),
						g.Attr("onclick", "collapseAllFolders()"),
						g.Text("Collapse All"),
					),
				),
				// Scrollable container for results and notes tree
				Div(
					Class("flex-1 overflow-y-auto"),
					// Results info
					g.If(searchQuery != "",
						P(
							Class("text-gray-600 mb-2 text-sm"),
							g.Textf("Found %d notes matching \"%s\"", countNotesInTree(displayTree), searchQuery),
						),
					),
					// Notes tree
					Div(
						ID("notes-list"),
						Class(""),
						g.If(displayTree != nil && len(displayTree.Children) > 0,
							Ul(
								Class(""),
								g.Group(g.Map(displayTree.Children, func(child *engine.TreeNode) gomponents.Node {
									return rs.renderTreeNode(child, slug)
								})),
							),
						),
						// Show message if no results found
						g.If(searchQuery != "" && (displayTree == nil || len(displayTree.Children) == 0),
							P(
								Class("text-gray-500 mt-4 text-sm"),
								g.Text("No notes found matching your search."),
							),
						),
					),
				),
			),
			// Main content area with the note
			Div(
				Class("flex-1 overflow-y-auto p-4 px-8"),
				H1(
					Class("text-3xl font-bold mb-4"),
					g.If(title != "", g.Text(title)),
				),
				g.If(len(matter) > 0,
					Ul(
						Class("bg-gray-100 p-4 rounded-lg mb-6"),
						g.Group(MapMapSorted(matter, func(key string, value any) gomponents.Node {
							return Li(
								g.Text(fmt.Sprintf("%s: %v", key, value)),
							)
						})),
					),
				),
				Div(
					Class("prose max-w-none"),
					g.Raw(string(markdown.Markdown(parsedContent))),
				),
				// Referenced By section
				g.If(len(referencedBy) > 0,
					Div(
						Class("mt-8 pt-6 border-t border-gray-200"),
						H3(
							Class("text-lg font-semibold mb-3 text-gray-700"),
							g.Text("Referenced by"),
						),
						Ul(
							Class("space-y-2"),
							g.Group(g.Map(referencedBy, func(ref model.NoteReference) gomponents.Node {
								return Li(
									A(
										Href("/"+ref.Slug),
										Class("text-blue-600 hover:text-blue-800 hover:underline"),
										g.Text(ref.Title),
									),
								)
							})),
						),
					),
				),
			),
			// Right sidebar with "On this page" table of contents
			Div(
				Class("w-64 bg-white border-l border-gray-200 p-4 flex flex-col h-full"),
				ID("toc-sidebar"),
				Div(
					Class("mb-4"),
					H3(
						Class("text-sm font-semibold text-gray-900 uppercase tracking-wide"),
						g.Text("On this page"),
					),
				),
				Div(
					Class("flex-1 overflow-y-auto"),
					Nav(
						ID("table-of-contents"),
						Class("space-y-1"),
						g.Group(renderTOC(tocItems)),
					),
				),
			),
		),
	), nil
}

// countNotesInTree counts the total number of notes in a template tree
func countNotesInTree(node *engine.TreeNode) int {
	if node == nil {
		return 0
	}

	count := 0
	if !node.IsFolder && node.Note != nil {
		count = 1
	}

	for _, child := range node.Children {
		count += countNotesInTree(child)
	}

	return count
}
