package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/origadmin/toolkits/crypto/hash"
	hashtypes "github.com/origadmin/toolkits/crypto/hash/types"

	"origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/migrate"
	"origadmin/application/origcms/internal/server"
	contentbiz "origadmin/application/origcms/internal/svc-content/biz"
	contentdata "origadmin/application/origcms/internal/svc-content/data"
	mediabiz "origadmin/application/origcms/internal/svc-media/biz"
	mediadata "origadmin/application/origcms/internal/svc-media/data"
	"origadmin/application/origcms/internal/svc-user/biz"
	"origadmin/application/origcms/internal/svc-user/data"
	systemdata "origadmin/application/origcms/internal/svc-system/data"
)

const (
	TestJWTSecret = "test-secret-for-integration-tests-only"
	TestJWTExpiry = 24 * time.Hour
)

// TestRole represents user roles for testing
type TestRole string

const (
	RoleGuest TestRole = "guest"
	RoleUser  TestRole = "user"
	RoleStaff TestRole = "staff"
	RoleAdmin TestRole = "admin"
)

// TestUser represents a test user
type TestUser struct {
	ID       string
	Username string
	Password string
	Role     TestRole
	IsStaff  bool
	Token    string
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

	// Initialize dependencies - use standard logger that implements log.Logger interface
	logger := log.DefaultLogger
	hasher, err := hash.NewCrypto(hashtypes.BCRYPT)
	if err != nil {
		t.Fatalf("failed to create hasher: %v", err)
	}

	// Add refresh token TTL parameter
	TestRefreshTokenExpiry := 7 * 24 * time.Hour
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
	uploadUC := mediabiz.NewUploadUseCase(uploadRepo, mediaRepo, profileRepo, taskRepo, mediaUC, storage, logger)
	uploadUC.SetPublisher(mockPub)

	// Content layer
	contentDB := contentdata.NewData(db)
	categoryRepo := contentdata.NewCategoryRepo(contentDB, logger)
	tagRepo := contentdata.NewTagRepo(contentDB, logger)
	commentRepo := contentdata.NewCommentRepo(contentDB, logger)
	playlistRepo := contentdata.NewPlaylistRepo(contentDB, logger)
	channelRepo := contentdata.NewChannelRepo(contentDB, logger)
	feedRepo := contentdata.NewFeedRepo(contentDB, logger)
	likeRepo := contentdata.NewLikeRepo(contentDB, logger)
	favoriteRepo := contentdata.NewFavoriteRepo(contentDB, logger)
	notificationRepo := contentdata.NewNotificationRepo(contentDB, logger)

	categoryTagUC := contentbiz.NewCategoryTagUseCase(categoryRepo, tagRepo, logger)
	commentUC := contentbiz.NewCommentUseCase(commentRepo, mediaUC, logger)
	playlistChannelUC := contentbiz.NewPlaylistChannelUseCase(playlistRepo, channelRepo, logger)
	feedUC := contentbiz.NewFeedUseCase(feedRepo, logger)
	likeFavoriteUC := contentbiz.NewLikeFavoriteUseCase(likeRepo, favoriteRepo, mediaUC, logger)
	notificationUC := contentbiz.NewNotificationUseCase(notificationRepo, logger)

	// Create handlers
	authHandler := server.NewAuthHandler(userUC, jwtMgr)
	userHandler := server.NewUserHandler(userUC, jwtMgr)
	mediaHandler := server.NewMediaHandler(jwtMgr, mediaUC, uploadUC, likeFavoriteUC)
	uploadHandler := server.NewUploadHandler(uploadUC, jwtMgr)
	categoryHandler := server.NewCategoryHandler(categoryTagUC)
	tagHandler := server.NewTagHandler(categoryTagUC)
	commentHandler := server.NewCommentHandler(commentUC, jwtMgr)
	playlistHandler := server.NewPlaylistHandler(playlistChannelUC, jwtMgr)
	feedHandler := server.NewFeedHandler(feedUC)
	notificationHandler := server.NewNotificationHandler(notificationUC, jwtMgr)
	channelHandler := server.NewChannelHandler(playlistChannelUC, jwtMgr)
	shareHandler := server.NewShareHandler(likeFavoriteUC, jwtMgr)

	// Stats repo for stats handler
	statsRepo := systemdata.NewStatsRepo(db)
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

	// Create MeHandler
	meHandler := server.NewMeHandler(userUC, likeFavoriteUC, playlistChannelUC, jwtMgr)

	// Register routes
	server.RegisterRoutes(router,
		authHandler,
		userHandler,
		mediaHandler,
		uploadHandler,
		categoryHandler,
		tagHandler,
		commentHandler,
		playlistHandler,
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
		role     TestRole
		username string
		password string
		isStaff  bool
		roleStr  string
	}{
		{RoleAdmin, "admin", "admin123", true, "admin"},
		{RoleStaff, "staff", "staff123", true, "staff"},
		{RoleUser, "user1", "user123", false, "user"},
		{RoleUser, "user2", "user456", false, "user"},
	}

	for _, u := range users {
		hashedPassword, _ := userUC.HashPassword(u.password)
		created, err := userUC.CreateUser(ctx, &types.User{
			Username: u.username,
			Email:    u.username + "@example.com",
			IsStaff:  u.isStaff,
		}, hashedPassword)
		if err != nil {
			t.Fatalf("failed to create test user %s: %v", u.username, err)
		}

		// Update user with proper details
		user, _ := userUC.GetUserEntity(ctx, created.Uuid)
		if user != nil {
			// Set role if needed
		}

		// Generate token
		token, err := ts.JWTMgr.Generate(created.Uuid, u.username, u.isStaff, u.roleStr)
		if err != nil {
			t.Fatalf("failed to generate token for %s: %v", u.username, err)
		}

		testUser := &TestUser{
			ID:       created.Uuid,
			Username: u.username,
			Password: u.password,
			Role:     u.role,
			IsStaff:  u.isStaff,
			Token:    token,
		}

		ts.Users[u.role] = testUser
		// Also store by username for specific user access
		ts.Users[TestRole(u.username)] = testUser
	}
}

