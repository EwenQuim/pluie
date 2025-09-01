package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected Config
	}{
		{
			name:    "Default values when no env vars set",
			envVars: map[string]string{},
			expected: Config{
				Port:                "9999",
				SiteTitle:           "Pluie",
				SiteIcon:            "/static/pluie.webp",
				SiteDescription:     "",
				PublicByDefault:     false,
				HomeNoteSlug:        "Index",
				HideYamlFrontmatter: false,
			},
		},
		{
			name: "Custom values from env vars",
			envVars: map[string]string{
				"PORT":                  "8080",
				"SITE_TITLE":            "My Custom Site",
				"SITE_ICON":             "/custom-icon.png",
				"SITE_DESCRIPTION":      "My custom description",
				"PUBLIC_BY_DEFAULT":     "true",
				"HOME_NOTE_SLUG":        "Home",
				"HIDE_YAML_FRONTMATTER": "true",
			},
			expected: Config{
				Port:                "8080",
				SiteTitle:           "My Custom Site",
				SiteIcon:            "/custom-icon.png",
				SiteDescription:     "My custom description",
				PublicByDefault:     true,
				HomeNoteSlug:        "Home",
				HideYamlFrontmatter: true,
			},
		},
		{
			name: "Mixed env vars with some defaults",
			envVars: map[string]string{
				"PORT":              "3000",
				"SITE_DESCRIPTION":  "Test description",
				"PUBLIC_BY_DEFAULT": "true",
			},
			expected: Config{
				Port:                "3000",
				SiteTitle:           "Pluie",
				SiteIcon:            "/static/pluie.webp",
				SiteDescription:     "Test description",
				PublicByDefault:     true,
				HomeNoteSlug:        "Index",
				HideYamlFrontmatter: false,
			},
		},
		{
			name: "Invalid boolean values should use defaults",
			envVars: map[string]string{
				"PUBLIC_BY_DEFAULT":     "invalid",
				"HIDE_YAML_FRONTMATTER": "not-a-bool",
			},
			expected: Config{
				Port:                "9999",
				SiteTitle:           "Pluie",
				SiteIcon:            "/static/pluie.webp",
				SiteDescription:     "",
				PublicByDefault:     false,
				HomeNoteSlug:        "Index",
				HideYamlFrontmatter: false,
			},
		},
		{
			name: "Empty string env vars should use defaults",
			envVars: map[string]string{
				"PORT":       "",
				"SITE_TITLE": "",
				"SITE_ICON":  "",
			},
			expected: Config{
				Port:                "9999",
				SiteTitle:           "Pluie",
				SiteIcon:            "/static/pluie.webp",
				SiteDescription:     "",
				PublicByDefault:     false,
				HomeNoteSlug:        "Index",
				HideYamlFrontmatter: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all environment variables first
			envVarsToClean := []string{
				"PORT", "SITE_TITLE", "SITE_ICON", "SITE_DESCRIPTION",
				"PUBLIC_BY_DEFAULT", "HOME_NOTE_SLUG", "HIDE_YAML_FRONTMATTER",
			}
			for _, key := range envVarsToClean {
				os.Unsetenv(key)
			}

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Clean up after test
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			// Load config
			cfg := Load()

			// Verify all fields
			if cfg.Port != tt.expected.Port {
				t.Errorf("Port = %q, want %q", cfg.Port, tt.expected.Port)
			}
			if cfg.SiteTitle != tt.expected.SiteTitle {
				t.Errorf("SiteTitle = %q, want %q", cfg.SiteTitle, tt.expected.SiteTitle)
			}
			if cfg.SiteIcon != tt.expected.SiteIcon {
				t.Errorf("SiteIcon = %q, want %q", cfg.SiteIcon, tt.expected.SiteIcon)
			}
			if cfg.SiteDescription != tt.expected.SiteDescription {
				t.Errorf("SiteDescription = %q, want %q", cfg.SiteDescription, tt.expected.SiteDescription)
			}
			if cfg.PublicByDefault != tt.expected.PublicByDefault {
				t.Errorf("PublicByDefault = %v, want %v", cfg.PublicByDefault, tt.expected.PublicByDefault)
			}
			if cfg.HomeNoteSlug != tt.expected.HomeNoteSlug {
				t.Errorf("HomeNoteSlug = %q, want %q", cfg.HomeNoteSlug, tt.expected.HomeNoteSlug)
			}
			if cfg.HideYamlFrontmatter != tt.expected.HideYamlFrontmatter {
				t.Errorf("HideYamlFrontmatter = %v, want %v", cfg.HideYamlFrontmatter, tt.expected.HideYamlFrontmatter)
			}
		})
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "Environment variable exists",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "Environment variable does not exist",
			key:          "NONEXISTENT_KEY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "Environment variable is empty string",
			key:          "EMPTY_KEY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up first
			os.Unsetenv(tt.key)

			// Set environment variable if provided
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvOrDefault(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvOrDefault(%q, %q) = %q, want %q", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		expected     bool
	}{
		{
			name:         "Valid true value",
			key:          "TEST_BOOL",
			defaultValue: false,
			envValue:     "true",
			expected:     true,
		},
		{
			name:         "Valid false value",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "false",
			expected:     false,
		},
		{
			name:         "Valid 1 value",
			key:          "TEST_BOOL",
			defaultValue: false,
			envValue:     "1",
			expected:     true,
		},
		{
			name:         "Valid 0 value",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "0",
			expected:     false,
		},
		{
			name:         "Invalid value uses default",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "invalid",
			expected:     true,
		},
		{
			name:         "Empty value uses default",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "",
			expected:     true,
		},
		{
			name:         "Nonexistent key uses default",
			key:          "NONEXISTENT_BOOL",
			defaultValue: false,
			envValue:     "",
			expected:     false,
		},
		{
			name:         "Case sensitive - True should work",
			key:          "TEST_BOOL",
			defaultValue: false,
			envValue:     "True",
			expected:     true,
		},
		{
			name:         "Case sensitive - FALSE should work",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "FALSE",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up first
			os.Unsetenv(tt.key)

			// Set environment variable if provided
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvBool(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvBool(%q, %v) = %v, want %v", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}
