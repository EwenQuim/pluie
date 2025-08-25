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
			Link(Rel("stylesheet"), Type("text/css"), Href("/static/style.css")),
			Link(Rel("stylesheet"), Type("text/css"), Href("/static/tailwind.min.css")),
			Script(Defer(), Src("https://cdn.jsdelivr.net/npm/htmx.org@2.0.6/dist/htmx.min.js")),
		),
		Body(
			ID("app"),
			Class("container mx-auto"),
			g.Attr("hx-boost"),
			g.Attr("hx-push-url"),
			g.Attr("hx-target", "#app"),
			g.Attr("hx-select", "#app"),
			g.Attr("hx-swap", "outerHTML"),

			H1(
				Class("text-4xl font-bold text-center my-4"),
				A(Href("/"), g.Text("Pluie")),
			),
			Main(
				node...,
			),
		),
	)
}
