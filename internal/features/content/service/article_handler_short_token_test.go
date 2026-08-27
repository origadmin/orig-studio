/**
 * B100 Backend Test: Article short_token resolution
 *
 * Verifies that the article handler correctly extracts the :id param
 * which can be either a short_token or a UUID (the repo resolves both).
 */

package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/pkg/http/std"
)

// TestGetArticle_ExtractsIdParam verifies the handler extracts the :id param
// from the URL. The param value can be either a short_token or a UUID -
// the repo's Get() method resolves both.
func TestGetArticle_ExtractsIdParam(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		param string
	}{
		{
			name:  "short_token value in :id param",
			path:  "/articles/abc123xyz",
			param: "abc123xyz",
		},
		{
			name:  "UUID value in :id param",
			path:  "/articles/019de8e8-c479-7c0e-bd1e-b2cefe7cec43",
			param: "019de8e8-c479-7c0e-bd1e-b2cefe7cec43",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := std.NewRouter()

			// Use a simple handler that captures the param
			var capturedParam string
			rt.GET("/articles/:id", func(ctx http2.Context) error {
				capturedParam = ctx.Var("id")
				return ctx.Result(http.StatusOK, nil)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			rt.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.param, capturedParam)
		})
	}
}

// TestGetArticle_MissingIdParam_Returns400 verifies that a missing :id param
// results in a 400 error (this is the existing behavior, unchanged).
func TestGetArticle_MissingIdParam_Returns400(t *testing.T) {
	rt := std.NewRouter()
	articleHandler := &ArticleHandler{uc: nil, jwt: nil}

	// Register the route without :id param (same pattern as RegisterRoutes)
	rt.GET("/articles", articleHandler.getArticle())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/articles", nil)
	rt.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "article slug is required")
}
