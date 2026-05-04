package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"origadmin/application/origcms/internal/features/content/biz"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/server"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// mockPlaylistRepo implements biz.PlaylistRepo for testing without a database.
type mockPlaylistRepo struct {
	playlists map[string]*biz.Playlist
	nextID    int
}

func newMockPlaylistRepo() *mockPlaylistRepo {
	return &mockPlaylistRepo{
		playlists: make(map[string]*biz.Playlist),
		nextID:    1,
	}
}

func (m *mockPlaylistRepo) Create(_ context.Context, p *biz.Playlist) (*biz.Playlist, error) {
	id := fmt.Sprintf("pl-%03d", m.nextID)
	m.nextID++
	result := &biz.Playlist{
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

func (m *mockPlaylistRepo) Get(_ context.Context, id string) (*biz.Playlist, error) {
	p, ok := m.playlists[id]
	if !ok {
		return nil, fmt.Errorf("playlist not found")
	}
	return p, nil
}

func (m *mockPlaylistRepo) GetByShortToken(_ context.Context, token string) (*biz.Playlist, error) {
	for _, p := range m.playlists {
		if p.ShortToken == token {
			return p, nil
		}
	}
	return nil, fmt.Errorf("playlist not found")
}

func (m *mockPlaylistRepo) Update(_ context.Context, p *biz.Playlist) (*biz.Playlist, error) {
	existing, ok := m.playlists[p.ID]
	if !ok {
		return nil, fmt.Errorf("playlist not found")
	}
	existing.Title = p.Title
	existing.Description = p.Description
	existing.IsPublic = p.IsPublic
	return existing, nil
}

func (m *mockPlaylistRepo) Delete(_ context.Context, id string) error {
	delete(m.playlists, id)
	return nil
}

func (m *mockPlaylistRepo) ListByUser(_ context.Context, userID string, page, pageSize int) ([]*biz.Playlist, int, error) {
	var result []*biz.Playlist
	for _, p := range m.playlists {
		if p.UserID == userID {
			result = append(result, p)
		}
	}
	return result, len(result), nil
}

func (m *mockPlaylistRepo) ListAll(_ context.Context, page, pageSize int) ([]*biz.Playlist, int, error) {
	var result []*biz.Playlist
	for _, p := range m.playlists {
		result = append(result, p)
	}
	return result, len(result), nil
}

func (m *mockPlaylistRepo) AddMedia(_ context.Context, playlistID, mediaID string) error {
	p, ok := m.playlists[playlistID]
	if !ok {
		return fmt.Errorf("playlist not found")
	}
	p.MediaItems = append(p.MediaItems, mediaID)
	return nil
}

func (m *mockPlaylistRepo) RemoveMedia(_ context.Context, playlistID, mediaID string) error {
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

func (m *mockPlaylistRepo) ReorderMedia(_ context.Context, playlistID string, mediaOrders map[string]int) error {
	return nil
}

func (m *mockPlaylistRepo) GetPlaylistMedia(_ context.Context, playlistID string) ([]string, error) {
	p, ok := m.playlists[playlistID]
	if !ok {
		return nil, fmt.Errorf("playlist not found")
	}
	return p.MediaItems, nil
}

func (m *mockPlaylistRepo) GetPlaylistMediaDetails(_ context.Context, playlistID string) ([]biz.PlaylistMediaItem, error) {
	return []biz.PlaylistMediaItem{}, nil
}

// mockSystemConfigRepo implements biz.SystemConfigRepo for testing.
type mockSystemConfigRepo struct {
	configs map[string]string
}

func newMockSystemConfigRepo() *mockSystemConfigRepo {
	return &mockSystemConfigRepo{configs: make(map[string]string)}
}

func (m *mockSystemConfigRepo) Get(_ context.Context, key string) (string, error) {
	val, ok := m.configs[key]
	if !ok {
		return "", fmt.Errorf("config not found: %s", key)
	}
	return val, nil
}

func (m *mockSystemConfigRepo) Set(_ context.Context, key, value string) error {
	m.configs[key] = value
	return nil
}

func (m *mockSystemConfigRepo) ListByCategory(_ context.Context, category string) (map[string]string, error) {
	return m.configs, nil
}

func (m *mockSystemConfigRepo) Delete(_ context.Context, key string) error {
	delete(m.configs, key)
	return nil
}

// mockUserRepo implements biz.UserRepo for testing.
type mockUserRepo struct{}

func (m *mockUserRepo) GetByUsername(_ context.Context, username string) (*biz.User, error) {
	return nil, fmt.Errorf("user not found")
}

// setupPlaylistTestHandler creates a PlaylistHandler with mock dependencies for testing.
func setupPlaylistTestHandler() *PlaylistHandler {
	repo := newMockPlaylistRepo()
	configRepo := newMockSystemConfigRepo()
	// Enable the videos module by default
	configRepo.configs["module_videos"] = "true"
	userRepo := &mockUserRepo{}
	uc := biz.NewPlaylistChannelUseCase(repo, nil, configRepo, userRepo, nil)
	return NewPlaylistHandler(uc, nil, nil)
}

// setAuthClaims injects mock claims into the gin context for authenticated endpoints.
func setAuthClaims(c *gin.Context, userID string, isStaff bool) {
	c.Set("claims", &auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID,
		},
		IsStaff: isStaff,
	})
}

// ---------------------------------------------------------------------------
// Tests: PlaylistHandler route registration
// ---------------------------------------------------------------------------

func TestPlaylistHandler_RouteRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	apiV1 := r.Group("/api/v1")
	playlists := apiV1.Group("/playlists")
	{
		playlists.GET("", func(c *gin.Context) {
			c.JSON(200, gin.H{"test": "list"})
		})
		playlists.GET("/:token", func(c *gin.Context) {
			token := c.Param("token")
			c.JSON(200, gin.H{"test": "detail", "token": token})
		})
	}

	routes := r.Routes()
	t.Logf("Registered routes (%d):", len(routes))
	for _, route := range routes {
		t.Logf("  %s %s -> %s", route.Method, route.Path, route.Handler)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/playlists", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "GET /api/v1/playlists should return 200")

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/playlists/rJvQd0-ll", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "GET /api/v1/playlists/rJvQd0-ll should return 200")
	assert.Contains(t, w.Body.String(), "rJvQd0-ll")

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/playlists-nonexistent/xxx", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code, "non-matching route should return 404")
}

func TestPlaylistHandler_TokenParamExtraction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	var capturedToken string
	r.GET("/api/v1/playlists/:token", func(c *gin.Context) {
		capturedToken = c.Param("token")
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/playlists/rJvQd0-ll", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "rJvQd0-ll", capturedToken)
}

// ---------------------------------------------------------------------------
// Tests: PlaylistHandler.getPlaylistByToken - validation
// ---------------------------------------------------------------------------

func TestPlaylistHandler_GetPlaylistByToken_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := setupPlaylistTestHandler()

	// Register without ModuleGuard (which needs settingUC)
	r.GET("/api/v1/playlists/:token", handler.getPlaylistByToken)

	// Empty token should be caught by Gin's routing (no match)
	// But if token is provided as empty string via param, handler should catch it
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/playlists/", nil)
	r.ServeHTTP(w, req)
	// Gin redirects /playlists/ to /playlists, so this is a 301 or 404
	assert.True(t, w.Code == http.StatusMovedPermanently || w.Code == http.StatusNotFound)
}

func TestPlaylistHandler_GetPlaylistByToken_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := setupPlaylistTestHandler()

	r.GET("/api/v1/playlists/:token", handler.getPlaylistByToken)

	// Non-existent token should return not found
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/playlists/nonexistent-token", nil)
	r.ServeHTTP(w, req)

	// Should return 404 since the playlist doesn't exist
	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(server.ErrNotFound), resp["code"])
}

