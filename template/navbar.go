package template

import (
	"fmt"
	"os"

	"github.com/EwenQuim/pluie/engine"
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
)

// navbarConfig holds configuration for rendering the navbar
type navbarConfig struct {
	siteTitle       string
	siteIcon        string
	siteDescription string
	currentSlug     string           // Current note slug for search form action
	searchQuery     string           // Current search query value
	displayTree     *engine.TreeNode // Optional filtered tree to display (if nil, uses full tree from notesService)
	mainContent     g.Node           // Main content area
}

// renderWithNavbar renders a page with consistent navbar structure
func (rs Resource) renderWithNavbar(notesService *engine.NotesService, config navbarConfig) g.Node {
	// Get site configuration from environment if not provided
	if config.siteTitle == "" {
		config.siteTitle = os.Getenv("SITE_TITLE")
		if config.siteTitle == "" {
			config.siteTitle = "Pluie"
		}
	}
	if config.siteIcon == "" {
		config.siteIcon = os.Getenv("SITE_ICON")
	}
	if config.siteDescription == "" {
		config.siteDescription = os.Getenv("SITE_DESCRIPTION")
	}

	return Div(
		Class("flex flex-col md:flex-row md:gap-2 h-screen w-screen justify-between"),
		// Mobile top bar (hidden on desktop)
		rs.renderMobileTopBar(config.siteTitle, config.siteIcon),
		// Mobile sidebar overlay
		rs.renderMobileSidebarOverlay(),
		// Left sidebar with notes list
		rs.renderLeftSidebar(notesService, config),
		// Main content area
		config.mainContent,
	)
}

// renderMobileTopBar renders the mobile top navigation bar
func (rs Resource) renderMobileTopBar(siteTitle, siteIcon string) g.Node {
	return Div(
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
	)
}

// renderMobileSidebarOverlay renders the overlay for mobile sidebar
func (rs Resource) renderMobileSidebarOverlay() g.Node {
	return Div(
		Class("fixed inset-0 bg-black bg-opacity-50 z-40 md:hidden opacity-0 invisible transition-all duration-300"),
		ID("mobile-sidebar-overlay"),
		g.Attr("onclick", "closeMobileSidebar()"),
	)
}

// renderLeftSidebar renders the left sidebar with navigation
func (rs Resource) renderLeftSidebar(notesService *engine.NotesService, config navbarConfig) g.Node {
	return Div(
		Class("w-3/4 md:w-1/4 max-w-md bg-white border-r border-gray-200 p-4 flex flex-col h-full md:relative fixed top-0 left-0 z-50 md:z-auto -translate-x-full md:translate-x-0 transition-transform duration-300 ease-in-out"),
		ID("mobile-sidebar"),
		// Site header with title and icon
		Div(
			Class("mb-6"),
			Div(
				Class("flex items-center gap-3 mb-2"),
				// Site icon
				g.If(config.siteIcon != "",
					Img(
						Src(config.siteIcon),
						Alt("Site Icon"),
						Class("w-8 h-8 object-contain rounded-md"),
					),
				),
				// Site title
				H1(
					Class("text-xl font-bold text-gray-900"),
					g.Text(config.siteTitle),
				),
			),
			// Site description
			g.If(config.siteDescription != "",
				P(
					Class("text-sm text-gray-500 italic"),
					g.Text(config.siteDescription),
				),
			),
		),
		// Navigation links
		Div(
			Class("mb-6"),
			// Simple search link
			A(
				Href("/-/search"),
				Class("w-full inline-flex border border-gray-300 items-center gap-2 px-3 py-2 text-sm text-gray-700 hover:text-gray-900 hover:bg-gray-50 rounded-md transition-colors"),
				g.Attr("hx-boost", "true"),
				Span(g.Text("🔍")),
				g.Text("Search (c+K)"),
			),
		),
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

		// Scrollable container for notes tree
		Div(
			Class("flex-1 overflow-y-auto"),
			func() g.Node {
				// Use displayTree if provided, otherwise use full tree from notesService
				tree := config.displayTree
				if tree == nil {
					tree = notesService.GetTree()
				}

				return g.Group([]g.Node{
					// Results info (only show if search query is present)
					g.If(config.searchQuery != "",
						P(
							Class("text-gray-600 mb-2 text-sm"),
							g.Text(fmt.Sprintf("Found %d notes matching \"%s\"", countNotesInTree(tree), config.searchQuery)),
						),
					),
					// Notes tree
					Div(
						ID("notes-list"),
						Class(""),
						g.If(tree != nil && len(tree.Children) > 0,
							Ul(
								Class(""),
								g.Group(g.Map(tree.Children, func(child *engine.TreeNode) g.Node {
									return rs.renderTreeNode(child, config.currentSlug)
								})),
							),
						),
					),
					// Show message if no results found
					g.If(config.searchQuery != "" && (tree == nil || len(tree.Children) == 0),
						P(
							Class("text-gray-500 mt-4 text-sm"),
							g.Text("No notes found matching your search."),
						),
					),
				})
			}(),
		),
	)
}
