package main

import (
	"cmp"
	"flag"
	"log/slog"
	"os"

	"github.com/EwenQuim/pluie/config"
	"github.com/EwenQuim/pluie/engine"
	"github.com/EwenQuim/pluie/template"
	"github.com/charmbracelet/log"
)

func main() {
	path := flag.String("path", ".", "Path to the obsidian folder")
	watch := flag.Bool("watch", true, "Enable file watching to auto-reload on changes")
	mode := flag.String("mode", "server", "Mode to run in: server or static")
	output := flag.String("output", "dist", "Output folder for static site generation")

	// Get default chat model from environment variable, fallback to tinyllama
	defaultChatModel := cmp.Or(os.Getenv("CHAT_MODEL"), "tinyllama")
	chatModel := flag.String("model", defaultChatModel, "Chat model to use for AI responses")

	flag.Parse()

	// Setup charmbracelet/log as slog handler
	logger := log.New(os.Stderr)
	logger.SetReportTimestamp(true)
	logger.SetReportCaller(false)

	// Use JSON logging if LOG_JSON=true
	if os.Getenv("LOG_JSON") == "true" {
		logger.SetFormatter(log.JSONFormatter)
	}

	slog.SetDefault(slog.New(logger))

	// Load configuration
	cfg := config.Load()

	// Load initial notes
	notesMap, tree, tagIndex, err := loadNotes(*path, cfg)
	if err != nil {
		slog.Error("Error loading notes", "error", err)
		return
	}

	notesService := engine.NewNotesService(notesMap, tree, tagIndex)

	// Initialize embedding progress tracker
	embeddingProgress := NewEmbeddingProgress()

	// Initialize Weaviate store for search (embeddings will be lazy-loaded on first search)
	wvStore, err := initializeWeaviateStore()
	if err != nil {
		slog.Warn("Failed to initialize Weaviate store, search and embeddings will not be available", "error", err)
		wvStore = nil
	}

	// Create embeddings manager
	embeddingsManager := NewEmbeddingsManager(wvStore, embeddingProgress, notesService)

	// Initialize chat client for AI responses
	chatClient, err := initializeChatClient(*chatModel)
	if err != nil {
		slog.Warn("Failed to initialize chat client, AI responses will not be available", "error", err)
		chatClient = nil
	}

	// Run in static mode if requested
	if *mode == "static" {
		err := generateStaticSite(notesService, cfg, *output)
		if err != nil {
			slog.Error("Error generating static site", "error", err)
			return
		}
		slog.Info("Static site generated successfully", "folder", *output)
		return
	}

	// Otherwise run in server mode
	server := &Server{
		NotesService:      notesService,
		rs:                template.Resource{},
		cfg:               cfg,
		chatClient:        chatClient,
		embeddingsManager: embeddingsManager,
	}

	// Start file watcher if enabled
	if *watch {
		_, err = watchFiles(server, *path, cfg)
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