func TestPlaylistHandler_FullServerSimulation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	apiV1 := r.Group("/api/v1")

	playlists := apiV1.Group("/playlists")
	{
		playlists.GET("", func(c *gin.Context) {
			c.JSON(200, gin.H{"test": "list"})
		})
		playlists.GET("/:token", func(c *gin.Context) {
			token := c.Param("token")
			c.JSON(200, gin.H{"test": "detail", "token": token})
		})
	}

	medias := apiV1.Group("/medias")
	{
		medias.GET("/:id/metadata", func(c *gin.Context) {
			c.JSON(200, gin.H{"test": "metadata"})
		})
	}

	admin := apiV1.Group("/admin")
	{
		adminPlaylists := admin.Group("/playlists")
		{
			adminPlaylists.GET("/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{"test": "admin-detail"})
			})
		}
	}

	testCases := []struct {
		method       string
		path         string
		expectedCode int
		description  string
	}{
		{"GET", "/api/v1/playlists", 200, "list playlists"},
		{"GET", "/api/v1/playlists/rJvQd0-ll", 200, "get playlist by token"},
		{"GET", "/api/v1/playlists/abc123", 200, "get playlist by token (simple)"},
		{"GET", "/api/v1/admin/playlists/550e8400-e29b-41d4-a716-446655440000", 200, "admin get playlist by id"},
		{"GET", "/api/v1/medias/123/metadata", 200, "get media metadata"},
		{"GET", "/api/v1/nonexistent", 404, "nonexistent route"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s %s", tc.method, tc.path), func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tc.method, tc.path, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, tc.expectedCode, w.Code, tc.description)
		})
	}
}

