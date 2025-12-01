package template

import (
	"fmt"
	"strings"

	"github.com/EwenQuim/pluie/model"
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
)

func (rs Resource) Layout(note *model.Note, node ...g.Node) g.Node {
	// Get base site configuration from Config
	baseSiteTitle := rs.cfg.SiteTitle
	siteIcon := rs.cfg.SiteIcon
	baseSiteDescription := rs.cfg.SiteDescription

	// Compute SEO data using the separated function
	seoData := ComputeSEOData(note, baseSiteTitle, baseSiteDescription)

	return HTML(
		Head(
			Meta(Charset("utf-8")),
			Meta(Name("viewport"), Content("width=device-width, initial-scale=1")),
			TitleEl(g.Text(seoData.PageTitle)),
			// Add favicon if icon is provided
			g.If(siteIcon != "",
				Link(Rel("icon"), Href(siteIcon)),
			),
			// Basic SEO meta tags
			Meta(Name("application-name"), Content(baseSiteTitle)),
			g.If(seoData.Description != "",
				Meta(Name("description"), Content(seoData.Description)),
			),
			Meta(Name("msapplication-TileImage"), Content(siteIcon)),

			// Keywords meta tag
			g.If(len(seoData.Keywords) > 0,
				Meta(Name("keywords"), Content(strings.Join(seoData.Keywords, ", "))),
			),

			// Canonical URL
			g.If(seoData.CanonicalURL != "",
				Link(Rel("canonical"), Href(seoData.CanonicalURL)),
			),

			// Open Graph meta tags
			Meta(g.Attr("property", "og:type"), Content(seoData.OGType)),
			Meta(g.Attr("property", "og:title"), Content(seoData.PageTitle)),
			g.If(seoData.Description != "",
				Meta(g.Attr("property", "og:description"), Content(seoData.Description)),
			),
			g.If(seoData.CanonicalURL != "",
				Meta(g.Attr("property", "og:url"), Content(seoData.CanonicalURL)),
			),
			Meta(g.Attr("property", "og:site_name"), Content(baseSiteTitle)),
			g.If(siteIcon != "",
				Meta(g.Attr("property", "og:image"), Content(siteIcon)),
			),

			// Twitter Card meta tags
			Meta(Name("twitter:card"), Content("summary")),
			Meta(Name("twitter:title"), Content(seoData.PageTitle)),
			g.If(seoData.Description != "",
				Meta(Name("twitter:description"), Content(seoData.Description)),
			),
			g.If(siteIcon != "",
				Meta(Name("twitter:image"), Content(siteIcon)),
			),

			// Additional SEO meta tags for articles
			g.If(note != nil && seoData.OGType == "article",
				g.Group([]g.Node{
					g.If(seoData.AuthorMeta != nil,
						Meta(Name("author"), Content(fmt.Sprintf("%v", seoData.AuthorMeta))),
					),
					g.If(seoData.DateMeta != nil,
						Meta(g.Attr("property", "article:published_time"), Content(fmt.Sprintf("%v", seoData.DateMeta))),
					),
					g.If(seoData.ModifiedMeta != nil,
						Meta(g.Attr("property", "article:modified_time"), Content(fmt.Sprintf("%v", seoData.ModifiedMeta))),
					),
				}),
			),

			Link(Rel("stylesheet"), Type("text/css"), Href("/static/tailwind.min.css")),
			Script(Defer(), Src("/static/htmx.js")),
			Script(Defer(), Src("/static/sse.js")),
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
