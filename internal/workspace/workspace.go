package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"tutugit/internal/assets"
)

// Workspace -> represents a logical grouping of commits.
type Workspace struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Commits     []string `json:"commits"`
	Status      string   `json:"status"`
}

// Meta -> holds all the tutugit metadata.
type Meta struct {
	Schema          string              `json:"$schema,omitempty"`
	Version         int                 `json:"version"`
	Workspaces      []Workspace         `json:"workspaces"`
	Tags            map[string][]string `json:"tags"`             // Commit SHA -> Tags
	ActiveWorkspace string              `json:"active_workspace"` // ID of the currently active workspace
	Impacts         map[string]string   `json:"impacts"`          // Commit SHA -> Impact Level (patch/minor/major)
}

// Manager -> handles the persistence of Tutugit metadata.
type Manager struct {
	RootPath string
}

// NewManager -> creates a new manager for the given repository root.
func NewManager(rootPath string) *Manager {
	return &Manager{RootPath: rootPath}
}

func (m *Manager) metaPath() string {
	return filepath.Join(m.RootPath, ".tutugit", "meta.json")
}

// Bootstrap -> initializes the tutugit metadata in the repository.
func (m *Manager) Bootstrap() error {
	path := m.metaPath()
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("tutugit already initialized in %s", path)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating .tutugit directory: %w", err)
	}

	// copy schemas to .tutugit/schemas
	if err := m.copySchemas(); err != nil {
		fmt.Printf("Warning: could not copy schemas: %v (falling back to remote references)\n", err)
	}

	meta := &Meta{
		Schema:  "./schemas/meta.schema.json",
		Version: 1,
		Workspaces: []Workspace{
			{
				ID:          "general",
				Name:        "General",
				Description: "Default project workspace",
				Commits:     []string{},
				Status:      "active",
			},
		},
		Tags:            make(map[string][]string),
		ActiveWorkspace: "general",
		Impacts:         make(map[string]string),
	}

	return m.Save(meta)
}

func (m *Manager) copySchemas() error {
	destDir := filepath.Join(m.RootPath, ".tutugit", "schemas")

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	files, err := assets.SchemasFS.ReadDir("schemas")
	if err != nil {
		return fmt.Errorf("failed to read embedded schemas: %w", err)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		
		src := "schemas/" + f.Name()
		dest := filepath.Join(destDir, f.Name())

		data, err := assets.SchemasFS.ReadFile(src)
		if err != nil {
			return fmt.Errorf("failed to read embedded schema %s: %w", f.Name(), err)
		}

		if err := os.WriteFile(dest, data, 0644); err != nil {
			return fmt.Errorf("failed to write schema %s: %w", f.Name(), err)
		}
	}

	return nil
}

// load reads the metadata from disk.
func (m *Manager) Load() (*Meta, error) {
	path := m.metaPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Meta{
			Workspaces: []Workspace{},
			Tags:       make(map[string][]string),
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read meta file: %w", err)
	}

	var meta Meta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse meta file: %w", err)
	}

	if meta.Tags == nil {
		meta.Tags = make(map[string][]string)
	}
	if meta.Impacts == nil {
		meta.Impacts = make(map[string]string)
	}
	if meta.Version == 0 {
		meta.Version = 1
	}

	return &meta, nil
}

// Save -> writes the metadata to disk.
func (m *Manager) Save(meta *Meta) error {
	path := m.metaPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating .tutugit directory: %w", err)
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to parse meta file: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write meta file: %w", err)
	}

	return nil
}

// AddCommitToWorkspace -> adds a commit SHA to a specific workspace.
func (m *Manager) AddCommitToWorkspace(workspaceID, commitSHA string) error {
	meta, err := m.Load()
	if err != nil {
		return err
	}

	found := false
	for i, w := range meta.Workspaces {
		if w.ID == workspaceID {
			for _, sha := range w.Commits {
				if sha == commitSHA {
					return nil
				}
			}
			meta.Workspaces[i].Commits = append(meta.Workspaces[i].Commits, commitSHA)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("workspace %s not found", workspaceID)
	}

	return m.Save(meta)
}

// AddTag -> associates a tag with a commit SHA.
func (m *Manager) AddTag(commitSHA, tag string) error {
	meta, err := m.Load()
	if err != nil {
		return err
	}

	if meta.Tags == nil {
		meta.Tags = make(map[string][]string)
	}

	tags := meta.Tags[commitSHA]
	for _, t := range tags {
		if t == tag {
			return nil
		}
	}

	meta.Tags[commitSHA] = append(tags, tag)
	return m.Save(meta)
}

// AddImpact -> associates a change impact level with a commit SHA.
func (m *Manager) AddImpact(commitSHA, level string) error {
	meta, err := m.Load()
	if err != nil {
		return err
	}

	if meta.Impacts == nil {
		meta.Impacts = make(map[string]string)
	}

	meta.Impacts[commitSHA] = level
	return m.Save(meta)
}

// CreateWorkspace -> logical workspace.
func (m *Manager) CreateWorkspace(id, name, desc string) error {
	meta, err := m.Load()
	if err != nil {
		return err
	}

	for _, w := range meta.Workspaces {
		if w.ID == id {
			return fmt.Errorf("workspace %s already exists", id)
		}
	}

	meta.Workspaces = append(meta.Workspaces, Workspace{
		ID:          id,
		Name:        name,
		Description: desc,
		Commits:     []string{},
		Status:      "active",
	})

	return m.Save(meta)
}

// SetActiveWorkspace -> sets the currently active workspace by ID.
func (m *Manager) SetActiveWorkspace(id string) error {
	meta, err := m.Load()
	if err != nil {
		return err
	}

	// Validate that the workspace exists
	if id != "" {
		found := false
		for _, w := range meta.Workspaces {
			if w.ID == id {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("workspace %s not found", id)
		}
	}

	meta.ActiveWorkspace = id
	return m.Save(meta)
}

// GetActiveWorkspaceName -> returns the name of the active workspace, or empty string.
func (m *Manager) GetActiveWorkspaceName(meta *Meta) string {
	for _, w := range meta.Workspaces {
		if w.ID == meta.ActiveWorkspace {
			return w.Name
		}
	}
	return ""
}
