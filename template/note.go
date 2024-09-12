package template

import (
	"github.com/EwenQuim/pluie/model"
	"github.com/maragudk/gomponents"
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"

	"github.com/go-fuego/fuego/extra/markdown"
)

func (rs Resource) Note(note model.Note) gomponents.Node {
	return Div(
		H1(g.Text(note.Title)),
		Div(
			g.Raw(string(markdown.Markdown(note.Content))),
		),
	)
}
