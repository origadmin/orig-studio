package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/timestamppb"

	media "origadmin/application/origstudio/api/gen/v1/media"
	types "origadmin/application/origstudio/api/gen/v1/types"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestOK_ProtoSingleResource(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	now := timestamppb.Now()
	resp := &media.GetMediaResponse{
		Media: &types.Media{
			Id:         "test-id-123",
			Title:      "Test Media",
			CreateTime: now,
			UpdateTime: now,
		},
	}

	OK(c, resp)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"create_time"`) {
		t.Errorf("expected snake_case create_time, got %s", body)
	}
	if !strings.Contains(body, `"update_time"`) {
		t.Errorf("expected snake_case update_time, got %s", body)
	}
	if strings.Contains(body, `"createTime"`) {
		t.Errorf("found camelCase createTime, should be snake_case create_time, got %s", body)
	}
	if strings.Contains(body, `"updateTime"`) {
		t.Errorf("found camelCase updateTime, should be snake_case update_time, got %s", body)
	}
	if !strings.Contains(body, `"media"`) {
		t.Errorf("expected media wrapper field, got %s", body)
	}
}

func TestPage_ProtoListResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	resp := &media.ListMediasResponse{
		Total:      100,
		Items:      []*types.Media{{Id: "m1", Title: "Media 1"}, {Id: "m2", Title: "Media 2"}},
		Page:       1,
		PageSize:   20,
		TotalPages: 5,
	}

	Page(c, resp, 0, 0, 0)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"page_size"`) {
		t.Errorf("expected snake_case page_size, got %s", body)
	}
	if !strings.Contains(body, `"total_pages"`) {
		t.Errorf("expected snake_case total_pages, got %s", body)
	}
	if !strings.Contains(body, `"items"`) {
		t.Errorf("expected items field, got %s", body)
	}
	if strings.Contains(body, `"pageSize"`) {
		t.Errorf("found camelCase pageSize, should be snake_case page_size, got %s", body)
	}
	if strings.Contains(body, `"totalPages"`) {
		t.Errorf("found camelCase totalPages, should be snake_case total_pages, got %s", body)
	}
}

func TestCreated_Proto(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	resp := &media.CreateMediaResponse{
		Media: &types.Media{Id: "new-id", Title: "New Media"},
	}

	Created(c, resp)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"media"`) {
		t.Errorf("expected media field in response, got %s", body)
	}
}

func TestOK_ProtoEmitUnpopulated(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	resp := &media.GetMediaResponse{
		Media: &types.Media{Id: "test-id"},
	}

	OK(c, resp)

	body := w.Body.String()
	if !strings.Contains(body, `"view_count"`) {
		t.Errorf("expected view_count field to be present, got %s", body)
	}
	if !strings.Contains(body, `"like_count"`) {
		t.Errorf("expected like_count field to be present, got %s", body)
	}
	if !strings.Contains(body, `"comment_count"`) {
		t.Errorf("expected comment_count field to be present, got %s", body)
	}
}

func TestOK_ProtoTimestampFormat(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	resp := &media.GetMediaResponse{
		Media: &types.Media{
			Id:         "ts-test",
			CreateTime: &timestamppb.Timestamp{Seconds: 1704067200, Nanos: 0},
		},
	}

	OK(c, resp)

	body := w.Body.String()
	if !strings.Contains(body, `"2024-01-01T00:00:00Z"`) {
		t.Errorf("expected RFC 3339 timestamp format, got %s", body)
	}
}
