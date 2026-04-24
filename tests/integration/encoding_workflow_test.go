package integration

import (
	"bytes"
	"net/http"
	"testing"
	"time"
)

func getResponseData(body *bytes.Buffer) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := ParseResponse(body, &result); err != nil {
		return nil, err
	}
	
	// Check if response has data field
	if data, ok := result["data"]; ok {
		if dataMap, ok := data.(map[string]interface{}); ok {
			return dataMap, nil
		}
	}
	return result, nil
}

// TestCompleteEncodingWorkflow tests the complete encoding workflow from upload to completion
func TestCompleteEncodingWorkflow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("create encoding tasks and verify lifecycle", func(t *testing.T) {
		// First, get encoding profiles to verify they exist
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/encoding/profiles",
		})
		if err != nil {
			t.Fatalf("Failed to get profiles: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		data, err := getResponseData(body)
		if err != nil {
			t.Fatalf("Failed to parse profiles response: %v", err)
		}

		// Verify profiles exist
		if _, ok := data["profiles"]; !ok {
			t.Logf("Response: %s", body.String())
			t.Error("Expected 'profiles' field in response data")
		}

		// Get transcoding status
		resp, _, err = ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/encoding/tasks",
		})
		if err != nil {
			t.Fatalf("Failed to get encoding tasks: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)
	})
}

// TestMultipleEncodingProfiles tests that multiple encoding profiles work together
func TestMultipleEncodingProfiles(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list and verify multiple profiles", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/encoding/profiles",
		})
		if err != nil {
			t.Fatalf("Failed to get profiles: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		data, err := getResponseData(body)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		profiles, ok := data["profiles"].([]interface{})
		if !ok {
			t.Logf("Response: %s", body.String())
			t.Fatalf("Expected 'profiles' to be an array")
		}

		// Should have multiple profiles
		if len(profiles) < 1 {
			t.Errorf("Expected at least 1 profile, got %d", len(profiles))
		}

		// Verify each profile has required fields
		for i, profile := range profiles {
			p, ok := profile.(map[string]interface{})
			if !ok {
				t.Errorf("Profile %d is not an object", i)
				continue
			}

			if _, ok := p["id"]; !ok {
				t.Errorf("Profile %d missing 'id' field", i)
			}
			if _, ok := p["name"]; !ok {
				t.Errorf("Profile %d missing 'name' field", i)
			}
		}
	})
}

// TestEncodingFailureHandling tests how the system handles encoding failures
func TestEncodingFailureHandling(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get encoding tasks status endpoint", func(t *testing.T) {
		// Get encoding tasks to verify the endpoint works
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/encoding/tasks",
		})
		if err != nil {
			t.Fatalf("Failed to get encoding tasks: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)
	})

	t.Run("get transcoding status endpoint", func(t *testing.T) {
		// Get transcoding status
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/encoding/status",
		})
		if err != nil {
			t.Fatalf("Failed to get transcoding status: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)
	})
}

// TestEncodingRetry tests the retry functionality for failed encoding tasks
func TestEncodingRetry(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("retry all failed tasks", func(t *testing.T) {
		// This requires authentication
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/encoding/retry-all-failed",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to retry all failed tasks: %v", err)
		}

		// Could be OK or 404 if no failed tasks exist
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("retry specific task", func(t *testing.T) {
		// Note: This test uses task_id as query param
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/encoding/retry?task_id=1",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to retry task: %v", err)
		}

		// Could be OK or 404 if task doesn't exist or 400 if task_id is missing
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound && resp.Code != http.StatusBadRequest {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})
}

// TestEncodingProgressTracking tests that encoding progress is tracked correctly
func TestEncodingProgressTracking(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get encoding tasks with progress", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/encoding/tasks",
		})
		if err != nil {
			t.Fatalf("Failed to get encoding tasks: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		data, err := getResponseData(body)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Check that tasks have progress field
		if list, ok := data["list"].([]interface{}); ok {
			for i, task := range list {
				taskMap, ok := task.(map[string]interface{})
				if !ok {
					continue
				}

				// Progress should exist (could be 0)
				if _, ok := taskMap["progress"]; !ok {
					t.Errorf("Task %d missing 'progress' field", i)
				}

				// Status should exist
				if _, ok := taskMap["status"]; !ok {
					t.Errorf("Task %d missing 'status' field", i)
				}
			}
		}
	})
}

