package config

import (
	"flag"
	"log/slog"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	// Runtime settings (can be overridden by CLI flags)
	Path   string
	Watch  bool
	Mode   string
	Output string

	// Server settings
	Port    string
	LogJSON bool

	// Site customization
	SiteTitle           string
	SiteIcon            string
	SiteDescription     string
	HideYamlFrontmatter bool

	// Privacy settings
	PublicByDefault bool
	HomeNoteSlug    string

	// AI/Chat settings
	ChatProvider  string // "ollama", "mistral", or "openai"
	ChatModel     string
	OllamaURL     string
	MistralAPIKey string
	OpenAIAPIKey  string

	// Embeddings/Weaviate settings
	WeaviateHost   string
	WeaviateScheme string
	WeaviateIndex  string
}

// LoadConfig parses CLI flags and creates Config with CLI flags > Env vars > Defaults priority
func LoadConfig(loadFlags bool) *Config {
	// 1. Set defaults
	cfg := &Config{
		Path:                ".",
		Watch:               true,
		Mode:                "server",
		Output:              "dist",
		ChatProvider:        "ollama",
		ChatModel:           "tinyllama",
		Port:                "9999",
		LogJSON:             false,
		SiteTitle:           "Pluie",
		SiteIcon:            "/static/pluie.webp",
		SiteDescription:     "",
		HideYamlFrontmatter: false,
		PublicByDefault:     false,
		HomeNoteSlug:        "Index",
		OllamaURL:           "http://ollama-models:11434",
		MistralAPIKey:       "",
		OpenAIAPIKey:        "",
		WeaviateHost:        "weaviate-embeddings:9035",
		WeaviateScheme:      "http",
		WeaviateIndex:       "Note",
	}

	// 2. Apply environment variables (override defaults)
	cfg.applyEnvironment()

	// 3. Apply CLI flags (override environment)
	if loadFlags {
		path := flag.String("path", "", "Path to the obsidian folder")
		watch := flag.Bool("watch", false, "Enable file watching to auto-reload on changes")
		mode := flag.String("mode", "", "Mode to run in: server or static")
		output := flag.String("output", "", "Output folder for static site generation")
		chatModel := flag.String("model", "", "Chat model to use for AI responses (overrides CHAT_MODEL env var)")
		flag.Parse()

		if *path != "" {
			cfg.Path = *path
		}
		if flag.Lookup("watch").Value.String() != flag.Lookup("watch").DefValue {
			cfg.Watch = *watch
		}
		if *mode != "" {
			cfg.Mode = *mode
		}
		if *output != "" {
			cfg.Output = *output
		}
		if *chatModel != "" {
			cfg.ChatModel = *chatModel
		}
	}

	// 4. Validate with warnings
	cfg.validate()

	slog.Info("Configuration loaded",
		slog.Any("config", cfg),
	)

	return cfg
}

// applyEnvironment loads configuration from environment variables
func (c *Config) applyEnvironment() {
	// Chat settings
	c.ChatProvider = getEnvOrDefault("CHAT_PROVIDER", c.ChatProvider)
	c.ChatModel = getEnvOrDefault("CHAT_MODEL", c.ChatModel)
	c.OllamaURL = getEnvOrDefault("OLLAMA_URL", c.OllamaURL)
	c.MistralAPIKey = getEnvOrDefault("MISTRAL_API_KEY", c.MistralAPIKey)
	c.OpenAIAPIKey = getEnvOrDefault("OPENAI_API_KEY", c.OpenAIAPIKey)

	// Server settings
	c.Port = getEnvOrDefault("PORT", c.Port)
	c.LogJSON = getEnvBool("LOG_JSON", c.LogJSON)

	// Site customization
	c.SiteTitle = getEnvOrDefault("SITE_TITLE", c.SiteTitle)
	c.SiteIcon = getEnvOrDefault("SITE_ICON", c.SiteIcon)
	c.SiteDescription = getEnvOrDefault("SITE_DESCRIPTION", c.SiteDescription)
	c.HideYamlFrontmatter = getEnvBool("HIDE_YAML_FRONTMATTER", c.HideYamlFrontmatter)
	c.HomeNoteSlug = getEnvOrDefault("HOME_NOTE_SLUG", c.HomeNoteSlug)

	// Privacy settings
	c.PublicByDefault = getEnvBool("PUBLIC_BY_DEFAULT", c.PublicByDefault)

	// Embeddings/Weaviate settings
	c.WeaviateHost = getEnvOrDefault("WEAVIATE_HOST", c.WeaviateHost)
	c.WeaviateScheme = getEnvOrDefault("WEAVIATE_SCHEME", c.WeaviateScheme)
	c.WeaviateIndex = getEnvOrDefault("WEAVIATE_INDEX", c.WeaviateIndex)
}

// validate checks configuration and warns about invalid values
func (c *Config) validate() {
	// Mode validation
	if c.Mode != "server" && c.Mode != "static" {
		slog.Warn("Invalid MODE, defaulting to 'server'", "provided", c.Mode)
		c.Mode = "server"
	}

	// Path validation
	if _, err := os.Stat(c.Path); os.IsNotExist(err) {
		slog.Warn("PATH does not exist, using current directory", "path", c.Path)
		c.Path = "."
	}

	// Chat provider validation
	if c.ChatProvider != "ollama" && c.ChatProvider != "mistral" && c.ChatProvider != "openai" {
		slog.Warn("Invalid CHAT_PROVIDER, defaulting to 'ollama'", "provided", c.ChatProvider)
		c.ChatProvider = "ollama"
	}

	// Weaviate scheme validation
	if c.WeaviateScheme != "http" && c.WeaviateScheme != "https" {
		slog.Warn("Invalid WEAVIATE_SCHEME, defaulting to 'http'", "provided", c.WeaviateScheme)
		c.WeaviateScheme = "http"
	}
}

// getEnvOrDefault returns the environment variable value or a default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool returns the environment variable as a boolean or a default if not set/invalid
func getEnvBool(key string, defaultValue bool) bool {
	if envValue := os.Getenv(key); envValue != "" {
		if parsed, err := strconv.ParseBool(envValue); err == nil {
			return parsed
		}
	}
	return defaultValue
}
