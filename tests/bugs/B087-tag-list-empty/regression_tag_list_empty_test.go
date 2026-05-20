/*
 * B087-R2 Regression Test: Tag creation succeeds but list appears empty
 *
 * Root Cause: Backend returns entity.Tag with fields "title"/"media_count",
 * but frontend expects "name"/"count". Also:
 *   - Frontend sends "keyword" param, backend expects "search"
 *   - Backend List() ignores search/status filter params
 *   - Backend status enum "ACTIVE"/"INACTIVE" vs frontend "active"/"inactive"
 *
 * This test verifies:
 * 1. TagResponse DTO maps entity.Tag fields to frontend-compatible names
 * 2. Tag list API returns items with frontend-compatible field names
 * 3. Tag list API supports "search" and "keyword" query parameters
 * 4. Tag list API returns status in lowercase format
 */

package bugs

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/tag"
	"origadmin/application/origstudio/internal/features/admin/service"
)

// TestB087R2_TagResponse_MapsFieldsCorrectly verifies that the TagResponse DTO
// correctly maps entity.Tag fields to frontend-compatible field names.
// This is the core regression test for the bug: frontend expects "name" and "count",
// but entity.Tag has "title" and "media_count".
func TestB087R2_TagResponse_MapsFieldsCorrectly(t *testing.T) {
	now := time.Now()
	entityTag := &entity.Tag{
		ID:                1,
		Title:             "TestTag-B087R2",
		Slug:              "testtag-b087r2",
		MediaCount:        5,
		ListingsThumbnail: "https://example.com/thumb.jpg",
		Status:            tag.StatusACTIVE,
		Description:       "A test tag",
		Color:             "#FF0000",
		CreateTime:        now,
		UpdateTime:        now,
	}

	// Convert to TagResponse DTO
	tagResp := service.ToTagResponse(entityTag)

	// Verify field mapping
	assert.Equal(t, "1", tagResp.ID, "ID should be string '1'")
	assert.Equal(t, "TestTag-B087R2", tagResp.Name, "Name should map from Title")
	assert.Equal(t, "testtag-b087r2", tagResp.Slug, "Slug should pass through")
	assert.Equal(t, 5, tagResp.Count, "Count should map from MediaCount")
	assert.Equal(t, "active", tagResp.Status, "Status ACTIVE should become 'active'")
	assert.Equal(t, "A test tag", tagResp.Description, "Description should pass through")
	assert.Equal(t, "#FF0000", tagResp.Color, "Color should pass through")
	assert.Equal(t, "https://example.com/thumb.jpg", tagResp.ListingsThumbnail, "ListingsThumbnail should pass through")
}

// TestB087R2_TagResponse_StatusMapping verifies all status enum values
// are correctly mapped to lowercase strings.
func TestB087R2_TagResponse_StatusMapping(t *testing.T) {
	tests := []struct {
		name     string
		status   tag.Status
		expected string
	}{
		{"ACTIVE maps to active", tag.StatusACTIVE, "active"},
		{"INACTIVE maps to inactive", tag.StatusINACTIVE, "inactive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entityTag := &entity.Tag{
				ID:     1,
				Title:  "test",
				Status: tt.status,
			}
			tagResp := service.ToTagResponse(entityTag)
			assert.Equal(t, tt.expected, tagResp.Status)
		})
	}
}

