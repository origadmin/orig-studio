package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	userv1 "origadmin/application/origstudio/api/gen/v1/user"
	ginpkgadapter "origadmin/application/origstudio/internal/pkg/http/gin"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	stdeadapter "origadmin/application/origstudio/internal/pkg/http/std"
	"origadmin/application/origstudio/internal/pkg/http/validate"
	"origadmin/application/origstudio/internal/server"
)

// ---------------------------------------------------------------------------
// The phase-1 parity test suite does NOT exercise the real UserService. It
// only validates that AuthHandler's BindJSON step produces identical 400
// responses for both the std (EE / test) and the gin (CE) adapter when given
// proto messages with protoc-gen-validate rules.
//
// To avoid wiring the full UserService graph, we build a minimal standalone
// handler with the same BindJSON / FailCtx pattern as the real AuthHandler
// and assert on its HTTP output.
// ---------------------------------------------------------------------------

type stubLoginHandler struct{}

// ServeLogin mirrors AuthHandler.Login's exact flow:
//
//	var req userv1.LoginRequest
//	if err := ctx.BindJSON(&req); err != nil { return server.FailCtx(ErrBadRequest, err.Error()) }
//	return server.OKCtx(...)
func (s *stubLoginHandler) ServeLogin(ctx http2.Context) error {
	var req userv1.LoginRequest
	if err := ctx.BindJSON(&req); err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
	}
	return server.OKCtx(ctx, &userv1.LoginResponse{AccessToken: "ok", RefreshToken: "r", ExpiresAt: 0})
}

func newStdLoginEnv() http.Handler {
	h := &stubLoginHandler{}
	r := stdeadapter.NewRouter()
	auth := r.Group("/auth")
	{
		auth.POST("/signin", h.ServeLogin)
	}
	return r
}

func newGinLoginEnv() http.Handler {
	h := &stubLoginHandler{}
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	rg := engine.Group("/")
	router := ginpkgadapter.NewRouterAdapter(rg)
	auth := router.Group("/auth")
	{
		auth.POST("/signin", h.ServeLogin)
	}
	return engine
}

func postLoginJSON(t *testing.T, router http.Handler, body string) (*httptest.ResponseRecorder, map[string]interface{}) {
	t.Helper()
	var reader *bytes.Buffer
	if body == "" {
		reader = bytes.NewBufferString("{}")
	} else {
		reader = bytes.NewBufferString(body)
	}
	req := httptest.NewRequest(http.MethodPost, "/auth/signin", reader)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp map[string]interface{}
	if w.Body.Len() > 0 {
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Logf("non-JSON body: %s", w.Body.String())
		}
	}
	return w, resp
}

// expectErrBadRequest asserts that the response equals server.ErrBadRequest
// with a message that mentions each required field name (in any order/case).
func expectErrBadRequest(t *testing.T, w *httptest.ResponseRecorder, resp map[string]interface{}, fields ...string) {
	t.Helper()
	assert.Equal(t, http.StatusBadRequest, w.Code, "expected 400, got %d: %s", w.Code, w.Body.String())
	if resp == nil {
		t.Fatalf("no JSON body: %s", w.Body.String())
	}
	// server.Fail writes Code = HTTP status (not the business error code), see
	// Fail/FailCtx source: c.JSON(getHTTPStatus(code), ErrorResponse{Code: httpStatus, ...})
	code, _ := resp["code"].(float64)
	assert.EqualValues(t, http.StatusBadRequest, int(code),
		"expected JSON code=HTTP 400, got body=%s", w.Body.String())
	reason, _ := resp["reason"].(string)
	assert.Equal(t, "BAD_REQUEST", reason, "expected reason=BAD_REQUEST, got body=%s", w.Body.String())
	msg, _ := resp["message"].(string)
	lower := strings.ToLower(msg)
	for _, f := range fields {
		assert.True(t, strings.Contains(lower, strings.ToLower(f)),
			"message must mention %q, got: %s", f, msg)
	}
}

// ---------------------------------------------------------------------------
// Parity tests
// ---------------------------------------------------------------------------

func TestLoginValidation_EmptyBody_StdAndGinReturn400(t *testing.T) {
	cases := []struct {
		name   string
		router http.Handler
	}{
		{"std (EE)", newStdLoginEnv()},
		{"gin (CE)", newGinLoginEnv()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w, resp := postLoginJSON(t, tc.router, "{}")
			expectErrBadRequest(t, w, resp, "username", "password")
		})
	}
}

func TestLoginValidation_PasswordTooShort_StdAndGinReturn400(t *testing.T) {
	cases := []struct {
		name   string
		router http.Handler
	}{
		{"std (EE)", newStdLoginEnv()},
		{"gin (CE)", newGinLoginEnv()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w, resp := postLoginJSON(t, tc.router, `{"username":"alice","password":"12345"}`)
			expectErrBadRequest(t, w, resp, "password")
			msg, _ := resp["message"].(string)
			// Username passes min_len 1 → message should NOT still mention it
			assert.False(t, strings.Contains(strings.ToLower(msg), "username"),
				"username is valid, should not appear in error, got: %s", msg)
		})
	}
}

func TestLoginValidation_ValidPayload_StdAndGinReturn200(t *testing.T) {
	cases := []struct {
		name   string
		router http.Handler
	}{
		{"std (EE)", newStdLoginEnv()},
		{"gin (CE)", newGinLoginEnv()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w, resp := postLoginJSON(t, tc.router, `{"username":"alice","password":"secret123"}`)
			assert.Equal(t, http.StatusOK, w.Code, "expected 200, got %d: %s", w.Code, w.Body.String())
			tok, _ := resp["access_token"].(string)
			assert.Equal(t, "ok", tok, "expected stub success body, got: %s", w.Body.String())
		})
	}
}

func TestLoginValidation_InvalidJSON_StdAndGinReturn400(t *testing.T) {
	cases := []struct {
		name   string
		router http.Handler
	}{
		{"std (EE)", newStdLoginEnv()},
		{"gin (CE)", newGinLoginEnv()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/auth/signin", strings.NewReader("{not json"))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			tc.router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code,
				"invalid JSON should become 400, got %d: %s", w.Code, w.Body.String())
		})
	}
}

// validate package integration check: the same LoginRequest used here should
// always produce a validate.ErrValidation for empty input. Failures here
// mean protoc-gen-validate didn't regenerate after proto edits.
func TestValidatePackage_LoginRequestEmptyProducesErrValidation(t *testing.T) {
	var req userv1.LoginRequest
	err := validate.Validate(&req)
	if !assert.Error(t, err) {
		t.Fatal("LoginRequest with empty username+password should fail validation")
	}
	assert.True(t, validate.IsErrValidation(err),
		"expected *validate.ErrValidation, got %T: %v", err, err)
}
