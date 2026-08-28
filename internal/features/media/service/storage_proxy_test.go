package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/data/enums"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/require"
)

// fakeStorage implements biz.Storage for testing StorageProxy. Only Download is used.
type fakeStorage struct {
	files map[string][]byte
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{files: make(map[string][]byte)}
}

type fakeReadCloser struct {
	*bytes.Reader
}

func (f *fakeReadCloser) Close() error { return nil }

func (fs *fakeStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	data, ok := fs.files[key]
	if !ok {
		return nil, os.ErrNotExist
	}
	return &fakeReadCloser{Reader: bytes.NewReader(data)}, nil
}

func (fs *fakeStorage) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
	return "", nil
}
func (fs *fakeStorage) DownloadToFile(ctx context.Context, key string, localPath string) error { return nil }
func (fs *fakeStorage) Delete(ctx context.Context, key string) error                           { return nil }
func (fs *fakeStorage) GetURL(ctx context.Context, key string) (string, error)                 { return "", nil }
func (fs *fakeStorage) StorePart(ctx context.Context, uploadID string, partNumber int, r io.Reader, size int64) (string, error) {
	return "", nil
}
func (fs *fakeStorage) MergeParts(ctx context.Context, uploadID string, totalParts int, finalPath string) error {
	return nil
}
func (fs *fakeStorage) DeleteParts(ctx context.Context, uploadID string) error { return nil }
func (fs *fakeStorage) PromoteToOriginal(ctx context.Context, tempPath string) (string, error) {
	return "", nil
}
func (fs *fakeStorage) CleanupTempParts(ctx context.Context, userID, uploadID string) error { return nil }
func (fs *fakeStorage) UploadDir(ctx context.Context, localDir string, keyPrefix string) error {
	return nil
}
func (fs *fakeStorage) DeletePrefix(ctx context.Context, keyPrefix string) error { return nil }
func (fs *fakeStorage) SyncStatus(ctx context.Context, key string) (enums.SyncStatus, error) {
	return enums.SyncStatusLocalOnly, nil
}

// buildTestProxy creates a StorageProxyService backed by a fake storage and a temp
// StoragePaths (required by constructor but not used by the proxy itself).
func buildTestProxy(t *testing.T, fs *fakeStorage) *StorageProxyService {
	t.Helper()
	tmpDir := t.TempDir()
	paths := conf.NewStoragePaths(tmpDir)
	_ = paths
	logger := log.NewStdLogger(os.Stdout)
	return NewStorageProxyService(fs, paths, logger)
}

func TestParseRange(t *testing.T) {
	const fileSize int64 = 1000

	tests := []struct {
		name      string
		header    string
		wantStart int64
		wantEnd   int64
		wantOk    bool
	}{
		{"no prefix", "0-499", 0, 0, false},
		{"simple range", "bytes=0-499", 0, 499, true},
		{"open-ended range", "bytes=500-", 500, 999, true},
		{"suffix length", "bytes=-200", 800, 999, true},
		{"suffix too large", "bytes=-2000", 0, 0, false},
		{"end beyond file", "bytes=900-2000", 0, 0, false},
		{"end < start", "bytes=500-400", 0, 0, false},
		{"start == size", "bytes=1000-", 0, 0, false},
		{"start negative via parse fail", "bytes=abc-499", 0, 0, false},
		{"single byte", "bytes=0-0", 0, 0, true},
		{"last byte", "bytes=999-999", 999, 999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, e, ok := parseRange(tt.header, fileSize)
			require.Equal(t, tt.wantOk, ok, "ok mismatch")
			if ok {
				require.Equal(t, tt.wantStart, s, "start mismatch")
				require.Equal(t, tt.wantEnd, e, "end mismatch")
				require.LessOrEqual(t, s, e)
				require.Less(t, e, fileSize)
			}
		})
	}
}

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		{"video.mp4", "video/mp4"},
		{"playlist.m3u8", "application/vnd.apple.mpegurl"},
		{"segment.ts", "video/mp2t"},
		{"image.jpg", "image/jpeg"},
		{"image.jpeg", "image/jpeg"},
		{"image.png", "image/png"},
		{"anim.gif", "image/gif"},
		{"pic.webp", "image/webp"},
		{"sub.vtt", "text/vtt; charset=utf-8"},
		{"v.webm", "video/webm"},
		{"a.mp3", "audio/mpeg"},
		{"a.aac", "audio/aac"},
		{"a.ogg", "audio/ogg"},
		{"a.wav", "audio/wav"},
		{"i.svg", "image/svg+xml"},
		{"favicon.ico", "image/x-icon"},
		{"data.json", "application/json"},
		{"style.css", "text/css"},
		{"app.js", "application/javascript"},
		{"unknown.bin", "application/octet-stream"},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			require.Equal(t, tt.expected, detectContentType(tt.key))
		})
	}
}

