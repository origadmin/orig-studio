package e2e

import (
	"net/http"
	"testing"

	"origadmin/application/origcms/tests/integration"
)

func TestCompleteUserWorkflow(t *testing.T) {
	ts := integration.SetupTestServer(t)
	defer ts.Cleanup()

	var authToken string

	t.Run("1. User Signup", func(t *testing.T) {
		signupBody := map[string]string{
			"username": "workflowuser",
			"password": "workflow123",
			"email":    "workflow@example.com",
			"nickname": "Workflow User",
		}

		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "POST",
			Path:   "/auth/signup",
			Body:   signupBody,
		})
		if err != nil {
			t.Fatalf("signup failed: %v", err)
		}

		if resp.Code != http.StatusCreated && resp.Code != http.StatusOK {
			t.Fatalf("signup returned %d, expected 201/200", resp.Code)
		}

		var result map[string]interface{}
		if err := integration.ParseResponse(body, &result); err != nil {
			t.Fatalf("failed to parse signup response: %v", err)
		}

		if token, ok := result["access_token"]; ok {
			authToken = token.(string)
		} else {
			t.Fatal("no access_token in signup response")
		}
	})

	if authToken == "" {
		t.Fatal("cannot continue workflow without auth token")
	}

	t.Run("2. Get Current User", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/auth/me",
			Token:  authToken,
		})
		if err != nil {
			t.Fatalf("get me failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := integration.ParseResponse(body, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if result["username"] != "workflowuser" {
			t.Errorf("expected username 'workflowuser', got %v", result["username"])
		}
	})

	t.Run("3. Browse Feed", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/feed",
		})
		if err != nil {
			t.Fatalf("get feed failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := integration.ParseResponse(body, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if _, ok := result["sections"]; !ok {
			t.Error("expected 'sections' in feed response")
		}
	})

	t.Run("4. Search Media", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/search?q=test",
		})
		if err != nil {
			t.Fatalf("search failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := integration.ParseResponse(body, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if _, ok := result["list"]; !ok {
			t.Error("expected 'list' in search response")
		}
	})

	t.Run("5. View Media Details", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/media/1",
		})
		if err != nil {
			t.Fatalf("get media failed: %v", err)
		}

		if resp.Code == http.StatusOK {
			var result map[string]interface{}
			if err := integration.ParseResponse(body, &result); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if _, ok := result["id"]; !ok {
				t.Error("expected 'id' in media response")
			}
		} else if resp.Code != http.StatusNotFound {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})

	t.Run("6. Check Like Status", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/media/1/like",
			Token:  authToken,
		})
		if err != nil {
			t.Fatalf("get like status failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := integration.ParseResponse(body, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if _, ok := result["is_liked"]; !ok {
			t.Error("expected 'is_liked' in response")
		}
	})

	t.Run("7. Like Media", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "POST",
			Path:   "/media/1/like",
			Token:  authToken,
		})
		if err != nil {
			t.Fatalf("like failed: %v", err)
		}

		if resp.Code == http.StatusOK {
			var result map[string]interface{}
			if err := integration.ParseResponse(body, &result); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if _, ok := result["is_liked"]; !ok {
				t.Error("expected 'is_liked' in response")
			}
		} else {
			t.Logf("Like returned status %d", resp.Code)
		}
	})

	t.Run("8. Check Favorite Status", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/media/1/favorite",
			Token:  authToken,
		})
		if err != nil {
			t.Fatalf("get favorite status failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := integration.ParseResponse(body, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if _, ok := result["is_favorited"]; !ok {
			t.Error("expected 'is_favorited' in response")
		}
	})

	t.Run("9. Favorite Media", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "POST",
			Path:   "/media/1/favorite",
			Token:  authToken,
		})
		if err != nil {
			t.Fatalf("favorite failed: %v", err)
		}

		if resp.Code == http.StatusOK {
			var result map[string]interface{}
			if err := integration.ParseResponse(body, &result); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if _, ok := result["is_favorited"]; !ok {
				t.Error("expected 'is_favorited' in response")
			}
		} else {
			t.Logf("Favorite returned status %d", resp.Code)
		}
	})

	t.Run("10. List Favorites", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/favorites",
			Token:  authToken,
		})
		if err != nil {
			t.Fatalf("list favorites failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := integration.ParseResponse(body, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if _, ok := result["list"]; !ok {
			t.Error("expected 'list' in favorites response")
		}
	})

	t.Run("11. Create Comment", func(t *testing.T) {
		commentBody := map[string]interface{}{
			"text":     "This is a great video!",
			"media_id": 1,
		}

		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "POST",
			Path:   "/comments",
			Body:   commentBody,
			Token:  authToken,
		})
		if err != nil {
			t.Fatalf("create comment failed: %v", err)
		}

		if resp.Code == http.StatusCreated || resp.Code == http.StatusOK {
			var result map[string]interface{}
			if err := integration.ParseResponse(body, &result); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if result["text"] != "This is a great video!" {
				t.Errorf("expected comment text, got %v", result["text"])
			}
		} else {
			t.Logf("Create comment returned status %d", resp.Code)
		}
	})

	t.Run("12. List Comments", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/comments?media_id=1",
		})
		if err != nil {
			t.Fatalf("list comments failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := integration.ParseResponse(body, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if _, ok := result["list"]; !ok {
			t.Error("expected 'list' in comments response")
		}
	})

	t.Run("13. Create Playlist", func(t *testing.T) {
		playlistBody := map[string]interface{}{
			"name":        "My Favorites",
			"description": "My favorite videos",
			"is_public":   true,
		}

		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "POST",
			Path:   "/playlists",
			Body:   playlistBody,
			Token:  authToken,
		})
		if err != nil {
			t.Fatalf("create playlist failed: %v", err)
		}

		if resp.Code == http.StatusCreated || resp.Code == http.StatusOK {
			var result map[string]interface{}
			if err := integration.ParseResponse(body, &result); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if _, ok := result["name"]; !ok {
				t.Error("expected 'name' in playlist response")
			}
		} else {
			t.Logf("Create playlist returned status %d", resp.Code)
		}
	})

	t.Run("14. Subscribe to User", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "POST",
			Path:   "/users/2/subscribe",
			Token:  authToken,
		})
		if err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}

		if resp.Code == http.StatusOK {
			var result map[string]interface{}
			if err := integration.ParseResponse(body, &result); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if _, ok := result["success"]; !ok {
				t.Error("expected 'success' in response")
			}
		} else {
			t.Logf("Subscribe returned status %d", resp.Code)
		}
	})

	t.Run("15. Get Share URL", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/media/1/share",
		})
		if err != nil {
			t.Fatalf("get share URL failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := integration.ParseResponse(body, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if _, ok := result["url"]; !ok {
			t.Error("expected 'url' in share response")
		}
	})

	t.Run("16. Logout", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "POST",
			Path:   "/auth/signout",
			Token:  authToken,
		})
		if err != nil {
			t.Fatalf("logout failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)
		integration.AssertJSON(t, body, map[string]interface{}{"message": "logged out"})
	})
}

