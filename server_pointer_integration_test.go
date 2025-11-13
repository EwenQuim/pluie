package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/model"
	"github.com/EwenQuim/pluie/template"

	"github.com/go-fuego/fuego"
)

func TestServerPointerIntegration(t *testing.T) {
	// Create initial note
	initialNote := model.Note{
		Title:    "Test Note",
		Slug:     "test-note",
		Content:  "Initial content",
		IsPublic: true,
	}

	// Create notesMap and tree
	notesMap := make(map[string]model.Note)
	notesMap[initialNote.Slug] = initialNote
	notes := []model.Note{initialNote}
	tree := engine.BuildTree(notes)

	// Create server instance
	notesService := engine.NewNotesService(&notesMap, tree, nil)
	server := Server{
		NotesService: notesService,
		rs:           template.Resource{},
		cfg: &config.Config{
			PublicByDefault: true,
			Port:            "8080",
		},
	}

	// Create HTTP server for testing
	fuegServer := fuego.NewServer()
	fuego.Get(fuegServer, "/{slug...}", server.getNote)

	// Test 1: Get initial note content
	req1 := httptest.NewRequest("GET", "/test-note", nil)
	w1 := httptest.NewRecorder()
	fuegServer.Mux.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w1.Code)
	}

	body1 := w1.Body.String()
	if !strings.Contains(body1, "Initial content") {
		t.Errorf("Expected response to contain 'Initial content', got: %s", body1)
	}

	// Test 2: Modify the note in the original notesMap
	modifiedNote := initialNote
	modifiedNote.Content = "Modified content after server start"
	notesMap[initialNote.Slug] = modifiedNote

	// Test 3: Get the note again - should show modified content
	req2 := httptest.NewRequest("GET", "/test-note", nil)
	w2 := httptest.NewRecorder()
	fuegServer.Mux.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w2.Code)
	}

	body2 := w2.Body.String()
	if !strings.Contains(body2, "Modified content after server start") {
		t.Errorf("Expected response to contain 'Modified content after server start', got: %s", body2)
	}

	// Verify the old content is no longer present
	if strings.Contains(body2, "Initial content") {
		t.Errorf("Response should not contain old 'Initial content', got: %s", body2)
	}

	// Test 4: Add a new note to the notesMap
	newNote := model.Note{
		Title:    "New Note",
		Slug:     "new-note",
		Content:  "This note was added after server initialization",
		IsPublic: true,
	}
	notesMap[newNote.Slug] = newNote

	// Test 5: Get the new note - should be accessible
	req3 := httptest.NewRequest("GET", "/new-note", nil)
	w3 := httptest.NewRecorder()
	fuegServer.Mux.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w3.Code)
	}

	body3 := w3.Body.String()
	if !strings.Contains(body3, "This note was added after server initialization") {
		t.Errorf("Expected response to contain new note content, got: %s", body3)
	}

	// Test 6: Verify original modified note is still accessible with modified content
	req4 := httptest.NewRequest("GET", "/test-note", nil)
	w4 := httptest.NewRecorder()
	fuegServer.Mux.ServeHTTP(w4, req4)

	if w4.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w4.Code)
	}

	body4 := w4.Body.String()
	if !strings.Contains(body4, "Modified content after server start") {
		t.Errorf("Expected response to still contain modified content, got: %s", body4)
	}
}

func TestServerPointerIntegrationWithPrivateNote(t *testing.T) {
	// Create initial public note
	publicNote := model.Note{
		Title:    "Public Note",
		Slug:     "public-note",
		Content:  "Public content",
		IsPublic: true,
	}

	// Create notesMap and tree
	notesMap := make(map[string]model.Note)
	notesMap[publicNote.Slug] = publicNote
	notes := []model.Note{publicNote}
	tree := engine.BuildTree(notes)

	// Create server instance
	notesService := engine.NewNotesService(&notesMap, tree, nil)
	server := Server{
		NotesService: notesService,
		rs:           template.Resource{},
		cfg: &config.Config{
			PublicByDefault: false, // Private by default
			Port:            "8080",
		},
	}

	// Create HTTP server for testing
	fuegServer := fuego.NewServer()
	fuego.Get(fuegServer, "/{slug...}", server.getNote)

	// Test 1: Access public note - should work
	req1 := httptest.NewRequest("GET", "/public-note", nil)
	w1 := httptest.NewRecorder()
	fuegServer.Mux.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w1.Code)
	}

	body1 := w1.Body.String()
	if !strings.Contains(body1, "Public content") {
		t.Errorf("Expected response to contain 'Public content', got: %s", body1)
	}

	// Test 2: Make the note private by modifying it in the notesMap
	privateNote := publicNote
	privateNote.IsPublic = false
	privateNote.Content = "Now private content"
	notesMap[publicNote.Slug] = privateNote

	// Test 3: Try to access the now-private note - should be denied
	req2 := httptest.NewRequest("GET", "/public-note", nil)
	w2 := httptest.NewRecorder()
	fuegServer.Mux.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("Expected status 200 (but with access denied content), got %d", w2.Code)
	}

	body2 := w2.Body.String()
	// Should not contain the private content
	if strings.Contains(body2, "Now private content") {
		t.Errorf("Response should not contain private content, got: %s", body2)
	}

	// Test 4: Make it public again
	publicAgainNote := privateNote
	publicAgainNote.IsPublic = true
	publicAgainNote.Content = "Public again content"
	notesMap[publicNote.Slug] = publicAgainNote

	// Test 5: Access should work again
	req3 := httptest.NewRequest("GET", "/public-note", nil)
	w3 := httptest.NewRecorder()
	fuegServer.Mux.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w3.Code)
	}

	body3 := w3.Body.String()
	if !strings.Contains(body3, "Public again content") {
		t.Errorf("Expected response to contain 'Public again content', got: %s", body3)
	}
}
