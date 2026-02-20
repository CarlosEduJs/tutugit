package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManager_LoadSave(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tutugit-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := NewManager(tmpDir)

	// Load non-existent returns defaults
	cfg, err := m.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Project.Name != "" {
		t.Errorf("Expected empty default project name, got %s", cfg.Project.Name)
	}

	// Save and reload
	cfg.Project.Name = "test-project"
	cfg.Project.Description = "A test project"
	if err := m.Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filepath.Join(tmpDir, ".tutugit", "config.yml")); os.IsNotExist(err) {
		t.Error("config.yml was not created")
	}

	// Reload
	loaded, err := m.Load()
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}
	if loaded.Project.Name != "test-project" {
		t.Errorf("Project name not persisted, got %s", loaded.Project.Name)
	}
	if loaded.Project.Description != "A test project" {
		t.Errorf("Project description not persisted, got %s", loaded.Project.Description)
	}
}

func TestManager_SaveWithSchema(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tutugit-config-schema-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := NewManager(tmpDir)
	cfg := DefaultConfig()
	cfg.Project.Name = "schema-test"

	if err := m.Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Read raw file to check for language server directive
	data, err := os.ReadFile(filepath.Join(tmpDir, ".tutugit", "config.yml"))
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	content := string(data)
	if !contains(content, "# yaml-language-server") {
		t.Error("Expected yaml-language-server directive in saved file")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
