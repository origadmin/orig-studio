package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	ginadapter "origadmin/application/origstudio/internal/pkg/http/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func registerHandler(r *gin.Engine, method, path string, h http2.HandlerFunc) {
	adapter := ginadapter.NewRouterAdapter(&r.RouterGroup)
	switch method {
	case http.MethodGet:
		adapter.GET(path, h)
	case http.MethodPost:
		adapter.POST(path, h)
	case http.MethodPut:
		adapter.PUT(path, h)
	case http.MethodDelete:
		adapter.DELETE(path, h)
	case http.MethodPatch:
		adapter.PATCH(path, h)
	}
}

// ---------------------------------------------------------------------------
// Tests: createArticle - archived state rejection (validation before DB call)
// ---------------------------------------------------------------------------

func TestCreateArticle_RejectArchivedState(t *testing.T) {
	r := setupTestRouter()
	handler := &ArticleHandler{uc: nil, jwt: nil}

	registerHandler(r, http.MethodPost, "/articles", handler.createArticle())

	body := map[string]interface{}{
		"title":   "Test Article",
		"content": "Test content",
		"state":   "archived",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/articles", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid state")
}

// ---------------------------------------------------------------------------
// Tests: createArticle - missing required fields (validation before DB call)
// ---------------------------------------------------------------------------

func TestCreateArticle_MissingTitle(t *testing.T) {
	r := setupTestRouter()
	handler := &ArticleHandler{uc: nil, jwt: nil}

	registerHandler(r, http.MethodPost, "/articles", handler.createArticle())

	body := map[string]interface{}{
		"content": "Test content",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/articles", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateArticle_MissingContent(t *testing.T) {
	r := setupTestRouter()
	handler := &ArticleHandler{uc: nil, jwt: nil}

	registerHandler(r, http.MethodPost, "/articles", handler.createArticle())

	body := map[string]interface{}{
		"title": "Test Article",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/articles", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: updateArticle - archived state rejection (validation before DB call)
// ---------------------------------------------------------------------------

func TestUpdateArticle_RejectArchivedState(t *testing.T) {
	r := setupTestRouter()
	handler := &ArticleHandler{uc: nil, jwt: nil}

	registerHandler(r, http.MethodPut, "/articles/:id", handler.updateArticle())

	body := map[string]interface{}{
		"state": "archived",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/articles/article-1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid state")
}

// ---------------------------------------------------------------------------
// Tests: updateArticle - missing article id (validation before DB call)
// ---------------------------------------------------------------------------

func TestUpdateArticle_MissingID(t *testing.T) {
	r := setupTestRouter()
	handler := &ArticleHandler{uc: nil, jwt: nil}

	// Route without :id param
	registerHandler(r, http.MethodPut, "/articles", handler.updateArticle())

	body := map[string]interface{}{
		"title": "Updated Title",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/articles", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "article id is required")
}

// ---------------------------------------------------------------------------
// Tests: updateArticleState - only draft/published allowed (validation before DB call)
// ---------------------------------------------------------------------------

func TestUpdateArticleState_RejectArchivedState(t *testing.T) {
	r := setupTestRouter()
	handler := &ArticleHandler{uc: nil, jwt: nil}

	registerHandler(r, http.MethodPatch, "/articles/:id/state", handler.updateArticleState())

	body := map[string]interface{}{
		"state": "archived",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/articles/article-1/state", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid state")
}

func TestUpdateArticleState_RejectInvalidState(t *testing.T) {
	r := setupTestRouter()
	handler := &ArticleHandler{uc: nil, jwt: nil}

	registerHandler(r, http.MethodPatch, "/articles/:id/state", handler.updateArticleState())

	body := map[string]interface{}{
		"state": "deleted",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/articles/article-1/state", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid state")
}

func TestUpdateArticleState_MissingState(t *testing.T) {
	r := setupTestRouter()
	handler := &ArticleHandler{uc: nil, jwt: nil}

	registerHandler(r, http.MethodPatch, "/articles/:id/state", handler.updateArticleState())

	body := map[string]interface{}{}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/articles/article-1/state", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: deleteArticle - missing article id (validation before DB call)
// ---------------------------------------------------------------------------

func TestDeleteArticle_MissingID(t *testing.T) {
	r := setupTestRouter()
	handler := &ArticleHandler{uc: nil, jwt: nil}

	registerHandler(r, http.MethodDelete, "/articles", handler.deleteArticle())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/articles", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "article id is required")
}

// ---------------------------------------------------------------------------
// Tests: listMyArticles - requires authentication (validation before DB call)
// ---------------------------------------------------------------------------

func TestListMyArticles_RequiresAuth(t *testing.T) {
	r := setupTestRouter()
	handler := &ArticleHandler{uc: nil, jwt: nil}

	registerHandler(r, http.MethodGet, "/articles/me", handler.listMyArticles())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/articles/me", nil)
	r.ServeHTTP(w, req)

	// Without claims, extractUserID returns "", should return 400
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "authentication required")
}
