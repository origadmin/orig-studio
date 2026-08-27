package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	std "origadmin/application/origstudio/internal/pkg/http/std"
)

func TestGetSystemInfo(t *testing.T) {
	handler := &AdminHandler{
		appVersion: "v1.0.0-test",
		dbDialect:  "sqlite3",
		startTime:  time.Now().Add(-5 * time.Minute),
	}

	r := std.NewRouter()
	r.GET("/api/v1/admin/settings/info", handler.getSystemInfo())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings/info", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	requiredFields := []string{"version", "goVersion", "database", "os", "uptime", "totalMemory", "usedMemory", "cpuUsage", "memoryUsage"}
	for _, field := range requiredFields {
		if _, ok := resp[field]; !ok {
			t.Errorf("missing required field: %s", field)
		}
	}

	if resp["version"] != "v1.0.0-test" {
		t.Errorf("expected version 'v1.0.0-test', got %v", resp["version"])
	}
	if resp["database"] != "SQLite" {
		t.Errorf("expected database 'SQLite', got %v", resp["database"])
	}
}

func TestGetSystemInfo_PostgresDialect(t *testing.T) {
	handler := &AdminHandler{
		appVersion: "v2.0.0",
		dbDialect:  "postgres",
		startTime:  time.Now(),
	}

	r := std.NewRouter()
	r.GET("/info", handler.getSystemInfo())

	req := httptest.NewRequest(http.MethodGet, "/info", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["database"] != "PostgreSQL" {
		t.Errorf("expected database 'PostgreSQL', got %v", resp["database"])
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{0, "0s"},
		{1 * time.Second, "1s"},
		{90 * time.Second, "1m 30s"},
		{3661 * time.Second, "1h 1m 1s"},
		{90061 * time.Second, "1d 1h 1m 1s"},
	}
	for _, tt := range tests {
		result := formatDuration(tt.input)
		if result != tt.expected {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KiB"},
		{1048576, "1.0 MiB"},
		{1073741824, "1.0 GiB"},
		{1536, "1.5 KiB"},
	}
	for _, tt := range tests {
		result := formatBytes(tt.input)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
