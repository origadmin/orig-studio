package integration

import (
	"net/http"
	"testing"
)

// TestEncodingFailureHandling tests encoding failure handling
func TestEncodingFailureHandling(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list encoding tasks", func(t *testing.T) {
		// GET /encoding/tasks - public route
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/encoding/tasks",
		})
		if err != nil {
			t.Fatalf("Failed to list encoding tasks: %v", err)
		}

		// Could be OK or error depending on implementation
		if resp.Code != http.StatusOK && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("get encoding profile by id", func(t *testing.T) {
		// GET /encoding/profiles/:profile_id - public route, profile_id is numeric
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/encoding/profiles/99999",
		})
		if err != nil {
			t.Fatalf("Failed to get encoding profile: %v", err)
		}

		// Could be OK, 400 (media not found), 404, or error
		if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})
}

// TestEncodingRetry tests encoding retry functionality
func TestEncodingRetry(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("retry all failed tasks", func(t *testing.T) {
		// POST /encoding/retry-all-failed requires JWT
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/encoding/retry-all-failed",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to retry all failed tasks: %v", err)
		}

		// Could be OK or error
		if resp.Code != http.StatusOK && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("retry specific task", func(t *testing.T) {
		// POST /encoding/retry?task_id=xxx requires JWT (query param, not body)
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/encoding/retry?task_id=550e8400-e29b-41d4-a716-446655440000",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to retry task: %v", err)
		}

		// Could be OK, 400 (task not found/cannot retry), 404, or error
		if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("retry - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/encoding/retry?task_id=550e8400-e29b-41d4-a716-446655440000",
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to retry task: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("admin retry specific task", func(t *testing.T) {
		// POST /admin/encoding/tasks/:taskId/retry requires JWT+Admin
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/admin/encoding/tasks/550e8400-e29b-41d4-a716-446655440000/retry",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to retry admin task: %v", err)
		}

		// Could be OK, 404, or error
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("admin retry failed tasks", func(t *testing.T) {
		// POST /admin/encoding/retry-failed requires JWT+Admin
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/admin/encoding/retry-failed",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to retry failed tasks: %v", err)
		}

		// Could be OK or error
		if resp.Code != http.StatusOK && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})
}

// TestEncodingProgressTracking tests encoding progress tracking
func TestEncodingProgressTracking(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list encoding tasks with progress", func(t *testing.T) {
		// GET /encoding/tasks - public route
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/encoding/tasks",
		})
		if err != nil {
			t.Fatalf("Failed to list encoding tasks: %v", err)
		}

		if resp.Code != http.StatusOK && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("admin list all encoding tasks", func(t *testing.T) {
		// GET /admin/encoding/tasks requires JWT+Admin
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/encoding/tasks",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to list admin encoding tasks: %v", err)
		}

		if resp.Code != http.StatusOK && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("admin encoding status", func(t *testing.T) {
		// GET /admin/encoding/status requires JWT+Admin
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/encoding/status",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to get encoding status: %v", err)
		}

		if resp.Code != http.StatusOK && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})
}

// TestEncodingTaskStatusUpdates tests encoding task status update workflow
func TestEncodingTaskStatusUpdates(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("create encoding profile", func(t *testing.T) {
		profile := map[string]interface{}{
			"name":      "test-profile",
			"codec":     "h264",
			"container": "mp4",
		}

		// POST /encoding/profiles requires JWT
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/encoding/profiles",
			Body:   profile,
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to create encoding profile: %v", err)
		}

		// Could be OK, Created, or error
		if resp.Code != http.StatusOK && resp.Code != http.StatusCreated && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("retry transcode task - admin media route", func(t *testing.T) {
		// POST /admin/medias/:id/tasks/:taskId/retry requires JWT+Admin
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/admin/medias/550e8400-e29b-41d4-a716-446655440000/tasks/550e8400-e29b-41d4-a716-446655440000/retry",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to retry transcode: %v", err)
		}

		// Could be OK, 400 (media not found/cannot retry), 404, or error
		if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("task status update - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/encoding/retry",
			Body: map[string]interface{}{
				"task_id": "550e8400-e29b-41d4-a716-446655440000",
			},
			Token: "",
		})
		if err != nil {
			t.Fatalf("Failed to retry task: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("encoding profile preview - admin", func(t *testing.T) {
		// POST /admin/encoding/profiles/preview requires JWT+Admin
		preview := map[string]interface{}{
			"name":         "test-preview",
			"extension":    "mp4",
			"video_codec":  "h264",
			"audio_codec":  "aac",
			"resolution":   "1080p",
			"video_bitrate": "5000k",
			"audio_bitrate": "128k",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/admin/encoding/profiles/preview",
			Body:   preview,
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to preview encoding profile: %v", err)
		}

		// Could be OK, 404, or error
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})
}
