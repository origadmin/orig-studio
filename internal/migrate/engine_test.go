package migrate

import (
	"context"
	"fmt"
	"testing"
)

func TestMemoryIDMapper(t *testing.T) {
	m := NewMemoryIDMapper()

	m.Set("user", "src-1", "tgt-1")
	m.Set("user", "src-2", "tgt-2")
	m.Set("media", "src-m1", "tgt-m1")

	id, ok := m.Map("user", "src-1")
	if !ok || id != "tgt-1" {
		t.Errorf("expected tgt-1, got %s, ok=%v", id, ok)
	}

	id, ok = m.Map("user", "src-999")
	if ok {
		t.Errorf("expected not found, got %s", id)
	}

	id, ok = m.Map("media", "src-m1")
	if !ok || id != "tgt-m1" {
		t.Errorf("expected tgt-m1, got %s, ok=%v", id, ok)
	}
}

func TestMemoryIDMapperPersistence(t *testing.T) {
	m := NewMemoryIDMapper()
	m.Set("user", "src-1", "tgt-1")
	m.Set("category", "cat-1", "100")

	path := t.TempDir() + "/id_map.json"
	if err := m.SaveToFile(path); err != nil {
		t.Fatalf("save: %v", err)
	}

	m2 := NewMemoryIDMapper()
	if err := m2.LoadFromFile(path); err != nil {
		t.Fatalf("load: %v", err)
	}

	id, ok := m2.Map("user", "src-1")
	if !ok || id != "tgt-1" {
		t.Errorf("expected tgt-1, got %s, ok=%v", id, ok)
	}

	id, ok = m2.Map("category", "cat-1")
	if !ok || id != "100" {
		t.Errorf("expected 100, got %s, ok=%v", id, ok)
	}
}

func TestMigrationStatePersistence(t *testing.T) {
	path := t.TempDir() + "/state.json"

	state := &MigrationState{
		ID: "test-migration",
		Progress: Progress{
			Phase:      PhaseUsers,
			Status:     StatusRunning,
			DoneItems:  42,
			TotalItems: 100,
		},
	}

	if err := SaveMigrationState(path, state); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := LoadMigrationState(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if loaded.ID != "test-migration" {
		t.Errorf("expected test-migration, got %s", loaded.ID)
	}
	if loaded.Progress.Phase != PhaseUsers {
		t.Errorf("expected PhaseUsers, got %s", loaded.Progress.Phase)
	}
	if loaded.Progress.DoneItems != 42 {
		t.Errorf("expected 42, got %d", loaded.Progress.DoneItems)
	}
}

func TestLoadNonexistentState(t *testing.T) {
	state, err := LoadMigrationState("/nonexistent/path/state.json")
	if err != nil {
		t.Errorf("expected nil error for nonexistent file, got %v", err)
	}
	if state != nil {
		t.Errorf("expected nil state, got %v", state)
	}
}

type mockIterator struct {
	items []int
	idx   int
	err   error
}

func (it *mockIterator) Next(ctx context.Context) bool {
	if it.idx < len(it.items) {
		it.idx++
		return true
	}
	return false
}
func (it *mockIterator) Err() error   { return it.err }
func (it *mockIterator) Close() error { return nil }

func TestPhaseConstants(t *testing.T) {
	phases := []Phase{PhaseDiscover, PhaseUsers, PhaseCategories, PhaseTags, PhaseMedia, PhaseComments, PhaseFiles, PhaseVerify, PhaseComplete}
	for _, p := range phases {
		if string(p) == "" {
			t.Errorf("phase should not be empty")
		}
	}
}

func TestStatusConstants(t *testing.T) {
	statuses := []Status{StatusPending, StatusRunning, StatusPaused, StatusCompleted, StatusFailed, StatusCancelled}
	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("status should not be empty")
		}
	}
}

func TestConsoleReporter(t *testing.T) {
	r := NewConsoleReporter()
	r.UpdateProgress(&Progress{
		Phase:      PhaseUsers,
		Status:     StatusRunning,
		DoneItems:  50,
		TotalItems: 100,
	})
	r.ReportError(PhaseUsers, "test-user", fmt.Errorf("test error"))
	r.ReportWarning(PhaseUsers, "test-user", "test warning")

	if len(r.errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(r.errors))
	}
	if len(r.warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(r.warnings))
	}
}
