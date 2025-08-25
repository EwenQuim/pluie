package template

import (
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
)

func (rs Resource) Layout(node ...g.Node) g.Node {
	return HTML(
		Head(
			Meta(Charset("utf-8")),
			Meta(Name("viewport"), Content("width=device-width, initial-scale=1")),
			Title("Pluie"),
			// Link(Rel("stylesheet"), Type("text/css"), Href("/static/style.css")),
			Link(Rel("stylesheet"), Type("text/css"), Href("/static/tailwind.min.css")),
			Script(Defer(), Src("https://cdn.jsdelivr.net/npm/htmx.org@2.0.6/dist/htmx.min.js")),
		),
		Body(
			ID("app"),
			Class("container mx-auto"),
			Main(
				node...,
			),
			Footer(
				Class("text-center mt-8 mb-4 text-gray-600"),
				P(
					Class("italic"),
					A(
						Class("underline"),
						Href("https://github.com/EwenQuim/pluie"),
						g.Text("Pluie"),
					),
					g.Text(", a simple note-taking app, powered by "),
					A(
						Class("underline"),
						Href("https://github.com/go-fuego/fuego"),
						g.Text("Fuego"),
					),
					g.Text("."),
				),
				P(
					g.Text("0% maintainance required, 100% obsidian compatible"),
				),
			),
		),
	)
}
