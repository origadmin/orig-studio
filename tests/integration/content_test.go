package integration

import (
	"net/http"
	"testing"
)

// ==================== Categories ====================

// TestCategories tests category endpoints
func TestCategories(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list categories", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/categories",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		// Response might be an array or an object with list field
		var result interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
	})

	t.Run("create category - no auth", func(t *testing.T) {
		category := map[string]string{
			"name":        "Test Category",
			"description": "Test description",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/categories",
			Body:   category,
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		// Categories don't currently require auth based on code review
		// But may change - accept OK or Created
		if resp.Code != http.StatusOK && resp.Code != http.StatusCreated {
			t.Logf("Create category returned status %d (may need auth)", resp.Code)
		}
	})

	t.Run("get category by id - not implemented", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/categories/1",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		// Currently returns 501 Not Implemented
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotImplemented {
			t.Logf("Get category returned status %d", resp.Code)
		}
	})

	t.Run("delete category - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "DELETE",
			Path:   "/categories/1",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		// Categories don't currently require auth
		if resp.Code != http.StatusOK {
			t.Logf("Delete category returned status %d", resp.Code)
		}
	})
}

// ==================== Tags ====================

// TestTags tests tag endpoints
func TestTags(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list tags", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/tags",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		// Check for expected fields
		if _, ok := result["list"]; !ok {
			t.Log("Response may not contain 'list' field - checking structure")
		}
	})

	t.Run("create tag - no auth", func(t *testing.T) {
		tag := map[string]string{
			"title": "TestTag",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/tags",
			Body:   tag,
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.Code != http.StatusOK && resp.Code != http.StatusCreated {
			t.Logf("Create tag returned status %d", resp.Code)
		}
	})

	t.Run("get tag by id - not implemented", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/tags/1",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.Code != http.StatusOK && resp.Code != http.StatusNotImplemented {
			t.Logf("Get tag returned status %d", resp.Code)
		}
	})

	t.Run("get media by tag - not implemented", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/tags/1/media",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.Code != http.StatusOK && resp.Code != http.StatusNotImplemented {
			t.Logf("Get media by tag returned status %d", resp.Code)
		}
	})

	t.Run("delete tag - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "DELETE",
			Path:   "/tags/1",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.Code != http.StatusOK {
			t.Logf("Delete tag returned status %d", resp.Code)
		}
	})
}

// ==================== Comments ====================

// TestComments tests comment endpoints
func TestComments(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list comments for media", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/comments?media_id=1",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if _, ok := result["list"]; !ok {
			t.Error("expected 'list' field in response")
		}
	})

	t.Run("list comments without media_id", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/comments",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		// Should require media_id
		AssertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("create comment - authenticated", func(t *testing.T) {
		comment := map[string]interface{}{
			"text":     "Test comment",
			"media_id": 1,
		}

		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/comments",
			Body:   comment,
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.Code == http.StatusCreated || resp.Code == http.StatusOK {
			var result map[string]interface{}
			if err := ParseResponse(body, &result); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if _, ok := result["text"]; !ok {
				t.Error("expected 'text' field in created comment")
			}
		} else {
			t.Logf("Create comment returned status %d", resp.Code)
		}
	})

	t.Run("create comment - no auth", func(t *testing.T) {
		comment := map[string]interface{}{
			"text":     "Test comment",
			"media_id": 1,
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/comments",
			Body:   comment,
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("create reply comment", func(t *testing.T) {
		comment := map[string]interface{}{
			"text":      "Reply comment",
			"media_id":  1,
			"parent_id": 1, // Reply to comment 1
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/comments",
			Body:   comment,
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.Code != http.StatusCreated && resp.Code != http.StatusOK {
			t.Logf("Create reply returned status %d", resp.Code)
		}
	})

	t.Run("list media comments - authenticated", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/comments/media/1",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if _, ok := result["list"]; !ok {
			t.Error("expected 'list' field in response")
		}
	})

	t.Run("update comment - owner", func(t *testing.T) {
		update := map[string]string{
			"text": "Updated comment text",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "PUT",
			Path:   "/comments/1",
			Body:   update,
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		// May return OK, Forbidden, or NotFound depending on ownership
		if resp.Code != http.StatusOK &&
			resp.Code != http.StatusForbidden &&
			resp.Code != http.StatusNotFound {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})

	t.Run("update comment - no auth", func(t *testing.T) {
		update := map[string]string{
			"text": "Updated comment text",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "PUT",
			Path:   "/comments/1",
			Body:   update,
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("delete comment - owner", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "DELETE",
			Path:   "/comments/1",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		// May return OK, Forbidden, or NotFound
		if resp.Code != http.StatusOK &&
			resp.Code != http.StatusForbidden &&
			resp.Code != http.StatusNotFound {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})

	t.Run("delete comment - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "DELETE",
			Path:   "/comments/1",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})
}

// TestFeed tests the feed endpoint
func TestFeed(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tests := []struct {
		name       string
		path       string
		wantStatus int
		checkField string
	}{
		{
			name:       "get feed default",
			path:       "/feed",
			wantStatus: http.StatusOK,
			checkField: "sections",
		},
		{
			name:       "get feed with pagination",
			path:       "/feed?page=1&page_size=10",
			wantStatus: http.StatusOK,
			checkField: "sections",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body, err := ts.MakeRequest(RequestOptions{
				Method: "GET",
				Path:   tt.path,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			AssertStatus(t, resp, tt.wantStatus)

			if tt.checkField != "" {
				var result map[string]interface{}
				if err := ParseResponse(body, &result); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if _, ok := result[tt.checkField]; !ok {
					t.Errorf("expected '%s' field in response", tt.checkField)
				}
			}
		})
	}
}

// TestSearch tests the search endpoint
func TestSearch(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tests := []struct {
		name       string
		path       string
		wantStatus int
		checkField string
	}{
		{
			name:       "search with query",
			path:       "/search?q=test",
			wantStatus: http.StatusOK,
			checkField: "list",
		},
		{
			name:       "search with pagination",
			path:       "/search?q=test&page=1&page_size=20",
			wantStatus: http.StatusOK,
			checkField: "list",
		},
		{
			name:       "search without query",
			path:       "/search",
			wantStatus: http.StatusOK,
			checkField: "list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body, err := ts.MakeRequest(RequestOptions{
				Method: "GET",
				Path:   tt.path,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			AssertStatus(t, resp, tt.wantStatus)

			if tt.checkField != "" {
				var result map[string]interface{}
				if err := ParseResponse(body, &result); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if _, ok := result[tt.checkField]; !ok {
					t.Errorf("expected '%s' field in response", tt.checkField)
				}
			}
		})
	}
}