// Cleanup closes the test server and cleans up resources
func (ts *TestServer) Cleanup() {
	if ts.Server != nil {
		ts.Server.Close()
	}
	if ts.DB != nil {
		ts.DB.Close()
	}
	// Clean up test uploads
	os.RemoveAll("./data/test-uploads")
}

// RequestOptions holds options for making HTTP requests
type RequestOptions struct {
	Method      string
	Path        string
	Body        interface{}
	Token       string
	ContentType string
}

// MakeRequest performs an HTTP request to the test server
func (ts *TestServer) MakeRequest(opts RequestOptions) (*http.Response, []byte, error) {
	url := ts.Server.URL + ts.BaseURL + opts.Path

	var body io.Reader
	if opts.Body != nil {
		switch v := opts.Body.(type) {
		case string:
			body = bytes.NewBufferString(v)
		case []byte:
			body = bytes.NewBuffer(v)
		default:
			jsonBody, err := json.Marshal(opts.Body)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to marshal body: %w", err)
			}
			body = bytes.NewBuffer(jsonBody)
		}
	}

	req, err := http.NewRequest(opts.Method, url, body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if opts.Token != "" {
		req.Header.Set("Authorization", "Bearer "+opts.Token)
	}
	if opts.ContentType != "" {
		req.Header.Set("Content-Type", opts.ContentType)
	} else if opts.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return resp, respBody, nil
}

// GetToken returns the token for a specific role or user
func (ts *TestServer) GetToken(role TestRole) string {
	if user, ok := ts.Users[role]; ok {
		return user.Token
	}
	return ""
}

// ParseResponse parses JSON response into the target struct
func ParseResponse(data []byte, target interface{}) error {
	return json.Unmarshal(data, target)
}

// AssertStatus checks if the response status matches expected
func AssertStatus(t *testing.T, resp *http.Response, expected int) {
	if resp.StatusCode != expected {
		t.Errorf("Expected status %d, got %d", expected, resp.StatusCode)
	}
}

// AssertJSON checks if the response body contains expected fields
func AssertJSON(t *testing.T, body []byte, checks map[string]interface{}) {
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
		return
	}

	for key, expected := range checks {
		actual, ok := result[key]
		if !ok {
			t.Errorf("Expected key '%s' not found in response", key)
			continue
		}
		if actual != expected {
			t.Errorf("For key '%s': expected %v (%T), got %v (%T)",
				key, expected, expected, actual, actual)
		}
	}
}

// CreateMultipartRequest creates a multipart form request for file uploads
func CreateMultipartRequest(t *testing.T, filePath string, fields map[string]string) ([]byte, string) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add file if provided
	if filePath != "" {
		f, err := os.Open(filePath)
		if err != nil {
			t.Fatalf("failed to open file: %v", err)
		}
		defer f.Close()

		fw, err := w.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}

		if _, err = io.Copy(fw, f); err != nil {
			t.Fatalf("failed to copy file: %v", err)
		}
	}

	// Add other fields
	for key, val := range fields {
		if err := w.WriteField(key, val); err != nil {
			t.Fatalf("failed to write field: %v", err)
		}
	}

	w.Close()

	return b.Bytes(), w.FormDataContentType()
}

// mockPublisher implements a mock message publisher for tests
type mockPublisher struct{}

func (m *mockPublisher) Publish(topic string, messages ...*message.Message) error {
	return nil
}

func (m *mockPublisher) Close() error {
	return nil
}
