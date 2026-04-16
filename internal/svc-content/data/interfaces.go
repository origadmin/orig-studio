/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package data

import (
	"context"

	"origadmin/application/origcms/internal/data/entity"
)

// ArticleRepo 文章仓库接口
type ArticleRepo interface {
	Create(ctx context.Context, article *entity.Article) error
	Update(ctx context.Context, article *entity.Article) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*entity.Article, error)
	List(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*entity.Article, int64, error)
}
