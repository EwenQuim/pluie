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
func (r Resource) Home() gomponents.Node {
	return HTML(
		Head(
			Meta(Charset("utf-8")),
			Meta(Name("viewport"), Content("width=device-width, initial-scale=1")),
			Title("Obsidian"),
			Link(Rel("stylesheet"), Type("text/css"), Href("/static/style.css")),
		),
		Body(
			H1(g.Text("Obsidian")),
			Ul(
				g.Group(g.Map(r.Notes, func(note model.Note) gomponents.Node {
					return Li(
						A(Href("/"+note.Slug), g.Text(note.Title)),
					)
				}),
				),
			),
		),
	)
}
