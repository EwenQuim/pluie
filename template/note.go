package template

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	"github.com/maragudk/gomponents"
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"

	"github.com/go-fuego/fuego/extra/markdown"
)

type Resource struct {
	// Resource is stateless and only contains rendering logic
	// Data is passed as parameters to methods
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
	sortKeysCaseInsensitive(keys)

	// Create nodes in sorted order
	var nodes []g.Node
	for _, k := range keys {
		nodes = append(nodes, cb(k, ts[k]))
	}
	return nodes
}

// sortKeysCaseInsensitive sorts a slice of strings alphabetically (case-insensitive)
func sortKeysCaseInsensitive(keys []string) {
	slices.SortFunc(keys, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})
}

// CSS class constants for consistent styling
const (
	folderButtonClass = "flex items-center text-left w-full px-2 py-1 text-gray-900 hover:text-black hover:bg-gray-50"
	chevronClass      = "mr-2 transition-transform duration-200 text-gray-400 text-xs"
	activeLinkClass   = "flex items-center px-2 py-1 text-purple-600 bg-purple-50 border-l border-purple-600 font-medium"
	inactiveLinkClass = "flex items-center px-2 py-1 text-gray-600 hover:text-gray-900 hover:bg-gray-50 border-l border-gray-300 hover:border-gray-800"
)

// renderTreeNode renders a single tree node with its children
func (rs Resource) renderTreeNode(node *engine.TreeNode, currentSlug string) gomponents.Node {
	if node == nil {
		return g.Text("")
	}

	if node.IsFolder {
		return rs.renderFolderNode(node, currentSlug)
	}
	return rs.renderNoteNode(node, currentSlug)
}

// renderFolderNode renders a folder tree node
func (rs Resource) renderFolderNode(node *engine.TreeNode, currentSlug string) gomponents.Node {
	return Li(
		Class(""),
		Div(
			Class("flex items-center py-1"),
			Button(
				Class(folderButtonClass),
				g.Attr("onclick", fmt.Sprintf("toggleFolder('%s')", node.Path)),
				rs.renderChevronIcon(node),
				Span(g.Text(node.Name)),
			),
		),
		rs.renderFolderChildren(node, currentSlug),
	)
}

// renderNoteNode renders a note tree node
func (rs Resource) renderNoteNode(node *engine.TreeNode, currentSlug string) gomponents.Node {
	isActive := node.Note != nil && node.Note.Slug == currentSlug
	linkClass := inactiveLinkClass
	if isActive {
		linkClass = activeLinkClass
	}

	return Li(
		A(
			Href("/"+node.Path),
			Class(linkClass),
			g.Attr("hx-boost", "true"),
			g.Attr("onclick", "handleMobileLinkClick()"),
			g.Text(node.Name),
		),
	)
}

// renderChevronIcon renders the folder chevron icon
func (rs Resource) renderChevronIcon(node *engine.TreeNode) gomponents.Node {
	return Span(
		Class(chevronClass),
		ID("chevron-"+node.Path),
		g.If(node.IsOpen, g.Text("▼")),
		g.If(!node.IsOpen, g.Text("▶")),
	)
}

// renderFolderChildren renders the children container for a folder
func (rs Resource) renderFolderChildren(node *engine.TreeNode, currentSlug string) gomponents.Node {
	if len(node.Children) == 0 {
		return g.Text("")
	}

	displayStyle := "display: none;"
	if node.IsOpen {
		displayStyle = "display: block;"
	}

	return Ul(
		Class("ml-4"),
		ID("folder-"+node.Path),
		g.Attr("style", displayStyle),
		g.Group(g.Map(node.Children, func(child *engine.TreeNode) gomponents.Node {
			return rs.renderTreeNode(child, currentSlug)
		})),
	)
}

// TOCItem represents a table of contents item
type TOCItem struct {
	ID    string
	Text  string
	Level int
}

