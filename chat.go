package main

import (
	"fmt"
	"log/slog"

	"github.com/EwenQuim/pluie/config"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/mistral"
	"github.com/tmc/langchaingo/llms/ollama"
)

// initializeChatClient creates a chat client based on the configured provider
func initializeChatClient(cfg *config.Config) (llms.Model, error) {
	slog.Info("Initializing chat client",
		"provider", cfg.ChatProvider,
		"model", cfg.ChatModel)

	switch cfg.ChatProvider {
	case "ollama":
		// Create Ollama client for local models
		return ollama.New(
			ollama.WithServerURL(cfg.OllamaURL),
			ollama.WithModel(cfg.ChatModel),
		)

	case "mistral":
		// Create Mistral API client
		if cfg.MistralAPIKey == "" {
			return nil, fmt.Errorf("MISTRAL_API_KEY is required when using mistral provider")
		}
		return mistral.New(
			mistral.WithAPIKey(cfg.MistralAPIKey),
			mistral.WithModel(cfg.ChatModel),
		)
	}

	return nil, fmt.Errorf("unsupported chat provider: %s", cfg.ChatProvider)
}
