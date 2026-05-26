package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAuthSignin tests the signin endpoint
func TestAuthSignin(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
		wantToken  bool
		checkField string
		checkValue interface{}
	}{
		{
			name:       "valid admin login",
			body:       map[string]string{"username": "admin", "password": "admin123"},
			wantStatus: http.StatusOK,
			wantToken:  true,
		},
		{
			name:       "valid user login",
			body:       map[string]string{"username": "user1", "password": "user123"},
			wantStatus: http.StatusOK,
			wantToken:  true,
		},
		{
			name:       "invalid password",
			body:       map[string]string{"username": "admin", "password": "wrongpassword"},
			wantStatus: http.StatusUnauthorized,
			wantToken:  false,
			checkField: "error",
			checkValue: "invalid credentials",
		},
		{
			name:       "non-existent user",
			body:       map[string]string{"username": "nonexistent", "password": "password"},
			wantStatus: http.StatusUnauthorized,
			wantToken:  false,
		},
		{
			name:       "missing username",
			body:       map[string]string{"password": "admin123"},
			wantStatus: http.StatusBadRequest,
			wantToken:  false,
		},
		{
			name:       "missing password",
			body:       map[string]string{"username": "admin"},
			wantStatus: http.StatusBadRequest,
			wantToken:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body, err := ts.MakeRequest(RequestOptions{
				Method: "POST",
				Path:   "/auth/signin",
				Body:   tt.body,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			AssertStatus(t, resp, tt.wantStatus)

			if tt.wantToken {
				var result map[string]interface{}
				if err := ParseResponse(body, &result); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if _, ok := result["access_token"]; !ok {
					t.Error("expected access_token in response")
				}
				if _, ok := result["user"]; !ok {
					t.Error("expected user in response")
				}
			}

			if tt.checkField != "" {
				AssertJSON(t, body, map[string]interface{}{tt.checkField: tt.checkValue})
			}
		})
	}
}

// TestAuthSignup tests the signup endpoint
func TestAuthSignup(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
		wantToken  bool
		checkField string
	}{
		{
			name:       "valid signup",
			body:       map[string]string{"username": "newuser", "password": "newpass123", "email": "new@example.com"},
			wantStatus: http.StatusCreated,
			wantToken:  true,
		},
		{
			name:       "duplicate username",
			body:       map[string]string{"username": "admin", "password": "password123"},
			wantStatus: http.StatusConflict,
			wantToken:  false,
			checkField: "error",
		},
		{
			name:       "short password",
			body:       map[string]string{"username": "shortpassuser", "password": "123"},
			wantStatus: http.StatusBadRequest,
			wantToken:  false,
		},
		{
			name:       "short username",
			body:       map[string]string{"username": "ab", "password": "password123"},
			wantStatus: http.StatusBadRequest,
			wantToken:  false,
		},
		{
			name:       "invalid email format",
			body:       map[string]string{"username": "bademail", "password": "password123", "email": "not-an-email"},
			wantStatus: http.StatusBadRequest,
			wantToken:  false,
		},
		{
			name:       "missing username",
			body:       map[string]string{"password": "password123"},
			wantStatus: http.StatusBadRequest,
			wantToken:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body, err := ts.MakeRequest(RequestOptions{
				Method: "POST",
				Path:   "/auth/signup",
				Body:   tt.body,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			AssertStatus(t, resp, tt.wantStatus)

			if tt.wantToken {
				var result map[string]interface{}
				if err := ParseResponse(body, &result); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if _, ok := result["access_token"]; !ok {
					t.Error("expected access_token in response")
				}
			}

			if tt.checkField != "" {
				var result map[string]interface{}
				if err := ParseResponse(body, &result); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if _, ok := result[tt.checkField]; !ok {
					t.Errorf("expected %s in error response", tt.checkField)
				}
			}
		})
	}
}

// TestAuthMe tests the /me endpoint (MeHandler, not /auth/me)
func TestAuthMe(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tests := []struct {
		name       string
		token      string
		wantStatus int
		checkUser  bool
	}{
		{
			name:       "valid token - admin",
			token:      ts.GetToken(RoleAdmin),
			wantStatus: http.StatusOK,
			checkUser:  true,
		},
		{
			name:       "valid token - user",
			token:      ts.GetToken(RoleUser),
			wantStatus: http.StatusOK,
			checkUser:  true,
		},
		{
			name:       "no token",
			token:      "",
			wantStatus: http.StatusUnauthorized,
			checkUser:  false,
		},
		{
			name:       "invalid token",
			token:      "invalid-token-string",
			wantStatus: http.StatusUnauthorized,
			checkUser:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body, err := ts.MakeRequest(RequestOptions{
				Method: "GET",
				Path:   "/me",
				Token:  tt.token,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			AssertStatus(t, resp, tt.wantStatus)

			if tt.checkUser {
				data, err := GetResponseData(body)
				if err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if _, ok := data["username"]; !ok && data["id"] == nil {
					t.Error("expected username or id in response data")
				}
			}
		})
	}
}

// TestAuthSignout tests the signout endpoint
func TestAuthSignout(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Signout is stateless, should return success
	resp, body, err := ts.MakeRequest(RequestOptions{
		Method: "POST",
		Path:   "/auth/signout",
		Token:  ts.GetToken(RoleUser),
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	AssertStatus(t, resp, http.StatusOK)
	AssertJSON(t, body, map[string]interface{}{"message": "logged out"})
}

// TestAuthTokenValidation tests JWT token validation scenarios
func TestAuthTokenValidation(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{
			name:       "missing auth header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong auth scheme",
			authHeader: "Basic dXNlcjpwYXNz",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "bearer without token",
			authHeader: "Bearer ",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "malformed token",
			authHeader: "Bearer not.a.valid.token",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", ts.BaseURL+"/me", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rec := httptest.NewRecorder()
			ts.Router.ServeHTTP(rec, req)

			AssertStatus(t, rec, tt.wantStatus)
		})
	}
}

// TestAuthRolePermissions tests different user roles can access auth endpoints
func TestAuthRolePermissions(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	roles := []struct {
		name     string
		token    string
		isStaff  bool
		username string
	}{
		{"admin", ts.GetToken(RoleAdmin), true, "admin"},
		{"editor", ts.GetToken(RoleEditor), false, "editor"},
		{"user", ts.GetToken(RoleUser), false, "user1"},
	}

	for _, role := range roles {
		t.Run(role.name+"_can_access_me", func(t *testing.T) {
			resp, body, err := ts.MakeRequest(RequestOptions{
				Method: "GET",
				Path:   "/me",
				Token:  role.token,
			})
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			AssertStatus(t, resp, http.StatusOK)

			data, err := GetResponseData(body)
			if err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			// Check is_staff field - MeHandler returns {code, message, data: {user object}}
			if isStaff, ok := data["is_staff"]; ok {
				if isStaff != role.isStaff {
					t.Errorf("expected is_staff=%v, got %v", role.isStaff, isStaff)
				}
			}
		})
	}
}
