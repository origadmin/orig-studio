/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/require"

	"origadmin/application/origcms/api/gen/v1/types"
	_ "github.com/sqlite3ent/sqlite3"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/migrate"
	"origadmin/application/origcms/internal/server"
	contentbiz "origadmin/application/origcms/internal/svc-content/biz"
	contentdata "origadmin/application/origcms/internal/svc-content/data"
	mediabiz "origadmin/application/origcms/internal/svc-media/biz"
	mediadata "origadmin/application/origcms/internal/svc-media/data"
	systemdata "origadmin/application/origcms/internal/svc-system/data"
	"origadmin/application/origcms/internal/svc-user/biz"
	"origadmin/application/origcms/internal/svc-user/data"
	"github.com/origadmin/toolkits/crypto/hash"
	hashtypes "github.com/origadmin/toolkits/crypto/hash/types"
)

const (
	TestJWTSecret          = "test-secret-key-for-testing-only"
	TestJWTExpiry          = 24 * time.Hour
	TestRefreshTokenExpiry = 7 * 24 * time.Hour
)

type TestRole string

const (
	RoleAdmin TestRole = "admin"
	RoleStaff TestRole = "staff"
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

type mockPublisher struct {
	messages []*mockMessage
}

type mockMessage struct {
	topic string
	data  []byte
}

func (m *mockPublisher) Publish(topic string, msgs ...*message.Message) error {
	for _, msg := range msgs {
		m.messages = append(m.messages, &mockMessage{topic: topic, data: msg.Payload})
	}
	return nil
}

func (m *mockPublisher) Close() error {
	return nil
}

type TestServer struct {
	Router  *gin.Engine
	DB      *entity.Client
	JWTMgr  *auth.Manager
	Hasher  hash.Crypto
	Users   map[TestRole]*TestUser
	BaseURL string
	Server  *httptest.Server
}

func SetupTestServer(t *testing.T) *TestServer {
	gin.SetMode(gin.TestMode)

	db, err := entity.Open("sqlite3", "file::memory:?cache=shared&_fk=1")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	ctx := context.Background()
	if err := db.Schema.Create(ctx, migrate.WithForeignKeys(false)); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	logger := log.NewStdLogger(nil)
	hasher, err := hash.NewCrypto(hashtypes.BCRYPT)
	if err != nil {
		t.Fatalf("failed to create hasher: %v", err)
	}

	jwtMgr := auth.NewManager(TestJWTSecret, TestJWTExpiry, TestRefreshTokenExpiry)

	if err := mediadata.SeedEncodeProfiles(ctx, db); err != nil {
		t.Fatalf("failed to seed encode profiles: %v", err)
	}

	userRepo := data.NewUserRepo(db)
	userUC := biz.NewUserUseCase(userRepo, hasher, logger)

	mediaRepo := mediadata.NewMediaRepo(db)
	profileRepo := mediadata.NewEncodeProfileRepo(db)
	taskRepo := mediadata.NewEncodingTaskRepo(db)
	reviewLogRepo := mediadata.NewReviewLogRepo(db)
	uploadRepo := mediadata.NewUploadRepo(db, logger)
	storage := mediadata.NewLocalStorage("./data/test-uploads")

	mockPub := &mockPublisher{}

	mediaUC := mediabiz.NewMediaUseCase(mediaRepo, profileRepo, taskRepo, reviewLogRepo, storage, mockPub, logger, nil)
	uploadUC := mediabiz.NewUploadUseCase(uploadRepo, mediaRepo, profileRepo, taskRepo, mediaUC, storage, 5*1024*1024, logger)
	uploadUC.SetPublisher(mockPub)

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

	authHandler := server.NewAuthHandler(userUC, jwtMgr)
	userHandler := server.NewUserHandler(userUC, jwtMgr)
	mediaHandler := server.NewMediaHandler(jwtMgr, mediaUC, uploadUC, likeFavoriteUC, playlistChannelUC, userUC, db)
	uploadHandler := server.NewUploadHandler(uploadUC, jwtMgr, logger)
	categoryHandler := server.NewCategoryHandler(categoryTagUC)
	tagHandler := server.NewTagHandler(categoryTagUC)
	feedHandler := server.NewFeedHandler(feedUC)
	notificationHandler := server.NewNotificationHandler(notificationUC, jwtMgr)
	channelHandler := server.NewChannelHandler(playlistChannelUC, jwtMgr, db)
	shareHandler := server.NewShareHandler(likeFavoriteUC, jwtMgr)
	meHandler := server.NewMeHandler(userUC, likeFavoriteUC, playlistChannelUC, jwtMgr)

	statsRepo := systemdata.NewStatsRepo(db)
	statsHandler := server.NewStatsHandler(mediaUC, likeFavoriteUC, statsRepo, jwtMgr)
	searchHandler := server.NewSearchHandler(mediaUC)

	router := gin.New()
	router.Use(gin.Recovery())

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

	srv := server.NewServer(
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
		nil,
		statsHandler,
		searchHandler,
		meHandler,
		nil,
		nil,
		nil,
		nil,
		nil,
		db,
		jwtMgr,
		"./data/uploads", // storageBasePath
	)
	srv.RegisterRoutes(router)

	ts := &TestServer{
		Router:  router,
		DB:      db,
		JWTMgr:  jwtMgr,
		Hasher:  hasher,
		Users:   make(map[TestRole]*TestUser),
		BaseURL: "/api/v1",
	}

	ts.createTestUsers(ctx, t, userUC)

	ts.Server = httptest.NewServer(router)

	return ts
}

func (ts *TestServer) createTestUsers(ctx context.Context, t *testing.T, userUC *biz.UserUseCase) {
	users := []struct {
		role    TestRole
		email   string
		name    string
		pass    string
		isStaff bool
		roleStr string
	}{
		{RoleAdmin, "admin@test.com", "admin", "admin123", true, "admin"},
		{RoleUser, "user@test.com", "user1", "user123", false, "user"},
		{RoleGuest, "guest@test.com", "guest", "guest123", false, "guest"},
	}

	for _, u := range users {
		hashedPass, err := userUC.HashPassword(u.pass)
		if err != nil {
			t.Fatalf("failed to hash password for %s: %v", u.role, err)
		}

		created, err := userUC.CreateUser(ctx, &types.User{
			Username: u.name,
			Email:    u.email,
			IsStaff:  u.isStaff,
		}, hashedPass)
		if err != nil {
			t.Fatalf("failed to create test user %s: %v", u.role, err)
		}

		if err := userUC.SetUserRole(ctx, created.Id, u.roleStr); err != nil {
			t.Fatalf("failed to set role for test user %s: %v", u.role, err)
		}

		token, err := ts.JWTMgr.Generate(created.Id, u.name, u.isStaff, u.roleStr)
		if err != nil {
			t.Fatalf("failed to generate token for test user %s: %v", u.role, err)
		}

		ts.Users[u.role] = &TestUser{
			ID:        created.Id,
			Email:     u.email,
			Username:  u.name,
			Password:  u.pass,
			Role:      u.roleStr,
			Token:     token,
			ExpiresAt: time.Now().Add(TestJWTExpiry),
		}
	}
}

func (ts *TestServer) GetAuthToken(role TestRole) string {
	return "Bearer " + ts.Users[role].Token
}

func (ts *TestServer) Cleanup() {
	if ts.Server != nil {
		ts.Server.Close()
	}
	if ts.DB != nil {
		ts.DB.Close()
	}
}

func AssertStatusCode(t *testing.T, resp *httptest.ResponseRecorder, expected int) {
	require.Equal(t, expected, resp.Code, "unexpected status code: %s", resp.Body.String())
}

func AssertSuccess(t *testing.T, resp *httptest.ResponseRecorder) {
	require.True(t, resp.Code >= 200 && resp.Code < 300, "expected success, got %d: %s", resp.Code, resp.Body.String())
}

func AssertError(t *testing.T, resp *httptest.ResponseRecorder) {
	require.True(t, resp.Code >= 400, "expected error, got %d", resp.Code)
}

type RequestOptions struct {
	Method string
	Path   string
	Body   interface{}
	Token  string
}

func (ts *TestServer) MakeRequest(opts RequestOptions) (*httptest.ResponseRecorder, *bytes.Buffer, error) {
	var bodyReader io.Reader
	if opts.Body != nil {
		jsonBody, err := json.Marshal(opts.Body)
		if err != nil {
			return nil, nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req := httptest.NewRequest(opts.Method, ts.BaseURL+opts.Path, bodyReader)
	if opts.Token != "" {
		req.Header.Set("Authorization", opts.Token)
	}
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	ts.Router.ServeHTTP(rec, req)
	return rec, rec.Body, nil
}

func (ts *TestServer) GetToken(role TestRole) string {
	return "Bearer " + ts.Users[role].Token
}

func AssertStatus(t *testing.T, resp *httptest.ResponseRecorder, expected int) {
	require.Equal(t, expected, resp.Code, "unexpected status code: %s", resp.Body.String())
}

func ParseResponse(body *bytes.Buffer, v interface{}) error {
	return json.Unmarshal(body.Bytes(), v)
}

func AssertJSON(t *testing.T, body *bytes.Buffer, expected map[string]interface{}) {
	var result map[string]interface{}
	if err := json.Unmarshal(body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	for key, expectedVal := range expected {
		actualVal, ok := result[key]
		if !ok {
			t.Errorf("expected key %q in response", key)
		} else if actualVal != expectedVal {
			t.Errorf("expected %q=%v, got %v", key, expectedVal, actualVal)
		}
	}
}
