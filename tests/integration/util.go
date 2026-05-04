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
	"origadmin/application/origcms/internal/infra/auth"
	authbiz "origadmin/application/origcms/internal/features/auth/biz"
	authdal "origadmin/application/origcms/internal/features/auth/dal"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/migrate"
	"origadmin/application/origcms/internal/handler"
	"origadmin/application/origcms/internal/server"
	adminbiz "origadmin/application/origcms/internal/features/admin/biz"
	admindal "origadmin/application/origcms/internal/features/admin/dal"
	adminservice "origadmin/application/origcms/internal/features/admin/service"
	authservice "origadmin/application/origcms/internal/features/auth/service"
	contentbiz "origadmin/application/origcms/internal/features/content/biz"
	contentdal "origadmin/application/origcms/internal/features/content/dal"
	contentservice "origadmin/application/origcms/internal/features/content/service"
	mediabiz "origadmin/application/origcms/internal/features/media/biz"
	mediadal "origadmin/application/origcms/internal/features/media/dal"
	mediaservice "origadmin/application/origcms/internal/features/media/service"
	systembiz "origadmin/application/origcms/internal/features/system/biz"
	systemdal "origadmin/application/origcms/internal/features/system/dal"
	systemservice "origadmin/application/origcms/internal/features/system/service"
	"origadmin/application/origcms/internal/features/user/biz"
	"origadmin/application/origcms/internal/features/user/dal"
	userservice "origadmin/application/origcms/internal/features/user/service"
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
	RoleAdmin  TestRole = "admin"
	RoleEditor TestRole = "editor"
	RoleUser   TestRole = "user"
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

	db, err := entity.Open("sqlite3", "file::memory:?_fk=1")
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

	if err := mediadal.SeedEncodeProfiles(ctx, db); err != nil {
		t.Fatalf("failed to seed encode profiles: %v", err)
	}

	userRepo := dal.NewUserRepo(db)
	userUC := biz.NewUserUseCase(userRepo, hasher, logger)

	mediaRepo := mediadal.NewMediaRepo(db)
	profileRepo := mediadal.NewEncodeProfileRepo(db)
	taskRepo := mediadal.NewEncodingTaskRepo(db)
	reviewLogRepo := mediadal.NewReviewLogRepo(db)
	uploadRepo := mediadal.NewUploadRepo(db, logger)
	storage := mediadal.NewLocalStorage("./data/test-uploads")

	mockPub := &mockPublisher{}

	mediaUC := mediabiz.NewMediaUseCase(mediaRepo, profileRepo, taskRepo, reviewLogRepo, storage, mockPub, logger, nil)
	uploadUC := mediabiz.NewUploadUseCase(uploadRepo, mediaRepo, profileRepo, taskRepo, mediaUC, storage, 5*1024*1024, logger)
	uploadUC.SetPublisher(mockPub)

	contentDB := contentdal.NewData(db)
	categoryRepo := contentdal.NewCategoryRepo(contentDB, logger)
	tagRepo := contentdal.NewTagRepo(contentDB, logger)
	channelRepo := contentdal.NewChannelRepo(contentDB, logger)
	feedRepo := contentdal.NewFeedRepo(contentDB, logger)
	likeRepo := contentdal.NewLikeRepo(contentDB, logger)
	favoriteRepo := contentdal.NewFavoriteRepo(contentDB, logger)
	notificationRepo := contentdal.NewNotificationRepo(contentDB, logger)

	categoryTagUC := contentbiz.NewCategoryTagUseCase(categoryRepo, tagRepo, logger)
	likeFavoriteUC := contentbiz.NewLikeFavoriteUseCase(likeRepo, favoriteRepo, mediaUC, logger)
	playlistRepo := contentdal.NewPlaylistRepo(contentDB, logger)
	systemConfigRepo := contentdal.NewSystemConfigRepo(contentDB, logger)
	channelUserRepo := contentdal.NewChannelUserRepo(contentDB, logger)
	playlistChannelUC := contentbiz.NewPlaylistChannelUseCase(playlistRepo, channelRepo, systemConfigRepo, channelUserRepo, logger)
	feedUC := contentbiz.NewFeedUseCase(feedRepo, logger)
	notificationUC := contentbiz.NewNotificationUseCase(notificationRepo, logger)

	authHandler := authservice.NewAuthHandler(userUC, jwtMgr)
	userHandler := userservice.NewUserHandler(userUC, jwtMgr)
	mediaHandler := mediaservice.NewMediaHandler(jwtMgr, mediaUC, uploadUC, likeFavoriteUC, playlistChannelUC, userUC, nil, nil, nil)
	uploadHandler := mediaservice.NewUploadHandler(uploadUC, jwtMgr, logger)
	categoryHandler := contentservice.NewCategoryHandler(categoryTagUC, jwtMgr)
	tagHandler := contentservice.NewTagHandler(categoryTagUC, jwtMgr)
	feedHandler := contentservice.NewFeedHandler(feedUC)
	notificationHandler := contentservice.NewNotificationHandler(notificationUC, jwtMgr)
	channelHandler := contentservice.NewChannelHandler(playlistChannelUC, jwtMgr, nil)
	shareHandler := contentservice.NewShareHandler(likeFavoriteUC, jwtMgr)
	meHandler := userservice.NewMeHandler(userUC, likeFavoriteUC, playlistChannelUC, nil, jwtMgr)

	statsRepo := systemdal.NewStatsRepo(db)
	statsHandler := systemservice.NewStatsHandler(mediaUC, likeFavoriteUC, statsRepo, jwtMgr)
	searchHandler := mediaservice.NewSearchHandler(mediaUC)

	// System handler
	settingRepo := systemdal.NewSettingRepo(db)
	settingUC := systembiz.NewSettingUseCase(settingRepo)
	systemHandler := systemservice.NewSystemHandler(jwtMgr, statsRepo, settingUC)

	// Admin handler (needs TagService from svc-admin)
	adminTagRepo := admindal.NewTagRepository(db)
	tagUC := adminbiz.NewTagUseCase(adminTagRepo)
	tagService := adminservice.NewTagService(tagUC)

	// Article use case for admin content management
	articleRepo := contentdal.NewArticleRepo(contentDB, logger)
	articleUC := contentbiz.NewArticleUseCase(articleRepo, logger)

	adminHandler := adminservice.NewAdminHandler(jwtMgr, mediaUC, nil, playlistChannelUC, tagService, settingUC, categoryTagUC, articleUC, userUC, nil)

	// Explore handler
	exploreHandler := contentservice.NewExploreHandler(db)

	// Comment like use case
	commentLikeRepo := contentdal.NewCommentLikeRepo(contentDB, logger)
	_ = contentbiz.NewCommentLikeUseCase(commentLikeRepo, logger)

	// Comment moderation handler
	commentModRepo := contentdal.NewCommentModerationRepo(contentDB, logger)
	commentReportRepo := contentdal.NewCommentReportRepo(contentDB, logger)
	commentModUC := contentbiz.NewCommentModerationUseCase(commentModRepo, commentReportRepo, settingUC, logger)
	commentModerationHandler := contentservice.NewCommentModerationHandler(commentModUC, jwtMgr)

	// Permission handler
	authData := authdal.NewData(db)
	permGroupRepo := authdal.NewPermissionGroupRepo(authData, logger)
	permMemberRepo := authdal.NewGroupMemberRepo(authData, logger)
	permUserRepo := authdal.NewUserPermRepo(authData, logger)
	permUC := authbiz.NewPermissionUseCase(permGroupRepo, permMemberRepo, permUserRepo, logger)
	permissionHandler := authservice.NewPermissionHandler(permUC, jwtMgr)

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
		[]handler.Module{
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
			systemHandler,
			statsHandler,
			searchHandler,
			meHandler,
			adminHandler,
			exploreHandler,
			commentModerationHandler,
			permissionHandler,
		},
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
		{RoleEditor, "editor@test.com", "editor", "editor123", false, "editor"},
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

// GetResponseData extracts the "data" field from a standard API response.
// Most API handlers return {"code": 0, "message": "ok", "data": {...}}.
// If the "data" field exists and is a map, it returns that map.
// Otherwise, it returns the top-level result (for handlers that don't use the wrapper).
func GetResponseData(body *bytes.Buffer) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := ParseResponse(body, &result); err != nil {
		return nil, err
	}

	// Check if response has data field
	if data, ok := result["data"]; ok {
		if dataMap, ok := data.(map[string]interface{}); ok {
			return dataMap, nil
		}
	}
	return result, nil
}