// removeObsidianCallouts removes Obsidian callout notations from content
func removeObsidianCallouts(content string) string {
	// Regular expression to match Obsidian callout lines: > [!KEYWORD].*
	calloutRegex := regexp.MustCompile(`(?m)^>\s*\[![\w\-]+\].*$`)

	// Remove callout lines
	return calloutRegex.ReplaceAllString(content, "")
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
			baseSlug := engine.SlugifyHeading(text)
			id := baseSlug

			// Handle duplicate slugs by appending a number
			if count, exists := usedSlugs[baseSlug]; exists {
				usedSlugs[baseSlug] = count + 1
				id = fmt.Sprintf("%s-%d", baseSlug, count+1)
			} else {
				usedSlugs[baseSlug] = 1
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
		// H1 = no indent, H2 = small indent, H3+ = progressively more indent
		var indentClass string
		var textSizeClass string
		var fontWeightClass string

		switch item.Level {
		case 1:
			indentClass = ""
			textSizeClass = "text-sm"
			fontWeightClass = "font-semibold"
		case 2:
			indentClass = ""
			textSizeClass = "text-sm"
			fontWeightClass = "font-medium"
		case 3:
			indentClass = "ml-3"
			textSizeClass = "text-xs"
			fontWeightClass = "font-normal"
		case 4:
			indentClass = "ml-6"
			textSizeClass = "text-xs"
			fontWeightClass = "font-normal"
		case 5:
			indentClass = "ml-9"
			textSizeClass = "text-xs"
			fontWeightClass = "font-normal"
		default: // H6 and beyond
			indentClass = "ml-12"
			textSizeClass = "text-xs"
			fontWeightClass = "font-normal"
		}

		node := A(
			Href("#"+item.ID),
			Class(fmt.Sprintf("block py-1 px-2 text-gray-600 hover:text-gray-700 hover:bg-gray-50 rounded-md transition-colors [&.active]:text-purple-600 [&.active]:bg-gray-100 [&.active]:font-medium %s %s %s", indentClass, textSizeClass, fontWeightClass)),
			g.Attr("onclick", "handleTOCClick(event, this)"),
			g.Text(item.Text),
		)

		nodes = append(nodes, node)
	}

	return nodes
}

// renderYamlProperty renders a YAML property with appropriate HTML based on its type
func renderYamlProperty(key string, value any) gomponents.Node {
	return Div(
		Class("flex flex-row items-center py-3 border-b border-gray-100 last:border-b-0 transition-colors duration-150 hover:bg-gray-50"),
		Dt(
			Class("text-sm font-medium text-gray-700 ml-4 mb-2 sm:mb-0 sm:w-1/3"),
			g.Text(key),
		),
		Dd(
			Class("sm:w-2/3 mr-4 sm:mr-2"),
			renderYamlValue(value),
		),
	)
}

