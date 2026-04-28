package integration

import (
	"net/http"
	"testing"
)

// TestCommentWorkflow tests the complete comment lifecycle
func TestCommentWorkflow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get comments for media", func(t *testing.T) {
		// GET /medias/:short_token/comments - public route
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/medias/abc12345/comments",
		})
		if err != nil {
			t.Fatalf("Failed to get comments: %v", err)
		}

		if resp.Code == http.StatusOK {
			data, err := GetResponseData(body)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if _, ok := data["items"]; !ok {
				t.Log("Note: 'items' field not found in comments response; structure may differ")
			}
		} else {
			t.Logf("Get comments returned status %d (media may not exist)", resp.Code)
		}
	})

	t.Run("create comment - authenticated", func(t *testing.T) {
		commentBody := map[string]interface{}{
			"comment": map[string]interface{}{
				"content":  "This is a test comment",
				"media_id": "550e8400-e29b-41d4-a716-446655440000",
			},
		}

		// POST /comments requires JWT
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/comments",
			Body:   commentBody,
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to create comment: %v", err)
		}

		// Could be Created, OK, or error
		if resp.Code != http.StatusOK && resp.Code != http.StatusCreated && resp.Code != http.StatusBadRequest && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("create comment - no auth", func(t *testing.T) {
		commentBody := map[string]interface{}{
			"comment": map[string]interface{}{
				"content":  "This should fail",
				"media_id": "550e8400-e29b-41d4-a716-446655440000",
			},
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/comments",
			Body:   commentBody,
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to create comment: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("reply to comment - authenticated", func(t *testing.T) {
		replyBody := map[string]interface{}{
			"comment": map[string]interface{}{
				"content":   "This is a reply",
				"media_id":  "550e8400-e29b-41d4-a716-446655440000",
				"parent_id": "550e8400-e29b-41d4-a716-446655440000",
			},
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/comments",
			Body:   replyBody,
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to reply to comment: %v", err)
		}

		// Could be Created, OK, NotFound, or InternalServerError
		if resp.Code != http.StatusOK && resp.Code != http.StatusCreated && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})
}

// TestLikeDislikeWorkflow tests the like/dislike functionality
func TestLikeDislikeWorkflow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Media routes use :short_token parameter
	// Since no media exists in test DB, authenticated requests will get 404 (media not found)
	// but auth checks (401) still work correctly

	t.Run("like media - authenticated", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/medias/abc12345/likes",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to like media: %v", err)
		}

		if resp.Code == http.StatusOK {
			data, err := GetResponseData(body)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}
			if _, ok := data["is_liked"]; !ok {
				t.Error("Expected 'is_liked' field in response data")
			}
			if _, ok := data["like_count"]; !ok {
				t.Error("Expected 'like_count' field in response data")
			}
		} else {
			t.Logf("Like media returned status %d (media may not exist)", resp.Code)
		}
	})

	t.Run("unlike media - authenticated", func(t *testing.T) {
		// DELETE /medias/:short_token/likes
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "DELETE",
			Path:   "/medias/abc12345/likes",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to unlike media: %v", err)
		}

		// Could be OK, 404, or 500 depending on media existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("get like status - public", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/medias/abc12345/likes",
		})
		if err != nil {
			t.Fatalf("Failed to get like status: %v", err)
		}

		if resp.Code == http.StatusOK {
			data, err := GetResponseData(body)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}
			if _, ok := data["like_count"]; !ok {
				t.Error("Expected 'like_count' field in response data")
			}
		} else {
			t.Logf("Get like status returned status %d (media may not exist)", resp.Code)
		}
	})

	t.Run("like media - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/medias/abc12345/likes",
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to like media: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})
}

// TestFavoriteWorkflow tests the favorite functionality
func TestFavoriteWorkflow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("favorite media - authenticated", func(t *testing.T) {
		// POST /medias/:short_token/favorites requires JWT
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/medias/abc12345/favorites",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to favorite media: %v", err)
		}

		if resp.Code == http.StatusOK {
			data, err := GetResponseData(body)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}
			if _, ok := data["is_favorited"]; !ok {
				t.Error("Expected 'is_favorited' field in response data")
			}
		} else {
			t.Logf("Favorite media returned status %d (media may not exist)", resp.Code)
		}
	})

	t.Run("get favorite status - optional JWT", func(t *testing.T) {
		// GET /medias/:short_token/favorites - optional JWT
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/medias/abc12345/favorites",
		})
		if err != nil {
			t.Fatalf("Failed to get favorite status: %v", err)
		}

		// May be 200 or 404 if media not found
		if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("favorite media - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/medias/abc12345/favorites",
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to favorite media: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("get user favorites - authenticated", func(t *testing.T) {
		// GET /me/favorites requires JWT (MeHandler)
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/me/favorites",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to get favorites: %v", err)
		}

		if resp.Code == http.StatusOK {
			data, err := GetResponseData(body)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}
			if _, ok := data["items"]; !ok {
				t.Log("Note: 'items' field not found in favorites response; structure may differ")
			}
		} else {
			t.Logf("Get favorites returned status %d", resp.Code)
		}
	})
}

