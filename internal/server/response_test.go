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

// TestOK_ProtoSingleResource verifies OK auto-detects proto.Message and serializes correctly.
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
	// Verify wrapper structure
	if !strings.Contains(body, `"code":0`) {
		t.Errorf("expected code:0 in response, got %s", body)
	}
	if !strings.Contains(body, `"message":"ok"`) {
		t.Errorf("expected message:ok in response, got %s", body)
	}
	// Verify snake_case field names (UseProtoNames=true)
	if !strings.Contains(body, `"create_time"`) {
		t.Errorf("expected snake_case create_time, got %s", body)
	}
	if !strings.Contains(body, `"update_time"`) {
		t.Errorf("expected snake_case update_time, got %s", body)
	}
	// Verify no camelCase (which would be createTime/updateTime)
	if strings.Contains(body, `"createTime"`) {
		t.Errorf("found camelCase createTime, should be snake_case create_time, got %s", body)
	}
	if strings.Contains(body, `"updateTime"`) {
		t.Errorf("found camelCase updateTime, should be snake_case update_time, got %s", body)
	}
	// Verify media wrapper
	if !strings.Contains(body, `"media"`) {
		t.Errorf("expected media wrapper field, got %s", body)
	}
}

// TestPage_ProtoListResponse verifies Page auto-detects proto.Message and serializes pagination correctly.
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
	// Verify pagination fields use snake_case
	if !strings.Contains(body, `"page_size"`) {
		t.Errorf("expected snake_case page_size, got %s", body)
	}
	if !strings.Contains(body, `"total_pages"`) {
		t.Errorf("expected snake_case total_pages, got %s", body)
	}
	// Verify data list field name is "medias" (Proto-defined), not "items"
	if !strings.Contains(body, `"medias"`) {
		t.Errorf("expected medias field, got %s", body)
	}
	// Verify no camelCase
	if strings.Contains(body, `"pageSize"`) {
		t.Errorf("found camelCase pageSize, should be snake_case page_size, got %s", body)
	}
	if strings.Contains(body, `"totalPages"`) {
		t.Errorf("found camelCase totalPages, should be snake_case total_pages, got %s", body)
	}
}

// TestCreated_Proto verifies Created auto-detects proto.Message and returns HTTP 201.
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
	if !strings.Contains(body, `"code":0`) {
		t.Errorf("expected code:0 in response, got %s", body)
	}
	if !strings.Contains(body, `"media"`) {
		t.Errorf("expected media field in response, got %s", body)
	}
}

// TestOK_ProtoEmitUnpopulated verifies that zero-value fields are emitted.
func TestOK_ProtoEmitUnpopulated(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Media with only ID set, all other fields zero
	resp := &media.GetMediaResponse{
		Media: &types.Media{Id: "test-id"},
	}

	OK(c, resp)

	body := w.Body.String()
	// protojson emits zero-value fields when EmitUnpopulated=true.
	// Note: protojson serializes int64/uint64 as strings per the protobuf JSON spec.
	// We verify the fields are present in the output.
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

// TestOK_ProtoTimestampFormat verifies timestamppb.Timestamp is serialized as RFC 3339.
func TestOK_ProtoTimestampFormat(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	resp := &media.GetMediaResponse{
		Media: &types.Media{
			Id:         "ts-test",
			CreateTime: &timestamppb.Timestamp{Seconds: 1704067200, Nanos: 0}, // 2024-01-01T00:00:00Z
		},
	}

	OK(c, resp)

	body := w.Body.String()
	// protojson serializes timestamppb as RFC 3339
	if !strings.Contains(body, `"2024-01-01T00:00:00Z"`) {
		t.Errorf("expected RFC 3339 timestamp format, got %s", body)
	}
}
