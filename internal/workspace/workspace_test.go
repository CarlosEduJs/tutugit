package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManager_Bootstrap(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tutugit-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := NewManager(tmpDir)

	// First bootstrap
	if err := m.Bootstrap(); err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	// Verify meta.json exists
	metaPath := filepath.Join(tmpDir, ".tutugit", "meta.json")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Error("meta.json was not created")
	}

	// Verify content
	meta, err := m.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(meta.Workspaces) != 1 || meta.Workspaces[0].ID != "general" {
		t.Errorf("Unexpected workspace structure: %+v", meta.Workspaces)
	}

	// Second bootstrap should fail
	if err := m.Bootstrap(); err == nil {
		t.Error("Second bootstrap should have failed because project is already initialized")
	}
}

func TestManager_MetadataOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tutugit-test-metadata-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := NewManager(tmpDir)
	m.Bootstrap()

	// Test AddImpact
	hash := "abc123"
	if err := m.AddImpact(hash, "major"); err != nil {
		t.Fatalf("AddImpact failed: %v", err)
	}

	meta, _ := m.Load()
	if meta.Impacts[hash] != "major" {
		t.Errorf("Impact not saved correctly, got %s", meta.Impacts[hash])
	}

	// Test CreateWorkspace
	if err := m.CreateWorkspace("test-ws", "Test WS", "A test workspace"); err != nil {
		t.Fatalf("CreateWorkspace failed: %v", err)
	}

	meta, _ = m.Load()
	found := false
	for _, w := range meta.Workspaces {
		if w.ID == "test-ws" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Workspace test-ws not found after creation")
	}

	// Test SetActiveWorkspace
	if err := m.SetActiveWorkspace("test-ws"); err != nil {
		t.Fatalf("SetActiveWorkspace failed: %v", err)
	}
	meta, _ = m.Load()
	if meta.ActiveWorkspace != "test-ws" {
		t.Errorf("Active workspace not updated, got %s", meta.ActiveWorkspace)
	}
}

func TestManager_AddCommitToWorkspace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tutugit-test-commit-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := NewManager(tmpDir)
	m.Bootstrap()

	// Add commit to general workspace
	commitSHA := "abc123def456"
	if err := m.AddCommitToWorkspace("general", commitSHA); err != nil {
		t.Fatalf("AddCommitToWorkspace failed: %v", err)
	}

	meta, _ := m.Load()
	found := false
	for _, w := range meta.Workspaces {
		if w.ID == "general" {
			for _, sha := range w.Commits {
				if sha == commitSHA {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Error("Commit not found in workspace")
	}

	// Test duplicate commit (should not error)
	if err := m.AddCommitToWorkspace("general", commitSHA); err != nil {
		t.Errorf("AddCommitToWorkspace with duplicate should not error: %v", err)
	}

	// Test non-existent workspace
	if err := m.AddCommitToWorkspace("non-existent", commitSHA); err == nil {
		t.Error("AddCommitToWorkspace to non-existent workspace should fail")
	}
}

func TestManager_AddTag(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tutugit-test-tag-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := NewManager(tmpDir)
	m.Bootstrap()

	commitSHA := "abc123"
	tag := "feature"

	// Add tag
	if err := m.AddTag(commitSHA, tag); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	meta, _ := m.Load()
	tags, exists := meta.Tags[commitSHA]
	if !exists || len(tags) == 0 {
		t.Error("Tag not found")
	}
	if tags[0] != tag {
		t.Errorf("Expected tag %s, got %s", tag, tags[0])
	}

	// Add duplicate tag (should not duplicate)
	if err := m.AddTag(commitSHA, tag); err != nil {
		t.Errorf("AddTag with duplicate should not error: %v", err)
	}

	meta, _ = m.Load()
	if len(meta.Tags[commitSHA]) != 1 {
		t.Errorf("Duplicate tag was added, got %d tags", len(meta.Tags[commitSHA]))
	}

	// Add second different tag
	if err := m.AddTag(commitSHA, "fix"); err != nil {
		t.Fatalf("AddTag second tag failed: %v", err)
	}

	meta, _ = m.Load()
	if len(meta.Tags[commitSHA]) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(meta.Tags[commitSHA]))
	}
}