// ---------------------------------------------------------------------------
// Tests: PlaylistHandler.listPlaylists - pagination validation
// ---------------------------------------------------------------------------

func TestPlaylistHandler_ListPlaylists_DefaultPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := setupPlaylistTestHandler()

	r.GET("/api/v1/playlists", handler.listPlaylists)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/playlists", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
}

func TestPlaylistHandler_ListPlaylists_CustomPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := setupPlaylistTestHandler()

	r.GET("/api/v1/playlists", handler.listPlaylists)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/playlists?page=2&page_size=5", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: PlaylistHandler.getPlaylistByToken - private playlist access control
// ---------------------------------------------------------------------------

func TestPlaylistHandler_GetPlaylistByToken_PublicPlaylist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	playlistRepo := newMockPlaylistRepo()
	configRepo := newMockSystemConfigRepo()
	configRepo.configs["module_videos"] = "true"
	uc := biz.NewPlaylistChannelUseCase(playlistRepo, nil, configRepo, &mockUserRepo{}, nil)
	handler := NewPlaylistHandler(uc, nil, nil)

	// Create a public playlist
	created, _ := playlistRepo.Create(context.Background(), &biz.Playlist{
		Title:    "Public Playlist",
		UserID:   "user-001",
		IsPublic: true,
	})

	r := gin.New()
	r.GET("/api/v1/playlists/:token", handler.getPlaylistByToken)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/playlists/"+created.ShortToken, nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
}

func TestPlaylistHandler_GetPlaylistByToken_PrivatePlaylist_NoAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	playlistRepo := newMockPlaylistRepo()
	configRepo := newMockSystemConfigRepo()
	configRepo.configs["module_videos"] = "true"
	uc := biz.NewPlaylistChannelUseCase(playlistRepo, nil, configRepo, &mockUserRepo{}, nil)
	handler := NewPlaylistHandler(uc, nil, nil)

	// Create a private playlist
	created, _ := playlistRepo.Create(context.Background(), &biz.Playlist{
		Title:    "Private Playlist",
		UserID:   "user-001",
		IsPublic: false,
	})

	r := gin.New()
	r.GET("/api/v1/playlists/:token", handler.getPlaylistByToken)

	// Access without auth - should return not found
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/playlists/"+created.ShortToken, nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(server.ErrNotFound), resp["code"])
}

