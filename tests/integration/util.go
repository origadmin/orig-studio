/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package integration

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/require"

	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/auth/hash"
	hashtypes "origadmin/application/origcms/internal/auth/hash/types"
	contentbiz "origadmin/application/origcms/internal/content/biz"
	contentdata "origadmin/application/origcms/internal/content/data"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/migrate"
	"origadmin/application/origcms/internal/data/system"
	"origadmin/application/origcms/internal/server"
	mediabiz "origadmin/application/origcms/internal/svc-media/biz"
	mediadata "origadmin/application/origcms/internal/svc-media/data"
	"origadmin/application/origcms/internal/svc-user/biz"
	"origadmin/application/origcms/internal/svc-user/data"
)

const (
	TestJWTSecret          = "test-secret-key-for-testing-only"
	TestJWTExpiry          = 24 * time.Hour
	TestRefreshTokenExpiry = 7 * 24 * time.Hour
)

type TestRole string

const (
	RoleAdmin TestRole = "admin"
	RoleUser  TestRole = "user"
	RoleGuest TestRole = "guest"
)

type TestUser struct {
	ID        string
	Email     string
	Username  string
	Password  string
	Role      string
	Token     string
	ExpiresAt time.Time
}

// mockPublisher implements pubsub.Publisher interface for testing
type mockPublisher struct {
	messages []*mockMessage
}

type mockMessage struct {
	topic string
	data  []byte
}

func (m *mockPublisher) Publish(topic string, data []byte) error {
	m.messages = append(m.messages, &mockMessage{topic: topic, data: data})
	return nil
}

// TestServer encapsulates the test environment
type TestServer struct {
	Router  *gin.Engine
	DB      *entity.Client
	JWTMgr  *auth.Manager
	Hasher  hash.Crypto
	Users   map[TestRole]*TestUser
	BaseURL string
	Server  *httptest.Server
}

// SetupTestServer creates a complete test server with database and all handlers
func SetupTestServer(t *testing.T) *TestServer {
	gin.SetMode(gin.TestMode)

	// Create in-memory SQLite database
	db, err := entity.Open("sqlite3", "file::memory:?cache=shared&_fk=1")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Run migrations
	ctx := context.Background()
	if err := db.Schema.Create(ctx, migrate.WithForeignKeys(false)); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	// Initialize dependencies
	logger := log.NewStdLogger(nil)
	hasher, err := hash.NewCrypto(hashtypes.BCRYPT)
	if err != nil {
		t.Fatalf("failed to create hasher: %v", err)
	}

	jwtMgr := auth.NewManager(TestJWTSecret, TestJWTExpiry, TestRefreshTokenExpiry)

	// Seed encode profiles
	if err := mediadata.SeedEncodeProfiles(ctx, db); err != nil {
		t.Fatalf("failed to seed encode profiles: %v", err)
	}

	// Initialize repositories and use cases
	userRepo := data.NewUserRepo(db)
	userUC := biz.NewUserUseCase(userRepo, hasher, logger)

	mediaRepo := mediadata.NewMediaRepo(db)
	profileRepo := mediadata.NewEncodeProfileRepo(db)
	taskRepo := mediadata.NewEncodingTaskRepo(db)
	uploadRepo := mediadata.NewUploadRepo(db, logger)
	storage := mediadata.NewLocalStorage("./data/test-uploads", logger)

	// Create mock pub/sub (no actual transcode in tests)
	mockPub := &mockPublisher{}

	mediaUC := mediabiz.NewMediaUseCase(mediaRepo, profileRepo, taskRepo, storage, mockPub, logger)
	uploadUC := mediabiz.NewUploadUseCase(uploadRepo, mediaRepo, profileRepo, taskRepo, mediaUC, storage, 5*1024*1024, logger)
	uploadUC.SetPublisher(mockPub)

	// Content layer
	contentDB := contentdata.NewData(db)
	categoryRepo := contentdata.NewCategoryRepo(contentDB, logger)
	tagRepo := contentdata.NewTagRepo(contentDB, logger)
	channelRepo := contentdata.NewChannelRepo(contentDB, logger)
	feedRepo := contentdata.NewFeedRepo(contentDB, logger)
	likeRepo := contentdata.NewLikeRepo(contentDB, logger)
	favoriteRepo := contentdata.NewFavoriteRepo(contentDB, logger)
	notificationRepo := contentdata.NewNotificationRepo(contentDB, logger)

	categoryTagUC := contentbiz.NewCategoryTagUseCase(categoryRepo, tagRepo, logger)
	likeFavoriteUC := contentbiz.NewLikeFavoriteUseCase(likeRepo, favoriteRepo, mediaUC, logger)
	playlistChannelUC := contentbiz.NewPlaylistChannelUseCase(nil, channelRepo, logger)
	feedUC := contentbiz.NewFeedUseCase(feedRepo, logger)
	notificationUC := contentbiz.NewNotificationUseCase(notificationRepo, logger)

	// Create handlers - only use handlers that actually exist
	authHandler := server.NewAuthHandler(userUC, jwtMgr)
	userHandler := server.NewUserHandler(userUC, jwtMgr)
	mediaHandler := server.NewMediaHandler(jwtMgr, mediaUC, uploadUC, likeFavoriteUC)
	uploadHandler := server.NewUploadHandler(uploadUC, jwtMgr)
	categoryHandler := server.NewCategoryHandler(categoryTagUC)
	tagHandler := server.NewTagHandler(categoryTagUC)
	feedHandler := server.NewFeedHandler(feedUC)
	notificationHandler := server.NewNotificationHandler(notificationUC, jwtMgr)
	channelHandler := server.NewChannelHandler(playlistChannelUC, jwtMgr)
	shareHandler := server.NewShareHandler(likeFavoriteUC, jwtMgr)
	meHandler := server.NewMeHandler(userUC, likeFavoriteUC, playlistChannelUC, jwtMgr)

	// Stats repo for stats handler
	statsRepo := system.NewStatsRepo(db)
	statsHandler := server.NewStatsHandler(mediaUC, likeFavoriteUC, statsRepo, jwtMgr)
	searchHandler := server.NewSearchHandler(mediaUC)

	// Setup router
	router := gin.New()
	router.Use(gin.Recovery())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Register routes using server.RegisterRoutes
	server.RegisterRoutes(router,
		authHandler,
		userHandler,
		mediaHandler,
		uploadHandler,
		categoryHandler,
		tagHandler,
		feedHandler,
		notificationHandler,
		channelHandler,
		shareHandler,
		statsHandler,
		searchHandler,
		meHandler,
	)

	// Create test server
	ts := &TestServer{
		Router:  router,
		DB:      db,
		JWTMgr:  jwtMgr,
		Hasher:  hasher,
		Users:   make(map[TestRole]*TestUser),
		BaseURL: "/api/v1",
	}

	// Create test users
	ts.createTestUsers(ctx, t, userUC)

	// Create httptest server
	ts.Server = httptest.NewServer(router)

	return ts
}

