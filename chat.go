package main

import (
	"fmt"
	"log/slog"

	"github.com/EwenQuim/pluie/config"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

// initializeChatClient creates an Ollama client for chat interactions
func initializeChatClient(cfg *config.Config) (llms.Model, error) {
	slog.Info("Initializing chat client",
		"url", cfg.OllamaURL,
		"model", cfg.ChatModel)

	// Create Ollama client for chat
	chatClient, err := ollama.New(
		ollama.WithServerURL(cfg.OllamaURL),
		ollama.WithModel(cfg.ChatModel),
	)
	if err != nil {
		return nil, fmt.Errorf("creating ollama chat client: %w", err)
	}

	slog.Info("Chat client initialized successfully")

	return chatClient, nil
}
