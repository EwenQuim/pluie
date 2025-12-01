package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/vectorstores/weaviate"
)

// initializeWeaviateStore creates and initializes the Weaviate store
func initializeWeaviateStore() (*weaviate.Store, error) {
	// Get Weaviate configuration from environment or use defaults
	wvHost := os.Getenv("WEAVIATE_HOST")
	if wvHost == "" {
		wvHost = "weaviate-embeddings:9035" // Default from docker-compose
	}

	wvScheme := os.Getenv("WEAVIATE_SCHEME")
	if wvScheme == "" {
		wvScheme = "http"
	}

	indexName := os.Getenv("WEAVIATE_INDEX")
	if indexName == "" {
		indexName = "Note"
	}

	slog.Info("Initializing Weaviate store",
		"host", wvHost,
		"scheme", wvScheme,
		"index", indexName,
		"embeddings_model", embeddingsModel)

	// Create Ollama client for embeddings
	embeddingsClient, err := ollama.New(
		ollama.WithServerURL("http://ollama-models:11434"),
		ollama.WithModel(embeddingsModel),
	)
	if err != nil {
		return nil, fmt.Errorf("creating ollama client: %w", err)
	}

	emb, err := embeddings.NewEmbedder(embeddingsClient)
	if err != nil {
		return nil, fmt.Errorf("creating embedder: %w", err)
	}

	// Create Weaviate store
	wvStore, err := weaviate.New(
		weaviate.WithEmbedder(emb),
		weaviate.WithScheme(wvScheme),
		weaviate.WithHost(wvHost),
		weaviate.WithIndexName(indexName),
		// Specify which metadata fields to retrieve during similarity search
		weaviate.WithQueryAttrs([]string{"text", "nameSpace", "title", "path", "slug"}),
	)
	if err != nil {
		return nil, fmt.Errorf("creating weaviate store: %w", err)
	}

	slog.Info("Weaviate store initialized successfully - embeddings will be created on first search access")

	return &wvStore, nil
}