// createTestUsers creates test users for each role
func (ts *TestServer) createTestUsers(ctx context.Context, t *testing.T, userUC *biz.UserUseCase) {
	users := []struct {
		role  TestRole
		email string
		name  string
		pass  string
	}{
		{RoleAdmin, "admin@test.com", "admin", "admin123"},
		{RoleUser, "user@test.com", "user1", "user123"},
		{RoleGuest, "guest@test.com", "guest", "guest123"},
	}

	for _, u := range users {
		// Create user
		_, err := userUC.Register(ctx, &biz.RegisterRequest{
			Email:    u.email,
			Username: u.name,
			Password: u.pass,
		})
		if err != nil {
			t.Fatalf("failed to create test user %s: %v", u.role, err)
		}

		// Login to get token
		resp, err := userUC.Login(ctx, &biz.LoginRequest{
			Email:    u.email,
			Password: u.pass,
		})
		if err != nil {
			t.Fatalf("failed to login test user %s: %v", u.role, err)
		}

		// Get user by email to get ID
		user, err := userUC.GetByEmail(ctx, u.email)
		if err != nil {
			t.Fatalf("failed to get test user %s: %v", u.role, err)
		}

		ts.Users[u.role] = &TestUser{
			ID:        user.Id,
			Email:     u.email,
			Username:  u.name,
			Password:  u.pass,
			Role:      string(u.role),
			Token:     resp.Token,
			ExpiresAt: time.Now().Add(TestJWTExpiry),
		}
	}
}

// GetAuthToken returns a valid auth token for the given role
func (ts *TestServer) GetAuthToken(role TestRole) string {
	return "Bearer " + ts.Users[role].Token
}

// Cleanup shuts down the test server
func (ts *TestServer) Cleanup() {
	if ts.Server != nil {
		ts.Server.Close()
	}
	if ts.DB != nil {
		ts.DB.Close()
	}
}

// AssertStatusCode asserts the response status code
func AssertStatusCode(t *testing.T, resp *httptest.ResponseRecorder, expected int) {
	require.Equal(t, expected, resp.Code, "unexpected status code: %s", resp.Body.String())
}

// AssertSuccess asserts a successful response (2xx)
func AssertSuccess(t *testing.T, resp *httptest.ResponseRecorder) {
	require.True(t, resp.Code >= 200 && resp.Code < 300, "expected success, got %d: %s", resp.Code, resp.Body.String())
}

// AssertError asserts an error response (4xx or 5xx)
func AssertError(t *testing.T, resp *httptest.ResponseRecorder) {
	require.True(t, resp.Code >= 400, "expected error, got %d", resp.Code)
}
