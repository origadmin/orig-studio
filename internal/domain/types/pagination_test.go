/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package types

import (
	"math"
	"testing"
)

func TestDefaultPaginationConfig(t *testing.T) {
	cfg := DefaultPaginationConfig()
	if cfg.DefaultPageSize != 20 {
		t.Errorf("DefaultPaginationConfig().DefaultPageSize = %d, want 20", cfg.DefaultPageSize)
	}
	if cfg.MaxPageSize != 100 {
		t.Errorf("DefaultPaginationConfig().MaxPageSize = %d, want 100", cfg.MaxPageSize)
	}
	if cfg.HardLimit != 1000 {
		t.Errorf("DefaultPaginationConfig().HardLimit = %d, want 1000", cfg.HardLimit)
	}
}

func TestNormalizePagination_NormalValues(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		wantPage     int
		wantPageSize int
	}{
		{"page=1, pageSize=20", 1, 20, 1, 20},
		{"page=5, pageSize=50", 5, 50, 5, 50},
		{"page=1, pageSize=1 (min)", 1, 1, 1, 1},
		{"page=1, pageSize=100 (max)", 1, 100, 1, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPage, gotPageSize := NormalizePagination(tt.page, tt.pageSize)
			if gotPage != tt.wantPage {
				t.Errorf("NormalizePagination() page = %d, want %d", gotPage, tt.wantPage)
			}
			if gotPageSize != tt.wantPageSize {
				t.Errorf("NormalizePagination() pageSize = %d, want %d", gotPageSize, tt.wantPageSize)
			}
		})
	}
}

func TestNormalizePagination_PageZero(t *testing.T) {
	page, pageSize := NormalizePagination(0, 20)
	if page != 1 {
		t.Errorf("NormalizePagination(0, 20) page = %d, want 1", page)
	}
	if pageSize != 20 {
		t.Errorf("NormalizePagination(0, 20) pageSize = %d, want 20", pageSize)
	}
}

func TestNormalizePagination_PageNegative(t *testing.T) {
	tests := []struct {
		name string
		page int
		want int
	}{
		{"page=-1", -1, 1},
		{"page=-100", -100, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, _ := NormalizePagination(tt.page, 20)
			if page != tt.want {
				t.Errorf("NormalizePagination(%d, 20) page = %d, want %d", tt.page, page, tt.want)
			}
		})
	}
}

func TestNormalizePagination_PageSizeZero(t *testing.T) {
	page, pageSize := NormalizePagination(1, 0)
	if page != 1 {
		t.Errorf("NormalizePagination(1, 0) page = %d, want 1", page)
	}
	if pageSize != 20 {
		t.Errorf("NormalizePagination(1, 0) pageSize = %d, want 20 (default)", pageSize)
	}
}

func TestNormalizePagination_PageSizeNegative(t *testing.T) {
	_, pageSize := NormalizePagination(1, -1)
	if pageSize != 20 {
		t.Errorf("NormalizePagination(1, -1) pageSize = %d, want 20 (default)", pageSize)
	}
}

func TestNormalizePagination_PageSizeOverMax(t *testing.T) {
	tests := []struct {
		name     string
		pageSize int
		want     int
	}{
		{"pageSize=101", 101, 100},
		{"pageSize=99999", 99999, 100},
		{"pageSize=1000", 1000, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, pageSize := NormalizePagination(1, tt.pageSize)
			if pageSize != tt.want {
				t.Errorf("NormalizePagination(1, %d) pageSize = %d, want %d", tt.pageSize, pageSize, tt.want)
			}
		})
	}
}

func TestNormalizePagination_BothInvalid(t *testing.T) {
	page, pageSize := NormalizePagination(-1, 0)
	if page != 1 {
		t.Errorf("NormalizePagination(-1, 0) page = %d, want 1", page)
	}
	if pageSize != 20 {
		t.Errorf("NormalizePagination(-1, 0) pageSize = %d, want 20", pageSize)
	}
}

