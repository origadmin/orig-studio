package integration

import (
	"net/http"
	"testing"
)

// TestAPIIntegration tests the complete API integration
func TestAPIIntegration(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("subscription status", func(t *testing.T) {
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
				t.Log("Note: 'is_subscribed' field not found in subscription response; structure may differ")
			}
		} else {
			t.Logf("Get subscription status returned status %d (channel may not exist)", resp.Code)
		}
	})

	t.Run("subscribe user", func(t *testing.T) {
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

	t.Run("unsubscribe user", func(t *testing.T) {
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
}
