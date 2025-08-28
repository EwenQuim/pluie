package template

import (
	"os"

	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
)

func (rs Resource) Layout(node ...g.Node) g.Node {
	// Get title, icon, and description from environment variables
	siteTitle := os.Getenv("SITE_TITLE")
	if siteTitle == "" {
		siteTitle = "Pluie"
	}

	siteIcon := os.Getenv("SITE_ICON")
	if siteIcon == "" {
		siteIcon = "/static/pluie.webp"
	}
	siteDescription := os.Getenv("SITE_DESCRIPTION")

	return HTML(
		Head(
			Meta(Charset("utf-8")),
			Meta(Name("viewport"), Content("width=device-width, initial-scale=1")),
			Title(siteTitle),
			// Add favicon if icon is provided
			g.If(siteIcon != "",
				Link(Rel("icon"), Href(siteIcon)),
			),
			// Add meta tags for site title, description, and icon
			Meta(Name("application-name"), Content(siteTitle)),
			g.If(siteDescription != "",
				Meta(Name("description"), Content(siteDescription)),
			),
			Meta(Name("msapplication-TileImage"), Content(siteIcon)),

			Link(Rel("stylesheet"), Type("text/css"), Href("/static/tailwind.min.css")),
			Script(Defer(), Src("https://cdn.jsdelivr.net/npm/htmx.org@2.0.6/dist/htmx.min.js")),
			Script(Defer(), Src("/static/app.js")),
		),
		Body(
			ID("app"),
			Class("scroll-smooth"),
			Main(
				node...,
			),
		),
	)
}
