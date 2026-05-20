/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package types provides common domain types including pagination and query utilities.
package types

import (
	"github.com/origadmin/runtime/log"
)

// PaginationConfig holds configurable pagination parameters.
// Values can be overridden via YAML configuration.
type PaginationConfig struct {
	DefaultPageSize int `json:"default_page_size" yaml:"default_page_size"` // Default: 20
	MaxPageSize     int `json:"max_page_size" yaml:"max_page_size"`         // Default: 100
	HardLimit       int `json:"hard_limit" yaml:"hard_limit"`               // Default: 1000
}

// DefaultPaginationConfig returns the default pagination configuration.
func DefaultPaginationConfig() PaginationConfig {
	return PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
		HardLimit:       1000,
	}
}

// globalConfig holds the active pagination configuration.
var globalConfig = DefaultPaginationConfig()

// InitPaginationConfig initializes the global pagination configuration.
// This should be called once during application startup.
func InitPaginationConfig(cfg PaginationConfig) {
	if cfg.DefaultPageSize <= 0 {
		cfg.DefaultPageSize = 20
	}
	if cfg.MaxPageSize <= 0 {
		cfg.MaxPageSize = 100
	}
	if cfg.HardLimit <= 0 {
		cfg.HardLimit = 1000
	}
	// Ensure MaxPageSize does not exceed HardLimit
	if cfg.MaxPageSize > cfg.HardLimit {
		cfg.MaxPageSize = cfg.HardLimit
	}
	globalConfig = cfg
}

// GetPaginationConfig returns the current global pagination configuration.
func GetPaginationConfig() PaginationConfig {
	return globalConfig
}

// NormalizePagination validates and corrects pagination parameters using the global config.
// page < 1 -> corrected to 1
// pageSize <= 0 -> corrected to DefaultPageSize
// pageSize > MaxPageSize -> corrected to MaxPageSize
func NormalizePagination(page, pageSize int) (int, int) {
	return NormalizePaginationWithConfig(page, pageSize, globalConfig)
}

// NormalizePaginationWithConfig validates and corrects pagination parameters using the given config.
func NormalizePaginationWithConfig(page, pageSize int, cfg PaginationConfig) (int, int) {
	origPage := page
	origPageSize := pageSize

	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = cfg.DefaultPageSize
	}
	if pageSize > cfg.MaxPageSize {
		pageSize = cfg.MaxPageSize
	}

	// Log correction when values were adjusted
	if origPage != page || origPageSize != pageSize {
		log.Warnf("pagination param corrected: page %d -> %d, page_size %d -> %d",
			origPage, page, origPageSize, pageSize)
	}

	return page, pageSize
}

// NormalizeQueryOption validates and corrects a complete QueryOption.
// If opt is nil, a new QueryOption with default values is returned.
func NormalizeQueryOption(opt *QueryOption) *QueryOption {
	if opt == nil {
		return &QueryOption{
			Page:     1,
			PageSize: int32(globalConfig.DefaultPageSize),
		}
	}

	origPage := opt.Page
	origPageSize := opt.PageSize

	if opt.Page < 1 {
		opt.Page = 1
	}

	// no_paging mode uses HardLimit
	if opt.NoPaging {
		if opt.PageSize <= 0 || int(opt.PageSize) > globalConfig.HardLimit {
			opt.PageSize = int32(globalConfig.HardLimit)
		}
	} else {
		if opt.PageSize <= 0 {
			opt.PageSize = int32(globalConfig.DefaultPageSize)
		}
		if int(opt.PageSize) > globalConfig.MaxPageSize {
			opt.PageSize = int32(globalConfig.MaxPageSize)
		}
	}

	// Log correction when values were adjusted
	if origPage != opt.Page || origPageSize != opt.PageSize {
		log.Warnf("pagination param corrected: page %d -> %d, page_size %d -> %d",
			origPage, opt.Page, origPageSize, opt.PageSize)
	}

	return opt
}

// CalculateTotalPages computes the total number of pages.
func CalculateTotalPages(totalSize int64, pageSize int) int32 {
	if pageSize <= 0 {
		return 0
	}
	return int32((totalSize + int64(pageSize) - 1) / int64(pageSize))
}

// NormalizeHTTPPagination validates and corrects page/pageSize parsed from HTTP query params.
// This is a convenience function for Gin HTTP handlers that parse page/pageSize from query strings.
// It returns the corrected page and pageSize values.
func NormalizeHTTPPagination(page, pageSize int) (int, int) {
	return NormalizePagination(page, pageSize)
}
