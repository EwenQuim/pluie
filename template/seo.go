package template

import (
	"fmt"
	"strings"

	"github.com/EwenQuim/pluie/model"
)

// SEOData contains all the computed SEO properties for a page
type SEOData struct {
	PageTitle    string
	Description  string
	CanonicalURL string
	OGType       string
	Keywords     []string
	AuthorMeta   interface{}
	DateMeta     interface{}
	ModifiedMeta interface{}
}

// ComputeSEOData extracts and computes SEO properties from a note and site configuration
func ComputeSEOData(note *model.Note, baseSiteTitle, baseSiteDescription string) SEOData {
	var seoData SEOData

	if note != nil {
		// Custom page title: "Note Title | Site Title"
		if note.Title != "" {
			seoData.PageTitle = fmt.Sprintf("%s | %s", note.Title, baseSiteTitle)
		} else {
			seoData.PageTitle = baseSiteTitle
		}

		// Extract description from note metadata or content
		if note.Metadata != nil {
			if desc, exists := note.Metadata["description"]; exists {
				if descStr, ok := desc.(string); ok && descStr != "" {
					seoData.Description = descStr
				}
			}

			// Extract metadata for SEO tags
			seoData.AuthorMeta = note.Metadata["author"]
			seoData.DateMeta = note.Metadata["date"]
			seoData.ModifiedMeta = note.Metadata["modified"]

			// Extract keywords from metadata
			if tags, exists := note.Metadata["tags"]; exists {
				if tagSlice, ok := tags.([]interface{}); ok {
					for _, tag := range tagSlice {
						if tagStr, ok := tag.(string); ok {
							seoData.Keywords = append(seoData.Keywords, tagStr)
						}
					}
				}
			}
		}

		// If no description in metadata, try to extract from content (first 160 chars)
		if seoData.Description == "" && note.Content != "" {
			content := strings.TrimSpace(note.Content)
			// Remove markdown syntax for a cleaner description
			content = strings.ReplaceAll(content, "#", "")
			content = strings.ReplaceAll(content, "*", "")
			content = strings.ReplaceAll(content, "_", "")
			content = strings.TrimSpace(content)

			if len(content) > 160 {
				seoData.Description = content[:157] + "..."
			} else {
				seoData.Description = content
			}
		}

		// Fallback to base site description
		if seoData.Description == "" {
			seoData.Description = baseSiteDescription
		}

		// Generate canonical URL
		if note.Slug != "" {
			seoData.CanonicalURL = fmt.Sprintf("/%s", note.Slug)
		}

		// Set Open Graph type
		seoData.OGType = "article"
	} else {
		// Default values when no note is provided
		seoData.PageTitle = baseSiteTitle
		seoData.Description = baseSiteDescription
		seoData.OGType = "website"
	}

	return seoData
}