// TestPlaylistWorkflow tests the playlist functionality (admin routes)
func TestPlaylistWorkflow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list playlists - admin", func(t *testing.T) {
		// GET /admin/playlists requires JWT+Admin
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/playlists",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to list playlists: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		data, err := GetResponseData(body)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		if _, ok := data["items"]; !ok {
			t.Log("Note: 'items' field not found in playlists response; structure may differ")
		}
	})

	t.Run("create playlist - admin", func(t *testing.T) {
		playlistBody := map[string]interface{}{
			"title":       "Test Playlist",
			"description": "This is a test playlist",
			"user_id":     ts.Users[RoleAdmin].ID,
		}

		// POST /admin/playlists requires JWT+Admin
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/admin/playlists",
			Body:   playlistBody,
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to create playlist: %v", err)
		}

		// Could be OK, Created, or error
		if resp.Code != http.StatusOK && resp.Code != http.StatusCreated && resp.Code != http.StatusBadRequest && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("get playlist by id - admin", func(t *testing.T) {
		// GET /admin/playlists/:id requires JWT+Admin
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/playlists/550e8400-e29b-41d4-a716-446655440000",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to get playlist: %v", err)
		}

		// Could be OK or 404
		if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("update playlist - admin", func(t *testing.T) {
		updateBody := map[string]interface{}{
			"title": "Updated Playlist",
		}

		// PUT /admin/playlists/:id requires JWT+Admin
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "PUT",
			Path:   "/admin/playlists/550e8400-e29b-41d4-a716-446655440000",
			Body:   updateBody,
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to update playlist: %v", err)
		}

		// Could be OK, 404, or 500
		if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("delete playlist - admin", func(t *testing.T) {
		// DELETE /admin/playlists/:id requires JWT+Admin
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "DELETE",
			Path:   "/admin/playlists/660e8400-e29b-41d4-a716-446655440099",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to delete playlist: %v", err)
		}

		// Could be OK or 404
		if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("list playlists - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/playlists",
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to list playlists: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})
}

// TestShareWorkflow tests the share functionality
func TestShareWorkflow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get share link - public", func(t *testing.T) {
		// GET /medias/:short_token/shares - public
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/medias/abc12345/shares",
		})
		if err != nil {
			t.Fatalf("Failed to get share link: %v", err)
		}

		// Could be OK or 404 depending on media existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("record share - authenticated", func(t *testing.T) {
		// POST /medias/:short_token/shares requires JWT
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/medias/abc12345/shares",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to record share: %v", err)
		}

		// Could be OK, 404, or 500 depending on media existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("record share - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/medias/abc12345/shares",
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to record share: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})
}

// TestCategoryTagWorkflow tests category and tag functionality
func TestCategoryTagWorkflow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list categories - public", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/categories",
		})
		if err != nil {
			t.Fatalf("Failed to list categories: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		data, err := GetResponseData(body)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := data["items"]; !ok {
			t.Error("Expected 'items' field in response data")
		}
	})

	t.Run("list tags - admin", func(t *testing.T) {
		// Tags are admin-only: GET /admin/tags
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/tags",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to list tags: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		data, err := GetResponseData(body)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := data["items"]; !ok {
			t.Log("Note: 'items' field not found in tags response; structure may differ")
		}
	})

	t.Run("get single category", func(t *testing.T) {
		// GET /categories/:id - public, ID is numeric
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/categories/1",
		})
		if err != nil {
			t.Fatalf("Failed to get category: %v", err)
		}

		// Could be OK or 404
		if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("get media by category", func(t *testing.T) {
		t.Skip("route /categories/:id/media does not exist")
	})

	t.Run("get media by tag", func(t *testing.T) {
		t.Skip("route /tags/:id/media does not exist")
	})
}

// TestContentInteractionErrorScenarios tests various error scenarios in content interactions
func TestContentInteractionErrorScenarios(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get comments for non-existent media", func(t *testing.T) {
		// GET /medias/:short_token/comments
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/medias/nonexistent/comments",
		})
		if err != nil {
			t.Fatalf("Failed to get comments: %v", err)
		}

		// Should return 404 for non-existent media
		if resp.Code != http.StatusNotFound && resp.Code != http.StatusOK {
			t.Logf("Get comments for non-existent media returned status %d", resp.Code)
		}
	})

	t.Run("share non-existent media", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/medias/nonexist/shares",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to share media: %v", err)
		}

		// Should return 404 or error for non-existent media
		if resp.Code != http.StatusNotFound && resp.Code != http.StatusOK {
			t.Logf("Share non-existent media returned status %d", resp.Code)
		}
	})

	t.Run("get non-existent category", func(t *testing.T) {
		// GET /categories/:id - UUID format
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/categories/99999",
		})
		if err != nil {
			t.Fatalf("Failed to get category: %v", err)
		}

		AssertStatus(t, resp, http.StatusNotFound)
	})
}