func TestNormalizePaginationWithConfig_CustomConfig(t *testing.T) {
	cfg := PaginationConfig{
		DefaultPageSize: 15,
		MaxPageSize:     50,
		HardLimit:       500,
	}
	tests := []struct {
		name         string
		page         int
		pageSize     int
		wantPage     int
		wantPageSize int
	}{
		{"custom default pageSize", 1, 0, 1, 15},
		{"custom max pageSize", 1, 60, 1, 50},
		{"custom normal value", 3, 30, 3, 30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPage, gotPageSize := NormalizePaginationWithConfig(tt.page, tt.pageSize, cfg)
			if gotPage != tt.wantPage {
				t.Errorf("NormalizePaginationWithConfig() page = %d, want %d", gotPage, tt.wantPage)
			}
			if gotPageSize != tt.wantPageSize {
				t.Errorf("NormalizePaginationWithConfig() pageSize = %d, want %d", gotPageSize, tt.wantPageSize)
			}
		})
	}
}

func TestNormalizeQueryOption_Nil(t *testing.T) {
	opt := NormalizeQueryOption(nil)
	if opt.Page != 1 {
		t.Errorf("NormalizeQueryOption(nil) Page = %d, want 1", opt.Page)
	}
	if opt.PageSize != 20 {
		t.Errorf("NormalizeQueryOption(nil) PageSize = %d, want 20", opt.PageSize)
	}
}

func TestNormalizeQueryOption_Empty(t *testing.T) {
	opt := NormalizeQueryOption(&QueryOption{})
	if opt.Page != 1 {
		t.Errorf("NormalizeQueryOption(&QueryOption{}) Page = %d, want 1", opt.Page)
	}
	if opt.PageSize != 20 {
		t.Errorf("NormalizeQueryOption(&QueryOption{}) PageSize = %d, want 20", opt.PageSize)
	}
}

func TestNormalizeQueryOption_NoPagingMode(t *testing.T) {
	opt := NormalizeQueryOption(&QueryOption{NoPaging: true, PageSize: 0})
	if opt.PageSize != 1000 {
		t.Errorf("NormalizeQueryOption(NoPaging=true, PageSize=0) PageSize = %d, want 1000 (HardLimit)", opt.PageSize)
	}
}

func TestNormalizeQueryOption_NoPagingModeOverHardLimit(t *testing.T) {
	opt := NormalizeQueryOption(&QueryOption{NoPaging: true, PageSize: 5000})
	if opt.PageSize != 1000 {
		t.Errorf("NormalizeQueryOption(NoPaging=true, PageSize=5000) PageSize = %d, want 1000 (HardLimit)", opt.PageSize)
	}
}

func TestCalculateTotalPages(t *testing.T) {
	tests := []struct {
		name      string
		totalSize int64
		pageSize  int
		want      int32
	}{
		{"0 items", 0, 20, 0},
		{"1 item, pageSize=20", 1, 20, 1},
		{"20 items, pageSize=20", 20, 20, 1},
		{"21 items, pageSize=20", 21, 20, 2},
		{"100 items, pageSize=20", 100, 20, 5},
		{"101 items, pageSize=20", 101, 20, 6},
		{"invalid pageSize=0", 100, 0, 0},
		{"invalid pageSize=-1", 100, -1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateTotalPages(tt.totalSize, tt.pageSize)
			if got != tt.want {
				t.Errorf("CalculateTotalPages(%d, %d) = %d, want %d", tt.totalSize, tt.pageSize, got, tt.want)
			}
		})
	}
}

func TestInitPaginationConfig(t *testing.T) {
	// Save original config
	originalConfig := globalConfig
	defer func() { globalConfig = originalConfig }()

	// Test with valid config
	cfg := PaginationConfig{
		DefaultPageSize: 25,
		MaxPageSize:     50,
		HardLimit:       500,
	}
	InitPaginationConfig(cfg)
	current := GetPaginationConfig()
	if current.DefaultPageSize != 25 {
		t.Errorf("InitPaginationConfig() DefaultPageSize = %d, want 25", current.DefaultPageSize)
	}
	if current.MaxPageSize != 50 {
		t.Errorf("InitPaginationConfig() MaxPageSize = %d, want 50", current.MaxPageSize)
	}
	if current.HardLimit != 500 {
		t.Errorf("InitPaginationConfig() HardLimit = %d, want 500", current.HardLimit)
	}

	// Test with zero values (should use defaults)
	InitPaginationConfig(PaginationConfig{})
	current = GetPaginationConfig()
	if current.DefaultPageSize != 20 {
		t.Errorf("InitPaginationConfig(zero) DefaultPageSize = %d, want 20", current.DefaultPageSize)
	}
	if current.MaxPageSize != 100 {
		t.Errorf("InitPaginationConfig(zero) MaxPageSize = %d, want 100", current.MaxPageSize)
	}
	if current.HardLimit != 1000 {
		t.Errorf("InitPaginationConfig(zero) HardLimit = %d, want 1000", current.HardLimit)
	}

	// Test MaxPageSize > HardLimit (should be clamped)
	InitPaginationConfig(PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     2000,
		HardLimit:       500,
	})
	current = GetPaginationConfig()
	if current.MaxPageSize != 500 {
		t.Errorf("InitPaginationConfig(MaxPageSize>HardLimit) MaxPageSize = %d, want 500", current.MaxPageSize)
	}
}

