package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/template"
)

// TestFileWatcherHTTPIntegration tests the file watcher with actual HTTP requests
func TestFileWatcherHTTPIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create an initial note file
	initialNote := `---
publish: true
---

# Test Note

This is the initial content.
`
	err := os.WriteFile(filepath.Join(tempDir, "test-note.md"), []byte(initialNote), 0644)
	if err != nil {
		t.Fatalf("Failed to create initial note: %v", err)
	}

	cfg := &config.Config{
		PublicByDefault: false,
		Port:            "0", // Use random available port
	}

	// Load initial notes
	notesMap, tree, tagIndex, err := loadNotes(tempDir, cfg)
	if err != nil {
		t.Fatalf("Failed to load notes: %v", err)
	}

	// Create server
	notesService := engine.NewNotesService(notesMap, tree, tagIndex)
	server := &Server{
		NotesService: notesService,
		rs:           template.Resource{},
		cfg:          cfg,
	}

	// Start file watcher
	watcher, err := watchFiles(server, tempDir, cfg)
	if err != nil {
		t.Fatalf("Failed to start file watcher: %v", err)
	}
	defer watcher.Close()

	// Start HTTP server in a goroutine
	httpServer := &http.Server{
		Addr:    ":18765", // Use a fixed test port
		Handler: createTestHandler(server),
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("HTTP server error: %v", err)
		}
	}()

	// Ensure server is shut down at the end
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		httpServer.Shutdown(ctx)
	}()

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	baseURL := "http://localhost:18765"

	// Test 1: Initial note should be served
	t.Run("ServeInitialNote", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/test-note")
		if err != nil {
			t.Fatalf("Failed to get initial note: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		bodyStr := string(body)
		if !strings.Contains(bodyStr, "Test Note") {
			t.Errorf("Response should contain 'Test Note', got: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "initial content") {
			t.Errorf("Response should contain 'initial content', got: %s", bodyStr)
		}
	})

	// Test 2: Modify the note and verify changes are reflected
	t.Run("ModifyNote", func(t *testing.T) {
		modifiedNote := `---
publish: true
---

# Modified Test Note

This content has been MODIFIED.
`
		err = os.WriteFile(filepath.Join(tempDir, "test-note.md"), []byte(modifiedNote), 0644)
		if err != nil {
			t.Fatalf("Failed to modify note: %v", err)
		}

		// Wait for file watcher to detect change and reload (with debounce)
		time.Sleep(1 * time.Second)

		resp, err := http.Get(baseURL + "/test-note")
		if err != nil {
			t.Fatalf("Failed to get modified note: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		bodyStr := string(body)
		if !strings.Contains(bodyStr, "Modified Test Note") {
			t.Errorf("Response should contain 'Modified Test Note', got: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "MODIFIED") {
			t.Errorf("Response should contain 'MODIFIED', got: %s", bodyStr)
		}
		if strings.Contains(bodyStr, "initial content") {
			t.Errorf("Response should not contain 'initial content' anymore, got: %s", bodyStr)
		}
	})

	// Test 3: Create a new note and verify it's served
	t.Run("CreateNewNote", func(t *testing.T) {
		newNote := `---
publish: true
---

# Brand New Note

This is a completely new note.
`
		err = os.WriteFile(filepath.Join(tempDir, "new-note.md"), []byte(newNote), 0644)
		if err != nil {
			t.Fatalf("Failed to create new note: %v", err)
		}

		// Wait for file watcher to detect change and reload
		time.Sleep(1 * time.Second)

		resp, err := http.Get(baseURL + "/new-note")
		if err != nil {
			t.Fatalf("Failed to get new note: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		bodyStr := string(body)
		if !strings.Contains(bodyStr, "Brand New Note") {
			t.Errorf("Response should contain 'Brand New Note', got: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "completely new note") {
			t.Errorf("Response should contain 'completely new note', got: %s", bodyStr)
		}
	})

	// Test 4: Delete a note and verify it's no longer served
	t.Run("DeleteNote", func(t *testing.T) {
		err = os.Remove(filepath.Join(tempDir, "new-note.md"))
		if err != nil {
			t.Fatalf("Failed to delete note: %v", err)
		}

		// Wait for file watcher to detect change and reload
		time.Sleep(1 * time.Second)

		resp, err := http.Get(baseURL + "/new-note")
		if err != nil {
			t.Fatalf("Failed to request deleted note: %v", err)
		}
		defer resp.Body.Close()

		// Should get 200 but with empty/not found content
		// (depending on how your server handles missing notes)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		bodyStr := string(body)
		if strings.Contains(bodyStr, "Brand New Note") {
			t.Errorf("Deleted note should not be served anymore, got: %s", bodyStr)
		}
	})

	// Test 5: Create note in a new subdirectory
	t.Run("CreateNoteInSubdirectory", func(t *testing.T) {
		subDir := filepath.Join(tempDir, "subfolder")
		err = os.Mkdir(subDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create subdirectory: %v", err)
		}

		// Wait for watcher to detect new directory
		time.Sleep(200 * time.Millisecond)

		subNote := `---
publish: true
---

# Subfolder Note

This note is in a subfolder.
`
		err = os.WriteFile(filepath.Join(subDir, "sub-note.md"), []byte(subNote), 0644)
		if err != nil {
			t.Fatalf("Failed to create note in subdirectory: %v", err)
		}

		// Wait for file watcher to detect change and reload
		time.Sleep(1 * time.Second)

		// Try different possible URL paths
		possiblePaths := []string{
			"/subfolder/sub-note",
			"/sub-note",
		}

		found := false
		for _, path := range possiblePaths {
			resp, err := http.Get(baseURL + path)
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					continue
				}

				bodyStr := string(body)
				if strings.Contains(bodyStr, "Subfolder Note") {
					found = true
					t.Logf("Found subfolder note at path: %s", path)
					break
				}
			}
		}

		if !found {
			t.Errorf("Subfolder note should be accessible")
		}
	})
}

// createTestHandler creates a minimal HTTP handler for testing
func createTestHandler(server *Server) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slug := strings.TrimPrefix(r.URL.Path, "/")
		if slug == "" {
			slug = server.NotesService.GetHomeSlug(server.cfg.HomeNoteSlug)
		}

		note, ok := server.NotesService.GetNote(slug)
		if !ok {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Note not found: %s", slug)
			return
		}

		// Additional security check: ensure note is public
		if !server.cfg.PublicByDefault && !note.IsPublic {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Note not found: %s", slug)
			return
		}

		// Return a simple representation of the note
		fmt.Fprintf(w, "Title: %s\n\n%s", note.Title, note.Content)
	})

	return mux
}