func TestManager_CreateWorkspace_Duplicate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tutugit-test-dup-ws-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := NewManager(tmpDir)
	m.Bootstrap()

	// Create workspace
	if err := m.CreateWorkspace("ws1", "Workspace 1", "First workspace"); err != nil {
		t.Fatalf("CreateWorkspace failed: %v", err)
	}

	// Try to create duplicate
	if err := m.CreateWorkspace("ws1", "Workspace 1 Dup", "Duplicate"); err == nil {
		t.Error("Creating duplicate workspace should fail")
	}
}

func TestManager_SetActiveWorkspace_Invalid(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tutugit-test-invalid-ws-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := NewManager(tmpDir)
	m.Bootstrap()

	// Try to set non-existent workspace as active
	if err := m.SetActiveWorkspace("non-existent"); err == nil {
		t.Error("Setting non-existent workspace as active should fail")
	}

	// Empty string should be allowed (deactivate)
	if err := m.SetActiveWorkspace(""); err != nil {
		t.Errorf("Setting empty workspace should not fail: %v", err)
	}
}

func TestManager_GetActiveWorkspaceName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tutugit-test-active-name-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := NewManager(tmpDir)
	m.Bootstrap()

	meta, _ := m.Load()

	// Get default active workspace name
	name := m.GetActiveWorkspaceName(meta)
	if name != "General" {
		t.Errorf("Expected 'General', got '%s'", name)
	}

	// Create and activate new workspace
	m.CreateWorkspace("test", "Test Workspace", "Test")
	m.SetActiveWorkspace("test")

	meta, _ = m.Load()
	name = m.GetActiveWorkspaceName(meta)
	if name != "Test Workspace" {
		t.Errorf("Expected 'Test Workspace', got '%s'", name)
	}

	// Test with invalid active workspace
	meta.ActiveWorkspace = "non-existent"
	name = m.GetActiveWorkspaceName(meta)
	if name != "" {
		t.Errorf("Expected empty string for non-existent workspace, got '%s'", name)
	}
}

func TestManager_Load_NonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tutugit-test-load-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := NewManager(tmpDir)

	// Load without bootstrap should return empty meta
	meta, err := m.Load()
	if err != nil {
		t.Fatalf("Load should not error on non-existent file: %v", err)
	}

	if len(meta.Workspaces) != 0 {
		t.Errorf("Expected empty workspaces, got %d", len(meta.Workspaces))
	}

	if meta.Tags == nil {
		t.Error("Tags map should be initialized")
	}
}

func TestManager_Save_Load_Roundtrip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tutugit-test-roundtrip-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := NewManager(tmpDir)

	// Create meta with custom data
	meta := &Meta{
		Version: 2,
		Workspaces: []Workspace{
			{
				ID:          "custom",
				Name:        "Custom WS",
				Description: "Custom workspace",
				Commits:     []string{"commit1", "commit2"},
				Status:      "active",
			},
		},
		Tags: map[string][]string{
			"commit1": {"feature", "important"},
		},
		ActiveWorkspace: "custom",
		Impacts: map[string]string{
			"commit1": "major",
		},
	}

	// Save
	if err := m.Save(meta); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load
	loaded, err := m.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify
	if loaded.Version != 2 {
		t.Errorf("Expected version 2, got %d", loaded.Version)
	}

	if len(loaded.Workspaces) != 1 {
		t.Fatalf("Expected 1 workspace, got %d", len(loaded.Workspaces))
	}

	ws := loaded.Workspaces[0]
	if ws.ID != "custom" || ws.Name != "Custom WS" {
		t.Errorf("Workspace data mismatch: %+v", ws)
	}

	if len(ws.Commits) != 2 {
		t.Errorf("Expected 2 commits, got %d", len(ws.Commits))
	}

	if len(loaded.Tags["commit1"]) != 2 {
		t.Errorf("Expected 2 tags for commit1, got %d", len(loaded.Tags["commit1"]))
	}

	if loaded.Impacts["commit1"] != "major" {
		t.Errorf("Expected impact 'major', got '%s'", loaded.Impacts["commit1"])
	}

	if loaded.ActiveWorkspace != "custom" {
		t.Errorf("Expected active workspace 'custom', got '%s'", loaded.ActiveWorkspace)
	}
}
