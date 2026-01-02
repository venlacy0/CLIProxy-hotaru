package donation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const bindingsFileName = "bindings.json"

// BindingStore manages user bindings with JSON file persistence.
type BindingStore struct {
	mu       sync.RWMutex
	filePath string
	bindings map[int]*UserBinding // keyed by LinuxDoID
}

// bindingsFile represents the JSON file structure.
type bindingsFile struct {
	Bindings []*UserBinding `json:"bindings"`
}

// NewBindingStore creates a new binding store with the given base directory.
func NewBindingStore(baseDir string) (*BindingStore, error) {
	filePath := filepath.Join(baseDir, bindingsFileName)
	store := &BindingStore{
		filePath: filePath,
		bindings: make(map[int]*UserBinding),
	}

	// Load existing bindings if file exists
	if err := store.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return store, nil
}

// load reads bindings from the JSON file.
func (s *BindingStore) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var file bindingsFile
	if err := json.Unmarshal(data, &file); err != nil {
		return err
	}

	s.bindings = make(map[int]*UserBinding)
	for _, binding := range file.Bindings {
		if binding != nil {
			s.bindings[binding.LinuxDoID] = binding
		}
	}

	return nil
}

// save writes bindings to the JSON file.
func (s *BindingStore) save() error {
	// Ensure directory exists
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Convert map to slice
	bindings := make([]*UserBinding, 0, len(s.bindings))
	for _, binding := range s.bindings {
		bindings = append(bindings, binding)
	}

	file := bindingsFile{Bindings: bindings}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0600)
}

// GetByLinuxDoID retrieves a binding by Linux Do user ID.
// Returns nil if no binding exists.
func (s *BindingStore) GetByLinuxDoID(linuxDoID int) *UserBinding {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.bindings[linuxDoID]
}

// Create creates a new binding.
// Returns an error if a binding already exists for the Linux Do user.
func (s *BindingStore) Create(binding *UserBinding) error {
	if binding == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Set bound time if not set
	if binding.BoundAt.IsZero() {
		binding.BoundAt = time.Now()
	}

	s.bindings[binding.LinuxDoID] = binding
	return s.save()
}

// Delete removes a binding by Linux Do user ID.
func (s *BindingStore) Delete(linuxDoID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.bindings, linuxDoID)
	return s.save()
}

// GetByNewAPIUserID retrieves a binding by new-api user ID.
// Returns nil if no binding exists.
func (s *BindingStore) GetByNewAPIUserID(newAPIUserID int) *UserBinding {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, binding := range s.bindings {
		if binding.NewAPIUserID == newAPIUserID {
			return binding
		}
	}
	return nil
}

// List returns all bindings.
func (s *BindingStore) List() []*UserBinding {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bindings := make([]*UserBinding, 0, len(s.bindings))
	for _, binding := range s.bindings {
		bindings = append(bindings, binding)
	}
	return bindings
}
