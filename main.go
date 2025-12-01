package main

import (
	"log/slog"
	"os"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/template"
	"github.com/charmbracelet/log"
)

func main() {
	// Load configuration (parses flags internally)
	cfg := config.LoadConfig(true)

	// Setup charmbracelet/log as slog handler
	logger := log.New(os.Stderr)
	logger.SetReportTimestamp(true)
	logger.SetReportCaller(false)

	// Use JSON logging based on config
	if cfg.LogJSON {
		logger.SetFormatter(log.JSONFormatter)
	}

	slog.SetDefault(slog.New(logger))

	// Load initial notes
	notesMap, tree, tagIndex, err := loadNotes(cfg.Path, cfg)
	if err != nil {
		slog.Error("Error loading notes", "error", err)
		return
	}

	notesService := engine.NewNotesService(notesMap, tree, tagIndex)

	// Initialize embedding progress tracker
	embeddingProgress := NewEmbeddingProgress()

	// Initialize Weaviate store for search (embeddings will be lazy-loaded on first search)
	wvStore, err := initializeWeaviateStore(cfg)
	if err != nil {
		slog.Warn("Failed to initialize Weaviate store, search and embeddings will not be available", "error", err)
		wvStore = nil
	}

	// Create embeddings manager
	embeddingsManager := NewEmbeddingsManager(wvStore, embeddingProgress, notesService)

	// Initialize chat client for AI responses
	chatClient, err := initializeChatClient(cfg)
	if err != nil {
		slog.Warn("Failed to initialize chat client, AI responses will not be available", "error", err)
		chatClient = nil
	}

	// Run in static mode if requested
	if cfg.Mode == "static" {
		err := generateStaticSite(notesService, cfg)
		if err != nil {
			slog.Error("Error generating static site", "error", err)
			return
		}
		slog.Info("Static site generated successfully", "folder", cfg.Output)
		return
	}

	// Otherwise run in server mode
	server := &Server{
		NotesService:      notesService,
		rs:                template.NewResource(cfg),
		cfg:               cfg,
		chatClient:        chatClient,
		embeddingsManager: embeddingsManager,
	}

	// Start file watcher if enabled
	if cfg.Watch {
		_, err = watchFiles(server, cfg.Path, cfg)
		if err != nil {
			slog.Error("Error starting file watcher", "error", err)
			// Continue anyway - the server can still work without file watching
		}
	}

	err = server.Start()
	if err != nil {
		panic(err)
	}

}