func TestAdminWorkflow(t *testing.T) {
	ts := integration.SetupTestServer(t)
	defer ts.Cleanup()

	adminToken := ts.GetToken(integration.RoleAdmin)

	t.Run("Admin View Dashboard Stats", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/stats/dashboard",
			Token:  adminToken,
		})
		if err != nil {
			t.Fatalf("get stats failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := integration.ParseResponse(body, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		fields := []string{"total_media", "total_users", "total_views"}
		for _, field := range fields {
			if _, ok := result[field]; !ok {
				t.Errorf("expected '%s' in stats response", field)
			}
		}
	})

	t.Run("Admin Create Category", func(t *testing.T) {
		categoryBody := map[string]interface{}{
			"name":        "New Category",
			"description": "Category created by admin",
		}

		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "POST",
			Path:   "/categories",
			Body:   categoryBody,
			Token:  adminToken,
		})
		if err != nil {
			t.Fatalf("create category failed: %v", err)
		}

		if resp.Code == http.StatusCreated || resp.Code == http.StatusOK {
			var result map[string]interface{}
			if err := integration.ParseResponse(body, &result); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if _, ok := result["name"]; !ok {
				t.Error("expected 'name' in category response")
			}
		} else {
			t.Logf("Create category returned status %d", resp.Code)
		}
	})

	t.Run("Admin List All Users", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/users",
			Token:  adminToken,
		})
		if err != nil {
			t.Fatalf("list users failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := integration.ParseResponse(body, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if _, ok := result["list"]; !ok {
			t.Error("expected 'list' in users response")
		}
		if _, ok := result["total"]; !ok {
			t.Error("expected 'total' in users response")
		}
	})
}

func TestContentCreatorWorkflow(t *testing.T) {
	ts := integration.SetupTestServer(t)
	defer ts.Cleanup()

	userToken := ts.GetToken(integration.RoleUser)

	t.Run("Create Channel", func(t *testing.T) {
		channelBody := map[string]interface{}{
			"title":          "My Awesome Channel",
			"description":    "Best content ever",
			"banner_logo":    "https://example.com/banner.jpg",
			"friendly_token": "awesomechannel",
		}

		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "POST",
			Path:   "/channels",
			Body:   channelBody,
			Token:  userToken,
		})
		if err != nil {
			t.Fatalf("create channel failed: %v", err)
		}

		if resp.Code == http.StatusCreated || resp.Code == http.StatusOK {
			var result map[string]interface{}
			if err := integration.ParseResponse(body, &result); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if _, ok := result["title"]; !ok {
				t.Error("expected 'title' in channel response")
			}
		} else {
			t.Logf("Create channel returned status %d", resp.Code)
		}
	})

	t.Run("List User Channels", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/channels/user/1",
		})
		if err != nil {
			t.Fatalf("list channels failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := integration.ParseResponse(body, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if _, ok := result["list"]; !ok {
			t.Error("expected 'list' in channels response")
		}
	})
}
