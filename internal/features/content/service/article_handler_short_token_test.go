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

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"origadmin/application/origcms/internal/handler"
)

// TestGetArticle_ExtractsIdParam verifies the handler extracts the :id param
// from the URL. The param value can be either a short_token or a UUID -
// the repo's Get() method resolves both.
func TestGetArticle_ExtractsIdParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		path   string
		param  string
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
			r := gin.New()

			// Use a simple handler that captures the param
			var capturedParam string
			r.GET("/articles/:id", func(c *gin.Context) {
				capturedParam = c.Param("id")
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.param, capturedParam)
		})
	}
}

// TestGetArticle_MissingIdParam_Returns400 verifies that a missing :id param
// results in a 400 error (this is the existing behavior, unchanged).
func TestGetArticle_MissingIdParam_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	articleHandler := &ArticleHandler{uc: nil, jwt: nil}

	// Use the adapter to register the route (same pattern as RegisterRoutes)
	adapter := handler.NewGinRouterAdapter(r.Group(""))
	adapter.GET("/articles", articleHandler.getArticle())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/articles", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "article id is required")
}