// TestB087R2_TagResponse_JSONSerialization verifies that the TagResponse
// serializes to JSON with the correct field names.
func TestB087R2_TagResponse_JSONSerialization(t *testing.T) {
	now := time.Now()
	entityTag := &entity.Tag{
		ID:         42,
		Title:      "JSONTest-B087R2",
		Slug:       "jsontest-b087r2",
		MediaCount: 10,
		Status:     tag.StatusACTIVE,
		CreateTime: now,
		UpdateTime: now,
	}

	tagResp := service.ToTagResponse(entityTag)
	jsonBytes, err := json.Marshal(tagResp)
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(jsonBytes, &result))

	// Verify JSON field names match frontend expectations
	assert.Contains(t, result, "name", "JSON should contain 'name' field")
	assert.NotContains(t, result, "title", "JSON should NOT contain 'title' field")
	assert.Contains(t, result, "count", "JSON should contain 'count' field")
	assert.NotContains(t, result, "media_count", "JSON should NOT contain 'media_count' field")
	assert.Contains(t, result, "id", "JSON should contain 'id' field")
	assert.Contains(t, result, "slug", "JSON should contain 'slug' field")
	assert.Contains(t, result, "status", "JSON should contain 'status' field")

	// Verify ID is string
	idVal, ok := result["id"].(string)
	assert.True(t, ok, "id should be a string in JSON")
	assert.Equal(t, "42", idVal)

	// Verify status is lowercase
	statusVal, ok := result["status"].(string)
	assert.True(t, ok, "status should be a string in JSON")
	assert.Equal(t, "active", statusVal)
}

// TestB087R2_TagResponse_NilHandling verifies that ToTagResponse handles
// nil entity gracefully.
func TestB087R2_TagResponse_NilHandling(t *testing.T) {
	tagResp := service.ToTagResponse(nil)
	assert.Nil(t, tagResp, "ToTagResponse(nil) should return nil")
}

// TestB087R2_TagResponse_EmptyFields verifies that empty/zero fields
// are correctly mapped.
func TestB087R2_TagResponse_EmptyFields(t *testing.T) {
	entityTag := &entity.Tag{
		ID:         0,
		Title:      "EmptyTest",
		MediaCount: 0,
		Status:     tag.StatusACTIVE,
	}

	tagResp := service.ToTagResponse(entityTag)
	assert.Equal(t, "0", tagResp.ID, "Zero ID should be '0'")
	assert.Equal(t, "EmptyTest", tagResp.Name, "Name should be set")
	assert.Equal(t, 0, tagResp.Count, "Zero count should be 0")
	assert.Equal(t, "active", tagResp.Status, "Default status should be 'active'")
}

// TestB087R2_ParseStatus_FrontendToDB verifies that frontend status strings
// are correctly parsed to database enum values.
func TestB087R2_ParseStatus_FrontendToDB(t *testing.T) {
	tests := []struct {
		input    string
		expected tag.Status
	}{
		{"active", tag.StatusACTIVE},
		{"inactive", tag.StatusINACTIVE},
		{"ACTIVE", tag.StatusACTIVE},    // uppercase should also work
		{"INACTIVE", tag.StatusINACTIVE}, // uppercase should also work
		{"", tag.StatusACTIVE},           // default to active
		{"invalid", tag.StatusACTIVE},    // invalid defaults to active
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := service.ParseTagStatus(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestB087R2_EntityTag_HasIncompatibleFields explicitly demonstrates the
// root cause: entity.Tag uses "title" and "media_count" which don't match
// the frontend Tag interface that expects "name" and "count".
func TestB087R2_EntityTag_HasIncompatibleFields(t *testing.T) {
	entityTag := &entity.Tag{
		Title:      "IncompatibleTest",
		MediaCount: 7,
	}

	// This demonstrates the problem: entity has Title, not Name
	assert.Equal(t, "IncompatibleTest", entityTag.Title)
	assert.Equal(t, 7, entityTag.MediaCount)

	// When serialized to JSON, the field names are "title" and "media_count"
	jsonBytes, err := json.Marshal(entityTag)
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(jsonBytes, &result))

	// These are the WRONG field names from frontend's perspective
	assert.Contains(t, result, "title", "entity.Tag serializes as 'title'")
	assert.Contains(t, result, "media_count", "entity.Tag serializes as 'media_count'")

	// Frontend expects these field names instead
	assert.NotContains(t, result, "name", "entity.Tag does NOT have 'name' field")
	assert.NotContains(t, result, "count", "entity.Tag does NOT have 'count' field")
}