func TestNormalizeHTTPPagination(t *testing.T) {
	// This is a convenience wrapper, just verify it delegates correctly
	page, pageSize := NormalizeHTTPPagination(0, 0)
	if page != 1 {
		t.Errorf("NormalizeHTTPPagination(0, 0) page = %d, want 1", page)
	}
	if pageSize != 20 {
		t.Errorf("NormalizeHTTPPagination(0, 0) pageSize = %d, want 20", pageSize)
	}

	page, pageSize = NormalizeHTTPPagination(3, 50)
	if page != 3 {
		t.Errorf("NormalizeHTTPPagination(3, 50) page = %d, want 3", page)
	}
	if pageSize != 50 {
		t.Errorf("NormalizeHTTPPagination(3, 50) pageSize = %d, want 50", pageSize)
	}
}

// AC-01: Comprehensive pagination parameter validation table
func TestNormalizePagination_AC01_FullTable(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		wantPage     int
		wantPageSize int
	}{
		{"page=0, pageSize=20", 0, 20, 1, 20},
		{"page=-1, pageSize=20", -1, 20, 1, 20},
		{"page=-100, pageSize=20", -100, 20, 1, 20},
		{"page=1 (normal), pageSize=20", 1, 20, 1, 20},
		{"page=5 (normal), pageSize=20", 5, 20, 5, 20},
		{"page=1, pageSize=0", 1, 0, 1, 20},
		{"page=1, pageSize=-1", 1, -1, 1, 20},
		{"page=1, pageSize=1 (min)", 1, 1, 1, 1},
		{"page=1, pageSize=100 (max)", 1, 100, 1, 100},
		{"page=1, pageSize=101 (over max)", 1, 101, 1, 100},
		{"page=1, pageSize=99999 (malicious)", 1, 99999, 1, 100},
		{"page=-1, pageSize=0 (both invalid)", -1, 0, 1, 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPage, gotPageSize := NormalizePagination(tt.page, tt.pageSize)
			if gotPage != tt.wantPage {
				t.Errorf("NormalizePagination(%d, %d) page = %d, want %d", tt.page, tt.pageSize, gotPage, tt.wantPage)
			}
			if gotPageSize != tt.wantPageSize {
				t.Errorf("NormalizePagination(%d, %d) pageSize = %d, want %d", tt.page, tt.pageSize, gotPageSize, tt.wantPageSize)
			}
		})
	}
}

// AC-03: INT32_MAX pageSize must be corrected to MaxPageSize
func TestNormalizePagination_PageSizeInt32Max(t *testing.T) {
	_, pageSize := NormalizePagination(1, math.MaxInt32)
	if pageSize != 100 {
		t.Errorf("NormalizePagination(1, INT32_MAX) pageSize = %d, want 100", pageSize)
	}
}

// AC-02: Default values when no pagination params provided
func TestNormalizePagination_DefaultValues(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		wantPage     int
		wantPageSize int
	}{
		{"no page and pageSize (both zero)", 0, 0, 1, 20},
		{"only page provided", 3, 0, 3, 20},
		{"only pageSize provided", 0, 50, 1, 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPage, gotPageSize := NormalizePagination(tt.page, tt.pageSize)
			if gotPage != tt.wantPage {
				t.Errorf("NormalizePagination() page = %d, want %d", gotPage, tt.wantPage)
			}
			if gotPageSize != tt.wantPageSize {
				t.Errorf("NormalizePagination() pageSize = %d, want %d", gotPageSize, tt.wantPageSize)
			}
		})
	}
}

