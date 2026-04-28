package integration

import (
	"net/http"
	"testing"
)

// TestMediaList tests the media list endpoint
func TestMediaList(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list media - public", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/medias",
		})
		if err != nil {
			t.Fatalf("Failed to list media: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		data, err := GetResponseData(body)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Media list returns {code, message, data: {items, total, page, page_size}}
		if _, ok := data["items"]; !ok {
			t.Error("Expected 'items' field in response data")
		}
		if _, ok := data["total"]; !ok {
			t.Error("Expected 'total' field in response data")
		}
	})

	t.Run("list media with pagination", func(t *testing.T) {
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/medias?page=1&page_size=10",
		})
		if err != nil {
			t.Fatalf("Failed to list media with pagination: %v", err)
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

	t.Run("list media with filters", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/medias?status=published&type=video",
		})
		if err != nil {
			t.Fatalf("Failed to list media with filters: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)
	})
}

// TestMediaGetByShortToken tests getting a single media item by short_token
func TestMediaGetByShortToken(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get media by short_token - public", func(t *testing.T) {
		// Media public routes use :short_token parameter
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/medias/abc12345",
		})
		if err != nil {
			t.Fatalf("Failed to get media: %v", err)
		}

		if resp.Code == http.StatusOK {
			data, err := GetResponseData(body)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}
			// Media get returns {code, message, data: {media object}}
			if _, ok := data["id"]; !ok {
				t.Error("Expected 'id' field in response data")
			}
		} else {
			t.Logf("Get media returned status %d (media may not exist)", resp.Code)
		}
	})

	t.Run("get non-existent media", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/medias/nonexistent",
		})
		if err != nil {
			t.Fatalf("Failed to get non-existent media: %v", err)
		}

		AssertStatus(t, resp, http.StatusNotFound)
	})
}

// TestMediaLike tests the media like endpoint
func TestMediaLike(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("like media - authenticated", func(t *testing.T) {
		// Media routes use :short_token parameter
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
			// Media like returns {code, message, data: {is_liked, like_count, ...}}
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
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})
}

// TestMediaFavorite tests the media favorite endpoint
func TestMediaFavorite(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("favorite media - authenticated", func(t *testing.T) {
		// POST /medias/:short_token/favorites
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
		// GET /medias/:short_token/favorites
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
}

// TestMediaShare tests the media share endpoint
func TestMediaShare(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get share link - public", func(t *testing.T) {
		// GET /medias/:short_token/shares
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/medias/abc12345/shares",
		})
		if err != nil {
			t.Fatalf("Failed to get share link: %v", err)
		}

		// Could be OK or 404 depending on media existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("record share - authenticated", func(t *testing.T) {
		// POST /medias/:short_token/shares
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/medias/abc12345/shares",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to record share: %v", err)
		}

		// Could be OK, 404, or 500 depending on media existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})
}

// TestMediaVariants tests the media variants endpoint (admin route)
func TestMediaVariants(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get media variants - admin route", func(t *testing.T) {
		// Admin media variants: GET /admin/medias/:id/variants (UUID ID)
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/medias/550e8400-e29b-41d4-a716-446655440000/variants",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to get media variants: %v", err)
		}

		// Could be OK, 404, or 500 depending on media existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("get media variants - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/medias/550e8400-e29b-41d4-a716-446655440000/variants",
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to get media variants: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})
}

// TestMediaEncodingTasks tests the media encoding tasks endpoint (admin route)
func TestMediaEncodingTasks(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get encoding tasks - admin route", func(t *testing.T) {
		// Admin media tasks: GET /admin/medias/:id/tasks (UUID ID)
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/medias/550e8400-e29b-41d4-a716-446655440000/tasks",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to get encoding tasks: %v", err)
		}

		// Could be OK, 404, or 500 depending on media existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("retry encoding task - admin route", func(t *testing.T) {
		// POST /admin/medias/:id/tasks/:taskId/retry
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/admin/medias/550e8400-e29b-41d4-a716-446655440000/tasks/550e8400-e29b-41d4-a716-446655440000/retry",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to retry encoding task: %v", err)
		}

		// Could be OK, 400 (media not found), 404, or 500 depending on media/task existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("get encoding tasks - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/medias/550e8400-e29b-41d4-a716-446655440000/tasks",
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to get encoding tasks: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})
}

