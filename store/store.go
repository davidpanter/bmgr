package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const defaultFile = "bindings.json"

type Binding struct {
	ID          string   `json:"id"`
	App         string   `json:"app"`
	Description string   `json:"description"`
	Keys        []string `json:"keys"`
	Alternates  []string `json:"alternates"`
	Tags        []string `json:"tags"`
	Notes       string   `json:"notes"`
	sourceFile  string
}

type fileData struct {
	Version  int       `json:"version"`
	Bindings []Binding `json:"bindings"`
}

type Store struct {
	Bindings []Binding
	dir      string
}

func ConfigDir() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not determine home directory: %w", err)
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "bmgr"), nil
}

func Load() (*Store, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, err
	}

	s := &Store{dir: dir, Bindings: make([]Binding, 0)}

	matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("could not list config files: %w", err)
	}

	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("could not read %s: %w", path, err)
		}

		var fd fileData
		if err := json.Unmarshal(data, &fd); err != nil {
			return nil, fmt.Errorf("could not parse %s: %w", path, err)
		}

		name := filepath.Base(path)
		for i := range fd.Bindings {
			fd.Bindings[i].sourceFile = name
		}
		s.Bindings = append(s.Bindings, fd.Bindings...)
	}

	return s, nil
}

func (s *Store) Save() error {
	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return fmt.Errorf("could not create config dir: %w", err)
	}

	// Group bindings by source file
	files := make(map[string][]Binding)
	for _, b := range s.Bindings {
		name := b.sourceFile
		if name == "" {
			name = defaultFile
		}
		files[name] = append(files[name], b)
	}

	// Track which files we wrote so we can remove empty ones
	written := make(map[string]bool)

	for name, bindings := range files {
		fd := fileData{Version: 1, Bindings: bindings}
		data, err := json.MarshalIndent(fd, "", "  ")
		if err != nil {
			return fmt.Errorf("could not marshal %s: %w", name, err)
		}

		path := filepath.Join(s.dir, name)
		tmp := path + ".tmp"
		if err := os.WriteFile(tmp, data, 0600); err != nil {
			return fmt.Errorf("could not write %s: %w", name, err)
		}
		if err := os.Rename(tmp, path); err != nil {
			os.Remove(tmp)
			return fmt.Errorf("could not save %s: %w", name, err)
		}
		written[name] = true
	}

	// Remove files that had bindings but now have none
	matches, _ := filepath.Glob(filepath.Join(s.dir, "*.json"))
	for _, path := range matches {
		name := filepath.Base(path)
		if !written[name] {
			os.Remove(path)
		}
	}

	return nil
}

func GenerateID() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("could not generate ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (b *Binding) SourceFile() string {
	return b.sourceFile
}

func (b *Binding) SetSourceFile(name string) {
	b.sourceFile = name
}

func (s *Store) Add(b Binding) error {
	s.Bindings = append(s.Bindings, b)
	return nil
}

func (s *Store) Remove(id string) error {
	for i, b := range s.Bindings {
		if b.ID == id {
			s.Bindings = append(s.Bindings[:i], s.Bindings[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("binding %q not found", id)
}

func (s *Store) FindByID(id string) *Binding {
	for i := range s.Bindings {
		if s.Bindings[i].ID == id {
			return &s.Bindings[i]
		}
	}
	return nil
}

func (s *Store) Replace(b Binding) error {
	for i, existing := range s.Bindings {
		if existing.ID == b.ID {
			s.Bindings[i] = b
			return nil
		}
	}
	return fmt.Errorf("binding %q not found", b.ID)
}
