package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	contentbiz "origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server"
)

// ---------------------------------------------------------------------------
// Mock types for user/service playlist tests
// ---------------------------------------------------------------------------

// meMockPlaylistRepo implements contentbiz.PlaylistRepo for testing.
type meMockPlaylistRepo struct {
	playlists map[string]*contentbiz.Playlist
	nextID    int
}

func newMeMockPlaylistRepo() *meMockPlaylistRepo {
	return &meMockPlaylistRepo{
		playlists: make(map[string]*contentbiz.Playlist),
		nextID:    1,
	}
}

func (m *meMockPlaylistRepo) Create(_ context.Context, p *contentbiz.Playlist) (*contentbiz.Playlist, error) {
	id := fmt.Sprintf("pl-%03d", m.nextID)
	m.nextID++
	result := &contentbiz.Playlist{
		ID:          id,
		Title:       p.Title,
		Description: p.Description,
		ShortToken:  fmt.Sprintf("st-%s", id),
		UserID:      p.UserID,
		IsPublic:    p.IsPublic,
		MediaItems:  []string{},
	}
	m.playlists[id] = result
	return result, nil
}

func (m *meMockPlaylistRepo) Get(_ context.Context, id string) (*contentbiz.Playlist, error) {
	p, ok := m.playlists[id]
	if !ok {
		return nil, fmt.Errorf("playlist not found")
	}
	return p, nil
}

func (m *meMockPlaylistRepo) GetByShortToken(_ context.Context, token string) (*contentbiz.Playlist, error) {
	for _, p := range m.playlists {
		if p.ShortToken == token {
			return p, nil
		}
	}
	return nil, fmt.Errorf("playlist not found")
}

func (m *meMockPlaylistRepo) Update(_ context.Context, p *contentbiz.Playlist) (*contentbiz.Playlist, error) {
	existing, ok := m.playlists[p.ID]
	if !ok {
		return nil, fmt.Errorf("playlist not found")
	}
	existing.Title = p.Title
	existing.Description = p.Description
	existing.IsPublic = p.IsPublic
	return existing, nil
}

func (m *meMockPlaylistRepo) Delete(_ context.Context, id string) error {
	delete(m.playlists, id)
	return nil
}

func (m *meMockPlaylistRepo) ListByUser(_ context.Context, userID string, page, pageSize int) ([]*contentbiz.Playlist, int, error) {
	var result []*contentbiz.Playlist
	for _, p := range m.playlists {
		if p.UserID == userID {
			result = append(result, p)
		}
	}
	return result, len(result), nil
}

func (m *meMockPlaylistRepo) ListAll(_ context.Context, page, pageSize int) ([]*contentbiz.Playlist, int, error) {
	var result []*contentbiz.Playlist
	for _, p := range m.playlists {
		result = append(result, p)
	}
	return result, len(result), nil
}

func (m *meMockPlaylistRepo) AddMedia(_ context.Context, playlistID, mediaID string) error {
	p, ok := m.playlists[playlistID]
	if !ok {
		return fmt.Errorf("playlist not found")
	}
	p.MediaItems = append(p.MediaItems, mediaID)
	return nil
}

func (m *meMockPlaylistRepo) RemoveMedia(_ context.Context, playlistID, mediaID string) error {
	p, ok := m.playlists[playlistID]
	if !ok {
		return fmt.Errorf("playlist not found")
	}
	for i, id := range p.MediaItems {
		if id == mediaID {
			p.MediaItems = append(p.MediaItems[:i], p.MediaItems[i+1:]...)
			break
		}
	}
	return nil
}

func (m *meMockPlaylistRepo) ReorderMedia(_ context.Context, playlistID string, mediaOrders map[string]int) error {
	return nil
}

func (m *meMockPlaylistRepo) GetPlaylistMedia(_ context.Context, playlistID string) ([]string, error) {
	p, ok := m.playlists[playlistID]
	if !ok {
		return nil, fmt.Errorf("playlist not found")
	}
	return p.MediaItems, nil
}

func (m *meMockPlaylistRepo) GetPlaylistMediaDetails(_ context.Context, playlistID string) ([]contentbiz.PlaylistMediaItem, error) {
	return []contentbiz.PlaylistMediaItem{}, nil
}

// meMockSystemConfigRepo implements contentbiz.SystemConfigRepo for testing.
type meMockSystemConfigRepo struct {
	configs map[string]string
}

