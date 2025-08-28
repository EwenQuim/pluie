package config

import (
	"log/slog"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	// Server settings
	Port string

	// Site customization
	SiteTitle       string
	SiteIcon        string
	SiteDescription string

	// Privacy settings
	PublicByDefault     bool
	HomeNoteSlug        string
	HideYamlFrontmatter bool
}

// Load creates a new Config with values from environment variables
func Load() *Config {
	cfg := &Config{
		Port:                getEnvOrDefault("PORT", "9999"),
		SiteTitle:           getEnvOrDefault("SITE_TITLE", "Pluie"),
		SiteIcon:            getEnvOrDefault("SITE_ICON", "/static/pluie.webp"),
		SiteDescription:     os.Getenv("SITE_DESCRIPTION"),
		PublicByDefault:     getEnvBool("PUBLIC_BY_DEFAULT", false),
		HomeNoteSlug:        getEnvOrDefault("HOME_NOTE_SLUG", "Index"),
		HideYamlFrontmatter: getEnvBool("HIDE_YAML_FRONTMATTER", false),
	}

	slog.Info("Configuration loaded",
		slog.Any("config", cfg),
	)

	return cfg
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
