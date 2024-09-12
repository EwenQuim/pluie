package template

import (
	"github.com/EwenQuim/pluie/model"
	"github.com/maragudk/gomponents"
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
)

type Resource struct {
	Notes []model.Note
}

// Displays the list of notes
func (rs Resource) List() gomponents.Node {
	return rs.Layout(
		H1(g.Text("Obsidian")),
		Ul(
			g.Group(g.Map(rs.Notes, func(note model.Note) gomponents.Node {
				return Li(
					A(Href("/"+note.Slug), g.Text(note.Title)),
				)
			}),
			),
		),
	)
}
