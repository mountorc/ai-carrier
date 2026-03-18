package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

var (
	uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

type Project struct {
	Project     string    `json:"project"`
	UUIDProject string    `json:"uuid_project"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	IsEnabled   bool      `json:"is_enabled"`
}

type ProjectConfig struct {
	Projects []Project `json:"projects"`
}

type ProjectManager struct {
	configPath string
	projects   map[string]Project
	uuidMap    map[string]Project
	mu         sync.RWMutex
}

func NewProjectManager(configPath string) (*ProjectManager, error) {
	pm := &ProjectManager{
		configPath: configPath,
		projects:   make(map[string]Project),
		uuidMap:    make(map[string]Project),
	}

	if err := pm.LoadConfig(); err != nil {
		return nil, err
	}

	return pm, nil
}

func (pm *ProjectManager) LoadConfig() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	absPath, err := filepath.Abs(pm.configPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config ProjectConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	pm.projects = make(map[string]Project)
	pm.uuidMap = make(map[string]Project)

	for _, p := range config.Projects {
		pm.projects[p.Project] = p
		pm.uuidMap[p.UUIDProject] = p
	}

	return nil
}

func (pm *ProjectManager) ReloadConfig() error {
	return pm.LoadConfig()
}

func (pm *ProjectManager) ValidateUUIDProject(uuidProject, project string) (*Project, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if !uuidRegex.MatchString(uuidProject) {
		return nil, fmt.Errorf("invalid UUID format")
	}

	p, exists := pm.uuidMap[uuidProject]
	if !exists {
		return nil, fmt.Errorf("uuid_project not found")
	}

	if !p.IsEnabled {
		return nil, fmt.Errorf("project is disabled")
	}

	if p.Project != project {
		return nil, fmt.Errorf("uuid_project does not match project")
	}

	return &p, nil
}

func (pm *ProjectManager) ValidateProject(project string) (*Project, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	p, exists := pm.projects[project]
	if !exists {
		return nil, fmt.Errorf("project not found")
	}

	if !p.IsEnabled {
		return nil, fmt.Errorf("project is disabled")
	}

	return &p, nil
}

func (pm *ProjectManager) GetProject(project string) (*Project, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	p, exists := pm.projects[project]
	if !exists {
		return nil, fmt.Errorf("project not found")
	}

	return &p, nil
}

func (pm *ProjectManager) GetAllProjects() []Project {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	projects := make([]Project, 0, len(pm.projects))
	for _, p := range pm.projects {
		projects = append(projects, p)
	}

	return projects
}
