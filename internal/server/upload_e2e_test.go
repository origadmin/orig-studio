/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/sqlite3ent/sqlite3"

	"github.com/go-kratos/kratos/v2/log"

	pb "origadmin/application/origcms/api/gen/v1/upload"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/svc-media/biz"
	"origadmin/application/origcms/internal/svc-media/data"
)

func TestUploadE2E(t *testing.T) {
	// 1. Setup Environment
	gin.SetMode(gin.TestMode)

	// Create required directories for the test
	// The handler uses 'data/uploads' and 'data/uploads/.chunks'
	require.NoError(t, os.MkdirAll("data/uploads/.chunks", 0o755))
	defer os.RemoveAll("data") // Cleanup after test

	// Initialize In-memory SQLite
	client, err := entity.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	defer client.Close()
	require.NoError(t, client.Schema.Create(context.Background()))

	// Initialize JWT Manager
	jwtMgr := auth.NewManager("secret-key", 24*time.Hour)

	// Setup svc-media dependencies
	logger := log.NewStdLogger(os.Stderr)
	uploadRepo := data.NewUploadRepo(client, logger)
	mediaRepo := data.NewMediaRepo(client)
	profileRepo := data.NewEncodeProfileRepo(client)
	taskRepo := data.NewEncodingTaskRepo(client)
	mediaUC := biz.NewMediaUseCase(mediaRepo, profileRepo, taskRepo, logger)
	storage := data.NewLocalStorage("data/uploads", logger)

	uploadUC := biz.NewUploadUseCase(uploadRepo, mediaRepo, profileRepo, taskRepo, mediaUC, storage, logger)

	// Setup Router
	router := gin.Default()
	RegisterRoutes(router, client, jwtMgr, uploadUC, mediaUC)

	// 2. Register & Login to get token
	username := "testuser"
	password := "password123"

	// Create user directly in DB
	user, err := client.User.Create().
		SetUsername(username).
		SetPassword(password).
		SetEmail("test@example.com").
		SetName("Test User").
		Save(context.Background())
	require.NoError(t, err)

	// Generate Token manually for testing (simulating a login)
	token, err := jwtMgr.Generate(int64(user.ID), username, true)
	require.NoError(t, err)
	authHeader := "Bearer " + token

	// 3. E2E Upload Flow
	filename := "e2e_video.mp4"
	fileSize := int64(1024 * 1024 * 3) // 3MB, so 2 parts (2MB + 1MB)

	// --- A. Initiate Multipart Upload ---
	initReq := pb.InitiateMultipartUploadRequest{
		Filename:    filename,
		FileSize:    fileSize,
		ContentType: "video/mp4",
		Title:       "My E2E Video",
	}
	body, _ := json.Marshal(&initReq)
	req, _ := http.NewRequest("POST", "/api/v1/uploads/multipart", bytes.NewBuffer(body))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var initResp pb.InitiateMultipartUploadResponse
	err = json.Unmarshal(w.Body.Bytes(), &initResp)
	require.NoError(t, err)
	uploadID := initResp.UploadId
	assert.NotEmpty(t, uploadID)
	assert.Equal(t, int32(2), initResp.TotalParts)

	// --- B. Upload Part 1 (2MB) ---
	part1Data := make([]byte, 2*1024*1024)
	for i := range part1Data {
		part1Data[i] = 'A'
	}

	req, _ = http.NewRequest("POST", fmt.Sprintf("/api/v1/uploads/%s/parts/1", uploadID), bytes.NewBuffer(part1Data))
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// --- C. List Parts (Breakpoint Check) ---
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/uploads/%s/parts", uploadID), nil)
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var listResp pb.ListPartsResponse
	err = json.Unmarshal(w.Body.Bytes(), &listResp)
	require.NoError(t, err)
	assert.Len(t, listResp.Parts, 1)
	assert.Equal(t, int32(1), listResp.Parts[0].PartNumber)

	// --- D. Upload Part 2 (1MB) ---
	part2Data := make([]byte, 1*1024*1024)
	for i := range part2Data {
		part2Data[i] = 'B'
	}
	req, _ = http.NewRequest("POST", fmt.Sprintf("/api/v1/uploads/%s/parts/2", uploadID), bytes.NewBuffer(part2Data))
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// --- E. Complete Upload ---
	completeReq := pb.CompleteMultipartUploadRequest{
		UploadId: uploadID,
		Sha256:   "dummy-sha256",
	}
	body, _ = json.Marshal(&completeReq)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/api/v1/uploads/%s/complete", uploadID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check for 500 errors and print body for debugging if failed
	if w.Code != http.StatusOK {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusOK, w.Code)

	var completeResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &completeResp)
	require.NoError(t, err)
	assert.NotNil(t, completeResp["media"])

	// --- F. Verify Media Record in DB ---
	mediaData, ok := completeResp["media"].(map[string]interface{})
	require.True(t, ok, "media field missing in response")
	mediaID := int64(mediaData["id"].(float64))
	dbMedia, err := client.Media.Get(context.Background(), int(mediaID))
	require.NoError(t, err)
	assert.Equal(t, "My E2E Video", dbMedia.Title)
	assert.Equal(t, "video/mp4", dbMedia.MimeType)

	// Final cleanup
	time.Sleep(100 * time.Millisecond)
}
