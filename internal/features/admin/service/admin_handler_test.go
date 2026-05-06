package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
)

func TestGetSystemInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &AdminHandler{
		appVersion: "v1.0.0-test",
		dbDialect:  "sqlite3",
		startTime:  time.Now().Add(-5 * time.Minute),
	}

	r := gin.New()
	adapter := ginadapter.NewStdRouterAdapter(&r.RouterGroup)
	adapter.GET("/api/v1/admin/settings/info", handler.getSystemInfo())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings/info", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Code != 0 {
		t.Errorf("expected code 0, got %d", resp.Code)
	}

	data := resp.Data

	// Verify required fields exist
	requiredFields := []string{"version", "goVersion", "database", "os", "uptime", "totalMemory", "usedMemory", "cpuUsage", "memoryUsage"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			t.Errorf("missing required field: %s", field)
		}
	}

	// Verify specific values
	if data["version"] != "v1.0.0-test" {
		t.Errorf("expected version 'v1.0.0-test', got %v", data["version"])
	}
	if data["database"] != "SQLite" {
		t.Errorf("expected database 'SQLite', got %v", data["database"])
	}
}

func TestGetSystemInfo_PostgresDialect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &AdminHandler{
		appVersion: "v2.0.0",
		dbDialect:  "postgres",
		startTime:  time.Now(),
	}

	r := gin.New()
	adapter := ginadapter.NewStdRouterAdapter(&r.RouterGroup)
	adapter.GET("/info", handler.getSystemInfo())

	req := httptest.NewRequest(http.MethodGet, "/info", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Data["database"] != "PostgreSQL" {
		t.Errorf("expected database 'PostgreSQL', got %v", resp.Data["database"])
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
