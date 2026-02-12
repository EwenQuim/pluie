package main

import (
	"testing"

	"github.com/EwenQuim/pluie/config"
)

func TestInitializeChatClientMissingMistralKey(t *testing.T) {
	cfg := &config.Config{
		ChatProvider:  "mistral",
		ChatModel:     "mistral-small",
		MistralAPIKey: "",
	}

	_, err := initializeChatClient(cfg)
	if err == nil {
		t.Fatal("expected error for missing Mistral API key")
	}
}

func TestInitializeChatClientMissingOpenAIKey(t *testing.T) {
	cfg := &config.Config{
		ChatProvider: "openai",
		ChatModel:    "gpt-4",
		OpenAIAPIKey: "",
	}

	_, err := initializeChatClient(cfg)
	if err == nil {
		t.Fatal("expected error for missing OpenAI API key")
	}
}

func TestInitializeChatClientUnknownProvider(t *testing.T) {
	cfg := &config.Config{
		ChatProvider: "unknown-provider",
		ChatModel:    "model",
	}

	_, err := initializeChatClient(cfg)
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}
