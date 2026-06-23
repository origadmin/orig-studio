/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import (
	"fmt"

	kratoslog "github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/features/media/biz"
)

// NewStorage creates a LocalStorage instance (used as the base for all storage types).
func NewStorage(sp *conf.StoragePaths) *LocalStorage {
	return NewLocalStorage(sp)
}

// NewStorageInterface creates the appropriate Storage implementation based on
// StorageConfig.Type. It always creates LocalStorage; for "s3" type it also
// creates S3Storage; for "hybrid" type it creates HybridStorage with async sync.
func NewStorageInterface(
	local *LocalStorage,
	sp *conf.StoragePaths,
	cfg *conf.StorageConfig,
	logger kratoslog.Logger,
) (biz.Storage, func(), error) {
	switch cfg.Type {
	case conf.StorageTypeS3:
		s3Storage, err := NewS3Storage(&cfg.S3, logger)
		if err != nil {
			return nil, func() {}, fmt.Errorf("create S3 storage: %w", err)
		}
		return s3Storage, func() {}, nil

	case conf.StorageTypeHybrid:
		s3Storage, err := NewS3Storage(&cfg.S3, logger)
		if err != nil {
			return nil, func() {}, fmt.Errorf("create S3 storage for hybrid: %w", err)
		}
		hs := NewHybridStorage(local, s3Storage, sp, cfg.Hybrid, logger)
		return hs, func() { hs.Close() }, nil

	case conf.StorageTypeLocal:
		fallthrough
	default:
		return local, func() {}, nil
	}
}
