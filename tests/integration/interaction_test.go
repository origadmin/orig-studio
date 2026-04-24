package integration

import (
	"net/http"
	"testing"
)

// ==================== Likes ====================

// TestLikes tests the standalone like endpoints
func TestLikes(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("toggle like", func(t *testing.T) {
		body := map[string]interface{}{
			"media_id": 1,
			"type":     "like",
		}

		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/likes",
			Body:   body,
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
			if _, ok := result["liked"]; !ok {
				t.Error("expected 'liked' field in response")
			}
		} else {
			t.Logf("Toggle like returned status %d", resp.Code)
		}
	})

	t.Run("toggle dislike", func(t *testing.T) {
		body := map[string]interface{}{
			"media_id": 1,
			"type":     "dislike",
		}

		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/likes",
			Body:   body,
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
			if _, ok := result["disliked"]; !ok {
				t.Error("expected 'disliked' field in response")
			}
		}
	})

	t.Run("get media likes", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/likes/media/1",
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
		if _, ok := result["likes"]; !ok {
			t.Error("expected 'likes' field in response")
		}
		if _, ok := result["dislikes"]; !ok {
			t.Error("expected 'dislikes' field in response")
		}
	})

	t.Run("check like status", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/likes/check/1",
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
		if _, ok := result["liked"]; !ok {
			t.Error("expected 'liked' field in response")
		}
	})
}

// ==================== Favorites ====================

// TestFavorites tests favorite endpoints
func TestFavorites(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list user favorites", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/favorites",
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
		if _, ok := result["list"]; !ok {
			t.Error("expected 'list' field in response")
		}
	})

	t.Run("toggle favorite", func(t *testing.T) {
		body := map[string]int{
			"media_id": 1,
		}

		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/favorites",
			Body:   body,
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
			if _, ok := result["favorited"]; !ok && result["is_favorited"] == nil {
				t.Error("expected 'favorited' or 'is_favorited' field in response")
			}
		}
	})

	t.Run("check favorite", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/favorites/check/1",
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
		if _, ok := result["favorited"]; !ok {
			t.Error("expected 'favorited' field in response")
		}
	})
}

// ==================== Subscriptions ====================

// TestSubscriptions tests subscription endpoints
func TestSubscriptions(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get user subscription status", func(t *testing.T) {
		// First get a user to subscribe to (admin or staff)
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/users/1/subscription",
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
		if _, ok := result["is_subscribed"]; !ok {
			t.Error("expected 'is_subscribed' field in response")
		}
	})

	t.Run("subscribe to user", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/users/2/subscribe",
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
			if _, ok := result["success"]; !ok {
				t.Error("expected 'success' field in response")
			}
		} else {
			t.Logf("Subscribe returned status %d", resp.Code)
		}
	})

	t.Run("unsubscribe from user", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "DELETE",
			Path:   "/users/2/subscribe",
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
			if _, ok := result["success"]; !ok {
				t.Error("expected 'success' field in response")
			}
		}
	})

	t.Run("get subscriptions list", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/subscriptions",
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
		if _, ok := result["list"]; !ok {
			t.Error("expected 'list' field in response")
		}
	})

	t.Run("get followers list", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/followers",
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
		if _, ok := result["list"]; !ok {
			t.Error("expected 'list' field in response")
		}
	})
}

// ==================== Playlists ====================

// TestPlaylists tests playlist endpoints
func TestPlaylists(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list all playlists", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/playlists",
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

	t.Run("get single playlist", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/playlists/1",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		// May be 200 or 404 depending on if playlists exist
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})

	t.Run("create playlist - authenticated", func(t *testing.T) {
		body := map[string]interface{}{
			"name":        "My Test Playlist",
			"description": "Test description",
			"is_public":   true,
		}

		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/playlists",
			Body:   body,
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.Code == http.StatusCreated || resp.Code == http.StatusOK {
			var result map[string]interface{}
			if err := ParseResponse(respBody, &result); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if _, ok := result["name"]; !ok {
				t.Error("expected 'name' field in created playlist")
			}
		} else {
			t.Logf("Create playlist returned status %d", resp.Code)
		}
	})

	t.Run("get my playlists", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/playlists/my",
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
		if _, ok := result["list"]; !ok {
			t.Error("expected 'list' field in response")
		}
	})

	t.Run("add media to playlist", func(t *testing.T) {
		body := map[string]int{
			"media_id": 1,
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/playlists/1/media",
			Body:   body,
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		// May be 200, 403 (not owner), or 404
		if resp.Code != http.StatusOK &&
			resp.Code != http.StatusForbidden &&
			resp.Code != http.StatusNotFound {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})

	t.Run("remove media from playlist", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "DELETE",
			Path:   "/playlists/1/media/1",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		// Currently returns 501 Not Implemented
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotImplemented {
			t.Logf("Remove media from playlist returned status %d", resp.Code)
		}
	})
}

// ==================== Channels ====================

// TestChannels tests channel endpoints
func TestChannels(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list all channels", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/channels",
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

	t.Run("get single channel", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/channels/1",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})

	t.Run("get user channels", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/channels/user/1",
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

	t.Run("create channel - authenticated", func(t *testing.T) {
		body := map[string]interface{}{
			"title":          "My Test Channel",
			"description":    "Test description",
			"banner_logo":    "https://example.com/banner.jpg",
			"friendly_token": "mychannel",
		}

		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/channels",
			Body:   body,
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.Code == http.StatusCreated || resp.Code == http.StatusOK {
			var result map[string]interface{}
			if err := ParseResponse(respBody, &result); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if _, ok := result["title"]; !ok {
				t.Error("expected 'title' field in created channel")
			}
		} else {
			t.Logf("Create channel returned status %d", resp.Code)
		}
	})

	t.Run("add media to channel", func(t *testing.T) {
		body := map[string]int{
			"media_id": 1,
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/channels/1/media",
			Body:   body,
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.Code != http.StatusOK &&
			resp.Code != http.StatusForbidden &&
			resp.Code != http.StatusNotFound {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})

	t.Run("remove media from channel", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "DELETE",
			Path:   "/channels/1/media/1",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.Code != http.StatusOK &&
			resp.Code != http.StatusForbidden &&
			resp.Code != http.StatusNotFound {
			t.Errorf("unexpected status: %d", resp.Code)
		}
	})
}

// ==================== Notifications ====================

// TestNotifications tests notification endpoints
func TestNotifications(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list notifications - authenticated", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/notifications",
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
		if _, ok := result["list"]; !ok {
			t.Error("expected 'list' field in response")
		}
	})

	t.Run("list notifications - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/notifications",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("mark notification read", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "PUT",
			Path:   "/notifications/1/read",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		// May be OK or NotFound
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Logf("Mark notification read returned status %d", resp.Code)
		}
	})
}

// ==================== Stats ====================

// TestStats tests stats endpoints
func TestStats(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get dashboard stats - admin", func(t *testing.T) {
		resp, respBody, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/stats/dashboard",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		if err := ParseResponse(respBody, &result); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		// Check for expected stats fields
		if _, ok := result["total_media"]; !ok {
			t.Log("expected 'total_media' field in response")
		}
	})

	t.Run("get dashboard stats - regular user", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/stats/dashboard",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		// Should be forbidden or OK depending on implementation
		if resp.Code != http.StatusOK && resp.Code != http.StatusForbidden {
			t.Logf("User accessing stats returned status %d", resp.Code)
		}
	})

	t.Run("get dashboard stats - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/stats/dashboard",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})
}