func newMeMockSystemConfigRepo() *meMockSystemConfigRepo {
	return &meMockSystemConfigRepo{configs: make(map[string]string)}
}

func (m *meMockSystemConfigRepo) Get(_ context.Context, key string) (string, error) {
	val, ok := m.configs[key]
	if !ok {
		return "", fmt.Errorf("config not found: %s", key)
	}
	return val, nil
}

func (m *meMockSystemConfigRepo) Set(_ context.Context, key, value string) error {
	m.configs[key] = value
	return nil
}

func (m *meMockSystemConfigRepo) ListByCategory(_ context.Context, category string) (map[string]string, error) {
	return m.configs, nil
}

func (m *meMockSystemConfigRepo) Delete(_ context.Context, key string) error {
	delete(m.configs, key)
	return nil
}

// meMockUserRepo implements contentbiz.UserRepo for testing.
type meMockUserRepo struct{}

func (m *meMockUserRepo) GetByUsername(_ context.Context, username string) (*contentbiz.User, error) {
	return nil, fmt.Errorf("user not found")
}

// ---------------------------------------------------------------------------
// Test environment setup
// ---------------------------------------------------------------------------

// mePlaylistTestEnv holds the test environment for MeHandler playlist tests.
type mePlaylistTestEnv struct {
	router       *gin.Engine
	handler      *MeHandler
	playlistRepo *meMockPlaylistRepo
}

// withClaims creates a middleware that injects test claims into the context.
func withMeClaims(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("claims", &auth.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: userID,
			},
		})
		c.Next()
	}
}

// newMePlaylistTestEnv creates a test environment without auth middleware.
func newMePlaylistTestEnv() *mePlaylistTestEnv {
	gin.SetMode(gin.TestMode)

	playlistRepo := newMeMockPlaylistRepo()
	configRepo := newMeMockSystemConfigRepo()
	configRepo.configs["module_videos"] = "true"
	userRepo := &meMockUserRepo{}

	playlistUC := contentbiz.NewPlaylistChannelUseCase(playlistRepo, nil, configRepo, userRepo, nil)

	handler := &MeHandler{
		playlistUC: playlistUC,
	}

	r := gin.New()
	me := r.Group("/me")
	{
		me.GET("/playlists", handler.GetPlaylists)
		me.POST("/playlists", handler.CreatePlaylist)
		me.PATCH("/playlists/:id", handler.UpdatePlaylist)
		me.DELETE("/playlists/:id", handler.DeletePlaylist)
		me.POST("/playlists/:id/media", handler.AddMediaToPlaylist)
		me.DELETE("/playlists/:id/media/:mediaId", handler.RemoveMediaFromPlaylist)
	}

	return &mePlaylistTestEnv{
		router:       r,
		handler:      handler,
		playlistRepo: playlistRepo,
	}
}

// newMePlaylistTestEnvWithAuth creates a test environment with JWT claims injection middleware.
func newMePlaylistTestEnvWithAuth(userID string) *mePlaylistTestEnv {
	gin.SetMode(gin.TestMode)

	playlistRepo := newMeMockPlaylistRepo()
	configRepo := newMeMockSystemConfigRepo()
	configRepo.configs["module_videos"] = "true"
	userRepo := &meMockUserRepo{}

	playlistUC := contentbiz.NewPlaylistChannelUseCase(playlistRepo, nil, configRepo, userRepo, nil)

	handler := &MeHandler{
		playlistUC: playlistUC,
	}

	r := gin.New()
	me := r.Group("/me")
	me.Use(withMeClaims(userID))
	{
		me.GET("/playlists", handler.GetPlaylists)
		me.POST("/playlists", handler.CreatePlaylist)
		me.PATCH("/playlists/:id", handler.UpdatePlaylist)
		me.DELETE("/playlists/:id", handler.DeletePlaylist)
		me.POST("/playlists/:id/media", handler.AddMediaToPlaylist)
		me.DELETE("/playlists/:id/media/:mediaId", handler.RemoveMediaFromPlaylist)
	}

	return &mePlaylistTestEnv{
		router:       r,
		handler:      handler,
		playlistRepo: playlistRepo,
	}
}

// ---------------------------------------------------------------------------
// Tests: GET /me/playlists - list user's playlists
// ---------------------------------------------------------------------------