// renderYamlValue renders a YAML value with appropriate HTML based on its type
func renderYamlValue(value any) gomponents.Node {
	switch v := value.(type) {
	case bool:
		// Render boolean as a checkbox-style indicator
		return Div(
			Class("flex items-center gap-2"),
			Div(
				Class(fmt.Sprintf("w-4 h-4 rounded border-2 flex items-center justify-center %s",
					func() string {
						if v {
							return "bg-green-100 border-green-500 text-green-700"
						}
						return "bg-gray-100 border-gray-300 text-gray-400"
					}())),
				g.If(v, g.Text("✓")),
			),
		)
	case []interface{}:
		// Render array as pills/tags
		if len(v) == 0 {
			return Span(
				Class("text-sm text-gray-500 italic"),
				g.Text("(empty list)"),
			)
		}
		return Div(
			Class("flex flex-wrap gap-1"),
			g.Group(g.Map(v, func(item interface{}) gomponents.Node {
				itemStr := fmt.Sprintf("%v", item)
				// Check if the item contains markdown links (parsed wikilinks)
				if strings.Contains(itemStr, "](") && strings.Contains(itemStr, "[") {
					return Span(
						Class("inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800 border border-blue-200"),
						g.Raw(string(markdown.Markdown(itemStr))),
					)
				}
				return Span(
					Class("inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800 border border-gray-200"),
					g.Text(itemStr),
				)
			})),
		)
	case map[string]interface{}:
		// Render object as nested key-value pairs
		if len(v) == 0 {
			return Span(
				Class("text-sm text-gray-500 italic"),
				g.Text("(empty object)"),
			)
		}
		return Div(
			Class("bg-gray-50 border border-gray-200 rounded-md p-3"),
			Dl(
				Class("space-y-2"),
				g.Group(MapMapSorted(v, func(nestedKey string, nestedValue any) gomponents.Node {
					return Div(
						Class("flex flex-col sm:flex-row sm:items-center"),
						Dt(
							Class("text-xs font-medium text-gray-600 sm:w-1/3"),
							g.Text(nestedKey),
						),
						Dd(
							Class("text-xs text-gray-800 font-mono bg-white px-2 py-1 rounded border sm:w-2/3 mt-1 sm:mt-0"),
							g.Text(fmt.Sprintf("%v", nestedValue)),
						),
					)
				})),
			),
		)
	case string:
		// Render string with special handling for URLs, emails, etc.
		str := strings.TrimSpace(v)
		if str == "" {
			return Span(
				Class("text-sm text-gray-500 italic"),
				g.Text("(empty)"),
			)
		}

		// Check if the string contains markdown links (parsed wikilinks)
		if strings.Contains(str, "](") && strings.Contains(str, "[") {
			return Div(
				Class("text-sm text-gray-900 bg-slate-50 hover:bg-slate-100 px-3 py-1 rounded border border-slate-200 hover:border-slate-300 font-mono transition-all duration-150"),
				g.Raw(string(markdown.Markdown(str))),
			)
		}

		// Check if it's a URL
		if strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://") {
			return A(
				Href(str),
				Target("_blank"),
				Rel("noopener noreferrer"),
				Class("inline-flex items-center gap-1 text-sm text-blue-600 hover:text-blue-800 hover:underline bg-blue-50 px-3 py-1 rounded border border-blue-200 transition-colors"),
				g.Text(str),
				Span(
					Class("text-xs"),
					g.Text("↗"),
				),
			)
		}

		// Check if it's an email
		if strings.Contains(str, "@") && strings.Contains(str, ".") {
			return A(
				Href("mailto:"+str),
				Class("inline-flex items-center gap-1 text-sm text-purple-600 hover:text-purple-800 hover:underline bg-purple-50 px-3 py-1 rounded border border-purple-200 transition-colors"),
				g.Text(str),
				Span(
					Class("text-xs"),
					g.Text("✉"),
				),
			)
		}

		// Check if it's a date-like string
		if regexp.MustCompile(`^\d{4}-\d{2}-\d{2}`).MatchString(str) {
			return Div(
				Class("inline-flex items-center gap-1 text-sm text-indigo-700 bg-indigo-50 px-3 py-1 rounded border border-indigo-200"),
				Span(
					Class("text-xs"),
					g.Text("📅"),
				),
				g.Text(str),
			)
		}

		// Regular string
		return Div(
			Class("text-sm text-gray-900 bg-slate-50 hover:bg-slate-100 px-3 py-1 rounded border border-slate-200 hover:border-slate-300 font-mono transition-all duration-150"),
			g.Text(str),
		)
	case int, int32, int64, float32, float64:
		// Render numbers with special styling
		return Div(
			Class("inline-flex items-center gap-1 text-sm text-orange-700 bg-orange-50 px-3 py-1 rounded border border-orange-200 font-mono"),
			Span(
				Class("text-xs"),
				g.Text("#"),
			),
			g.Text(fmt.Sprintf("%v", v)),
		)
	default:
		// Fallback for unknown types
		return Div(
			Class("text-sm text-gray-900 bg-slate-50 hover:bg-slate-100 px-3 py-1 rounded border border-slate-200 hover:border-slate-300 font-mono transition-all duration-150"),
			g.Text(fmt.Sprintf("%v", value)),
		)
	}
}

