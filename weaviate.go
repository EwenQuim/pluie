package main

import (
	"fmt"
	"log/slog"

	"github.com/EwenQuim/pluie/config"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/vectorstores/weaviate"
)

// initializeWeaviateStore creates and initializes the Weaviate store
func initializeWeaviateStore(cfg *config.Config) (*weaviate.Store, error) {
	slog.Info("Initializing Weaviate store",
		"host", cfg.WeaviateHost,
		"scheme", cfg.WeaviateScheme,
		"index", cfg.WeaviateIndex,
		"embeddings_model", embeddingsModel)

	// Create Ollama client for embeddings
	embeddingsClient, err := ollama.New(
		ollama.WithServerURL(cfg.OllamaURL),
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
		weaviate.WithScheme(cfg.WeaviateScheme),
		weaviate.WithHost(cfg.WeaviateHost),
		weaviate.WithIndexName(cfg.WeaviateIndex),
		// Specify which metadata fields to retrieve during similarity search
		weaviate.WithQueryAttrs([]string{"text", "nameSpace", "title", "path", "slug"}),
	)
	if err != nil {
		return nil, fmt.Errorf("creating weaviate store: %w", err)
	}

	slog.Info("Weaviate store initialized successfully - embeddings will be created on first search access")

	return &wvStore, nil
}
