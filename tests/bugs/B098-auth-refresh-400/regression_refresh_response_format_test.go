package bugs

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"origadmin/application/origstudio/tests/integration"
)

// TestB096_AuthEndpointsUseUnifiedResponseFormat verifies that ALL auth
// endpoints (signin, signup, refresh) use the server.OK()/server.Created()
// unified response format {"code":0,"message":"ok","data":{...}} per C016.
//
// Bug B096: Login/RegisterUser used c.JSON() directly, returning flat
// TokenResponse without the {code, message, data} envelope. This violated
// the C016 unified API response convention. Meanwhile, the frontend
// attemptRefresh() used raw axios (no interceptor) and did not unwrap the
// code/data envelope, causing setAuth() to receive the wrong shape.
//
// Fix: Backend uses server.OK()/server.Created() for all auth endpoints.
// Frontend attemptRefresh/interceptor manually unwrap ApiResponse<Token>.
func TestB096_AuthEndpointsUseUnifiedResponseFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	gin.SetMode(gin.TestMode)

	ts := integration.SetupTestServer(t)
	defer ts.Cleanup()

	// Step 1: Login to get tokens
	loginResp, loginBody, err := ts.MakeRequest(integration.RequestOptions{
		Method: "POST",
		Path:   "/auth/signin",
		Body:   map[string]string{"username": "admin", "password": "admin123"},
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, loginResp.Code)

	// Parse login response - must be in unified format {code, message, data}
	var loginResult map[string]interface{}
	require.NoError(t, json.Unmarshal(loginBody.Bytes(), &loginResult))

	// Login MUST use server.OK() format: {code:0, message:"ok", data:{access_token:...}}
	assert.Equal(t, float64(0), loginResult["code"], "login response must have code=0 (server.OK format)")
	assert.Equal(t, "ok", loginResult["message"], "login response must have message='ok'")

	loginData, ok := loginResult["data"].(map[string]interface{})
	require.True(t, ok, "login response must have 'data' field as object (server.OK format)")

	accessToken, hasAccessToken := loginData["access_token"]
	assert.True(t, hasAccessToken, "login data must have 'access_token'")
	assert.NotEmpty(t, accessToken)

	refreshToken, hasRefreshToken := loginData["refresh_token"]
	assert.True(t, hasRefreshToken, "login data must have 'refresh_token'")
	assert.NotEmpty(t, refreshToken)

	// Step 2: Call /auth/refresh with the refresh_token
	refreshResp, refreshBody, err := ts.MakeRequest(integration.RequestOptions{
		Method: "POST",
		Path:   "/auth/refresh",
		Body:   map[string]string{"refresh_token": refreshToken.(string)},
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, refreshResp.Code)

	// Parse refresh response - must also be in unified format
	var refreshResult map[string]interface{}
	require.NoError(t, json.Unmarshal(refreshBody.Bytes(), &refreshResult))

	// Refresh MUST use server.OK() format: {code:0, message:"ok", data:{access_token:...}}
	assert.Equal(t, float64(0), refreshResult["code"], "refresh response must have code=0 (server.OK format)")
	assert.Equal(t, "ok", refreshResult["message"], "refresh response must have message='ok'")

	refreshData, ok := refreshResult["data"].(map[string]interface{})
	require.True(t, ok, "refresh response must have 'data' field as object (server.OK format)")

	newAccessToken, hasNewAccessToken := refreshData["access_token"]
	assert.True(t, hasNewAccessToken, "refresh data must have 'access_token'")
	assert.NotEmpty(t, newAccessToken)

	newRefreshToken, hasNewRefreshToken := refreshData["refresh_token"]
	assert.True(t, hasNewRefreshToken, "refresh data must have 'refresh_token'")
	assert.NotEmpty(t, newRefreshToken)

	// Both data objects must have token_type and expires_in
	assert.Contains(t, loginData, "token_type", "login data must have 'token_type'")
	assert.Contains(t, loginData, "expires_in", "login data must have 'expires_in'")
	assert.Contains(t, refreshData, "token_type", "refresh data must have 'token_type'")
	assert.Contains(t, refreshData, "expires_in", "refresh data must have 'expires_in'")
}

// TestB096_LoginAndRefreshResponseFormatsMatch verifies that /auth/signin and
// /auth/refresh return the same response structure (both server.OK() wrapped).
func TestB096_LoginAndRefreshResponseFormatsMatch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	gin.SetMode(gin.TestMode)

	ts := integration.SetupTestServer(t)
	defer ts.Cleanup()

	// Login
	loginResp, loginBody, err := ts.MakeRequest(integration.RequestOptions{
		Method: "POST",
		Path:   "/auth/signin",
		Body:   map[string]string{"username": "admin", "password": "admin123"},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, loginResp.Code)

	var loginResult map[string]interface{}
	require.NoError(t, json.Unmarshal(loginBody.Bytes(), &loginResult))

	loginData := loginResult["data"].(map[string]interface{})
	refreshToken, _ := loginData["refresh_token"].(string)
	require.NotEmpty(t, refreshToken)

	// Refresh
	refreshResp, refreshBody, err := ts.MakeRequest(integration.RequestOptions{
		Method: "POST",
		Path:   "/auth/refresh",
		Body:   map[string]string{"refresh_token": refreshToken},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, refreshResp.Code)

	var refreshResult map[string]interface{}
	require.NoError(t, json.Unmarshal(refreshBody.Bytes(), &refreshResult))

	refreshData := refreshResult["data"].(map[string]interface{})

	// Both responses must have the same top-level envelope keys
	envelopeKeys := []string{"code", "message", "data"}
	for _, key := range envelopeKeys {
		assert.Contains(t, loginResult, key, "login response must have '%s' at top level", key)
		assert.Contains(t, refreshResult, key, "refresh response must have '%s' at top level", key)
	}

	// Both data objects must have the same essential auth fields
	essentialKeys := []string{"access_token", "refresh_token", "token_type", "expires_in"}
	for _, key := range essentialKeys {
		assert.Contains(t, loginData, key, "login data must have '%s'", key)
		assert.Contains(t, refreshData, key, "refresh data must have '%s'", key)
	}
}

// TestB096_RefreshTokenMissingFieldReturns400 verifies that /auth/refresh
// returns 400 when refresh_token is missing from the request body.
func TestB096_RefreshTokenMissingFieldReturns400(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	gin.SetMode(gin.TestMode)

	ts := integration.SetupTestServer(t)
	defer ts.Cleanup()

	// Call /auth/refresh without refresh_token
	resp, _, err := ts.MakeRequest(integration.RequestOptions{
		Method: "POST",
		Path:   "/auth/refresh",
		Body:   map[string]string{},
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

// TestB096_RefreshTokenInvalidTokenReturns401 verifies that /auth/refresh
// returns 401 when the refresh_token is invalid.
func TestB096_RefreshTokenInvalidTokenReturns401(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	gin.SetMode(gin.TestMode)

	ts := integration.SetupTestServer(t)
	defer ts.Cleanup()

	// Call /auth/refresh with an invalid refresh_token
	resp, _, err := ts.MakeRequest(integration.RequestOptions{
		Method: "POST",
		Path:   "/auth/refresh",
		Body:   map[string]string{"refresh_token": "invalid-token-string"},
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}