// NoteWithList displays a note with the list of all notes on the left side
func (rs Resource) NoteWithList(notesService *engine.NotesService, note *model.Note, searchQuery string) (gomponents.Node, error) {

	matter := map[string]any{}
	var content []byte
	var slug string
	var title string
	var referencedBy []model.NoteReference

	if note != nil {
		// Parse wikilinks in metadata before using it
		matter = engine.ParseWikiLinksInMetadata(note.Metadata, notesService.GetTree())
		// Also parse tag links in metadata
		matter = engine.ParseTagLinksInMetadata(matter)
		slug = note.Slug
		title = note.Title
		referencedBy = note.ReferencedBy
		content = []byte(note.Content)
	} else {
		title = "404 : Not found"
		content = []byte("This note does not exist or is private.")
	}

	// Parse wiki-style links before markdown processing
	parsedContent := engine.ParseWikiLinks(string(content), notesService.GetTree())

	// Parse hashtags to clickable links
	parsedContent = engine.ParseHashtagLinks(parsedContent)

	parsedContent = engine.ProcessMarkdownLinks(parsedContent)

	// Remove Obsidian callout notations from the content
	parsedContent = removeObsidianCallouts(parsedContent)

	// Extract headings for table of contents
	tocItems := extractHeadings(parsedContent)

	// Filter tree based on search query
	displayTree := notesService.GetTree()
	if searchQuery != "" {
		displayTree = engine.FilterTreeBySearch(notesService.GetTree(), searchQuery)
	}

	// Get site title, icon, and description from environment variables
	siteTitle := os.Getenv("SITE_TITLE")
	if siteTitle == "" {
		siteTitle = "Pluie"
	}
	siteIcon := os.Getenv("SITE_ICON")
	siteDescription := os.Getenv("SITE_DESCRIPTION")

	return rs.Layout(
		note,
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
							Placeholder("Search page..."),
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
				Class("flex-1 container overflow-y-auto p-4 md:px-8"),
				H1(
					Class("text-3xl md:text-4xl font-bold mb-4 mt-2"),
					g.If(title != "", g.Text(title)),
				),
				g.If(len(matter) > 0 && os.Getenv("HIDE_YAML_FRONTMATTER") != "true",
					Div(
						Class("mb-6 opacity-80"),
						// YAML front matter header with toggle button
						Div(
							Class("flex items-center justify-between bg-gradient-to-br from-slate-50 to-slate-100 hover:from-slate-100 hover:to-slate-200 border border-slate-200 rounded-t-lg px-4 py-3 transition-all duration-200"),
							Div(
								Class("flex items-center gap-2"),
								Span(
									Class("text-xs font-mono text-gray-500 uppercase tracking-wide"),
									g.Textf("%d properties", len(matter)),
								),
							),
							Button(
								Class("flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900 transition-colors"),
								g.Attr("onclick", "toggleYamlFrontmatter()"),
								g.Attr("id", "yaml-toggle-btn"),
								Span(g.Text("Show")),
							),
						),
						// YAML front matter content (hidden by default)
						Div(
							Class("bg-white border-l border-r border-b border-gray-200 rounded-b-lg transition-all duration-300 overflow-hidden"),
							g.Attr("id", "yaml-content"),
							g.Attr("style", "display: none;"),
							Div(
								Dl(
									Class("grid grid-cols-1"),
									g.Group(MapMapSorted(matter, func(key string, value any) gomponents.Node {
										return renderYamlProperty(key, value)
									})),
								),
							),
						),
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
				Class("w-64 bg-white border-l border-gray-200 p-4 hidden md:flex flex-col h-full"),
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

// TagList displays all notes that contain a specific tag
func (rs Resource) TagList(notesService *engine.NotesService, tag string, notes []model.Note) (gomponents.Node, error) {
	// Get site title, icon, and description from environment variables
	siteTitle := os.Getenv("SITE_TITLE")
	if siteTitle == "" {
		siteTitle = "Pluie"
	}
	siteIcon := os.Getenv("SITE_ICON")
	siteDescription := os.Getenv("SITE_DESCRIPTION")

	var title string
	var content gomponents.Node

	if tag == "" {
		title = "Tag not found"
		content = Div(
			Class("prose max-w-none"),
			P(g.Text("No tag specified.")),
		)
	} else if len(notes) == 0 {
		title = fmt.Sprintf("Tag: #%s", tag)
		content = Div(
			Class("prose max-w-none"),
			P(g.Textf("No notes found with tag #%s.", tag)),
		)
	} else {
		title = fmt.Sprintf("Tag: #%s (%d notes)", tag, len(notes))
		content = Div(
			Class("prose max-w-none"),
			P(
				Class("text-gray-600 mb-6"),
				g.Textf("Found %d notes with tag #%s:", len(notes), tag),
			),
			Div(
				Class("grid gap-4 md:grid-cols-2 lg:grid-cols-3"),
				g.Group(g.Map(notes, func(note model.Note) gomponents.Node {
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
								g.Group(g.Map(notesService.GetTree().Children, func(child *engine.TreeNode) gomponents.Node {
									return rs.renderTreeNode(child, "")
								})),
							),
						),
					),
				),
			),
			// Main content area with the tag results
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

// renderNoteCard renders a single note as a card for the tag list view
func (rs Resource) renderNoteCard(note model.Note) gomponents.Node {
	// Extract first few lines of content for description
	description := extractDescription(note.Content)

	return Div(
		Class("bg-white border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow"),
		A(
			Href("/"+note.Slug),
			Class("block"),
			g.Attr("hx-boost", "true"),
			H3(
				Class("text-lg font-semibold text-gray-900 mb-2 hover:text-blue-600"),
				g.Text(note.Title),
			),
			g.If(description != "",
				P(
					Class("text-sm text-gray-600 line-clamp-3"),
					g.Text(description),
				),
			),
		),
	)
}

// extractDescription extracts the first few lines of content for a description
func extractDescription(content string) string {
	// Remove markdown headers and get first paragraph
	lines := strings.Split(content, "\n")
	var description strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and headers
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Stop at first substantial line and use it as description
		if len(line) > 10 {
			description.WriteString(line)
			break
		}
	}

	desc := description.String()
	// Truncate if too long
	if len(desc) > 150 {
		desc = desc[:150] + "..."
	}

	return desc
}