// AC-05: Custom configuration takes effect
func TestNormalizePaginationWithConfig_AC05_CustomConfig(t *testing.T) {
	cfg := PaginationConfig{
		DefaultPageSize: 15,
		MaxPageSize:     50,
		HardLimit:       500,
	}
	tests := []struct {
		name         string
		page         int
		pageSize     int
		wantPage     int
		wantPageSize int
	}{
		{"custom default pageSize (pageSize=0)", 1, 0, 1, 15},
		{"custom max pageSize (pageSize=60)", 1, 60, 1, 50},
		{"custom normal value (pageSize=30)", 3, 30, 3, 30},
		{"custom hard limit does not affect normal mode", 1, 600, 1, 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPage, gotPageSize := NormalizePaginationWithConfig(tt.page, tt.pageSize, cfg)
			if gotPage != tt.wantPage {
				t.Errorf("NormalizePaginationWithConfig() page = %d, want %d", gotPage, tt.wantPage)
			}
			if gotPageSize != tt.wantPageSize {
				t.Errorf("NormalizePaginationWithConfig() pageSize = %d, want %d", gotPageSize, tt.wantPageSize)
			}
		})
	}
}

// AC-09: Pagination parameter correction logging
// The NormalizePagination and NormalizeQueryOption functions log warnings when
// parameters are corrected. We verify the correction logic works correctly;
// the actual log output is verified by manual inspection or integration tests.
func TestNormalizePagination_CorrectionDetection(t *testing.T) {
	tests := []struct {
		name          string
		page          int
		pageSize      int
		expectPageMod bool
		expectSizeMod bool
	}{
		{"no correction needed", 1, 20, false, false},
		{"page corrected", -1, 20, true, false},
		{"pageSize corrected (zero)", 1, 0, false, true},
		{"pageSize corrected (over max)", 1, 200, false, true},
		{"both corrected", -1, 0, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPage, gotPageSize := NormalizePagination(tt.page, tt.pageSize)
			pageWasModified := (gotPage != tt.page)
			sizeWasModified := (gotPageSize != tt.pageSize)
			if pageWasModified != tt.expectPageMod {
				t.Errorf("page modification: got=%v, want=%v (input=%d, output=%d)",
					pageWasModified, tt.expectPageMod, tt.page, gotPage)
			}
			if sizeWasModified != tt.expectSizeMod {
				t.Errorf("pageSize modification: got=%v, want=%v (input=%d, output=%d)",
					sizeWasModified, tt.expectSizeMod, tt.pageSize, gotPageSize)
			}
		})
	}
}

// AC-10: no_paging mode safety limit
func TestNormalizeQueryOption_AC10_NoPagingMode(t *testing.T) {
	tests := []struct {
		name         string
		noPaging     bool
		pageSize     int32
		wantPageSize int32
	}{
		{"no_paging=true, pageSize=0", true, 0, 1000},
		{"no_paging=true, pageSize=5000 (over HardLimit)", true, 5000, 1000},
		{"no_paging=true, pageSize=500 (within HardLimit)", true, 500, 500},
		{"no_paging=false, pageSize=0", false, 0, 20},
		{"no_paging=false, pageSize=200 (over MaxPageSize)", false, 200, 100},
		{"no_paging=false, pageSize=50 (normal)", false, 50, 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := NormalizeQueryOption(&QueryOption{NoPaging: tt.noPaging, PageSize: tt.pageSize})
			if opt.PageSize != tt.wantPageSize {
				t.Errorf("NormalizeQueryOption(NoPaging=%v, PageSize=%d) PageSize = %d, want %d",
					tt.noPaging, tt.pageSize, opt.PageSize, tt.wantPageSize)
			}
		})
	}
}

// AC-10: no_paging mode with custom HardLimit
func TestNormalizeQueryOption_AC10_NoPagingCustomHardLimit(t *testing.T) {
	originalConfig := globalConfig
	defer func() { globalConfig = originalConfig }()

	InitPaginationConfig(PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
		HardLimit:       500,
	})

	opt := NormalizeQueryOption(&QueryOption{NoPaging: true, PageSize: 0})
	if opt.PageSize != 500 {
		t.Errorf("NormalizeQueryOption(NoPaging=true, PageSize=0) with HardLimit=500: PageSize = %d, want 500", opt.PageSize)
	}

	opt = NormalizeQueryOption(&QueryOption{NoPaging: true, PageSize: 600})
	if opt.PageSize != 500 {
		t.Errorf("NormalizeQueryOption(NoPaging=true, PageSize=600) with HardLimit=500: PageSize = %d, want 500", opt.PageSize)
	}
}

