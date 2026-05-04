package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

// ---------------------------------------------------------------------------
// Tests: createArticle - archived state rejection (validation before DB call)
// ---------------------------------------------------------------------------

func TestCreateArticle_RejectArchivedState(t *testing.T) {
	r := setupTestRouter()
	handler := &ArticleHandler{uc: nil, jwt: nil}

	r.POST("/articles", handler.createArticle())

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

	r.POST("/articles", handler.createArticle())

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

	r.POST("/articles", handler.createArticle())

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

	r.PUT("/articles/:id", handler.updateArticle())

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
	r.PUT("/articles", handler.updateArticle())

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

	r.PATCH("/articles/:id/state", handler.updateArticleState())

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

	r.PATCH("/articles/:id/state", handler.updateArticleState())

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

	r.PATCH("/articles/:id/state", handler.updateArticleState())

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

	r.DELETE("/articles", handler.deleteArticle())

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

	r.GET("/articles/me", handler.listMyArticles())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/articles/me", nil)
	r.ServeHTTP(w, req)

	// Without claims, extractUserID returns "", should return 400
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "authentication required")
}
