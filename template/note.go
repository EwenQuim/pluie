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

func MapMap[T any](ts map[string]T, cb func(k string, v T) g.Node) []g.Node {
	var nodes []g.Node
	for k, v := range ts {
		nodes = append(nodes, cb(k, v))
	}
	return nodes
}

func (rs Resource) Note(note model.Note) (gomponents.Node, error) {
	matter := map[string]any{}
	content, err := frontmatter.Parse(strings.NewReader(note.Content), &matter)
	if err != nil {
		content = []byte(note.Content)
		fmt.Println("Error parsing frontmatter:", err)
	}

	return rs.Layout(
		H1(g.Text(note.Title)),
		Ul(
			Class("bg-gray-100 p-4 rounded-lg"),
			g.Group(MapMap(matter, func(key string, value any) gomponents.Node {
				return Li(
					g.Text(fmt.Sprintf("%s: %v", key, value)),
				)
			})),
		),

		Div(
			Class("prose lg:prose-xl"),
			g.Raw(string(markdown.Markdown(string(content)))),
		),
	), nil
}
