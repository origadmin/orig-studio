package integration

import (
	"fmt"
	"net/http"
	"testing"
)

// TestMediaList tests the media list endpoint
func TestMediaList(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tests := []struct {
		name       string
		path       string
		wantStatus int
		checkList  bool
	}{
		{
			name:       "list all media",
			path:       "/media",
			wantStatus: http.StatusOK,
			checkList:  true,
		},
		{
			name:       "list with pagination",
			path:       "/media?page=1&page_size=10",
			wantStatus: http.StatusOK,
			checkList:  true,
		},
		{
			name:       "list with type filter",
			path:       "/media?type=video",
			wantStatus: http.StatusOK,
			checkList:  true,
		},
		{
			name:       "list with search keyword",
			path:       "/media?keyword=test",
			wantStatus: http.StatusOK,
			checkList:  true,
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

			if tt.checkList {
				var result map[string]interface{}
				if err := ParseResponse(body, &result); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if _, ok := result["list"]; !ok {
					t.Error("expected 'list' field in response")
				}
				if _, ok := result["total"]; !ok {
					t.Error("expected 'total' field in response")
				}
			}
		})
	}
}

// TestMediaGetByID tests getting a single media by ID
func TestMediaGetByID(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tests := []struct {
		name       string
		id         string
		wantStatus int
		wantFound  bool
	}{
		{
			name:       "get existing media",
			id:         "1",
			wantStatus: http.StatusOK,
			wantFound:  true,
		},
		{
			name:       "non-existent media",
			id:         "99999",
			wantStatus: http.StatusNotFound,
			wantFound:  false,
		},
		{
			name:       "invalid id format",
			id:         "abc",
			wantStatus: http.StatusBadRequest,
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body, err := ts.MakeRequest(RequestOptions{
				Method: "GET",
				Path:   "/media/" + tt.id,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			AssertStatus(t, resp, tt.wantStatus)

			if tt.wantFound {
				var result map[string]interface{}
				if err := ParseResponse(body, &result); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if _, ok := result["id"]; !ok {
					t.Error("expected 'id' field in response")
				}
			}
		})
	}
}

// TestMediaCreateUpdateDelete tests media CRUD operations
func TestMediaCreateUpdateDelete(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Create media metadata (without actual file upload for now)
	// Note: File upload requires multipart form and is tested separately

	// Test update media (requires auth)
	t.Run("update media - owner", func(t *testing.T) {
		// First, we need a media to update - we'll use the existing one
		updateBody := map[string]interface{}{
			"title":       "Updated Title",
			"description": "Updated description",
		}

		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "PUT",
			Path:   "/media/1",
			Body:   updateBody,
			Token:  ts.GetToken(RoleUser), // user1 should be owner or admin/staff
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		// Might be 200, 403 (if not owner), or 404
		if resp.StatusCode != http.StatusOK &&
			resp.StatusCode != http.StatusForbidden &&
			resp.StatusCode != http.StatusNotFound {
			t.Errorf("unexpected status: %d", resp.StatusCode)
		}

		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			if err := ParseResponse(body, &result); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if result["title"] != "Updated Title" {
				t.Errorf("expected title to be updated, got %v", result["title"])
			}
		}
	})

	// Test update media without auth
	t.Run("update media - no auth", func(t *testing.T) {
		updateBody := map[string]interface{}{
			"title": "Updated Title",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "PUT",
			Path:   "/media/1",
			Body:   updateBody,
			Token:  "", // No auth
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	// Test delete media (requires auth)
	t.Run("delete media - owner", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "DELETE",
			Path:   "/media/1",
			Token:  ts.GetToken(RoleAdmin), // Use admin to ensure we can delete
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		// Might be 200, 403, or 404 depending on data state
		if resp.StatusCode != http.StatusOK &&
			resp.StatusCode != http.StatusForbidden &&
			resp.StatusCode != http.StatusNotFound {
			t.Errorf("unexpected status: %d", resp.StatusCode)
		}
	})
}

// TestMediaLike tests media like/dislike functionality
func TestMediaLike(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tests := []struct {
		name       string
		method     string
		path       string
		body       interface{}
		token      string
		wantStatus int
	}{
		{
			name:       "like media - authenticated",
			method:     "POST",
			path:       "/media/1/like",
			body:       nil,
			token:      ts.GetToken(RoleUser),
			wantStatus: http.StatusOK,
		},
		{
			name:       "dislike media - authenticated",
			method:     "POST",
			path:       "/media/1/dislike",
			body:       nil,
			token:      ts.GetToken(RoleUser),
			wantStatus: http.StatusOK,
		},
		{
			name:       "like media - no auth",
			method:     "POST",
			path:       "/media/1/like",
			body:       nil,
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "get like status - public",
			method:     "GET",
			path:       "/media/1/like",
			body:       nil,
			token:      "",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body, err := ts.MakeRequest(RequestOptions{
				Method: tt.method,
				Path:   tt.path,
				Body:   tt.body,
				Token:  tt.token,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			AssertStatus(t, resp, tt.wantStatus)

			if tt.wantStatus == http.StatusOK {
				var result map[string]interface{}
				if err := ParseResponse(body, &result); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				// Check for expected fields
				if _, ok := result["is_liked"]; !ok && tt.method == "GET" {
					t.Error("expected 'is_liked' field in response")
				}
			}
		})
	}
}

// TestMediaFavorite tests media favorite functionality
func TestMediaFavorite(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tests := []struct {
		name       string
		method     string
		path       string
		token      string
		wantStatus int
	}{
		{
			name:       "favorite media - authenticated",
			method:     "POST",
			path:       "/media/1/favorite",
			token:      ts.GetToken(RoleUser),
			wantStatus: http.StatusOK,
		},
		{
			name:       "favorite media - no auth",
			method:     "POST",
			path:       "/media/1/favorite",
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "get favorite status - public",
			method:     "GET",
			path:       "/media/1/favorite",
			token:      "",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body, err := ts.MakeRequest(RequestOptions{
				Method: tt.method,
				Path:   tt.path,
				Token:  tt.token,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			AssertStatus(t, resp, tt.wantStatus)

			if tt.wantStatus == http.StatusOK {
				var result map[string]interface{}
				if err := ParseResponse(body, &result); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if _, ok := result["is_favorited"]; !ok {
					t.Error("expected 'is_favorited' field in response")
				}
			}
		})
	}
}

// TestMediaShare tests media share functionality
func TestMediaShare(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tests := []struct {
		name       string
		method     string
		path       string
		token      string
		wantStatus int
	}{
		{
			name:       "get share URL - public",
			method:     "GET",
			path:       "/media/1/share",
			token:      "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "record share - authenticated",
			method:     "POST",
			path:       "/media/1/share",
			token:      ts.GetToken(RoleUser),
			wantStatus: http.StatusOK,
		},
		{
			name:       "record share - no auth",
			method:     "POST",
			path:       "/media/1/share",
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body, err := ts.MakeRequest(RequestOptions{
				Method: tt.method,
				Path:   tt.path,
				Token:  tt.token,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			AssertStatus(t, resp, tt.wantStatus)

			if tt.wantStatus == http.StatusOK {
				var result map[string]interface{}
				if err := ParseResponse(body, &result); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if tt.method == "GET" {
					if _, ok := result["url"]; !ok {
						t.Error("expected 'url' field in response")
					}
				} else {
					if _, ok := result["success"]; !ok {
						t.Error("expected 'success' field in response")
					}
				}
			}
		})
	}
}

// TestMediaVariants tests media variants endpoint
func TestMediaVariants(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tests := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{
			name:       "get variants for existing media",
			id:         "1",
			wantStatus: http.StatusOK,
		},
		{
			name:       "get variants for non-existent media",
			id:         "99999",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _, err := ts.MakeRequest(RequestOptions{
				Method: "GET",
				Path:   fmt.Sprintf("/media/%s/variants", tt.id),
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			AssertStatus(t, resp, tt.wantStatus)
		})
	}
}

// TestMediaTranscodingStatus tests transcoding status endpoints
func TestMediaTranscodingStatus(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "get transcoding status",
			path:       "/media/transcoding/status",
			wantStatus: http.StatusOK,
		},
		{
			name:       "get encoding tasks flat",
			path:       "/media/encoding/tasks",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _, err := ts.MakeRequest(RequestOptions{
				Method: "GET",
				Path:   tt.path,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			AssertStatus(t, resp, tt.wantStatus)
		})
	}
}

// TestMediaEncodeProfiles tests encoding profile endpoints
func TestMediaEncodeProfiles(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tests := []struct {
		name       string
		method     string
		path       string
		body       interface{}
		token      string
		wantStatus int
	}{
		{
			name:       "list profiles - public",
			method:     "GET",
			path:       "/media/profiles",
			body:       nil,
			token:      "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "get single profile - public",
			method:     "GET",
			path:       "/media/profiles/1",
			body:       nil,
			token:      "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "create profile - no auth",
			method:     "POST",
			path:       "/media/profiles",
			body:       map[string]interface{}{"name": "Test Profile"},
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _, err := ts.MakeRequest(RequestOptions{
				Method: tt.method,
				Path:   tt.path,
				Body:   tt.body,
				Token:  tt.token,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			AssertStatus(t, resp, tt.wantStatus)
		})
	}
}
