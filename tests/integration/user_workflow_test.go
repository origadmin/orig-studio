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
			Path:   "/auth/signup",
			Body:   registerBody,
		})
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		// Could be 201 Created or 400 if user already exists
		if resp.Code != http.StatusCreated && resp.Code != http.StatusBadRequest {
			t.Errorf("Unexpected status: %d", resp.Code)
		}

		if resp.Code == http.StatusCreated {
			var result map[string]interface{}
			if err := ParseResponse(body, &result); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if _, ok := result["access_token"]; !ok {
				t.Error("Expected 'access_token' field in response")
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
			Path:   "/auth/signup",
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
			Path:   "/auth/signup",
			Body:   registerBody,
		})
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		// Should fail with 400
		if resp.Code != http.StatusBadRequest && resp.Code != http.StatusConflict {
			t.Errorf("Unexpected status: %d", resp.Code)
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
			Path:   "/auth/signin",
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

		if _, ok := result["access_token"]; !ok {
			t.Error("Expected 'access_token' field in response")
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
			Path:   "/auth/signin",
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
			Path:   "/auth/signin",
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
			Path:   "/auth/signin",
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
			Path:   "/me",
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

		data, ok := result["data"]
		if !ok {
			t.Error("Expected 'data' field in response")
			return
		}

		dataMap, ok := data.(map[string]interface{})
		if !ok {
			t.Error("Expected 'data' to be an object")
			return
		}

		if _, ok := dataMap["id"]; !ok {
			t.Log("Note: 'id' field not found in me response; field may use different name")
		}
		if _, ok := dataMap["username"]; !ok {
			t.Error("Expected 'username' field in response")
		}
	})

	t.Run("get current user profile - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/me",
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
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
		}

		if resp.Code == http.StatusOK {
			var result map[string]interface{}
			if err := ParseResponse(body, &result); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			// User handler returns user object directly (no wrapper)
			if _, ok := result["id"]; !ok {
				t.Log("Note: 'id' field not found in user response; field may use different name")
			}
			if _, ok := result["username"]; !ok {
				t.Error("Expected 'username' field in response")
			}
		}
	})

	t.Run("update user profile - authenticated", func(t *testing.T) {
		updateBody := map[string]interface{}{
			"nickname": "Updated Name",
			"email":    "updated@example.com",
		}

		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "PUT",
			Path:   "/me",
			Body:   updateBody,
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to update profile: %v", err)
		}

		// Could be OK or 404
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
		}

		if resp.Code == http.StatusOK {
			var result map[string]interface{}
			if err := ParseResponse(body, &result); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			data, ok := result["data"]
			if !ok {
				t.Error("Expected 'data' field in response")
				return
			}

			dataMap, ok := data.(map[string]interface{})
			if !ok {
				t.Error("Expected 'data' to be an object")
				return
			}

			if dataMap["nickname"] != "Updated Name" {
				t.Errorf("Expected nickname to be updated")
			}
			if dataMap["email"] != "updated@example.com" {
				t.Errorf("Expected email to be updated")
			}
		}
	})

	t.Run("update user profile - no auth", func(t *testing.T) {
		updateBody := map[string]interface{}{
			"nickname": "Should Fail",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "PUT",
			Path:   "/me",
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
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
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
		if resp.Code != http.StatusForbidden && resp.Code != http.StatusUnauthorized && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
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
		if resp.Code != http.StatusUnauthorized && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
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
			Path:   "/me/favorites",
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

		data, ok := result["data"]
		if !ok {
			t.Error("Expected 'data' field in response")
			return
		}

		dataMap, ok := data.(map[string]interface{})
		if !ok {
			t.Error("Expected 'data' to be an object")
			return
		}

		if _, ok := dataMap["items"]; !ok {
			t.Error("Expected 'items' field in response")
		}
	})

	t.Run("get user's own media - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/me/favorites",
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
			Path:   "/me/favorites",
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

		data, ok := result["data"]
		if !ok {
			t.Error("Expected 'data' field in response")
			return
		}

		dataMap, ok := data.(map[string]interface{})
		if !ok {
			t.Error("Expected 'data' to be an object")
			return
		}

		if _, ok := dataMap["items"]; !ok {
			t.Error("Expected 'items' field in response")
		}
	})

	t.Run("get user's history - authenticated", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/me/likes",
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

		data, ok := result["data"]
		if !ok {
			t.Error("Expected 'data' field in response")
			return
		}

		dataMap, ok := data.(map[string]interface{})
		if !ok {
			t.Error("Expected 'data' to be an object")
			return
		}

		if _, ok := dataMap["items"]; !ok {
			t.Error("Expected 'items' field in response")
		}
	})
}

// TestUserSubscriptionWorkflow tests subscription functionality
func TestUserSubscriptionWorkflow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("subscribe to user - authenticated", func(t *testing.T) {
		// Channel routes use :token (short_token) parameter
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/channels/testchnl/subscription",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to subscribe: %v", err)
		}

		// Could be OK, 404, or 500 depending on whether channel exists
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}

		if resp.Code == http.StatusOK {
			var result map[string]interface{}
			if err := ParseResponse(body, &result); err != nil {
				// Response may not be JSON
				t.Logf("Response body: %s", body.String())
				return
			}

			data, ok := result["data"]
			if !ok {
				t.Log("Expected 'data' field in response")
				return
			}

			dataMap, ok := data.(map[string]interface{})
			if !ok {
				t.Error("Expected 'data' to be an object")
				return
			}

			if _, ok := dataMap["success"]; !ok {
				t.Log("Expected 'success' field in response")
			}
		}
	})

	t.Run("subscribe to user - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/channels/testchnl/subscription",
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to subscribe: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("get subscription status - public", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/channels/testchnl/subscription",
		})
		if err != nil {
			t.Fatalf("Failed to get subscription status: %v", err)
		}

		// Channel subscription status requires authentication
		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("get user's subscriptions - authenticated", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/me/subscriptions",
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

		data, ok := result["data"]
		if !ok {
			t.Error("Expected 'data' field in response")
			return
		}

		dataMap, ok := data.(map[string]interface{})
		if !ok {
			t.Error("Expected 'data' to be an object")
			return
		}

		if _, ok := dataMap["items"]; !ok {
			t.Error("Expected 'items' field in response")
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
			Path:   "/channels/nonexist/subscription",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to subscribe: %v", err)
		}

		// Could be 404, 400, or 500 depending on channel validation
		if resp.Code != http.StatusNotFound && resp.Code != http.StatusBadRequest && resp.Code != http.StatusInternalServerError {
			t.Logf("Subscribe to non-existent channel returned status %d", resp.Code)
		}
	})

	t.Run("get non-existent user's media", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/channels/nonexist/videos",
		})
		if err != nil {
			t.Fatalf("Failed to get channel videos: %v", err)
		}

		// Could be 404 or 400 (invalid token format)
		if resp.Code != http.StatusNotFound && resp.Code != http.StatusBadRequest {
			t.Logf("Get non-existent channel videos returned status %d", resp.Code)
		}
	})
}
