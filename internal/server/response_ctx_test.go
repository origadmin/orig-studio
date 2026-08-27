package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ginhttp "github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/timestamppb"

	http2 "origadmin/application/origstudio/internal/pkg/http"
	ginadapter "origadmin/application/origstudio/internal/pkg/http/gin"
	std "origadmin/application/origstudio/internal/pkg/http/std"
	media "origadmin/application/origstudio/api/gen/v1/media"
	types "origadmin/application/origstudio/api/gen/v1/types"
)

func init() {
	ginhttp.SetMode(ginhttp.TestMode)
}

// dispatchGin runs h through a gin adapter router and returns the recorder.
func dispatchGin(t *testing.T, h http2.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	r := ginhttp.New()
	group := r.Group("/api/v1")
	adapter := ginadapter.NewRouterAdapter(group)
	adapter.GET("/test", h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	r.ServeHTTP(w, req)
	return w
}

// dispatchStd runs h through a std router and returns the recorder.
func dispatchStd(t *testing.T, h http2.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	rt := std.NewRouter()
	group := rt.Group("/api/v1")
	group.GET("/test", h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	rt.ServeHTTP(w, req)
	return w
}

// TestOKCtx_Proto_GinPath verifies OKCtx emits snake_case protojson via gin adapter.
func TestOKCtx_Proto_GinPath(t *testing.T) {
	now := timestamppb.Now()
	resp := &media.GetMediaResponse{
		Media: &types.Media{
			Id:         "gin-id",
			Title:      "Gin Path",
			CreateTime: now,
			UpdateTime: now,
		},
	}
	w := dispatchGin(t, func(ctx http2.Context) error {
		return OKCtx(ctx, resp)
	})
	if w.Code != http.StatusOK {
		t.Errorf("gin path: expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"create_time"`) {
		t.Errorf("gin path: expected snake_case create_time, got %s", body)
	}
	if strings.Contains(body, `"createTime"`) {
		t.Errorf("gin path: found camelCase, got %s", body)
	}
	if !strings.Contains(body, `"gin-id"`) {
		t.Errorf("gin path: expected id value, got %s", body)
	}
}

// TestOKCtx_Proto_StdPath verifies OKCtx emits snake_case protojson via std adapter.
func TestOKCtx_Proto_StdPath(t *testing.T) {
	now := timestamppb.Now()
	resp := &media.GetMediaResponse{
		Media: &types.Media{
			Id:         "std-id",
			Title:      "Std Path",
			CreateTime: now,
			UpdateTime: now,
		},
	}
	w := dispatchStd(t, func(ctx http2.Context) error {
		return OKCtx(ctx, resp)
	})
	if w.Code != http.StatusOK {
		t.Errorf("std path: expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"create_time"`) {
		t.Errorf("std path: expected snake_case create_time, got %s", body)
	}
	if strings.Contains(body, `"createTime"`) {
		t.Errorf("std path: found camelCase, got %s", body)
	}
	if !strings.Contains(body, `"std-id"`) {
		t.Errorf("std path: expected id value, got %s", body)
	}
}

// TestOKCtx_Proto_EmitUnpopulated verifies zero-value fields are emitted (protojson EmitUnpopulated).
func TestOKCtx_Proto_EmitUnpopulated(t *testing.T) {
	resp := &media.GetMediaResponse{
		Media: &types.Media{Id: "unpop-test"},
	}
	w := dispatchStd(t, func(ctx http2.Context) error {
		return OKCtx(ctx, resp)
	})
	body := w.Body.String()
	if !strings.Contains(body, `"view_count"`) {
		t.Errorf("expected view_count field present, got %s", body)
	}
	if !strings.Contains(body, `"like_count"`) {
		t.Errorf("expected like_count field present, got %s", body)
	}
}

// TestOKCtx_NonProto_Map verifies non-proto data uses encoding/json (gin.H equivalent).
func TestOKCtx_NonProto_Map(t *testing.T) {
	w := dispatchStd(t, func(ctx http2.Context) error {
		return OKCtx(ctx, map[string]any{"items": []string{"a", "b"}, "total": 2})
	})
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"items"`) || !strings.Contains(body, `"total"`) {
		t.Errorf("expected items/total fields, got %s", body)
	}
}

// TestFailCtx verifies error response shape.
func TestFailCtx(t *testing.T) {
	w := dispatchStd(t, func(ctx http2.Context) error {
		return FailCtx(ctx, ErrBadRequest, "invalid input")
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"BAD_REQUEST"`) {
		t.Errorf("expected reason BAD_REQUEST, got %s", body)
	}
	if !strings.Contains(body, `"invalid input"`) {
		t.Errorf("expected message, got %s", body)
	}
}

// TestPageCtx_Proto verifies PageCtx with proto message.
func TestPageCtx_Proto(t *testing.T) {
	resp := &media.ListMediasResponse{
		Total:      2,
		Items:      []*types.Media{{Id: "m1"}, {Id: "m2"}},
		Page:       1,
		PageSize:   20,
		TotalPages: 1,
	}
	w := dispatchStd(t, func(ctx http2.Context) error {
		PageCtx(ctx, resp, 0, 0, 0)
		return nil
	})
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"page_size"`) {
		t.Errorf("expected snake_case page_size, got %s", body)
	}
	if !strings.Contains(body, `"total_pages"`) {
		t.Errorf("expected snake_case total_pages, got %s", body)
	}
}

// TestPageCtx_NonProto verifies PageCtx with non-proto items builds PageData shell.
func TestPageCtx_NonProto(t *testing.T) {
	w := dispatchStd(t, func(ctx http2.Context) error {
		PageCtx(ctx, []string{"a", "b"}, 2, 1, 20)
		return nil
	})
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	// PageData uses json tags: items/total/page/page_size (snake)
	if !strings.Contains(body, `"page_size"`) {
		t.Errorf("expected page_size field, got %s", body)
	}
	if !strings.Contains(body, `"total"`) {
		t.Errorf("expected total field, got %s", body)
	}
}

// TestCreatedCtx_Proto verifies 201 status.
func TestCreatedCtx_Proto(t *testing.T) {
	resp := &media.CreateMediaResponse{
		Media: &types.Media{Id: "new-id", Title: "New"},
	}
	w := dispatchStd(t, func(ctx http2.Context) error {
		return CreatedCtx(ctx, resp)
	})
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"new-id"`) {
		t.Errorf("expected id value, got %s", w.Body.String())
	}
}

// TestOKCtx_Proto_TimestampFormat verifies RFC3339 timestamp via protojson.
func TestOKCtx_Proto_TimestampFormat(t *testing.T) {
	resp := &media.GetMediaResponse{
		Media: &types.Media{
			Id:         "ts-test",
			CreateTime: &timestamppb.Timestamp{Seconds: 1704067200, Nanos: 0},
		},
	}
	w := dispatchStd(t, func(ctx http2.Context) error {
		return OKCtx(ctx, resp)
	})
	if !strings.Contains(w.Body.String(), `"2024-01-01T00:00:00Z"`) {
		t.Errorf("expected RFC3339 timestamp, got %s", w.Body.String())
	}
}

// TestOKCtx_Proto_GinVsStd_Parity verifies gin and std paths produce identical protojson output.
func TestOKCtx_Proto_GinVsStd_Parity(t *testing.T) {
	resp := &media.GetMediaResponse{
		Media: &types.Media{
			Id:         "parity-id",
			Title:      "Parity",
			CreateTime: &timestamppb.Timestamp{Seconds: 1704067200, Nanos: 0},
		},
	}
	wGin := dispatchGin(t, func(ctx http2.Context) error {
		return OKCtx(ctx, resp)
	})
	wStd := dispatchStd(t, func(ctx http2.Context) error {
		return OKCtx(ctx, resp)
	})
	if wGin.Body.String() != wStd.Body.String() {
		t.Errorf("gin and std paths differ:\ngin: %s\nstd: %s", wGin.Body.String(), wStd.Body.String())
	}
}
