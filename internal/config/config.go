package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the project-level configuration for tutugit.
type Config struct {
	Schema  string  `yaml:"$schema,omitempty"`
	Project Project `yaml:"project"`
}

// Project holds basic project metadata.
type Project struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// Manager handles loading and saving the config file.
type Manager struct {
	RootPath string
}

// NewManager creates a new config manager for the given repository root.
func NewManager(rootPath string) *Manager {
	return &Manager{RootPath: rootPath}
}

func (m *Manager) configPath() string {
	return filepath.Join(m.RootPath, ".tutugit", "config.yml")
}

// Load reads the config from disk. Returns sensible defaults if the file doesn't exist.
func (m *Manager) Load() (*Config, error) {
	path := m.configPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse config: %w", err)
	}

	return &cfg, nil
}

// Save writes the config to disk.
func (m *Manager) Save(cfg *Config) error {
	path := m.configPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("could not create .tutugit directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("could not serialize config: %w", err)
	}

	// Prepend language server directive if schema is set
	if cfg.Schema != "" {
		directive := fmt.Sprintf("# yaml-language-server: $schema=%s\n", cfg.Schema)
		data = append([]byte(directive), data...)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("could not write config: %w", err)
	}

	return nil
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Schema: "./schemas/config.schema.json",
		Project: Project{
			Name:        "",
			Description: "",
		},
	}
}