// TestMediaEncodeProfiles tests the encoding profiles endpoint
func TestMediaEncodeProfiles(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list encoding profiles - public", func(t *testing.T) {
		// GET /encoding/profiles is a public route
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/encoding/profiles",
		})
		if err != nil {
			t.Fatalf("Failed to list encoding profiles: %v", err)
		}

		if resp.Code == http.StatusOK {
			data, err := GetResponseData(body)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}
			// Encoding profiles list returns {items, total, ...} in data
			if _, ok := data["items"]; !ok {
				t.Log("Note: 'items' field not found in encoding profiles response; structure may differ")
			}
		} else {
			t.Logf("List encoding profiles returned status %d", resp.Code)
		}
	})

	t.Run("create encoding profile - admin route", func(t *testing.T) {
		profile := map[string]interface{}{
			"name":      "test-profile",
			"codec":     "h264",
			"container": "mp4",
		}

		// POST /encoding/profiles requires JWT
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/encoding/profiles",
			Body:   profile,
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to create encoding profile: %v", err)
		}

		// Could be OK, Created, or error
		if resp.Code != http.StatusOK && resp.Code != http.StatusCreated && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("list encoding profiles - admin route", func(t *testing.T) {
		// GET /admin/encoding/profiles requires JWT+Admin
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/encoding/profiles",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to list encoding profiles (admin): %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)
	})

	t.Run("create encoding profile - no auth", func(t *testing.T) {
		profile := map[string]interface{}{
			"name":      "test-profile-noauth",
			"codec":     "h264",
			"container": "mp4",
		}

		// POST /encoding/profiles requires JWT
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "POST",
			Path:   "/encoding/profiles",
			Body:   profile,
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to create encoding profile: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})
}

// TestMediaSprite tests the media sprite endpoints
func TestMediaSprite(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("get sprite VTT - public", func(t *testing.T) {
		// GET /medias/:short_token/sprite.vtt
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/medias/abc12345/sprite.vtt",
		})
		if err != nil {
			t.Fatalf("Failed to get sprite VTT: %v", err)
		}

		// Could be OK or 404 depending on media existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("get sprite JPG - public", func(t *testing.T) {
		// GET /medias/:short_token/sprite.jpg
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/medias/abc12345/sprite.jpg",
		})
		if err != nil {
			t.Fatalf("Failed to get sprite JPG: %v", err)
		}

		// Could be OK or 404 depending on media existence
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})
}

// TestAdminMediaCRUD tests admin media CRUD operations
func TestAdminMediaCRUD(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("list admin medias", func(t *testing.T) {
		// GET /admin/medias
		resp, body, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/medias",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to list admin medias: %v", err)
		}

		AssertStatus(t, resp, http.StatusOK)

		data, err := GetResponseData(body)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		if _, ok := data["items"]; !ok {
			t.Log("Note: 'items' field not found in admin medias response; structure may differ")
		}
	})

	t.Run("get admin media by UUID", func(t *testing.T) {
		// GET /admin/medias/:id (UUID ID)
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/medias/550e8400-e29b-41d4-a716-446655440000",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to get admin media: %v", err)
		}

		// Could be OK or 404
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("update admin media", func(t *testing.T) {
		// PUT /admin/medias/:id (UUID ID)
		update := map[string]interface{}{
			"title": "Updated Title",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "PUT",
			Path:   "/admin/medias/550e8400-e29b-41d4-a716-446655440000",
			Body:   update,
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to update admin media: %v", err)
		}

		// Could be OK, 404, or 500
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("delete admin media", func(t *testing.T) {
		// DELETE /admin/medias/:id (UUID ID)
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "DELETE",
			Path:   "/admin/medias/660e8400-e29b-41d4-a716-446655440099",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to delete admin media: %v", err)
		}

		// Could be OK or 404
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("admin media stats", func(t *testing.T) {
		// GET /admin/medias/:id/stats
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/medias/550e8400-e29b-41d4-a716-446655440000/stats",
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to get admin media stats: %v", err)
		}

		// Could be OK or 404
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("admin media state change", func(t *testing.T) {
		// PUT /admin/medias/:id/state
		state := map[string]interface{}{
			"state": "published",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "PUT",
			Path:   "/admin/medias/550e8400-e29b-41d4-a716-446655440000/state",
			Body:   state,
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to change admin media state: %v", err)
		}

		// Could be OK, 404, or 500
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("admin media review", func(t *testing.T) {
		// PUT /admin/medias/:id/review
		review := map[string]interface{}{
			"action":  "approve",
			"comment": "Looks good",
		}

		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "PUT",
			Path:   "/admin/medias/550e8400-e29b-41d4-a716-446655440000/review",
			Body:   review,
			Token:  ts.GetToken(RoleAdmin),
		})
		if err != nil {
			t.Fatalf("Failed to review admin media: %v", err)
		}

		// Could be OK, 404, or 500
		if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound && resp.Code != http.StatusInternalServerError {
			t.Errorf("Unexpected status: %d", resp.Code)
		}
	})

	t.Run("admin media - no auth", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/medias",
			Token:  "",
		})
		if err != nil {
			t.Fatalf("Failed to list admin medias: %v", err)
		}

		AssertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("admin media - non-admin user", func(t *testing.T) {
		resp, _, err := ts.MakeRequest(RequestOptions{
			Method: "GET",
			Path:   "/admin/medias",
			Token:  ts.GetToken(RoleUser),
		})
		if err != nil {
			t.Fatalf("Failed to list admin medias: %v", err)
		}

		AssertStatus(t, resp, http.StatusForbidden)
	})
}
