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
	Tree  *TreeNode
}

// TreeNode represents a node in the file tree structure (imported from main package)
type TreeNode struct {
	Name     string      `json:"name"`     // Display name (folder name or note title)
	Path     string      `json:"path"`     // Full path from root
	IsFolder bool        `json:"isFolder"` // True if this is a folder, false if it's a note
	Note     *model.Note `json:"note"`     // Reference to the note if this is a note node
	Children []*TreeNode `json:"children"` // Child nodes (subfolders and notes)
	IsOpen   bool        `json:"isOpen"`   // Whether the folder is expanded in the UI
}

func MapMap[T any](ts map[string]T, cb func(k string, v T) g.Node) []g.Node {
	var nodes []g.Node
	for k, v := range ts {
		nodes = append(nodes, cb(k, v))
	}
	return nodes
}

// filterTreeBySearch filters the tree to only show nodes that match the search query
func (rs Resource) filterTreeBySearch(root *TreeNode, query string) *TreeNode {
	if root == nil || query == "" {
		return root
	}

	query = strings.ToLower(query)

	// Create a new root for filtered results
	filteredRoot := &TreeNode{
		Name:     root.Name,
		Path:     root.Path,
		IsFolder: true,
		Children: make([]*TreeNode, 0),
		IsOpen:   true,
	}

	// Recursively filter children
	rs.filterChildren(root, filteredRoot, query)

	return filteredRoot
}

// filterChildren recursively filters children based on search query
func (rs Resource) filterChildren(source *TreeNode, target *TreeNode, query string) {
	for _, child := range source.Children {
		if child.IsFolder {
			// Check if folder name matches the search query
			folderMatches := strings.Contains(strings.ToLower(child.Name), query)

			// Create temp folder to check for matching descendants
			tempFolder := &TreeNode{
				Name:     child.Name,
				Path:     child.Path,
				IsFolder: true,
				Children: make([]*TreeNode, 0),
				IsOpen:   true, // Open folders in search results
			}

			// Recursively filter children
			rs.filterChildren(child, tempFolder, query)

			// Include folder if:
			// 1. The folder name itself matches, OR
			// 2. The folder has matching descendants
			if folderMatches || len(tempFolder.Children) > 0 {
				// If folder name matches, include ALL its contents
				if folderMatches {
					tempFolder = rs.copyEntireSubtree(child)
					tempFolder.IsOpen = true // Ensure matched folders are open
				}
				target.Children = append(target.Children, tempFolder)
			}
		} else {
			// For notes, check if title matches
			if strings.Contains(strings.ToLower(child.Name), query) {
				noteNode := &TreeNode{
					Name:     child.Name,
					Path:     child.Path,
					IsFolder: false,
					Note:     child.Note,
					Children: make([]*TreeNode, 0),
				}
				target.Children = append(target.Children, noteNode)
			}
		}
	}
}

// copyEntireSubtree creates a complete copy of a subtree with all its contents
func (rs Resource) copyEntireSubtree(source *TreeNode) *TreeNode {
	if source == nil {
		return nil
	}

	copy := &TreeNode{
		Name:     source.Name,
		Path:     source.Path,
		IsFolder: source.IsFolder,
		Note:     source.Note,
		IsOpen:   true, // Open all folders in search results
		Children: make([]*TreeNode, len(source.Children)),
	}

	for i, child := range source.Children {
		copy.Children[i] = rs.copyEntireSubtree(child)
	}

	return copy
}