func TestPlaylistHandler_GetPlaylistByToken_PrivatePlaylist_OwnerAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	playlistRepo := newMockPlaylistRepo()
	configRepo := newMockSystemConfigRepo()
	configRepo.configs["module_videos"] = "true"
	uc := biz.NewPlaylistChannelUseCase(playlistRepo, nil, configRepo, &mockUserRepo{}, nil)
	handler := NewPlaylistHandler(uc, nil, nil)

	// Create a private playlist
	created, _ := playlistRepo.Create(context.Background(), &biz.Playlist{
		Title:    "Private Playlist",
		UserID:   "user-001",
		IsPublic: false,
	})

	r := gin.New()
	r.GET("/api/v1/playlists/:token", func(c *gin.Context) {
		c.Set("claims", &auth.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "user-001",
			},
		})
		c.Next()
	}, handler.getPlaylistByToken)

	// Access as owner - should succeed
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/playlists/"+created.ShortToken, nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPlaylistHandler_GetPlaylistByToken_PrivatePlaylist_NonOwnerAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	playlistRepo := newMockPlaylistRepo()
	configRepo := newMockSystemConfigRepo()
	configRepo.configs["module_videos"] = "true"
	uc := biz.NewPlaylistChannelUseCase(playlistRepo, nil, configRepo, &mockUserRepo{}, nil)
	handler := NewPlaylistHandler(uc, nil, nil)

	// Create a private playlist owned by user-001
	created, _ := playlistRepo.Create(context.Background(), &biz.Playlist{
		Title:    "Private Playlist",
		UserID:   "user-001",
		IsPublic: false,
	})

	r := gin.New()
	r.GET("/api/v1/playlists/:token", func(c *gin.Context) {
		c.Set("claims", &auth.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "user-002",
			},
		})
		c.Next()
	}, handler.getPlaylistByToken)

	// Access as non-owner - should return not found
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/playlists/"+created.ShortToken, nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(server.ErrNotFound), resp["code"])
}

// ---------------------------------------------------------------------------
// Tests: Biz layer - PlaylistChannelUseCase
// ---------------------------------------------------------------------------

