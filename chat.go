package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/EwenQuim/pluie/model"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

const (
	chatModel = "mistral" // Mistral model for chat/summarization
)

// initializeChatClient creates an Ollama client for chat interactions
func initializeChatClient() (llms.Model, error) {
	// Get Ollama configuration from environment or use defaults
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://ollama-models:11434" // Default from docker-compose
	}

	slog.Info("Initializing chat client",
		"url", ollamaURL,
		"model", chatModel)

	// Create Ollama client for chat
	chatClient, err := ollama.New(
		ollama.WithServerURL(ollamaURL),
		ollama.WithModel(chatModel),
	)
	if err != nil {
		return nil, fmt.Errorf("creating ollama chat client: %w", err)
	}

	return chatClient, nil
}

// generateChatResponse generates an AI summary based on the user query and search results
func generateChatResponse(ctx context.Context, chatClient llms.Model, query string, results []model.Note) (string, error) {
	if len(results) == 0 {
		return "", nil
	}

	// Build context from search results
	var contextBuilder strings.Builder
	contextBuilder.WriteString("Here are the relevant notes from my knowledge base:\n\n")

	for i, note := range results {
		contextBuilder.WriteString(fmt.Sprintf("Document %d - %s:\n", i+1, note.Title))
		// Limit content to first 500 characters to avoid token limits
		content := note.Content
		if len(content) > 500 {
			content = content[:500] + "..."
		}
		contextBuilder.WriteString(content)
		contextBuilder.WriteString("\n\n---\n\n")
	}

	// Build the prompt
	prompt := fmt.Sprintf(`You are a helpful assistant that answers questions based on a user's personal notes.

User's Question: %s

%s

Based on the documents provided above, please answer the user's question. Be concise, accurate, and cite which documents you're referencing when relevant. If the documents don't contain enough information to fully answer the question, acknowledge that and answer what you can.`, query, contextBuilder.String())

	slog.Info("Generating chat response", "query", query, "num_docs", len(results))

	// Generate response
	response, err := llms.GenerateFromSinglePrompt(ctx, chatClient, prompt)
	if err != nil {
		return "", fmt.Errorf("generating chat response: %w", err)
	}

	slog.Info("Chat response generated", "response_length", len(response))

	return response, nil
}
