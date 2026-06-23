/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/conf"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
)

// NewSpriteUseCaseWithConfig creates a new sprite use case with paths from StoragePaths.
// This is the wire-compatible constructor that injects StoragePaths.
func NewSpriteUseCaseWithConfig(
	mediaRepo MediaRepo,
	settingUC *systembiz.SettingUseCase,
	sp *conf.StoragePaths,
	logger log.Logger,
) *SpriteUseCase {
	return NewSpriteUseCase(mediaRepo, settingUC, sp, logger)
}