func TestStorageProxy_FullFile(t *testing.T) {
	fs := newFakeStorage()
	content := []byte("hello world, this is a test media file content for range tests")
	fs.files["originals/admin/2026/06/test.mp4"] = content

	proxy := buildTestProxy(t, fs)
	srv := httptest.NewServer(proxy)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/files/originals/admin/2026/06/test.mp4")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "video/mp4", resp.Header.Get("Content-Type"))
	require.Equal(t, "bytes", resp.Header.Get("Accept-Ranges"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, content, body)
}

func TestStorageProxy_RangeRequest(t *testing.T) {
	fs := newFakeStorage()
	content := bytes.Repeat([]byte("ABCDEFGHIJ"), 200) // 2000 bytes
	require.Len(t, content, 2000)
	fs.files["originals/u/2026/06/video.mp4"] = content

	proxy := buildTestProxy(t, fs)
	srv := httptest.NewServer(proxy)
	defer srv.Close()

	// Range: bytes=0-999 (first 1000 bytes)
	req, _ := http.NewRequest("GET", srv.URL+"/files/originals/u/2026/06/video.mp4", nil)
	req.Header.Set("Range", "bytes=0-999")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusPartialContent, resp.StatusCode)
	require.Equal(t, "bytes 0-999/2000", resp.Header.Get("Content-Range"))
	cl, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	require.NoError(t, err)
	require.Equal(t, 1000, cl)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Len(t, body, 1000)
	require.Equal(t, content[:1000], body)

	// Range: bytes=1500- (open-ended, last 500 bytes)
	req2, _ := http.NewRequest("GET", srv.URL+"/files/originals/u/2026/06/video.mp4", nil)
	req2.Header.Set("Range", "bytes=1500-")
	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()
	require.Equal(t, http.StatusPartialContent, resp2.StatusCode)
	require.Equal(t, "bytes 1500-1999/2000", resp2.Header.Get("Content-Range"))
	body2, _ := io.ReadAll(resp2.Body)
	require.Equal(t, content[1500:], body2)

	// Range: bytes=-300 (suffix)
	req3, _ := http.NewRequest("GET", srv.URL+"/files/originals/u/2026/06/video.mp4", nil)
	req3.Header.Set("Range", "bytes=-300")
	resp3, err := http.DefaultClient.Do(req3)
	require.NoError(t, err)
	defer resp3.Body.Close()
	require.Equal(t, http.StatusPartialContent, resp3.StatusCode)
	require.Equal(t, "bytes 1700-1999/2000", resp3.Header.Get("Content-Range"))
}

func TestStorageProxy_NotFound(t *testing.T) {
	fs := newFakeStorage()
	proxy := buildTestProxy(t, fs)
	srv := httptest.NewServer(proxy)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/files/originals/missing/file.mp4")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestStorageProxy_MethodNotAllowed(t *testing.T) {
	fs := newFakeStorage()
	proxy := buildTestProxy(t, fs)
	srv := httptest.NewServer(proxy)
	defer srv.Close()

	req, _ := http.NewRequest("POST", srv.URL+"/files/originals/x.mp4", nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestStorageProxy_PathTraversal(t *testing.T) {
	fs := newFakeStorage()
	proxy := buildTestProxy(t, fs)
	srv := httptest.NewServer(proxy)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/files/../secret")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestPromoteToOriginal_Integration verifies that LocalStorage correctly
// delegates to StoragePaths.PromoteToOriginal, preserving the original time directory.
// This is the core fix for cross-month upload failures.
func TestPromoteToOriginal_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	paths := conf.NewStoragePaths(tmpDir)

	// Simulate writing a file to a June path, as if session was created on June 30.
	juneTime := time.Date(2026, 6, 30, 23, 59, 0, 0, time.UTC)
	tempKey := paths.RelativeTempAt("admin", "video.mp4", juneTime)
	fullTemp := paths.FullPath(tempKey)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullTemp), 0755))
	content := []byte("cross-month-upload-content")
	require.NoError(t, os.WriteFile(fullTemp, content, 0644))

	// Call PromoteToOriginal — it must derive year/month from the tempKey itself.
	promoted, err := paths.PromoteToOriginal(tempKey)
	require.NoError(t, err)
	require.Equal(t, "originals/admin/2026/06/video.mp4", promoted)

	// Even if we now "advance time" to July by checking, the path must remain June.
	require.True(t, strings.Contains(promoted, "/2026/06/"),
		fmt.Sprintf("promoted path must use June directory, got %s", promoted))

	promotedFull := paths.FullPath(promoted)
	data, err := os.ReadFile(promotedFull)
	require.NoError(t, err)
	require.Equal(t, content, data)
}