func TestPlaylistChannelUseCase_CreatePlaylist(t *testing.T) {
	repo := newMockPlaylistRepo()
	configRepo := newMockSystemConfigRepo()
	uc := biz.NewPlaylistChannelUseCase(repo, nil, configRepo, &mockUserRepo{}, nil)

	p, err := uc.CreatePlaylist(context.Background(), &biz.Playlist{
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
	repo := newMockPlaylistRepo()
	configRepo := newMockSystemConfigRepo()
	uc := biz.NewPlaylistChannelUseCase(repo, nil, configRepo, &mockUserRepo{}, nil)

	created, _ := uc.CreatePlaylist(context.Background(), &biz.Playlist{
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
	repo := newMockPlaylistRepo()
	configRepo := newMockSystemConfigRepo()
	uc := biz.NewPlaylistChannelUseCase(repo, nil, configRepo, &mockUserRepo{}, nil)

	_, err := uc.GetPlaylist(context.Background(), "nonexistent-id")
	assert.Error(t, err)
}

func TestPlaylistChannelUseCase_GetPlaylistByShortToken(t *testing.T) {
	repo := newMockPlaylistRepo()
	configRepo := newMockSystemConfigRepo()
	uc := biz.NewPlaylistChannelUseCase(repo, nil, configRepo, &mockUserRepo{}, nil)

	created, _ := uc.CreatePlaylist(context.Background(), &biz.Playlist{
		Title:    "Test Playlist",
		UserID:   "user-001",
		IsPublic: true,
	})

	found, err := uc.GetPlaylistByShortToken(context.Background(), created.ShortToken)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestPlaylistChannelUseCase_UpdatePlaylist(t *testing.T) {
	repo := newMockPlaylistRepo()
	configRepo := newMockSystemConfigRepo()
	uc := biz.NewPlaylistChannelUseCase(repo, nil, configRepo, &mockUserRepo{}, nil)

	created, _ := uc.CreatePlaylist(context.Background(), &biz.Playlist{
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
	repo := newMockPlaylistRepo()
	configRepo := newMockSystemConfigRepo()
	uc := biz.NewPlaylistChannelUseCase(repo, nil, configRepo, &mockUserRepo{}, nil)

	created, _ := uc.CreatePlaylist(context.Background(), &biz.Playlist{
		Title:    "Original Title",
		UserID:   "user-001",
		IsPublic: true,
	})

	created.Title = "Hacked Title"
	_, err := uc.UpdatePlaylist(context.Background(), created, "user-002", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestPlaylistChannelUseCase_DeletePlaylist(t *testing.T) {
	repo := newMockPlaylistRepo()
	configRepo := newMockSystemConfigRepo()
	uc := biz.NewPlaylistChannelUseCase(repo, nil, configRepo, &mockUserRepo{}, nil)

	created, _ := uc.CreatePlaylist(context.Background(), &biz.Playlist{
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
	repo := newMockPlaylistRepo()
	configRepo := newMockSystemConfigRepo()
	uc := biz.NewPlaylistChannelUseCase(repo, nil, configRepo, &mockUserRepo{}, nil)

	created, _ := uc.CreatePlaylist(context.Background(), &biz.Playlist{
		Title:    "To Delete",
		UserID:   "user-001",
		IsPublic: true,
	})

	err := uc.DeletePlaylist(context.Background(), created.ID, "user-002", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestPlaylistChannelUseCase_AddMediaToPlaylist(t *testing.T) {
	repo := newMockPlaylistRepo()
	configRepo := newMockSystemConfigRepo()
	uc := biz.NewPlaylistChannelUseCase(repo, nil, configRepo, &mockUserRepo{}, nil)

	created, _ := uc.CreatePlaylist(context.Background(), &biz.Playlist{
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
	repo := newMockPlaylistRepo()
	configRepo := newMockSystemConfigRepo()
	uc := biz.NewPlaylistChannelUseCase(repo, nil, configRepo, &mockUserRepo{}, nil)

	created, _ := uc.CreatePlaylist(context.Background(), &biz.Playlist{
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
	repo := newMockPlaylistRepo()
	configRepo := newMockSystemConfigRepo()
	uc := biz.NewPlaylistChannelUseCase(repo, nil, configRepo, &mockUserRepo{}, nil)

	uc.CreatePlaylist(context.Background(), &biz.Playlist{
		Title: "Playlist 1", UserID: "user-001", IsPublic: true,
	})
	uc.CreatePlaylist(context.Background(), &biz.Playlist{
		Title: "Playlist 2", UserID: "user-001", IsPublic: false,
	})
	uc.CreatePlaylist(context.Background(), &biz.Playlist{
		Title: "Other User's Playlist", UserID: "user-002", IsPublic: true,
	})

	list, total, err := uc.ListUserPlaylists(context.Background(), "user-001", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Equal(t, 2, len(list))
}

func TestPlaylistChannelUseCase_ListAllPlaylists(t *testing.T) {
	repo := newMockPlaylistRepo()
	configRepo := newMockSystemConfigRepo()
	uc := biz.NewPlaylistChannelUseCase(repo, nil, configRepo, &mockUserRepo{}, nil)

	uc.CreatePlaylist(context.Background(), &biz.Playlist{
		Title: "Playlist 1", UserID: "user-001", IsPublic: true,
	})
	uc.CreatePlaylist(context.Background(), &biz.Playlist{
		Title: "Playlist 2", UserID: "user-002", IsPublic: true,
	})

	list, total, err := uc.ListPlaylists(context.Background(), 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Equal(t, 2, len(list))
}
