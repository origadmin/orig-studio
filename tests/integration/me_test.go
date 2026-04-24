package integration

import (
	"net/http"
	"testing"
)

// TestMeHandler tests the MeHandler routes
func TestMeHandler(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Test GET /api/v1/me
	resp, body, err := ts.MakeRequest(RequestOptions{
		Method: "GET",
		Path:   "/me",
		Token:  ts.GetToken(RoleUser),
	})
	if err != nil {
		t.Fatalf("Failed to get me: %v", err)
	}

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
		t.Logf("Response body: %s", body.String())
	}
}
