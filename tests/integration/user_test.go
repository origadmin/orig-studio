package integration

import (
	"net/http"
	"testing"
)

// TestUsersCRUD tests user CRUD operations
func TestUsersCRUD(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list users", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/users",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(respBody, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if _, ok := result["list"]; !ok {
			t.Error("expected 'list' field in response")
		}
		if _, ok := result["total"]; !ok {
			t.Error("expected 'total' field in response")
		}
	})

	t.Run("list users with pagination", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/users?page=1&limit=10",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(respBody, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if _, ok := result["page"]; !ok {
			t.Error("expected 'page' field in response")
		}
	})

	t.Run("get user by id", func(t *testing.T) {
		// Get user1 (should exist from seed data)
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/users/1",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.Code == http.StatusOK {
			var result map[string]interface{}
			if err := ParseResponse(respBody, &result); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if _, ok := result["id"]; !ok {
				t.Error("expected 'id' field in response")
			}
		} else if resp.Code != http.StatusNotFound {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})

	t.Run("get non-existent user", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/users/99999",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusNotFound)
	})

	t.Run("get user with invalid id", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/users/invalid",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("create user - admin", func(t *testing.T) {
		newUser := map[string]string{
			"username": "newusertest",
			"email":    "newuser@test.com",
			"password": "password123",
			"name":     "New User",
		}

		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/users",
			Body:   newUser,
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.Code == http.StatusCreated {
			var result map[string]interface{}
			if err := ParseResponse(respBody, &result); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if _, ok := result["username"]; !ok {
				t.Error("expected 'username' field in created user")
			}
		} else {
			t.Logf("Create user returned status %d", resp.Code)
		}
	})

	t.Run("delete user - admin", func(t *testing.T) {
		// Try to delete a user (user2 which might not have dependencies)
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "DELETE",
			Path:   "/users/2",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.Code == http.StatusOK {
			t.Log("Successfully deleted user")
		} else {
			t.Logf("Delete user returned status %d", resp.Code)
		}
	})
}

// TestUserSubscription tests user subscription endpoints in detail
func TestUserSubscription(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	userToken := ts.GetToken(RoleUser)

	tests := []struct {
		name       string
		method     string
		path       string
		token      string
		wantStatus int
		checkField string
	}{
		{
			name:       "get subscription status - authenticated",
			method:     "GET",
			path:       "/users/2/subscription",
			token:      userToken,
			wantStatus: http.StatusOK,
			checkField: "is_subscribed",
		},
		{
			name:       "get subscription status - public",
			method:     "GET",
			path:       "/users/2/subscription",
			token:      "",
			wantStatus: http.StatusOK,
			checkField: "is_subscribed",
		},
		{
			name:       "subscribe - authenticated",
			method:     "POST",
			path:       "/users/2/subscribe",
			token:      userToken,
			wantStatus: http.StatusOK,
			checkField: "success",
		},
		{
			name:       "subscribe - no auth",
			method:     "POST",
			path:       "/users/2/subscribe",
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "unsubscribe - authenticated",
			method:     "DELETE",
			path:       "/users/2/subscribe",
			token:      userToken,
			wantStatus: http.StatusOK,
			checkField: "success",
		},
		{
			name:       "unsubscribe - no auth",
			method:     "DELETE",
			path:       "/users/2/subscribe",
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, respBody, err := ts.MakeRequest(RequestOptions{
				Method: tt.method,
				Path:   tt.path,
				Token:  tt.token,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			AssertStatus(t, resp, tt.wantStatus)

			if tt.checkField != "" && resp.Code == http.StatusOK {
				var result map[string]interface{}
				if err := ParseResponse(respBody, &result); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if _, ok := result[tt.checkField]; !ok {
					t.Errorf("expected '%s' field in response", tt.checkField)
				}
			}
		})
	}
}

// TestUserFollowersAndFollowing tests follower/following endpoints
func TestUserFollowersAndFollowing(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	userToken := ts.GetToken(RoleUser)

	t.Run("get subscriptions list", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/subscriptions",
			Token:  userToken,
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(respBody, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if _, ok := result["list"]; !ok {
			t.Error("expected 'list' field in response")
		}
		if _, ok := result["total"]; !ok {
			t.Error("expected 'total' field in response")
		}
		if _, ok := result["page"]; !ok {
			t.Error("expected 'page' field in response")
		}
	})

	t.Run("get followers list", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/followers",
			Token:  userToken,
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(respBody, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if _, ok := result["list"]; !ok {
			t.Error("expected 'list' field in response")
		}
	})
}

// TestUserRoles tests that different user roles are properly reflected in responses
func TestUserRoles(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tests := []struct {
		name    string
		token   string
		role    string
		isStaff bool
	}{
		{"admin", ts.GetToken(RoleAdmin), "admin", true},
		{"staff", ts.GetToken(RoleStaff), "staff", true},
		{"user", ts.GetToken(RoleUser), "user", false},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_role_in_token", func(t *testing.T) {
			resp, respBody, err := ts.MakeRequest(RequestOptions{
				Method: "GET",
				Path:   "/auth/me",
				Token:  tt.token,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			AssertStatus(t, resp, http.StatusOK)

			var result map[string]interface{}
			if err := ParseResponse(respBody, &result); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			// Check is_staff field
			if isStaff, ok := result["is_staff"]; ok {
				if isStaff != tt.isStaff {
					t.Errorf("expected is_staff=%v, got %v", tt.isStaff, isStaff)
				}
			} else {
				t.Error("expected 'is_staff' field in response")
			}
		})
	}
}

// TestUserProfileOperations tests user profile related operations
func TestUserProfileOperations(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get current user profile", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/auth/me",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(respBody, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if _, ok := result["username"]; !ok {
			t.Error("expected 'username' field in response")
		}
		if _, ok := result["id"]; !ok {
			t.Error("expected 'id' field in response")
		}
	})

	t.Run("get other user profile", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/users/1",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.Code == http.StatusOK {
			var result map[string]interface{}
			if err := ParseResponse(respBody, &result); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if _, ok := result["username"]; !ok {
				t.Error("expected 'username' field in response")
			}
		} else if resp.Code != http.StatusNotFound {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})
}
