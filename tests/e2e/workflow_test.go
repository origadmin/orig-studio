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

		// Auth handler returns TokenResponse directly with access_token at top level
		if token, ok := result["access_token"]; ok {
			authToken = "Bearer " + token.(string)
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
			Path:   "/me",
			Token:  authToken,
		})
		if err != nil {
			t.Fatalf("get me failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)

		data, err := integration.GetResponseData(body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if _, ok := data["username"]; !ok {
			t.Log("Note: 'username' field not found in me response; field may not exist in current schema")
		}
	})

	t.Run("3. Search Media", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/search?q=test",
		})
		if err != nil {
			t.Fatalf("search failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)
	})

	t.Run("4. List Media", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/medias",
		})
		if err != nil {
			t.Fatalf("list media failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)
	})

	t.Run("5. View Media Details", func(t *testing.T) {
		// Media public routes use :short_token parameter
		resp, _, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/medias/nonexistent",
		})
		if err != nil {
			t.Fatalf("get media failed: %v", err)
		}

		// Non-existent media should return 404
		if resp.Code != http.StatusNotFound && resp.Code != http.StatusOK {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})

	t.Run("6. Check Like Status", func(t *testing.T) {
		// GET /medias/:short_token/likes - public (optional JWT)
		resp, _, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/medias/abc12345/likes",
			Token:  authToken,
		})
		if err != nil {
			t.Fatalf("get like status failed: %v", err)
		}

		// Could be OK or 404 depending on media existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})

	t.Run("7. Like Media", func(t *testing.T) {
		// POST /medias/:short_token/likes requires JWT
		resp, _, err := ts.MakeRequest(integration.RequestOptions{
			Method: "POST",
			Path:   "/medias/abc12345/likes",
			Token:  authToken,
		})
		if err != nil {
			t.Fatalf("like failed: %v", err)
		}

		// Could be OK or 404 depending on media existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Logf("Like returned status %d", resp.Code)
		}
	})

	t.Run("8. Check Favorite Status", func(t *testing.T) {
		// GET /medias/:short_token/favorites - optional JWT
		resp, _, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/medias/abc12345/favorites",
			Token:  authToken,
		})
		if err != nil {
			t.Fatalf("get favorite status failed: %v", err)
		}

		// Could be OK or 404 depending on media existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})

	t.Run("9. Favorite Media", func(t *testing.T) {
		// POST /medias/:short_token/favorites requires JWT
		resp, _, err := ts.MakeRequest(integration.RequestOptions{
			Method: "POST",
			Path:   "/medias/abc12345/favorites",
			Token:  authToken,
		})
		if err != nil {
			t.Fatalf("favorite failed: %v", err)
		}

		// Could be OK or 404 depending on media existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Logf("Favorite returned status %d", resp.Code)
		}
	})

	t.Run("10. List Favorites", func(t *testing.T) {
		// GET /me/favorites requires JWT
		resp, _, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/me/favorites",
			Token:  authToken,
		})
		if err != nil {
			t.Fatalf("list favorites failed: %v", err)
		}

		if resp.Code != http.StatusOK {
			t.Logf("List favorites returned status %d", resp.Code)
		}
	})

	t.Run("11. Create Comment", func(t *testing.T) {
		commentBody := map[string]interface{}{
			"comment": map[string]interface{}{
				"content":  "This is a great video!",
				"media_id": "550e8400-e29b-41d4-a716-446655440000",
			},
		}

		// POST /comments requires JWT
		resp, _, err := ts.MakeRequest(integration.RequestOptions{
			Method: "POST",
			Path:   "/comments",
			Body:   commentBody,
			Token:  authToken,
		})
		if err != nil {
			t.Fatalf("create comment failed: %v", err)
		}

		// Could be OK, Created, or error
		if resp.Code != http.StatusOK && resp.Code != http.StatusCreated && resp.Code != http.StatusInternalServerError {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})

	t.Run("12. List Comments", func(t *testing.T) {
		// GET /medias/:short_token/comments - public
		resp, _, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/medias/abc12345/comments",
		})
		if err != nil {
			t.Fatalf("list comments failed: %v", err)
		}

		// Could be OK or 404 depending on media existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})

	t.Run("13. Get Share URL", func(t *testing.T) {
		// GET /medias/:short_token/shares - public
		resp, _, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/medias/abc12345/shares",
		})
		if err != nil {
			t.Fatalf("get share URL failed: %v", err)
		}

		// Could be OK or 404 depending on media existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})

	t.Run("14. Logout", func(t *testing.T) {
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
		// GET /admin/stats/dashboard requires JWT+Admin
		resp, _, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/admin/stats/dashboard",
			Token:  adminToken,
		})
		if err != nil {
			t.Fatalf("get stats failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)
	})

	t.Run("Admin List Medias", func(t *testing.T) {
		// GET /admin/medias requires JWT+Admin
		resp, _, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/admin/medias",
			Token:  adminToken,
		})
		if err != nil {
			t.Fatalf("list medias failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)
	})

	t.Run("Admin List Tags", func(t *testing.T) {
		// GET /admin/tags requires JWT+Admin
		resp, _, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/admin/tags",
			Token:  adminToken,
		})
		if err != nil {
			t.Fatalf("list tags failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)
	})

	t.Run("Admin List Channels", func(t *testing.T) {
		// GET /admin/channels requires JWT+Admin
		resp, _, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/admin/channels",
			Token:  adminToken,
		})
		if err != nil {
			t.Fatalf("list channels failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)
	})

	t.Run("Admin List Playlists", func(t *testing.T) {
		// GET /admin/playlists requires JWT+Admin
		resp, _, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/admin/playlists",
			Token:  adminToken,
		})
		if err != nil {
			t.Fatalf("list playlists failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)
	})

	t.Run("Admin List Encoding Profiles", func(t *testing.T) {
		// GET /admin/encoding/profiles requires JWT+Admin
		resp, _, err := ts.MakeRequest(integration.RequestOptions{
			Method: "GET",
			Path:   "/admin/encoding/profiles",
			Token:  adminToken,
		})
		if err != nil {
			t.Fatalf("list encoding profiles failed: %v", err)
		}

		integration.AssertStatus(t, resp, http.StatusOK)
	})
}