// QueryOptionFromRequest integration test
func TestQueryOptionFromRequest_Integration(t *testing.T) {
	// Mock request implementing PaginatingRequest and KeywordRequest
	req := &mockListRequest{page: 3, pageSize: 50, keyword: "test"}
	opt := QueryOptionFromRequest(req)
	if opt.Page != 3 {
		t.Errorf("QueryOptionFromRequest() Page = %d, want 3", opt.Page)
	}
	if opt.PageSize != 50 {
		t.Errorf("QueryOptionFromRequest() PageSize = %d, want 50", opt.PageSize)
	}
	if opt.Keyword != "test" {
		t.Errorf("QueryOptionFromRequest() Keyword = %q, want %q", opt.Keyword, "test")
	}
}

// QueryOptionFromRequest with invalid pagination params should normalize
func TestQueryOptionFromRequest_InvalidPagination(t *testing.T) {
	req := &mockListRequest{page: -1, pageSize: 99999, keyword: ""}
	opt := QueryOptionFromRequest(req)
	if opt.Page != 1 {
		t.Errorf("QueryOptionFromRequest(invalid) Page = %d, want 1", opt.Page)
	}
	if opt.PageSize != 100 {
		t.Errorf("QueryOptionFromRequest(invalid) PageSize = %d, want 100", opt.PageSize)
	}
}

// QueryOptionFromRequest with nil-like request should return defaults
func TestQueryOptionFromRequest_NoPaginationInterface(t *testing.T) {
	req := &mockNoPageRequest{}
	opt := QueryOptionFromRequest(req)
	if opt.Page != 1 {
		t.Errorf("QueryOptionFromRequest(no pagination) Page = %d, want 1", opt.Page)
	}
	if opt.PageSize != 20 {
		t.Errorf("QueryOptionFromRequest(no pagination) PageSize = %d, want 20", opt.PageSize)
	}
}

// NormalizeQueryOption preserves other fields
func TestNormalizeQueryOption_PreservesOtherFields(t *testing.T) {
	opt := &QueryOption{
		Page:       -1,
		PageSize:   0,
		Keyword:    "search",
		PageToken:  "token123",
		PagingMode: "offset",
		OnlyCount:  true,
		NoPaging:   false,
		OrderBy:    []string{"created_at desc"},
		Status:     int32Ptr(1),
	}
	normalized := NormalizeQueryOption(opt)
	if normalized.Keyword != "search" {
		t.Errorf("NormalizeQueryOption() Keyword = %q, want %q", normalized.Keyword, "search")
	}
	if normalized.PageToken != "token123" {
		t.Errorf("NormalizeQueryOption() PageToken = %q, want %q", normalized.PageToken, "token123")
	}
	if normalized.PagingMode != "offset" {
		t.Errorf("NormalizeQueryOption() PagingMode = %q, want %q", normalized.PagingMode, "offset")
	}
	if !normalized.OnlyCount {
		t.Errorf("NormalizeQueryOption() OnlyCount = false, want true")
	}
	if len(normalized.OrderBy) != 1 || normalized.OrderBy[0] != "created_at desc" {
		t.Errorf("NormalizeQueryOption() OrderBy = %v, want [created_at desc]", normalized.OrderBy)
	}
	if normalized.Status == nil || *normalized.Status != 1 {
		t.Errorf("NormalizeQueryOption() Status = %v, want 1", normalized.Status)
	}
}

// AC-03: Full pageSize boundary table
func TestNormalizePagination_PageSizeBoundary_AC03(t *testing.T) {
	tests := []struct {
		name     string
		pageSize int
		want     int
	}{
		{"pageSize=10 (normal)", 10, 10},
		{"pageSize=100 (boundary max)", 100, 100},
		{"pageSize=101 (over max)", 101, 100},
		{"pageSize=10000 (far over max)", 10000, 100},
		{"pageSize=INT32_MAX", math.MaxInt32, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, pageSize := NormalizePagination(1, tt.pageSize)
			if pageSize != tt.want {
				t.Errorf("NormalizePagination(1, %d) pageSize = %d, want %d", tt.pageSize, pageSize, tt.want)
			}
		})
	}
}

// Helper types for QueryOptionFromRequest tests

type mockListRequest struct {
	page     int32
	pageSize int32
	keyword  string
}

func (r *mockListRequest) GetPage() int32     { return r.page }
func (r *mockListRequest) GetPageSize() int32  { return r.pageSize }
func (r *mockListRequest) GetKeyword() string  { return r.keyword }

type mockNoPageRequest struct{}

func (r *mockNoPageRequest) GetKeyword() string { return "test" }

// Helper to create int32 pointer
func int32Ptr(v int32) *int32 { return &v }
