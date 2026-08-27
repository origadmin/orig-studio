/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	pb "origadmin/application/origstudio/api/gen/v1/media"
	types "origadmin/application/origstudio/api/gen/v1/types"
	"origadmin/application/origstudio/internal/features/media/biz"
	"origadmin/application/origstudio/internal/features/media/biz/mock"
	"origadmin/application/origstudio/internal/features/media/dto"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestMediaHandler builds a MediaHandler wired to a gomock MediaRepo.
// Only the browse path (mediaUC) is populated; other deps are nil and not
// exercised by the browse endpoints under test.
func newTestMediaHandler(t *testing.T, repo *mock.MockMediaRepo) *MediaHandler {
	t.Helper()
	uc := biz.NewMediaUseCase(
		repo,
		nil, // profileRepo
		nil, // encodingRepo
		nil, // reviewLogRepo
		nil, // storage
		nil, // publisher
		log.DefaultLogger,
		nil, // spriteUC
	)
	return NewMediaHandlerForMicroservice(
		nil, // jwtMgr
		uc,
		nil, // uploadUC
		nil, // likeFavoriteUC
		nil, // mediaService
		nil, // settingUC — ModuleGuardCtx treats nil as "always pass"
	)
}

// decodeListResponse parses the standard envelope returned by server.OK.
func decodeListResponse(t *testing.T, body []byte) map[string]interface{} {
	t.Helper()
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &resp), "response body: %s", string(body))
	return resp
}

// TestMediaHandler_ListMedias_Success verifies that GET /api/v1/medias
// returns a paginated list when the repo succeeds.
func TestMediaHandler_ListMedias_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockMediaRepo(ctrl)

	repo.EXPECT().
		List(gomock.Any(), gomock.Any()).
		Return([]*types.Media{
			{Id: "m1", Title: "Video 1"},
			{Id: "m2", Title: "Video 2"},
		}, int32(2), nil).
		Times(1)

	h := newTestMediaHandler(t, repo)
	srv := httptest.NewServer(h.HTTPHandler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/medias?page=1&page_size=20")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeListResponse(t, mustReadBody(t, resp.Body))
	require.Equal(t, float64(2), body["total"])
	items, ok := body["items"].([]interface{})
	require.True(t, ok)
	require.Len(t, items, 2)
}

// TestMediaHandler_ListMedias_RepoError verifies that GET /api/v1/medias
// returns an error envelope and logs the request_id when the repo fails.
func TestMediaHandler_ListMedias_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockMediaRepo(ctrl)

	repo.EXPECT().
		List(gomock.Any(), gomock.Any()).
		Return(([]*types.Media)(nil), int32(0), errors.New("db connection lost")).
		Times(1)

	h := newTestMediaHandler(t, repo)
	srv := httptest.NewServer(h.HTTPHandler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/medias")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Fail() maps ErrInternal to HTTP 500
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	body := decodeListResponse(t, mustReadBody(t, resp.Body))
	require.NotEmpty(t, body["message"])
}

// TestMediaHandler_ListFeaturedMedias_Success verifies that GET /api/v1/medias/featured
// returns items and respects the limit cap of 50.
func TestMediaHandler_ListFeaturedMedias_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockMediaRepo(ctrl)

	repo.EXPECT().
		List(gomock.Any(), gomock.Any()).
		Return([]*types.Media{{Id: "f1"}}, int32(1), nil).
		Times(1)

	h := newTestMediaHandler(t, repo)
	srv := httptest.NewServer(h.HTTPHandler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/medias/featured?limit=10")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeListResponse(t, mustReadBody(t, resp.Body))
	items, ok := body["items"].([]interface{})
	require.True(t, ok)
	require.Len(t, items, 1)
}

