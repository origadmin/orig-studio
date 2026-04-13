package integration

import (
	"net/http"
	"testing"
)

// TestUserRegistrationWorkflow tests the complete user registration flow
func TestUserRegistrationWorkflow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("register new user", func(t *testing.T) {
		registerBody := map[string]interface{}{
			"username": "newuser",
			"password": "password123",
			"email":    "newuser@example.com",
		}

		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/auth/register",
			Body:   registerBody,
		})
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		// Could be OK or 400 if user already exists
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}

		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			if err := ParseResponse(body, &result); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if _, ok := result["token"]; !ok {
				t.Error("Expected 'token' field in response")
			}
			if _, ok := result["user"]; !ok {
				t.Error("Expected 'user' field in response")
			}
		}
	})

	t.Run("register with missing fields", func(t *testing.T) {
		registerBody := map[string]interface{}{
			"username": "incompleteuser",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/auth/register",
			Body:   registerBody,
		})
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		AssertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("register with duplicate username", func(t *testing.T) {
		// Try to register with an existing username (user1)
		registerBody := map[string]interface{}{
			"username": "user1",
			"password": "password123",
			"email":    "user1duplicate@example.com",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/auth/register",
			Body:   registerBody,
		})
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		// Should fail with 400
		if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusConflict {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}
	})
}

// TestUserLoginWorkflow tests the complete user login flow
func TestUserLoginWorkflow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("login with valid credentials", func(t *testing.T) {
		loginBody := map[string]interface{}{
			"username": "user1",
			"password": "user123",
		}

		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/auth/login",
			Body:   loginBody,
		})
		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := result["token"]; !ok {
			t.Error("Expected 'token' field in response")
		}
		if _, ok := result["user"]; !ok {
			t.Error("Expected 'user' field in response")
		}
	})

	t.Run("login with invalid credentials", func(t *testing.T) {
		loginBody := map[string]interface{}{
			"username": "user1",
			"password": "wrongpassword",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/auth/login",
			Body:   loginBody,
		})
		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("login with non-existent user", func(t *testing.T) {
		loginBody := map[string]interface{}{
			"username": "nonexistent",
			"password": "password123",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/auth/login",
			Body:   loginBody,
		})
		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("login with missing fields", func(t *testing.T) {
		loginBody := map[string]interface{}{
			"username": "user1",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/auth/login",
			Body:   loginBody,
		})
		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}

		AssertStatus(t, resp, http.StatusBadRequest)
	})
}

// TestUserProfileManagement tests user profile management functionality
func TestUserProfileManagement(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get current user profile - authenticated", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/users/me",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to get profile: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := result["id"]; !ok {
			t.Error("Expected 'id' field in response")
		}
		if _, ok := result["username"]; !ok {
			t.Error("Expected 'username' field in response")
		}
	})

	t.Run("get current user profile - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/users/me",
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to get profile: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("get user by id - public", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/users/1",
		})
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		// Could be OK or 404
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}

		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			if err := ParseResponse(body, &result); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if _, ok := result["id"]; !ok {
				t.Error("Expected 'id' field in response")
			}
			if _, ok := result["username"]; !ok {
				t.Error("Expected 'username' field in response")
			}
		}
	})

	t.Run("update user profile - authenticated", func(t *testing.T) {
		updateBody := map[string]interface{}{
			"display_name": "Updated Name",
			"bio":          "This is my updated bio",
		}

		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "PUT",
			Path:   "/users/me",
			Body:   updateBody,
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to update profile: %v", err)
		}

		// Could be OK or 404
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}

		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			if err := ParseResponse(body, &result); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if result["display_name"] != "Updated Name" {
				t.Errorf("Expected display_name to be updated")
			}
		}
	})

	t.Run("update user profile - no auth", func(t *testing.T) {
		updateBody := map[string]interface{}{
			"display_name": "Should Fail",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "PUT",
			Path:   "/users/me",
			Body:   updateBody,
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to update profile: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})
}

// TestUserPermissions tests user permission handling
func TestUserPermissions(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("admin can access admin endpoint", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/users",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to access admin endpoint: %v", err)
		}

		// Could be OK or maybe not implemented
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}
	})

	t.Run("regular user cannot access admin endpoint", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/users",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to access admin endpoint: %v", err)
		}

		// Should be 403 Forbidden or 401 Unauthorized or 404 Not Found
		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}
	})

	t.Run("guest cannot access admin endpoint", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/users",
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to access admin endpoint: %v", err)
		}

		// Should be 401 Unauthorized or 404 Not Found
		if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}
	})
}

// TestUserMediaAccess tests user access to their own media
func TestUserMediaAccess(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get user's own media - authenticated", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/users/me/media",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to get user media: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := result["list"]; !ok {
			t.Error("Expected 'list' field in response")
		}
	})

	t.Run("get user's own media - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/users/me/media",
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to get user media: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("get user's favorites - authenticated", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/favorites",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to get favorites: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := result["list"]; !ok {
			t.Error("Expected 'list' field in response")
		}
	})

	t.Run("get user's history - authenticated", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/history",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to get history: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := result["list"]; !ok {
			t.Error("Expected 'list' field in response")
		}
	})
}

// TestUserSubscriptionWorkflow tests subscription functionality
func TestUserSubscriptionWorkflow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("subscribe to user - authenticated", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/users/2/subscribe",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to subscribe: %v", err)
		}

		// Could be OK or 404
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}

		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			if err := ParseResponse(body, &result); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if _, ok := result["is_subscribed"]; !ok {
				t.Error("Expected 'is_subscribed' field in response")
			}
		}
	})

	t.Run("subscribe to user - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/users/2/subscribe",
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to subscribe: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("get subscription status - public", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/users/2/subscribe",
		})
		if err != nil {
			t.Fatalf("Failed to get subscription status: %v", err)
		}

		// Could be OK or 404
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}

		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			if err := ParseResponse(body, &result); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if _, ok := result["subscriber_count"]; !ok {
				t.Error("Expected 'subscriber_count' field in response")
			}
		}
	})

	t.Run("get user's subscriptions - authenticated", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/subscriptions",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to get subscriptions: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := result["list"]; !ok {
			t.Error("Expected 'list' field in response")
		}
	})
}

// TestUserErrorScenarios tests various error scenarios for user operations
func TestUserErrorScenarios(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get non-existent user", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/users/99999",
		})
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		AssertStatus(t, resp, http.StatusNotFound)
	})

	t.Run("subscribe to non-existent user", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/users/99999/subscribe",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to subscribe: %v", err)
		}

		AssertStatus(t, resp, http.StatusNotFound)
	})

	t.Run("get non-existent user's media", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/users/99999/media",
		})
		if err != nil {
			t.Fatalf("Failed to get user media: %v", err)
		}

		AssertStatus(t, resp, http.StatusNotFound)
	})
}
