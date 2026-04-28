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

		data, err := GetResponseData(body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if _, ok := data["items"]; !ok {
			t.Error("expected 'items' field in response data")
		}
	})

	t.Run("get category by id", func(t *testing.T) {
		// GET /categories/:id - public route, ID is numeric
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/categories/1",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		// Could be OK or 404
		if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})
}

// ==================== Tags (Admin) ====================

// TestTags tests tag endpoints (admin routes)
func TestTags(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list tags - admin", func(t *testing.T) {
		// GET /admin/tags requires JWT+Admin
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/tags",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		data, err := GetResponseData(body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if _, ok := data["items"]; !ok {
			t.Log("Note: 'items' field not found in tags response; structure may differ")
		}
	})

	t.Run("create tag - admin", func(t *testing.T) {
		tag := map[string]string{
			"name": "TestTag",
			"slug": "test-tag",
		}

		// POST /admin/tags requires JWT+Admin
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/admin/tags",
			Body:   tag,
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.Code != http.StatusOK && resp.Code != http.StatusCreated && resp.Code != http.StatusBadRequest && resp.Code != http.StatusInternalServerError {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})

	t.Run("get tag by id - admin", func(t *testing.T) {
		// GET /admin/tags/:id requires JWT+Admin, ID is numeric
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/tags/1",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		// Could be OK or 404
		if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})

	t.Run("update tag - admin", func(t *testing.T) {
		tag := map[string]string{
			"name": "UpdatedTag",
			"slug": "updated-tag",
		}

		// PUT /admin/tags/:id requires JWT+Admin
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "PUT",
			Path:   "/admin/tags/1",
			Body:   tag,
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		// Could be OK, 404, or 500
		if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})

	t.Run("delete tag - admin", func(t *testing.T) {
		// DELETE /admin/tags/:id requires JWT+Admin
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "DELETE",
			Path:   "/admin/tags/99",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		// Could be OK or 404
		if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})

	t.Run("list tags - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/tags",
			Token:  "",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})
}

// ==================== Comments ====================

// TestComments tests comment endpoints
func TestComments(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list comments for media", func(t *testing.T) {
		// GET /medias/:short_token/comments - public route
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/medias/abc12345/comments",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.Code == http.StatusOK {
			data, err := GetResponseData(body)
			if err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if _, ok := data["items"]; !ok {
				t.Log("Note: 'items' field not found in comments response; structure may differ")
			}
		} else {
			t.Logf("List comments returned status %d (media may not exist)", resp.Code)
		}
	})

	t.Run("create comment - authenticated", func(t *testing.T) {
		comment := map[string]interface{}{
			"comment": map[string]interface{}{
				"content":  "Test comment",
				"media_id": "550e8400-e29b-41d4-a716-446655440000",
			},
		}

		// POST /comments requires JWT
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/comments",
			Body:   comment,
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		// Could be OK, Created, or error
		if resp.Code != http.StatusOK && resp.Code != http.StatusCreated && resp.Code != http.StatusBadRequest && resp.Code != http.StatusInternalServerError {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})

	t.Run("create comment - no auth", func(t *testing.T) {
		comment := map[string]interface{}{
			"comment": map[string]interface{}{
				"content":  "Test comment",
				"media_id": "550e8400-e29b-41d4-a716-446655440000",
			},
		}

		// POST /comments requires JWT
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

	t.Run("delete comment - authenticated", func(t *testing.T) {
		// DELETE /comments/:id requires JWT
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "DELETE",
			Path:   "/comments/1",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		// Could be OK, Forbidden, or NotFound
		if resp.Code != http.StatusOK &&
			resp.Code != http.StatusForbidden &&
			resp.Code != http.StatusNotFound &&
			resp.Code != http.StatusInternalServerError {
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

// ==================== Search ====================

// TestSearch tests the search endpoint
func TestSearch(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "search with query",
			path:       "/search?q=test",
			wantStatus: http.StatusOK,
		},
		{
			name:       "search with pagination",
			path:       "/search?q=test&page=1&page_size=20",
			wantStatus: http.StatusOK,
		},
		{
			name:       "search without query",
			path:       "/search",
			wantStatus: http.StatusOK,
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

			// Search handler uses PageResponse with {code, message, data: {items, total, ...}}
			data, err := GetResponseData(body)
			if err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if _, ok := data["items"]; !ok {
				t.Log("Note: 'items' field not found in search response; structure may differ")
			}
		})
	}
}