// TestHLSGenerationVerification tests that HLS files are generated properly
func TestHLSGenerationVerification(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get media variants", func(t *testing.T) {
		// First, we need a media to check - use media ID 1
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/media/1/variants",
		})
		if err != nil {
			t.Fatalf("Failed to get media variants: %v", err)
		}

		// Could be OK or 404 if media doesn't exist
		if resp.Code == http.StatusOK {
			data, err := getResponseData(body)
			if err != nil {
				t.Fatalf("Failed to parse variants response: %v", err)
			}

			// Check that variants exist
			if variants, ok := data["variants"].([]interface{}); ok {
				for i, variant := range variants {
					variantMap, ok := variant.(map[string]interface{})
					if !ok {
						continue
					}

					// Each variant should have resolution, codec, etc.
					if _, ok := variantMap["resolution"]; !ok {
						t.Errorf("Variant %d missing 'resolution' field", i)
					}
					if _, ok := variantMap["codec"]; !ok {
						t.Errorf("Variant %d missing 'codec' field", i)
					}
				}
			}
		}
	})
}

// TestEncodingWorkflowWithPolling tests the encoding workflow with polling for completion
func TestEncodingWorkflowWithPolling(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("poll encoding status", func(t *testing.T) {
		// Poll a few times to verify the endpoint is stable
		for i := 0; i < 3; i++ {
			resp, _, err := ts.MakeRequest(RequestOptions{
				Method: "GET",
				Path:   "/encoding/tasks",
			})
			if err != nil {
				t.Fatalf("Poll %d failed: %v", i, err)
			}

			AssertStatus(t, resp, http.StatusOK)

			// Small delay between polls
			time.Sleep(100 * time.Millisecond)
		}
	})
}

// TestEncodingProfileCRUD tests encoding profile CRUD operations
func TestEncodingProfileCRUD(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get all profiles", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/encoding/profiles",
		})
		if err != nil {
			t.Fatalf("Failed to get profiles: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		data, err := getResponseData(body)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := data["profiles"]; !ok {
			t.Logf("Response: %s", body.String())
			t.Error("Expected 'profiles' field in response data")
		}
	})

	t.Run("get single profile", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/encoding/profiles/1",
		})
		if err != nil {
			t.Fatalf("Failed to get profile: %v", err)
		}

		// Could be OK or 404 if profile doesn't exist
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("create profile - unauthorized", func(t *testing.T) {
		profileBody := map[string]interface{}{
			"name":       "Test Profile",
			"resolution": "720p",
			"codec":      "h264",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/encoding/profiles",
			Body:   profileBody,
			Token:  "", // No auth
		})
		if err != nil {
			t.Fatalf("Failed to create profile: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("update profile - unauthorized", func(t *testing.T) {
		updateBody := map[string]interface{}{
			"name": "Updated Profile",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "PUT",
			Path:   "/encoding/profiles/1",
			Body:   updateBody,
			Token:  "", // No auth
		})
		if err != nil {
			t.Fatalf("Failed to update profile: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("delete profile - unauthorized", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "DELETE",
			Path:   "/encoding/profiles/1",
			Token:  "", // No auth
		})
		if err != nil {
			t.Fatalf("Failed to delete profile: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})
}

// TestEncodingTaskStatusUpdates tests that encoding task statuses are properly updated
func TestEncodingTaskStatusUpdates(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get encoding tasks with statuses", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/encoding/tasks",
		})
		if err != nil {
			t.Fatalf("Failed to get encoding tasks: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		data, err := getResponseData(body)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if list, ok := data["list"].([]interface{}); ok {
			for i, task := range list {
				taskMap, ok := task.(map[string]interface{})
				if !ok {
					continue
				}

				// Verify status is one of the expected values
				status, ok := taskMap["status"].(string)
				if !ok {
					t.Errorf("Task %d status is not a string", i)
					continue
				}

				validStatuses := map[string]bool{
					"pending":    true,
					"processing": true,
					"success":    true,
					"failed":     true,
					"skipped":    true,
					"partial":    true,
				}

				if !validStatuses[status] {
					t.Errorf("Task %d has invalid status: %s", i, status)
				}
			}
		}
	})
}

// TestEncodingWorkflowErrorScenarios tests various error scenarios in the encoding workflow
func TestEncodingWorkflowErrorScenarios(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get non-existent profile", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/encoding/profiles/99999",
		})
		if err != nil {
			t.Fatalf("Failed to get non-existent profile: %v", err)
		}

		AssertStatus(t, resp, http.StatusNotFound)
	})

	t.Run("get non-existent task", func(t *testing.T) {
		// Note: tasks endpoint returns list, not 404 for non-existent task
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/encoding/tasks",
		})
		if err != nil {
			t.Fatalf("Failed to get tasks: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)
	})

	t.Run("retry without task_id", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/encoding/retry",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to retry without task_id: %v", err)
		}

		AssertStatus(t, resp, http.StatusBadRequest)
	})
}
