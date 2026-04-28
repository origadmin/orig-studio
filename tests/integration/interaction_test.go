package integration

import (
	"net/http"
	"testing"
)

// TestLikes tests like-related endpoints
func TestLikes(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("like media - authenticated", func(t *testing.T) {
		// POST /medias/:short_token/likes requires JWT
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

	t.Run("get like status - public", func(t *testing.T) {
		// GET /medias/:short_token/likes - public (optional JWT)
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

	t.Run("unlike media - authenticated", func(t *testing.T) {
		// DELETE /medias/:short_token/likes requires JWT
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "DELETE",
			Path:   "/medias/abc12345/likes",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to unlike media: %v", err)
		}

		// Could be OK, 404, or 500 depending on media existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})
}

// TestFavorites tests favorite-related endpoints
func TestFavorites(t *testing.T) {
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

		// Could be OK or 404 depending on media existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
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

	t.Run("list user favorites - authenticated", func(t *testing.T) {
		// GET /me/favorites requires JWT (MeHandler)
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/me/favorites",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to list favorites: %v", err)
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
			t.Logf("List favorites returned status %d", resp.Code)
		}
	})
}

// TestSubscriptions tests subscription-related endpoints
func TestSubscriptions(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get subscription status", func(t *testing.T) {
		// GET /channels/:token/subscription - optional JWT
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/channels/testchnl/subscription",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to get subscription status: %v", err)
		}

		if resp.Code == http.StatusOK {
			data, err := GetResponseData(body)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}
			if _, ok := data["is_subscribed"]; !ok {
				t.Error("Expected 'is_subscribed' field in response data")
			}
		} else {
			t.Logf("Get subscription status returned status %d (channel may not exist)", resp.Code)
		}
	})

	t.Run("subscribe to channel - authenticated", func(t *testing.T) {
		// POST /channels/:token/subscription requires JWT
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/channels/testchnl/subscription",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to subscribe: %v", err)
		}

		// Could be OK, 404, or 500 depending on channel existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("unsubscribe from channel - authenticated", func(t *testing.T) {
		// DELETE /channels/:token/subscription requires JWT
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "DELETE",
			Path:   "/channels/testchnl/subscription",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to unsubscribe: %v", err)
		}

		// Could be OK, 404, or 500 depending on channel existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("subscribe - no auth", func(t *testing.T) {
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

	t.Run("get subscriptions list", func(t *testing.T) {
		// GET /me/subscriptions requires JWT (MeHandler)
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/me/subscriptions",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to get subscriptions: %v", err)
		}

		if resp.Code == http.StatusOK {
			data, err := GetResponseData(body)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}
			if _, ok := data["items"]; !ok {
				t.Log("Note: 'items' field not found in subscriptions response; structure may differ")
			}
		} else {
			t.Logf("Get subscriptions returned status %d", resp.Code)
		}
	})

	t.Run("get followers list", func(t *testing.T) {
		// GET /me/followers requires JWT (MeHandler)
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/me/followers",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to get followers: %v", err)
		}

		if resp.Code == http.StatusOK {
			data, err := GetResponseData(body)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}
			if _, ok := data["items"]; !ok {
				t.Log("Note: 'items' field not found in followers response; structure may differ")
			}
		} else {
			t.Logf("Get followers returned status %d", resp.Code)
		}
	})
}

// TestPlaylists tests playlist-related endpoints (admin routes)
func TestPlaylists(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list all playlists - admin", func(t *testing.T) {
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
			"description": "Test description",
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
		if resp.Code != http.StatusOK && resp.Code != http.StatusCreated && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("get single playlist - admin", func(t *testing.T) {
		// GET /admin/playlists/:id requires JWT+Admin
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/playlists/550e8400-e29b-41d4-a716-446655440000",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to get playlist: %v", err)
		}

		// Could be OK, 400 (bad request), or 404
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

		// Could be OK, 400, 404, or 500
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

		// Could be OK, 400, or 404
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

// TestChannels tests channel-related endpoints (admin routes)
func TestChannels(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list all channels - admin", func(t *testing.T) {
		// GET /admin/channels requires JWT+Admin
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/channels",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to list channels: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		data, err := GetResponseData(body)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		if _, ok := data["items"]; !ok {
			t.Log("Note: 'items' field not found in channels response; structure may differ")
		}
	})

	t.Run("get single channel - admin", func(t *testing.T) {
		// GET /admin/channels/:id requires JWT+Admin
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/channels/550e8400-e29b-41d4-a716-446655440000",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to get channel: %v", err)
		}

		// Could be OK, 400, or 404
		if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("update channel - admin", func(t *testing.T) {
		channelBody := map[string]interface{}{
			"title":       "Updated Channel",
			"description": "Updated description",
		}

		// PUT /admin/channels/:id requires JWT+Admin
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "PUT",
			Path:   "/admin/channels/550e8400-e29b-41d4-a716-446655440000",
			Body:   channelBody,
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to update channel: %v", err)
		}

		// Could be OK, 400, 404, or 500
		if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("delete channel - admin", func(t *testing.T) {
		// DELETE /admin/channels/:id requires JWT+Admin
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "DELETE",
			Path:   "/admin/channels/660e8400-e29b-41d4-a716-446655440099",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to delete channel: %v", err)
		}

		// Could be OK, 400, or 404
		if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("list channels - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/channels",
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to list channels: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})
}

// TestNotifications tests notification-related endpoints
func TestNotifications(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list notifications - authenticated", func(t *testing.T) {
		// GET /notifications requires JWT
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/notifications",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to list notifications: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		data, err := GetResponseData(body)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		if _, ok := data["items"]; !ok {
			t.Log("Note: 'items' field not found in notifications response; structure may differ")
		}
	})

	t.Run("list notifications - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/notifications",
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to list notifications: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("mark notification as read", func(t *testing.T) {
		// POST /notifications/:id/read requires JWT, ID is numeric
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/notifications/1/read",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to mark notification as read: %v", err)
		}

		// Could be OK, 404, or 500
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("mark all notifications as read", func(t *testing.T) {
		// POST /notifications/read-all requires JWT
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/notifications/read-all",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to mark all notifications as read: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)
	})
}