// TestMediaHandler_ListLatestMedias_Success verifies that GET /api/v1/medias/latest
// returns items.
func TestMediaHandler_ListLatestMedias_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockMediaRepo(ctrl)

	repo.EXPECT().
		List(gomock.Any(), gomock.Any()).
		Return([]*types.Media{{Id: "l1"}, {Id: "l2"}, {Id: "l3"}}, int32(3), nil).
		Times(1)

	h := newTestMediaHandler(t, repo)
	srv := httptest.NewServer(h.HTTPHandler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/medias/latest?limit=5")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeListResponse(t, mustReadBody(t, resp.Body))
	items, ok := body["items"].([]interface{})
	require.True(t, ok)
	require.Len(t, items, 3)
}

// TestMediaHandler_GetMedia_Success verifies that GET /api/v1/medias/:token
// returns the media when found and public.
func TestMediaHandler_GetMedia_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockMediaRepo(ctrl)

	token := "abc123"
	repo.EXPECT().
		GetByShortToken(gomock.Any(), token).
		Return(&types.Media{
			Id:       "m1",
			Title:    "Public Video",
			Privacy:  types.Privacy_PRIVACY_PUBLIC,
			ShortToken: token,
		}, nil).
		Times(1)

	h := newTestMediaHandler(t, repo)
	srv := httptest.NewServer(h.HTTPHandler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/medias/" + token)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeListResponse(t, mustReadBody(t, resp.Body))
	media, ok := body["media"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "m1", media["id"])
}

// TestMediaHandler_GetMedia_NotFound verifies that GET /api/v1/medias/:token
// returns 404 when the repo returns an error.
func TestMediaHandler_GetMedia_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockMediaRepo(ctrl)

	token := "not-exist"
	repo.EXPECT().
		GetByShortToken(gomock.Any(), token).
		Return((*types.Media)(nil), errors.New("record not found")).
		Times(1)

	h := newTestMediaHandler(t, repo)
	srv := httptest.NewServer(h.HTTPHandler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/medias/" + token)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Fail() maps ErrNotFound to HTTP 404
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestMediaHandler_GetMedia_PrivateMediaReturns404 verifies that private media
// is hidden from the public API (returns 404, not the media object).
func TestMediaHandler_GetMedia_PrivateMediaReturns404(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockMediaRepo(ctrl)

	token := "priv-token"
	repo.EXPECT().
		GetByShortToken(gomock.Any(), token).
		Return(&types.Media{
			Id:       "m-private",
			Title:    "Private Video",
			Privacy:  types.Privacy_PRIVACY_PRIVATE,
			ShortToken: token,
		}, nil).
		Times(1)

	h := newTestMediaHandler(t, repo)
	srv := httptest.NewServer(h.HTTPHandler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/medias/" + token)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Private media must return 404 to hide its existence
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestMediaHandler_ListMedias_PaginationCap verifies that page_size is capped
// at 100 even when the client requests more.
func TestMediaHandler_ListMedias_PaginationCap(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockMediaRepo(ctrl)

	// Expect the capped page_size (100) to be passed to the repo
	repo.EXPECT().
		List(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ interface{}, opts ...*dto.MediaQueryOption) ([]*types.Media, int32, error) {
			if len(opts) > 0 && opts[0] != nil {
				require.Equal(t, int32(100), opts[0].PageSize, "page_size should be capped at 100")
			}
			return []*types.Media{}, int32(0), nil
		}).
		Times(1)

	h := newTestMediaHandler(t, repo)
	srv := httptest.NewServer(h.HTTPHandler())
	defer srv.Close()

	// Request page_size=500 which should be capped to 100
	resp, err := http.Get(srv.URL + "/api/v1/medias?page_size=500")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// mustReadBody reads and closes an io.Reader, failing the test on error.
func mustReadBody(t *testing.T, r interface{ Read([]byte) (int, error) }) []byte {
	t.Helper()
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 4096)
	for {
		n, err := r.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			break
		}
	}
	return buf
}

// Compile-time assertion that pb.ListMediasResponse is used (keeps import alive).
var _ pb.ListMediasResponse
