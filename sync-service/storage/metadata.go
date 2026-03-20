package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ProjectConfig struct {
	RepoURL           string    `json:"repo_url"`
	Branch            string    `json:"branch"`
	GitHubToken       string    `json:"github_token"`
	LastCommit        string    `json:"last_commit,omitempty"`
	LastSync          time.Time `json:"last_sync,omitempty"`
	LastSyncDirection string    `json:"last_sync_direction,omitempty"`
}

type Store struct {
	mu       sync.RWMutex
	path     string
	projects map[string]ProjectConfig
}

func NewStore(path string) (*Store, error) {
	s := &Store{
		path:     path,
		projects: make(map[string]ProjectConfig),
	}

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil {
		if err := json.Unmarshal(data, &s.projects); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *Store) Load(projectID string) (ProjectConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cfg, ok := s.projects[projectID]
	return cfg, ok
}

func (s *Store) Save(projectID string, cfg ProjectConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.projects[projectID] = cfg
	return s.writeLocked()
}

func (s *Store) Delete(projectID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.projects, projectID)
	return s.writeLocked()
}

func (s *Store) writeLocked() error {
	data, err := json.MarshalIndent(s.projects, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
