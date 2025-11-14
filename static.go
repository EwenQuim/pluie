package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	"github.com/EwenQuim/pluie/static"
	"github.com/EwenQuim/pluie/template"
)

const distFolder = "dist"

// generateStaticSite generates a static version of the site in the /dist folder
func generateStaticSite(notesService *engine.NotesService, cfg *config.Config) error {
	rs := template.Resource{}

	// Create dist folder
	if err := os.RemoveAll(distFolder); err != nil {
		return fmt.Errorf("failed to remove existing dist folder: %w", err)
	}
	if err := os.MkdirAll(distFolder, 0755); err != nil {
		return fmt.Errorf("failed to create dist folder: %w", err)
	}

	slog.Info("Generating static site", "folder", distFolder)

	// Copy static assets
	if err := copyStaticAssets(); err != nil {
		return fmt.Errorf("failed to copy static assets: %w", err)
	}

	// Get home note slug
	homeNoteSlug := getHomeNoteSlug(notesService, cfg)

	// Generate home page (index.html)
	if err := generateHomePage(notesService, rs, homeNoteSlug); err != nil {
		return fmt.Errorf("failed to generate home page: %w", err)
	}

	// Generate all note pages
	if err := generateNotePages(notesService, rs, cfg); err != nil {
		return fmt.Errorf("failed to generate note pages: %w", err)
	}

	// Generate tag pages
	if err := generateTagPages(notesService, rs); err != nil {
		return fmt.Errorf("failed to generate tag pages: %w", err)
	}

	slog.Info("Static site generation complete")
	return nil
}

// getHomeNoteSlug determines the home note slug based on priority
func getHomeNoteSlug(notesService *engine.NotesService, cfg *config.Config) string {
	// Priority 1: Check HOME_NOTE_SLUG configuration
	if cfg.HomeNoteSlug != "" {
		if _, exists := notesService.GetNote(cfg.HomeNoteSlug); exists {
			return cfg.HomeNoteSlug
		}
	}

	// Priority 2: First note in alphabetical order
	tree := notesService.GetTree()
	if tree != nil {
		notes := engine.GetAllNotesFromTree(tree)
		if len(notes) > 0 {
			sort.Slice(notes, func(i, j int) bool {
				return notes[i].Slug < notes[j].Slug
			})
			return notes[0].Slug
		}
	}

	return ""
}

// generateHomePage generates the home page at /dist/index.html
func generateHomePage(notesService *engine.NotesService, rs template.Resource, homeNoteSlug string) error {
	slog.Info("Generating home page", "slug", homeNoteSlug)

	var note *model.Note
	if homeNoteSlug != "" {
		if n, ok := notesService.GetNote(homeNoteSlug); ok {
			note = &n
		}
	}

	// Render the home page
	node, err := rs.NoteWithList(notesService, note, "")
	if err != nil {
		return fmt.Errorf("failed to render home page: %w", err)
	}

	// Write to index.html
	indexPath := filepath.Join(distFolder, "index.html")
	if err := writeNodeToFile(node, indexPath); err != nil {
		return fmt.Errorf("failed to write index.html: %w", err)
	}

	slog.Info("Home page generated", "path", indexPath)
	return nil
}

// generateNotePages generates HTML pages for all public notes
func generateNotePages(notesService *engine.NotesService, rs template.Resource, cfg *config.Config) error {
	tree := notesService.GetTree()
	if tree == nil {
		slog.Warn("No tree found, skipping note pages")
		return nil
	}

	notes := engine.GetAllNotesFromTree(tree)
	slog.Info("Generating note pages", "count", len(notes))

	for _, note := range notes {
		// Skip private notes if not public by default
		if !cfg.PublicByDefault && !note.IsPublic {
			slog.Debug("Skipping private note", "slug", note.Slug)
			continue
		}

		// Render the note page
		node, err := rs.NoteWithList(notesService, &note, "")
		if err != nil {
			return fmt.Errorf("failed to render note %s: %w", note.Slug, err)
		}

		// Write to {slug}/index.html or {slug}.html
		notePath := filepath.Join(distFolder, note.Slug, "index.html")

		// Create directory if needed
		noteDir := filepath.Dir(notePath)
		if err := os.MkdirAll(noteDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for note %s: %w", note.Slug, err)
		}

		if err := writeNodeToFile(node, notePath); err != nil {
			return fmt.Errorf("failed to write note %s: %w", note.Slug, err)
		}

		slog.Debug("Note page generated", "slug", note.Slug, "path", notePath)
	}

	slog.Info("Note pages generated", "count", len(notes))
	return nil
}

// generateTagPages generates HTML pages for all tags
func generateTagPages(notesService *engine.NotesService, rs template.Resource) error {
	tagIndex := notesService.GetTagIndex()
	allTags := tagIndex.GetAllTags()

	slog.Info("Generating tag pages", "count", len(allTags))

	for _, tag := range allTags {
		// Get all notes that contain this tag
		notesWithTag := tagIndex.GetNotesWithTag(tag)

		// Render the tag page
		node, err := rs.TagList(notesService, tag, notesWithTag)
		if err != nil {
			return fmt.Errorf("failed to render tag %s: %w", tag, err)
		}

		// Write to /-/tag/{tag}/index.html
		// Sanitize tag for file path
		sanitizedTag := strings.ReplaceAll(tag, "/", "-")
		tagPath := filepath.Join(distFolder, "-", "tag", sanitizedTag, "index.html")

		// Create directory if needed
		tagDir := filepath.Dir(tagPath)
		if err := os.MkdirAll(tagDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for tag %s: %w", tag, err)
		}

		if err := writeNodeToFile(node, tagPath); err != nil {
			return fmt.Errorf("failed to write tag %s: %w", tag, err)
		}

		slog.Debug("Tag page generated", "tag", tag, "path", tagPath)
	}

	slog.Info("Tag pages generated", "count", len(allTags))
	return nil
}

// copyStaticAssets copies the static assets to /dist/static
func copyStaticAssets() error {
	slog.Info("Copying static assets")

	staticDir := filepath.Join(distFolder, "static")
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return fmt.Errorf("failed to create static directory: %w", err)
	}

	// Get all static files from the embedded FS
	entries, err := static.StaticFiles.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read static directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Read file from embedded FS
		content, err := static.StaticFiles.ReadFile(entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read static file %s: %w", entry.Name(), err)
		}

		// Write to dist/static
		destPath := filepath.Join(staticDir, entry.Name())
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write static file %s: %w", entry.Name(), err)
		}

		slog.Debug("Static asset copied", "file", entry.Name())
	}

	slog.Info("Static assets copied")
	return nil
}

// writeNodeToFile renders a gomponents.Node to an HTML file
func writeNodeToFile(node interface{ Render(io.Writer) error }, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := node.Render(file); err != nil {
		return fmt.Errorf("failed to render node: %w", err)
	}

	return nil
}
