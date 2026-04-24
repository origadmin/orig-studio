package integration

import (
	"net/http"
	"testing"
)

// TestPermissionGuest tests unauthenticated access (guest)
func TestPermissionGuest(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Public endpoints that should work without auth
	publicEndpoints := []struct {
		name   string
		method string
		path   string
	}{
		{"get feed", "GET", "/feed"},
		{"get health", "GET", "/health"},
		{"list media", "GET", "/media"},
		{"search", "GET", "/search?q=test"},
		{"list categories", "GET", "/categories"},
		{"list tags", "GET", "/tags"},
		{"list comments", "GET", "/comments?media_id=1"},
		{"get share URL", "GET", "/media/1/share"},
		{"get like status", "GET", "/media/1/like"},
		{"get favorite status", "GET", "/media/1/favorite"},
	}

	for _, ep := range publicEndpoints {
		t.Run(ep.name, func(t *testing.T) {
			resp, _, err := ts.MakeRequest(RequestOptions{
				Method: ep.method,
				Path:   ep.path,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			// Should not be 401 Unauthorized
			if resp.Code == http.StatusUnauthorized {
				t.Errorf("public endpoint %s returned 401", ep.path)
			}
		})
	}

	// Protected endpoints that should require auth
	protectedEndpoints := []struct {
		name   string
		method string
		path   string
		body   interface{}
	}{
		{"create media", "POST", "/media/upload", nil},
		{"update media", "PUT", "/media/1", map[string]string{"title": "Test"}},
		{"delete media", "DELETE", "/media/1", nil},
		{"create comment", "POST", "/comments", map[string]interface{}{"text": "test", "media_id": 1}},
		{"like media", "POST", "/media/1/like", nil},
		{"favorite media", "POST", "/media/1/favorite", nil},
		{"record share", "POST", "/media/1/share", nil},
		{"create playlist", "POST", "/playlists", map[string]string{"name": "Test"}},
		{"create channel", "POST", "/channels", map[string]string{"title": "Test"}},
		{"create category", "POST", "/categories", map[string]string{"name": "Test"}},
		{"create tag", "POST", "/tags", map[string]string{"title": "Test"}},
		{"get me", "GET", "/auth/me", nil},
		{"subscribe user", "POST", "/users/1/subscribe", nil},
		{"unsubscribe user", "DELETE", "/users/1/subscribe", nil},
		{"list favorites", "GET", "/favorites", nil},
		{"toggle favorite", "POST", "/favorites", map[string]int{"media_id": 1}},
		{"toggle like", "POST", "/likes", map[string]interface{}{"media_id": 1, "type": "like"}},
		{"list notifications", "GET", "/notifications", nil},
		{"stats dashboard", "GET", "/stats/dashboard", nil},
	}

	for _, ep := range protectedEndpoints {
		t.Run(ep.name+"_requires_auth", func(t *testing.T) {
			resp, _, err := ts.MakeRequest(RequestOptions{
				Method: ep.method,
				Path:   ep.path,
				Body:   ep.body,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			// Should be 401 Unauthorized
			if resp.Code != http.StatusUnauthorized {
				t.Errorf("protected endpoint %s returned %d, expected 401", ep.path, resp.Code)
			}
		})
	}
}

// TestPermissionUser tests authenticated user permissions
func TestPermissionUser(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	userToken := ts.GetToken(RoleUser)

	// Operations a regular user should be able to do
	allowedOperations := []struct {
		name       string
		method     string
		path       string
		body       interface{}
		wantStatus int // Expected status (not necessarily 200, could be 404 if resource doesn't exist)
	}{
		{"get own profile", "GET", "/auth/me", nil, http.StatusOK},
		{"create comment", "POST", "/comments", map[string]interface{}{"text": "test", "media_id": 1}, 0}, // 0 means any non-401/403
		{"like media", "POST", "/media/1/like", nil, 0},
		{"favorite media", "POST", "/media/1/favorite", nil, 0},
		{"list favorites", "GET", "/favorites", nil, http.StatusOK},
		{"toggle favorite", "POST", "/favorites", map[string]int{"media_id": 1}, 0},
		{"toggle like", "POST", "/likes", map[string]interface{}{"media_id": 1, "type": "like"}, 0},
		{"create playlist", "POST", "/playlists", map[string]interface{}{"name": "My Playlist"}, 0},
		{"create channel", "POST", "/channels", map[string]string{"title": "My Channel"}, 0},
		{"subscribe user", "POST", "/users/2/subscribe", nil, 0},
	}

	for _, op := range allowedOperations {
		t.Run("user_can_"+op.name, func(t *testing.T) {
			resp, _, err := ts.MakeRequest(RequestOptions{
				Method: op.method,
				Path:   op.path,
				Body:   op.body,
				Token:  userToken,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			// Should not be 401 Unauthorized or 403 Forbidden
			if resp.Code == http.StatusUnauthorized {
				t.Errorf("user was denied access (401) to %s", op.path)
			}
			if resp.Code == http.StatusForbidden {
				t.Errorf("user was forbidden (403) from %s", op.path)
			}
		})
	}
}

// TestPermissionStaff tests staff/admin elevated permissions
func TestPermissionStaff(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	staffToken := ts.GetToken(RoleStaff)
	adminToken := ts.GetToken(RoleAdmin)

	tests := []struct {
		name       string
		token      string
		role       string
		method     string
		path       string
		body       interface{}
		wantStatus int
	}{
		{
			name:       "staff access stats dashboard",
			token:      staffToken,
			role:       "staff",
			method:     "GET",
			path:       "/stats/dashboard",
			wantStatus: http.StatusOK,
		},
		{
			name:       "admin access stats dashboard",
			token:      adminToken,
			role:       "admin",
			method:     "GET",
			path:       "/stats/dashboard",
			wantStatus: http.StatusOK,
		},
		{
			name:       "admin can update any media",
			token:      adminToken,
			role:       "admin",
			method:     "PUT",
			path:       "/media/1",
			body:       map[string]string{"title": "Admin Updated"},
			wantStatus: 0, // Should not be 403
		},
		{
			name:       "staff can update any media",
			token:      staffToken,
			role:       "staff",
			method:     "PUT",
			path:       "/media/1",
			body:       map[string]string{"title": "Staff Updated"},
			wantStatus: 0, // Should not be 403
		},
	}

	for _, tt := range tests {
		t.Run(tt.role+"_"+tt.name, func(t *testing.T) {
			resp, _, err := ts.MakeRequest(RequestOptions{
				Method: tt.method,
				Path:   tt.path,
				Body:   tt.body,
				Token:  tt.token,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			if tt.wantStatus == 0 {
				// Just check it's not 403
				if resp.Code == http.StatusForbidden {
					t.Errorf("%s was forbidden from %s", tt.role, tt.path)
				}
			} else if resp.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, resp.Code)
			}
		})
	}
}

// TestPermissionCrossUser tests that users cannot modify other users' resources
func TestPermissionCrossUser(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	user1Token := ts.GetToken(RoleUser)
	// user2Token would be for user2 - we need to access it differently

	// Try to perform operations on resources that belong to user2 (or admin)
	// These should be forbidden (403) or not found (404) for user1
	restrictedOperations := []struct {
		name   string
		method string
		path   string
		body   interface{}
	}{
		{"delete other user's playlist", "DELETE", "/playlists/1", nil},
		{"update other user's channel", "PUT", "/channels/1", map[string]string{"title": "Hacked"}},
		{"delete other user's comment", "DELETE", "/comments/1", nil},
	}

	for _, op := range restrictedOperations {
		t.Run(op.name, func(t *testing.T) {
			resp, _, err := ts.MakeRequest(RequestOptions{
				Method: op.method,
				Path:   op.path,
				Body:   op.body,
				Token:  user1Token,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			// Should be 403 Forbidden or 404 Not Found
			if resp.Code != http.StatusForbidden &&
				resp.Code != http.StatusNotFound {
				t.Errorf("expected 403 or 404, got %d for %s", resp.Code, op.name)
			}
		})
	}
}

// TestTokenExpired tests behavior with expired/invalid tokens
func TestTokenExpired(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{
			name:       "completely invalid token",
			token:      "this.is.not.valid",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "empty token string",
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "malformed token",
			token:      "bearer-token-without-bearer-prefix",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _, err := ts.MakeRequest(RequestOptions{
				Method: "GET",
				Path:   "/auth/me",
				Token:  tt.token,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			AssertStatus(t, resp, tt.wantStatus)
		})
	}
}

// TestPermissionRoleHierarchy tests role hierarchy (admin > staff > user > guest)
func TestPermissionRoleHierarchy(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	roles := []struct {
		name  string
		token string
		level int // Higher = more permissions
	}{
		{"guest", "", 0},
		{"user", ts.GetToken(RoleUser), 1},
		{"staff", ts.GetToken(RoleStaff), 2},
		{"admin", ts.GetToken(RoleAdmin), 3},
	}

	// Test access to protected resource
	for _, role := range roles {
		t.Run(role.name+"_can_access_me", func(t *testing.T) {
			resp, _, err := ts.MakeRequest(RequestOptions{
				Method: "GET",
				Path:   "/auth/me",
				Token:  role.token,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			if role.level == 0 {
				// Guest should be 401
				AssertStatus(t, resp, http.StatusUnauthorized)
			} else {
				// All authenticated users should get 200
				AssertStatus(t, resp, http.StatusOK)
			}
		})
	}
}
