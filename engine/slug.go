package engine

import (
	"net/url"
	"regexp"
	"strings"
)

// SlugifyOptions defines options for slugification behavior
type SlugifyOptions struct {
	// PreserveSlashes keeps forward slashes in the slug (useful for paths)
	PreserveSlashes bool
	// URLEncode applies URL encoding to the result
	URLEncode bool
	// RemoveExtension removes file extensions like .md
	RemoveExtension bool
	// TrimSlashes removes leading and trailing slashes
	TrimSlashes bool
	// PreserveCase keeps the original case instead of converting to lowercase
	PreserveCase bool
}

// DefaultNoteSlugOptions returns the default options for note slugification
func DefaultNoteSlugOptions() SlugifyOptions {
	return SlugifyOptions{
		PreserveSlashes: true,
		URLEncode:       true,
		RemoveExtension: true,
		TrimSlashes:     true,
	}
}

// DefaultHeadingSlugOptions returns the default options for heading slugification
func DefaultHeadingSlugOptions() SlugifyOptions {
	return SlugifyOptions{
		PreserveSlashes: false,
		URLEncode:       false,
		RemoveExtension: false,
		TrimSlashes:     false,
	}
}

// Slugify converts a string to a URL-friendly slug with configurable options
func Slugify(text string, options SlugifyOptions) string {
	if text == "" {
		return ""
	}

	slug := text

	// Remove file extension if requested
	if options.RemoveExtension {
		slug = strings.TrimSuffix(slug, ".md")
	}

	// Convert to lowercase unless preserving case
	if !options.PreserveCase {
		slug = strings.ToLower(slug)
	}

	if options.PreserveSlashes {
		// For paths: replace spaces with dashes but preserve forward slashes
		slug = strings.ReplaceAll(slug, " ", "-")

		// Clean up multiple consecutive dashes
		slug = cleanMultipleDashes(slug)

		// Clean leading/trailing slashes and dashes if requested
		if options.TrimSlashes {
			slug = strings.Trim(slug, "/-")
		}

		// URL encode while preserving forward slashes
		if options.URLEncode {
			slug = url.PathEscape(slug)
			slug = strings.ReplaceAll(slug, "%2F", "/")
		}
	} else {
		// For simple slugs: replace all non-alphanumeric characters with dashes
		slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")

		// Remove leading and trailing dashes
		slug = strings.Trim(slug, "-")
	}

	return slug
}

// SlugifyNote creates a slug for a note using note-specific options
func SlugifyNote(text string) string {
	return Slugify(text, DefaultNoteSlugOptions())
}

// SlugifyHeading creates a slug for a heading using heading-specific options
func SlugifyHeading(text string) string {
	return Slugify(text, DefaultHeadingSlugOptions())
}

// SlugifyNoteWithCaseLogic creates a slug for a note with special case-preserving logic
// This function preserves case when creating from titles but converts existing slugs to lowercase
func SlugifyNoteWithCaseLogic(text, existingSlug string) string {
	usingTitle := existingSlug == ""

	options := SlugifyOptions{
		PreserveSlashes: true,
		URLEncode:       true,
		RemoveExtension: true,
		TrimSlashes:     true,
		PreserveCase:    usingTitle, // Preserve case only when creating from title
	}

	// Use the provided text (either title or existing slug)
	return Slugify(text, options)
}

// cleanMultipleDashes removes consecutive dashes using a single pass
func cleanMultipleDashes(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	prevDash := false
	for _, r := range s {
		if r == '-' {
			if !prevDash {
				result.WriteRune(r)
				prevDash = true
			}
		} else {
			result.WriteRune(r)
			prevDash = false
		}
	}

	return result.String()
}
