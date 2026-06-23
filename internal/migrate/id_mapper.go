package migrate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type MemoryIDMapper struct {
	maps map[string]map[string]string
	mu   sync.RWMutex
}

func NewMemoryIDMapper() *MemoryIDMapper {
	return &MemoryIDMapper{
		maps: make(map[string]map[string]string),
	}
}

func (m *MemoryIDMapper) Map(sourceType, sourceID string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if typeMap, ok := m.maps[sourceType]; ok {
		if targetID, ok := typeMap[sourceID]; ok {
			return targetID, true
		}
	}
	return "", false
}

func (m *MemoryIDMapper) Set(sourceType, sourceID, targetID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.maps[sourceType] == nil {
		m.maps[sourceType] = make(map[string]string)
	}
	m.maps[sourceType][sourceID] = targetID
}

func (m *MemoryIDMapper) SaveToFile(path string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, err := json.MarshalIndent(m.maps, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (m *MemoryIDMapper) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return json.Unmarshal(data, &m.maps)
}

type MigrationState struct {
	ID         string            `json:"id"`
	Source     SourceConfig      `json:"source"`
	Target     TargetConfig      `json:"target"`
	Progress   Progress          `json:"progress"`
	PhaseState map[string]string `json:"phase_state"`
	CreatedAt  string            `json:"created_at"`
	UpdatedAt  string            `json:"updated_at"`
}

func LoadMigrationState(path string) (*MigrationState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var state MigrationState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func SaveMigrationState(path string, state *MigrationState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
