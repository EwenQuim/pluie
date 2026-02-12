package main

import (
	"fmt"
	"log/slog"

	"github.com/EwenQuim/pluie/config"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/mistral"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/vectorstores/weaviate"
)

// createEmbeddingClient creates an LLM client for the configured embedding provider
func createEmbeddingClient(cfg *config.Config) (embeddings.EmbedderClient, error) {
	switch cfg.EmbeddingProvider {
	case "ollama":
		return ollama.New(
			ollama.WithServerURL(cfg.OllamaURL),
			ollama.WithModel(cfg.EmbeddingModel),
		)
	case "openai":
		if cfg.OpenAIAPIKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY is required when using openai embedding provider")
		}
		return openai.New(
			openai.WithToken(cfg.OpenAIAPIKey),
			openai.WithModel(cfg.EmbeddingModel),
		)
	case "mistral":
		if cfg.MistralAPIKey == "" {
			return nil, fmt.Errorf("MISTRAL_API_KEY is required when using mistral embedding provider")
		}
		return mistral.New(
			mistral.WithAPIKey(cfg.MistralAPIKey),
			mistral.WithModel(cfg.EmbeddingModel),
		)
	default:
		return nil, fmt.Errorf("unsupported embedding provider: %s", cfg.EmbeddingProvider)
	}
}

// initializeWeaviateStore creates and initializes the Weaviate store
func initializeWeaviateStore(cfg *config.Config) (*weaviate.Store, error) {
	slog.Info("Initializing Weaviate store",
		"host", cfg.WeaviateHost,
		"scheme", cfg.WeaviateScheme,
		"index", cfg.WeaviateIndex,
		"embedding_provider", cfg.EmbeddingProvider,
		"embedding_model", cfg.EmbeddingModel)

	embeddingsClient, err := createEmbeddingClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating embedding client: %w", err)
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
