package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/template"
	"github.com/charmbracelet/log"
)

// version is set at build time via -ldflags
var version = "dev (local build)"

func main() {
	// Load configuration (parses flags internally)
	cfg := config.LoadConfig(true)

	if cfg.Version {
		fmt.Println("pluie " + version)
		return
	}

	// Setup charmbracelet/log as slog handler
	logger := log.New(os.Stderr)
	logger.SetReportTimestamp(true)
	logger.SetReportCaller(false)

	// Use JSON logging based on config
	if cfg.LogJSON {
		logger.SetFormatter(log.JSONFormatter)
	}

	slog.SetDefault(slog.New(logger))
	slog.Info("Starting pluie", "version", version)

	// Set up signal-based context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
	embeddingsManager := NewEmbeddingsManager(ctx, wvStore, embeddingProgress, notesService, cfg.EmbeddingsTrackingFile, cfg.EmbeddingModel)

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
		_, err = watchFiles(ctx, server, cfg.Path, cfg)
		if err != nil {
			slog.Error("Error starting file watcher", "error", err)
			// Continue anyway - the server can still work without file watching
		}
	}

	err = server.Start(ctx)
	if err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
