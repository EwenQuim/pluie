package static

import (
	"embed"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	handler := Handler()

	if handler == nil {
		t.Fatal("Handler() returned nil")
	}
}

func TestStaticFilesServing(t *testing.T) {
	handler := Handler()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "Serve app.js",
			path:           "/app.js",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Serve favicon.ico",
			path:           "/favicon.ico",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Serve pluie.webp",
			path:           "/pluie.webp",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Non-existent file returns 404",
			path:           "/nonexistent.txt",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Root path",
			path:           "/",
			expectedStatus: http.StatusOK, // Should serve directory listing or index
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d for path %s", tt.expectedStatus, w.Code, tt.path)
			}
		})
	}
}

func TestStaticFilesEmbedded(t *testing.T) {
	// Test that StaticFiles embed.FS is not nil and contains expected files
	if StaticFiles == (embed.FS{}) {
		t.Error("StaticFiles embed.FS appears to be empty")
	}

	// Test reading some expected files
	expectedFiles := []string{
		"app.js",
		"favicon.ico",
		"pluie.webp",
	}

	for _, filename := range expectedFiles {
		t.Run("File exists: "+filename, func(t *testing.T) {
			data, err := StaticFiles.ReadFile(filename)
			if err != nil {
				t.Errorf("Failed to read embedded file %s: %v", filename, err)
			}
			if len(data) == 0 {
				t.Errorf("Embedded file %s is empty", filename)
			}
		})
	}
}
