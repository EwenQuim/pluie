package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/engine"
)

func testStaticConfig(vaultDir, outputDir string) *config.Config {
	return &config.Config{
		Path:            vaultDir,
		Mode:            "static",
		Output:          outputDir,
		SiteTitle:       "Test",
		SiteIcon:        "/static/pluie.webp",
		PublicByDefault: true,
		HomeNoteSlug:    "Index",
	}
}

func TestValidateOutputPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{name: "root path", path: "/", wantErr: true},
		{name: "system dir /etc", path: "/etc", wantErr: true},
		{name: "system dir /home", path: "/home", wantErr: true},
		{name: "system dir /var", path: "/var", wantErr: true},
		{name: "valid relative path", path: "dist", wantErr: false},
		{name: "valid nested path", path: "output/site", wantErr: false},
		{name: "valid absolute path", path: "/tmp/pluie-test/dist", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOutputPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateOutputPath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestGenerateStaticSiteEndToEnd(t *testing.T) {
	// Create temporary vault with test notes
	vaultDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "output")

	// Create test notes
	indexContent := `---
public: true
---
# Welcome
This is the index note.
`
	testNoteContent := `---
public: true
title: Test Note
tags:
  - test
---
Some content here.
`

	os.WriteFile(filepath.Join(vaultDir, "Index.md"), []byte(indexContent), 0644)
	os.WriteFile(filepath.Join(vaultDir, "test-note.md"), []byte(testNoteContent), 0644)

	// Load notes from temp vault
	cfg := testStaticConfig(vaultDir, outputDir)
	notesMap, tree, tagIndex, err := loadNotes(vaultDir, cfg)
	if err != nil {
		t.Fatalf("loadNotes error: %v", err)
	}

	notesService := engine.NewNotesService(notesMap, tree, tagIndex)

	// Generate static site
	err = generateStaticSite(notesService, cfg)
	if err != nil {
		t.Fatalf("generateStaticSite error: %v", err)
	}

	// Verify output files exist
	expectedFiles := []string{
		"index.html",
		"static",
	}
	for _, f := range expectedFiles {
		path := filepath.Join(outputDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected output file/dir %q does not exist", f)
		}
	}

	// Verify index.html has content
	indexHTML, err := os.ReadFile(filepath.Join(outputDir, "index.html"))
	if err != nil {
		t.Fatalf("reading index.html: %v", err)
	}
	if len(indexHTML) == 0 {
		t.Error("index.html should not be empty")
	}
}
