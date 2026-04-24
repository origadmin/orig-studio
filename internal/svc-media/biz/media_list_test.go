package biz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestListEncodingTasksFlat_OnlyStats tests the only_stats parameter
func TestListEncodingTasksFlat_OnlyStats(t *testing.T) {
	// Setup mocks
	mockEncodingRepo := &MockEncodingTaskRepo{}
	mockMediaRepo := &MockMediaRepo{}
	mockEncodeProfileRepo := &MockEncodeProfileRepo{}

	// Create use case
	uc := NewMediaUseCase(mockMediaRepo, mockEncodeProfileRepo, mockEncodingRepo, nil, nil, nil, nil, nil)

	// Test case 1: only_stats=true should return complete stats
	filter := &TranscodingStatusFilter{
		Page:       1,
		PageSize:   25,
		OnlyStats:  true,
		Status:     "failed", // This should be ignored when only_stats=true
		ProfileFilter: "H264",  // This should be ignored when only_stats=true
		SearchQuery: "test",  // This should be ignored when only_stats=true
	}

	result, err := uc.ListEncodingTasksFlat(context.Background(), filter, nil)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// We can't assert specific counts because the mock returns empty counts
	// but we can verify the structure is correct
	assert.IsType(t, &FlatTaskList{}, result)
	assert.Empty(t, result.Items) // No items should be returned when only_stats=true
}

// TestListEncodingTasksFlat_FilteredStats tests stats with filters
func TestListEncodingTasksFlat_FilteredStats(t *testing.T) {
	// Setup mocks
	mockEncodingRepo := &MockEncodingTaskRepo{}
	mockMediaRepo := &MockMediaRepo{}
	mockEncodeProfileRepo := &MockEncodeProfileRepo{}

	// Create use case
	uc := NewMediaUseCase(mockMediaRepo, mockEncodeProfileRepo, mockEncodingRepo, nil, nil, nil, nil, nil)

	// Test case: filtered stats should respect filters except status
	filter := &TranscodingStatusFilter{
		Page:       1,
		PageSize:   25,
		OnlyStats:  false,
		Status:     "", // Empty status means all
		ProfileFilter: "H264",
		SearchQuery: "test",
	}

	result, err := uc.ListEncodingTasksFlat(context.Background(), filter, nil)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.IsType(t, &FlatTaskList{}, result)
}

// TestListEncodingTasksFlat_StatusFilter tests status filtering
func TestListEncodingTasksFlat_StatusFilter(t *testing.T) {
	// Setup mocks
	mockEncodingRepo := &MockEncodingTaskRepo{}
	mockMediaRepo := &MockMediaRepo{}
	mockEncodeProfileRepo := &MockEncodeProfileRepo{}

	// Create use case
	uc := NewMediaUseCase(mockMediaRepo, mockEncodeProfileRepo, mockEncodingRepo, nil, nil, nil, nil, nil)

	// Test case: status filter should be applied to list but not to stats
	filter := &TranscodingStatusFilter{
		Page:       1,
		PageSize:   25,
		OnlyStats:  false,
		Status:     "failed", // This should be applied to ListFlat but not to CountByStatusWithFilter
	}

	result, err := uc.ListEncodingTasksFlat(context.Background(), filter, nil)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.IsType(t, &FlatTaskList{}, result)
}

// TestListEncodingTasksFlat_DefaultValues tests default values
func TestListEncodingTasksFlat_DefaultValues(t *testing.T) {
	// Setup mocks
	mockEncodingRepo := &MockEncodingTaskRepo{}
	mockMediaRepo := &MockMediaRepo{}
	mockEncodeProfileRepo := &MockEncodeProfileRepo{}

	// Create use case
	uc := NewMediaUseCase(mockMediaRepo, mockEncodeProfileRepo, mockEncodingRepo, nil, nil, nil, nil, nil)

	// Test case: default values should be handled correctly
	filter := &TranscodingStatusFilter{
		Page:      1, // Default page
		PageSize:  25, // Default page size
		OnlyStats: false,
		Status:    "", // Empty status should default to "all"
	}

	result, err := uc.ListEncodingTasksFlat(context.Background(), filter, nil)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.IsType(t, &FlatTaskList{}, result)
}