package main

import (
	"log/slog"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	"github.com/adrg/frontmatter"
)

type Explorer struct {
	BasePath string
}

func (e Explorer) getFolderNotes(currentPath string) ([]model.Note, error) {
	start := time.Now()
	if e.shouldSkipPath(currentPath) {
		return nil, nil
	}

	dir, err := os.ReadDir(e.BasePath + "/" + currentPath)
	if err != nil {
		return nil, err
	}

	folderMetadata := e.collectFolderMetadata(dir, currentPath)
	notes := e.processDirectoryEntries(dir, currentPath, folderMetadata)

	slog.Debug("explored", "notes", len(notes), "folder", currentPath, "in", time.Since(start))
	return notes, nil
}

// shouldSkipPath determines if a path should be skipped during exploration
func (e Explorer) shouldSkipPath(currentPath string) bool {
	return strings.HasPrefix(currentPath, "/.") ||
		strings.Contains(currentPath, "node_modules") ||
		strings.Contains(currentPath, ".git")
}

// collectFolderMetadata collects metadata from .pluie files in the directory
func (e Explorer) collectFolderMetadata(dir []os.DirEntry, currentPath string) map[string]map[string]any {
	folderMetadata := make(map[string]map[string]any)

	for _, entry := range dir {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".pluie") {
			if metadata := e.parsePluieFile(currentPath, entry.Name()); metadata != nil {
				folderPath := strings.Trim(currentPath, "/")
				folderMetadata[folderPath] = metadata
			}
		}
	}

	return folderMetadata
}

// parsePluieFile parses a .pluie metadata file
func (e Explorer) parsePluieFile(currentPath, fileName string) map[string]any {
	metadataBytes, err := os.ReadFile(path.Join(e.BasePath, currentPath, fileName))
	if err != nil {
		return nil
	}

	var metadata map[string]any
	_, err = frontmatter.Parse(strings.NewReader(string(metadataBytes)), &metadata)
	if err != nil {
		return nil
	}

	return metadata
}

// processDirectoryEntries processes all entries in a directory using concurrency
func (e Explorer) processDirectoryEntries(dir []os.DirEntry, currentPath string, folderMetadata map[string]map[string]any) []model.Note {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var notes []model.Note

	// Process directories and markdown files concurrently
	for _, entry := range dir {
		wg.Go(func() {
			if entry.IsDir() {
				subfolderNotes, err := e.getFolderNotes(currentPath + "/" + entry.Name())
				if err == nil && len(subfolderNotes) > 0 {
					mu.Lock()
					notes = append(notes, subfolderNotes...)
					mu.Unlock()
				}
			} else if strings.HasSuffix(entry.Name(), ".md") {
				if note := e.processMarkdownFile(currentPath, entry.Name(), folderMetadata); note != nil {
					mu.Lock()
					notes = append(notes, *note)
					mu.Unlock()
				}
			}

		})
	}

	wg.Wait()
	return notes
}

// processMarkdownFile processes a single markdown file
func (e Explorer) processMarkdownFile(currentPath, fileName string, folderMetadata map[string]map[string]any) *model.Note {
	contentBytes, err := os.ReadFile(path.Join(e.BasePath, currentPath, fileName))
	if err != nil {
		return nil
	}

	// Parse frontmatter
	var metadata map[string]any
	parsedContent, err := frontmatter.Parse(strings.NewReader(string(contentBytes)), &metadata)
	var finalContent string
	if err != nil {
		finalContent = string(contentBytes)
		metadata = make(map[string]any)
	} else {
		finalContent = string(parsedContent)
	}

	// Remove comment blocks between %% markers before displaying
	finalContent = engine.RemoveCommentBlocks(finalContent)

	// Extract title from frontmatter or filename
	title := e.extractTitle(fileName, metadata)

	note := model.Note{
		Title:    title,
		Content:  finalContent,
		Slug:     path.Join(currentPath, fileName),
		Path:     path.Join(currentPath, fileName),
		Metadata: metadata,
	}
	note.BuildSlug()
	note.DetermineIsPublic(folderMetadata)

	return &note
}

// extractTitle extracts the title from frontmatter or falls back to filename
func (e Explorer) extractTitle(fileName string, metadata map[string]any) string {
	title := strings.TrimSuffix(fileName, ".md")
	if frontmatterTitle, exists := metadata["title"]; exists {
		if titleStr, ok := frontmatterTitle.(string); ok && titleStr != "" {
			title = titleStr
		}
	}
	return title
}

// filterPublicNotes filters notes based on public/private visibility
func filterPublicNotes(notes []model.Note, publicByDefault bool) []model.Note {
	if publicByDefault {
		return notes
	}

	publicNotes := make([]model.Note, 0, len(notes))
	for _, note := range notes {
		if note.IsPublic {
			publicNotes = append(publicNotes, note)
		}
	}
	slog.Info("filtered notes", "publicNotes", len(publicNotes), "totalNotes", len(notes))
	return publicNotes
}