func TestMeHandler_GetPlaylists_RequiresAuth(t *testing.T) {
	env := newMePlaylistTestEnv()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/me/playlists", nil)
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(server.ErrUnauthorized), resp["code"])
}

func TestMeHandler_GetPlaylists_Success(t *testing.T) {
	env := newMePlaylistTestEnvWithAuth("user-001")

	// Create a playlist first
	env.playlistRepo.Create(context.Background(), &contentbiz.Playlist{
		Title:    "Test Playlist",
		UserID:   "user-001",
		IsPublic: true,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/me/playlists", nil)
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok, "response should have data field")
	items, ok := data["items"].([]interface{})
	assert.True(t, ok, "data should have items array")
	assert.Equal(t, 1, len(items), "should have 1 playlist")
}

func TestMeHandler_GetPlaylists_EmptyList(t *testing.T) {
	env := newMePlaylistTestEnvWithAuth("user-001")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/me/playlists", nil)
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
}

func TestMeHandler_GetPlaylists_WithPagination(t *testing.T) {
	env := newMePlaylistTestEnvWithAuth("user-001")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/me/playlists?page=1&page_size=10", nil)
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: POST /me/playlists - create playlist
// ---------------------------------------------------------------------------

func TestMeHandler_CreatePlaylist_RequiresAuth(t *testing.T) {
	env := newMePlaylistTestEnv()

	body, _ := json.Marshal(map[string]string{
		"title": "Test Playlist",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/me/playlists", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMeHandler_CreatePlaylist_MissingTitle(t *testing.T) {
	env := newMePlaylistTestEnvWithAuth("user-001")

	body, _ := json.Marshal(map[string]string{
		"description": "A test playlist",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/me/playlists", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(server.ErrBadRequest), resp["code"])
	assert.Contains(t, resp["message"], "title is required")
}

func TestMeHandler_CreatePlaylist_Success(t *testing.T) {
	env := newMePlaylistTestEnvWithAuth("user-001")

	body, _ := json.Marshal(map[string]string{
		"title":       "My Playlist",
		"description": "A test playlist",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/me/playlists", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok, "response should have data field")
	playlist, ok := data["playlist"].(map[string]interface{})
	assert.True(t, ok, "data should have playlist field")
	assert.Equal(t, "My Playlist", playlist["title"])
	assert.Equal(t, "user-001", playlist["user_id"])
}

func TestMeHandler_CreatePlaylist_EmptyTitle(t *testing.T) {
	env := newMePlaylistTestEnvWithAuth("user-001")

	body, _ := json.Marshal(map[string]string{
		"title": "",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/me/playlists", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMeHandler_CreatePlaylist_InvalidJSON(t *testing.T) {
	env := newMePlaylistTestEnvWithAuth("user-001")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/me/playlists", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: PATCH /me/playlists/:id - update playlist
// ---------------------------------------------------------------------------

func TestMeHandler_UpdatePlaylist_RequiresAuth(t *testing.T) {
	env := newMePlaylistTestEnv()

	body, _ := json.Marshal(map[string]string{
		"title": "Updated Title",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/me/playlists/pl-001", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMeHandler_UpdatePlaylist_NotFound(t *testing.T) {
	env := newMePlaylistTestEnvWithAuth("user-001")

	body, _ := json.Marshal(map[string]string{
		"title": "Updated Title",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/me/playlists/nonexistent-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(server.ErrNotFound), resp["code"])
}

func TestMeHandler_UpdatePlaylist_Success(t *testing.T) {
	env := newMePlaylistTestEnvWithAuth("user-001")

	// Create a playlist first
	created, _ := env.playlistRepo.Create(context.Background(), &contentbiz.Playlist{
		Title:    "Original Title",
		UserID:   "user-001",
		IsPublic: true,
	})

	body, _ := json.Marshal(map[string]interface{}{
		"title":       "Updated Title",
		"description": "Updated description",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/me/playlists/"+created.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok)
	playlist, ok := data["playlist"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Updated Title", playlist["title"])
}

func TestMeHandler_UpdatePlaylist_InvalidJSON(t *testing.T) {
	env := newMePlaylistTestEnvWithAuth("user-001")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/me/playlists/pl-001", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: DELETE /me/playlists/:id - delete playlist
// ---------------------------------------------------------------------------

func TestMeHandler_DeletePlaylist_RequiresAuth(t *testing.T) {
	env := newMePlaylistTestEnv()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/me/playlists/pl-001", nil)
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMeHandler_DeletePlaylist_Success(t *testing.T) {
	env := newMePlaylistTestEnvWithAuth("user-001")

	// Create a playlist first
	created, _ := env.playlistRepo.Create(context.Background(), &contentbiz.Playlist{
		Title:    "To Delete",
		UserID:   "user-001",
		IsPublic: true,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/me/playlists/"+created.ID, nil)
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])

	// Verify the playlist is deleted
	_, err := env.playlistRepo.Get(context.Background(), created.ID)
	assert.Error(t, err, "playlist should be deleted")
}

func TestMeHandler_DeletePlaylist_NotOwned(t *testing.T) {
	env := newMePlaylistTestEnvWithAuth("user-002")

	// Create a playlist owned by user-001
	created, _ := env.playlistRepo.Create(context.Background(), &contentbiz.Playlist{
		Title:    "Other User's Playlist",
		UserID:   "user-001",
		IsPublic: true,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/me/playlists/"+created.ID, nil)
	env.router.ServeHTTP(w, req)

	// Should fail because user-002 doesn't own this playlist
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: POST /me/playlists/:id/media - add media to playlist
// ---------------------------------------------------------------------------

func TestMeHandler_AddMedia_RequiresAuth(t *testing.T) {
	env := newMePlaylistTestEnv()

	body, _ := json.Marshal(map[string]string{
		"media_id": "media-001",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/me/playlists/pl-001/media", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMeHandler_AddMedia_MissingMediaID(t *testing.T) {
	env := newMePlaylistTestEnvWithAuth("user-001")

	body, _ := json.Marshal(map[string]string{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/me/playlists/pl-001/media", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["message"], "media_id is required")
}

func TestMeHandler_AddMedia_Success(t *testing.T) {
	env := newMePlaylistTestEnvWithAuth("user-001")

	// Create a playlist first
	created, _ := env.playlistRepo.Create(context.Background(), &contentbiz.Playlist{
		Title:    "My Playlist",
		UserID:   "user-001",
		IsPublic: true,
	})

	body, _ := json.Marshal(map[string]string{
		"media_id": "media-001",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/me/playlists/"+created.ID+"/media", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])

	// Verify the media was added
	mediaItems, _ := env.playlistRepo.GetPlaylistMedia(context.Background(), created.ID)
	assert.Contains(t, mediaItems, "media-001")
}

func TestMeHandler_AddMedia_InvalidJSON(t *testing.T) {
	env := newMePlaylistTestEnvWithAuth("user-001")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/me/playlists/pl-001/media", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: DELETE /me/playlists/:id/media/:mediaId - remove media from playlist
// ---------------------------------------------------------------------------

func TestMeHandler_RemoveMedia_RequiresAuth(t *testing.T) {
	env := newMePlaylistTestEnv()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/me/playlists/pl-001/media/media-001", nil)
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMeHandler_RemoveMedia_Success(t *testing.T) {
	env := newMePlaylistTestEnvWithAuth("user-001")

	// Create a playlist with a media item
	created, _ := env.playlistRepo.Create(context.Background(), &contentbiz.Playlist{
		Title:    "My Playlist",
		UserID:   "user-001",
		IsPublic: true,
	})
	env.playlistRepo.AddMedia(context.Background(), created.ID, "media-001")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/me/playlists/"+created.ID+"/media/media-001", nil)
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])

	// Verify the media was removed
	mediaItems, _ := env.playlistRepo.GetPlaylistMedia(context.Background(), created.ID)
	assert.NotContains(t, mediaItems, "media-001")
}

// ---------------------------------------------------------------------------
// Tests: Biz layer - PlaylistChannelUseCase
// ---------------------------------------------------------------------------

func TestPlaylistChannelUseCase_CreatePlaylist(t *testing.T) {
	repo := newMeMockPlaylistRepo()
	configRepo := newMeMockSystemConfigRepo()
	uc := contentbiz.NewPlaylistChannelUseCase(repo, nil, configRepo, &meMockUserRepo{}, nil)

	p, err := uc.CreatePlaylist(context.Background(), &contentbiz.Playlist{
		Title:    "Test Playlist",
		UserID:   "user-001",
		IsPublic: true,
	})

	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "Test Playlist", p.Title)
	assert.Equal(t, "user-001", p.UserID)
	assert.True(t, p.IsPublic)
	assert.NotEmpty(t, p.ID)
	assert.NotEmpty(t, p.ShortToken)
}

func TestPlaylistChannelUseCase_GetPlaylist(t *testing.T) {
	repo := newMeMockPlaylistRepo()
	configRepo := newMeMockSystemConfigRepo()
	uc := contentbiz.NewPlaylistChannelUseCase(repo, nil, configRepo, &meMockUserRepo{}, nil)

	created, _ := uc.CreatePlaylist(context.Background(), &contentbiz.Playlist{
		Title:    "Test Playlist",
		UserID:   "user-001",
		IsPublic: true,
	})

	found, err := uc.GetPlaylist(context.Background(), created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, created.Title, found.Title)
}

func TestPlaylistChannelUseCase_GetPlaylist_NotFound(t *testing.T) {
	repo := newMeMockPlaylistRepo()
	configRepo := newMeMockSystemConfigRepo()
	uc := contentbiz.NewPlaylistChannelUseCase(repo, nil, configRepo, &meMockUserRepo{}, nil)

	_, err := uc.GetPlaylist(context.Background(), "nonexistent-id")
	assert.Error(t, err)
}

func TestPlaylistChannelUseCase_GetPlaylistByShortToken(t *testing.T) {
	repo := newMeMockPlaylistRepo()
	configRepo := newMeMockSystemConfigRepo()
	uc := contentbiz.NewPlaylistChannelUseCase(repo, nil, configRepo, &meMockUserRepo{}, nil)

	created, _ := uc.CreatePlaylist(context.Background(), &contentbiz.Playlist{
		Title:    "Test Playlist",
		UserID:   "user-001",
		IsPublic: true,
	})

	found, err := uc.GetPlaylistByShortToken(context.Background(), created.ShortToken)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestPlaylistChannelUseCase_UpdatePlaylist(t *testing.T) {
	repo := newMeMockPlaylistRepo()
	configRepo := newMeMockSystemConfigRepo()
	uc := contentbiz.NewPlaylistChannelUseCase(repo, nil, configRepo, &meMockUserRepo{}, nil)

	created, _ := uc.CreatePlaylist(context.Background(), &contentbiz.Playlist{
		Title:    "Original Title",
		UserID:   "user-001",
		IsPublic: true,
	})

	created.Title = "Updated Title"
	updated, err := uc.UpdatePlaylist(context.Background(), created, "user-001", false)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", updated.Title)
}

func TestPlaylistChannelUseCase_UpdatePlaylist_PermissionDenied(t *testing.T) {
	repo := newMeMockPlaylistRepo()
	configRepo := newMeMockSystemConfigRepo()
	uc := contentbiz.NewPlaylistChannelUseCase(repo, nil, configRepo, &meMockUserRepo{}, nil)

	created, _ := uc.CreatePlaylist(context.Background(), &contentbiz.Playlist{
		Title:    "Original Title",
		UserID:   "user-001",
		IsPublic: true,
	})

	created.Title = "Hacked Title"
	_, err := uc.UpdatePlaylist(context.Background(), created, "user-002", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestPlaylistChannelUseCase_UpdatePlaylist_AdminOverride(t *testing.T) {
	repo := newMeMockPlaylistRepo()
	configRepo := newMeMockSystemConfigRepo()
	uc := contentbiz.NewPlaylistChannelUseCase(repo, nil, configRepo, &meMockUserRepo{}, nil)

	created, _ := uc.CreatePlaylist(context.Background(), &contentbiz.Playlist{
		Title:    "Original Title",
		UserID:   "user-001",
		IsPublic: true,
	})

	created.Title = "Admin Updated Title"
	updated, err := uc.UpdatePlaylist(context.Background(), created, "admin-001", true)
	assert.NoError(t, err)
	assert.Equal(t, "Admin Updated Title", updated.Title)
}

func TestPlaylistChannelUseCase_DeletePlaylist(t *testing.T) {
	repo := newMeMockPlaylistRepo()
	configRepo := newMeMockSystemConfigRepo()
	uc := contentbiz.NewPlaylistChannelUseCase(repo, nil, configRepo, &meMockUserRepo{}, nil)

	created, _ := uc.CreatePlaylist(context.Background(), &contentbiz.Playlist{
		Title:    "To Delete",
		UserID:   "user-001",
		IsPublic: true,
	})

	err := uc.DeletePlaylist(context.Background(), created.ID, "user-001", false)
	assert.NoError(t, err)

	_, err = uc.GetPlaylist(context.Background(), created.ID)
	assert.Error(t, err)
}

func TestPlaylistChannelUseCase_DeletePlaylist_PermissionDenied(t *testing.T) {
	repo := newMeMockPlaylistRepo()
	configRepo := newMeMockSystemConfigRepo()
	uc := contentbiz.NewPlaylistChannelUseCase(repo, nil, configRepo, &meMockUserRepo{}, nil)

	created, _ := uc.CreatePlaylist(context.Background(), &contentbiz.Playlist{
		Title:    "To Delete",
		UserID:   "user-001",
		IsPublic: true,
	})

	err := uc.DeletePlaylist(context.Background(), created.ID, "user-002", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestPlaylistChannelUseCase_AddMediaToPlaylist(t *testing.T) {
	repo := newMeMockPlaylistRepo()
	configRepo := newMeMockSystemConfigRepo()
	uc := contentbiz.NewPlaylistChannelUseCase(repo, nil, configRepo, &meMockUserRepo{}, nil)

	created, _ := uc.CreatePlaylist(context.Background(), &contentbiz.Playlist{
		Title:    "My Playlist",
		UserID:   "user-001",
		IsPublic: true,
	})

	err := uc.AddMediaToPlaylist(context.Background(), created.ID, "media-001", "user-001", false)
	assert.NoError(t, err)

	mediaItems, _ := repo.GetPlaylistMedia(context.Background(), created.ID)
	assert.Contains(t, mediaItems, "media-001")
}

func TestPlaylistChannelUseCase_RemoveMediaFromPlaylist(t *testing.T) {
	repo := newMeMockPlaylistRepo()
	configRepo := newMeMockSystemConfigRepo()
	uc := contentbiz.NewPlaylistChannelUseCase(repo, nil, configRepo, &meMockUserRepo{}, nil)

	created, _ := uc.CreatePlaylist(context.Background(), &contentbiz.Playlist{
		Title:    "My Playlist",
		UserID:   "user-001",
		IsPublic: true,
	})

	uc.AddMediaToPlaylist(context.Background(), created.ID, "media-001", "user-001", false)

	err := uc.RemoveMediaFromPlaylist(context.Background(), created.ID, "media-001", "user-001", false)
	assert.NoError(t, err)

	mediaItems, _ := repo.GetPlaylistMedia(context.Background(), created.ID)
	assert.NotContains(t, mediaItems, "media-001")
}

func TestPlaylistChannelUseCase_ListUserPlaylists(t *testing.T) {
	repo := newMeMockPlaylistRepo()
	configRepo := newMeMockSystemConfigRepo()
	uc := contentbiz.NewPlaylistChannelUseCase(repo, nil, configRepo, &meMockUserRepo{}, nil)

	uc.CreatePlaylist(context.Background(), &contentbiz.Playlist{
		Title: "Playlist 1", UserID: "user-001", IsPublic: true,
	})
	uc.CreatePlaylist(context.Background(), &contentbiz.Playlist{
		Title: "Playlist 2", UserID: "user-001", IsPublic: false,
	})
	uc.CreatePlaylist(context.Background(), &contentbiz.Playlist{
		Title: "Other User's Playlist", UserID: "user-002", IsPublic: true,
	})

	list, total, err := uc.ListUserPlaylists(context.Background(), "user-001", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Equal(t, 2, len(list))
}

func TestPlaylistChannelUseCase_ListAllPlaylists(t *testing.T) {
	repo := newMeMockPlaylistRepo()
	configRepo := newMeMockSystemConfigRepo()
	uc := contentbiz.NewPlaylistChannelUseCase(repo, nil, configRepo, &meMockUserRepo{}, nil)

	uc.CreatePlaylist(context.Background(), &contentbiz.Playlist{
		Title: "Playlist 1", UserID: "user-001", IsPublic: true,
	})
	uc.CreatePlaylist(context.Background(), &contentbiz.Playlist{
		Title: "Playlist 2", UserID: "user-002", IsPublic: true,
	})

	list, total, err := uc.ListPlaylists(context.Background(), 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Equal(t, 2, len(list))
}