// renderTreeNode renders a single tree node with its children
func (rs Resource) renderTreeNode(node *TreeNode, currentSlug string) gomponents.Node {
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
					Class("flex items-center text-left w-full px-2 py-1 text-gray-700 hover:bg-gray-100 rounded"),
					g.Attr("onclick", fmt.Sprintf("toggleFolder('%s')", node.Path)),
					// Folder icon (chevron)
					Span(
						Class("mr-2 transition-transform duration-200"),
						ID("chevron-"+node.Path),
						g.If(node.IsOpen,
							g.Text("▼"),
						),
						g.If(!node.IsOpen,
							g.Text("▶"),
						),
					),
					// Folder icon
					Span(
						Class("mr-2"),
						g.Text("📁"),
					),
					g.Text(node.Name),
				),
			),
			// Children container
			g.If(len(node.Children) > 0,
				Ul(
					Class("ml-4 space-y-1"),
					ID("folder-"+node.Path),
					g.If(node.IsOpen,
						g.Attr("style", "display: block;"),
					),
					g.If(!node.IsOpen,
						g.Attr("style", "display: none;"),
					),
					g.Group(g.Map(node.Children, func(child *TreeNode) gomponents.Node {
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
			linkClass = "flex items-center px-2 py-1 text-blue-800 bg-blue-100 rounded-md font-medium"
		} else {
			linkClass = "flex items-center px-2 py-1 text-blue-600 hover:text-blue-800 hover:bg-blue-50 rounded-md"
		}

		return Li(
			A(
				Href("/"+node.Path),
				Class(linkClass),
				// Note icon
				Span(
					Class("mr-2"),
					g.Text("📄"),
				),
				g.Text(node.Name),
			),
		)
	}
}

// countNotesInTree counts the total number of notes in a tree
func (rs Resource) countNotesInTree(node *TreeNode) int {
	if node == nil {
		return 0
	}

	count := 0
	if !node.IsFolder && node.Note != nil {
		count = 1
	}

	for _, child := range node.Children {
		count += rs.countNotesInTree(child)
	}

	return count
}

// NoteWithList displays a note with the list of all notes on the left side
func (rs Resource) NoteWithList(note *model.Note, searchQuery string) (gomponents.Node, error) {

	matter := map[string]any{}
	var content []byte
	var slug string
	var title string
	var referencedBy []model.NoteReference

	if note != nil {
		var err error
		content, err = frontmatter.Parse(strings.NewReader(note.Content), &matter)
		if err != nil {
			content = []byte(note.Content)
			fmt.Println("Error parsing frontmatter:", err)
		}
		slug = note.Slug
		title = note.Title
		referencedBy = note.ReferencedBy
	} else {
		content = []byte("Note is nil")
	}

	// Parse wiki-style links before markdown processing
	parsedContent := rs.parseWikiLinks(string(content))

	// Filter tree based on search query
	var displayTree *TreeNode
	if searchQuery != "" {
		displayTree = rs.filterTreeBySearch(rs.Tree, searchQuery)
	} else {
		displayTree = rs.Tree
	}

	return rs.Layout(
		Div(
			Class("flex gap-6 h-screen"),
			// Left sidebar with notes list
			Div(
				Class("w-1/4 bg-gray-50 p-4 rounded-lg flex flex-col h-full"),
				H2(
					Class("text-xl font-bold mb-4"),
					g.Text("Notes"),
				),
				// Search form
				Form(
					Method("GET"),
					Action("/"+slug),
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
							g.Attr("hx-get", "/"+slug),
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
				// Fold/Unfold all buttons
				Div(
					Class("mb-4 flex gap-2"),
					Button(
						Class("px-3 py-1 text-sm bg-gray-200 hover:bg-gray-300 text-gray-700 rounded-md transition-colors"),
						g.Attr("onclick", "expandAllFolders()"),
						g.Text("Expand All"),
					),
					Button(
						Class("px-3 py-1 text-sm bg-gray-200 hover:bg-gray-300 text-gray-700 rounded-md transition-colors"),
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
							g.Textf("Found %d notes matching \"%s\"", rs.countNotesInTree(displayTree), searchQuery),
						),
					),
					// Notes tree
					Div(
						ID("notes-list"),
						Class(""),
						g.If(displayTree != nil && len(displayTree.Children) > 0,
							Ul(
								Class("space-y-1"),
								g.Group(g.Map(displayTree.Children, func(child *TreeNode) gomponents.Node {
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
			// Right content area with the note
			Div(
				Class("flex-1 overflow-y-auto p-4"),
				H1(
					Class("text-3xl font-bold mb-4"),
					g.If(title != "", g.Text(title)),
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
		),
	), nil
}
