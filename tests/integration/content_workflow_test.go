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
		// Get comments for media ID 1
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/media/1/comments",
		})
		if err != nil {
			t.Fatalf("Failed to get comments: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := result["list"]; !ok {
			t.Error("Expected 'list' field in response")
		}
		if _, ok := result["total"]; !ok {
			t.Error("Expected 'total' field in response")
		}
	})

	t.Run("create comment - authenticated", func(t *testing.T) {
		commentBody := map[string]interface{}{
			"body": "This is a test comment",
		}

		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/media/1/comments",
			Body:   commentBody,
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to create comment: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := result["id"]; !ok {
			t.Error("Expected 'id' field in response")
		}
	})

	t.Run("create comment - no auth", func(t *testing.T) {
		commentBody := map[string]interface{}{
			"body": "This should fail",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/media/1/comments",
			Body:   commentBody,
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to create comment: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("reply to comment - authenticated", func(t *testing.T) {
		// This test assumes comment ID 1 exists
		replyBody := map[string]interface{}{
			"body": "This is a reply",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/media/1/comments/1/replies",
			Body:   replyBody,
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to reply to comment: %v", err)
		}

		// Could be OK or 404 if comment doesn't exist
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}
	})
}

// TestLikeDislikeWorkflow tests the like/dislike functionality
func TestLikeDislikeWorkflow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("like media - authenticated", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/media/1/like",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to like media: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := result["is_liked"]; !ok {
			t.Error("Expected 'is_liked' field in response")
		}
		if _, ok := result["like_count"]; !ok {
			t.Error("Expected 'like_count' field in response")
		}
	})

	t.Run("dislike media - authenticated", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/media/1/dislike",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to dislike media: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := result["is_disliked"]; !ok {
			t.Error("Expected 'is_disliked' field in response")
		}
	})

	t.Run("get like status - public", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/media/1/like",
		})
		if err != nil {
			t.Fatalf("Failed to get like status: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := result["like_count"]; !ok {
			t.Error("Expected 'like_count' field in response")
		}
	})

	t.Run("like media - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/media/1/like",
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
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/media/1/favorite",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to favorite media: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := result["is_favorited"]; !ok {
			t.Error("Expected 'is_favorited' field in response")
		}
	})

	t.Run("get favorite status - public", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/media/1/favorite",
		})
		if err != nil {
			t.Fatalf("Failed to get favorite status: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)
	})

	t.Run("favorite media - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/media/1/favorite",
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to favorite media: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("get user favorites - authenticated", func(t *testing.T) {
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
}

// TestPlaylistWorkflow tests the playlist functionality
func TestPlaylistWorkflow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("create playlist - authenticated", func(t *testing.T) {
		playlistBody := map[string]interface{}{
			"title":       "Test Playlist",
			"description": "This is a test playlist",
		}

		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/playlists",
			Body:   playlistBody,
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to create playlist: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := result["id"]; !ok {
			t.Error("Expected 'id' field in response")
		}
	})

	t.Run("get playlists - public", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/playlists",
		})
		if err != nil {
			t.Fatalf("Failed to get playlists: %v", err)
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

	t.Run("get single playlist", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/playlists/1",
		})
		if err != nil {
			t.Fatalf("Failed to get playlist: %v", err)
		}

		// Could be OK or 404
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}
	})

	t.Run("add media to playlist - authenticated", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/playlists/1/items",
			Body: map[string]interface{}{
				"media_id": 1,
			},
			Token: ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to add to playlist: %v", err)
		}

		// Could be OK or 404
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}
	})

	t.Run("create playlist - no auth", func(t *testing.T) {
		playlistBody := map[string]interface{}{
			"title": "Should Fail",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/playlists",
			Body:   playlistBody,
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to create playlist: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})
}

// TestShareWorkflow tests the share functionality
func TestShareWorkflow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get share URL - public", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/media/1/share",
		})
		if err != nil {
			t.Fatalf("Failed to get share URL: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := result["url"]; !ok {
			t.Error("Expected 'url' field in response")
		}
	})

	t.Run("record share - authenticated", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/media/1/share",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to record share: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := result["success"]; !ok {
			t.Error("Expected 'success' field in response")
		}
	})

	t.Run("record share - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/media/1/share",
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

		var result map[string]interface{}
		if err := ParseResponse(body, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := result["list"]; !ok {
			t.Error("Expected 'list' field in response")
		}
	})

	t.Run("list tags - public", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/tags",
		})
		if err != nil {
			t.Fatalf("Failed to list tags: %v", err)
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

	t.Run("get single category", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/categories/1",
		})
		if err != nil {
			t.Fatalf("Failed to get category: %v", err)
		}

		// Could be OK or 404
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.StatusCode)
		}
	})

	t.Run("get media by category", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/categories/1/media",
		})
		if err != nil {
			t.Fatalf("Failed to get media by category: %v", err)
		}

		// Could be OK or 404
		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			if err := ParseResponse(body, &result); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if _, ok := result["list"]; !ok {
				t.Error("Expected 'list' field in response")
			}
		}
	})

	t.Run("get media by tag", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/tags/1/media",
		})
		if err != nil {
			t.Fatalf("Failed to get media by tag: %v", err)
		}

		// Could be OK or 404
		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			if err := ParseResponse(body, &result); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if _, ok := result["list"]; !ok {
				t.Error("Expected 'list' field in response")
			}
		}
	})
}

// TestContentInteractionErrorScenarios tests various error scenarios in content interactions
func TestContentInteractionErrorScenarios(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get comments for non-existent media", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/media/99999/comments",
		})
		if err != nil {
			t.Fatalf("Failed to get comments: %v", err)
		}

		AssertStatus(t, resp, http.StatusNotFound)
	})

	t.Run("like non-existent media", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/media/99999/like",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to like media: %v", err)
		}

		AssertStatus(t, resp, http.StatusNotFound)
	})

	t.Run("get non-existent playlist", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/playlists/99999",
		})
		if err != nil {
			t.Fatalf("Failed to get playlist: %v", err)
		}

		AssertStatus(t, resp, http.StatusNotFound)
	})

	t.Run("get non-existent category", func(t *testing.T) {
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
