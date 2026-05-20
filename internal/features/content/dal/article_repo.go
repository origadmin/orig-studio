/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/api/gen/v1/types"
	"origadmin/application/origstudio/internal/dal/convpb"
	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/article"
	"origadmin/application/origstudio/internal/features/content/biz"
)

type articleRepo struct {
	data *Data
	log  *log.Helper
}

// NewArticleRepo creates a new article repository.
func NewArticleRepo(data *Data, logger log.Logger) biz.ArticleRepo {
	return &articleRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "article.data")),
	}
}

func (r *articleRepo) Create(ctx context.Context, a *types.Article) (*types.Article, error) {
	builder := r.data.db.Article.Create().
		SetTitle(a.Title).
		SetContent(a.Content).
		SetState(a.State)

	if a.Summary != "" {
		builder = builder.SetSummary(a.Summary)
	}
	if a.Slug != "" {
		builder = builder.SetSlug(a.Slug)
	}
	if a.Featured {
		builder = builder.SetFeatured(a.Featured)
	}
	if len(a.Tags) > 0 {
		builder = builder.SetTags(a.Tags)
	}
	if a.UserId != "" {
		builder = builder.SetUserID(a.UserId)
	}
	if a.CategoryId != 0 {
		builder = builder.SetCategoryID(a.CategoryId)
	}
	if a.MediaId != "" {
		builder = builder.SetMediaID(a.MediaId)
	}
	if a.Thumbnail != "" {
		builder = builder.SetThumbnail(a.Thumbnail)
	}
	if a.PublishedAt != nil && !a.PublishedAt.AsTime().IsZero() {
		builder = builder.SetPublishedAt(a.PublishedAt.AsTime())
	}

	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create article: %w", err)
	}
	return convpb.ConvertArticleToArticlePB(ent), nil
}

func (r *articleRepo) Update(ctx context.Context, a *types.Article) (*types.Article, error) {
	builder := r.data.db.Article.UpdateOneID(a.Id).
		SetTitle(a.Title).
		SetContent(a.Content).
		SetState(a.State)

	if a.Summary != "" {
		builder = builder.SetSummary(a.Summary)
	}
	if a.Slug != "" {
		builder = builder.SetSlug(a.Slug)
	}
	if a.Featured {
		builder = builder.SetFeatured(a.Featured)
	}
	if a.Tags != nil {
		builder = builder.SetTags(a.Tags)
	}
	if a.CategoryId != 0 {
		builder = builder.SetCategoryID(a.CategoryId)
	}
	if a.MediaId != "" {
		builder = builder.SetMediaID(a.MediaId)
	}
	builder = builder.SetThumbnail(a.Thumbnail)
	if a.PublishedAt != nil && !a.PublishedAt.AsTime().IsZero() {
		builder = builder.SetPublishedAt(a.PublishedAt.AsTime())
	}

	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update article: %w", err)
	}
	return convpb.ConvertArticleToArticlePB(ent), nil
}

func (r *articleRepo) Delete(ctx context.Context, id string) error {
	return r.data.db.Article.DeleteOneID(id).Exec(ctx)
}

func (r *articleRepo) Get(ctx context.Context, id string) (*types.Article, error) {
	var ent *entity.Article
	var err error

	ent, err = r.data.db.Article.Query().
		Where(article.IDEQ(id)).
		Only(ctx)
	if err == nil {
		return convpb.ConvertArticleToArticlePB(ent), nil
	}

	ent, err = r.data.db.Article.Query().
		Where(article.ShortTokenEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get article: %w", err)
	}
	return convpb.ConvertArticleToArticlePB(ent), nil
}

func (r *articleRepo) GetBySlug(ctx context.Context, slug string) (*types.Article, error) {
	ent, err := r.data.db.Article.Query().
		Where(article.SlugEQ(slug)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get article by slug: %w", err)
	}
	return convpb.ConvertArticleToArticlePB(ent), nil
}

func (r *articleRepo) GetByShortToken(ctx context.Context, shortToken string) (*types.Article, error) {
	ent, err := r.data.db.Article.Query().
		Where(article.ShortTokenEQ(shortToken)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get article by short_token: %w", err)
	}
	return convpb.ConvertArticleToArticlePB(ent), nil
}

func (r *articleRepo) List(ctx context.Context, filters *types.ListArticlesRequest) (*types.ListArticlesResponse, error) {
	query := r.data.db.Article.Query()

	if filters.State != "" {
		query = query.Where(article.StateEQ(filters.State))
	}
	if filters.Featured {
		query = query.Where(article.FeaturedEQ(filters.Featured))
	}
	if filters.UserId != "" {
		query = query.Where(article.UserIDEQ(filters.UserId))
	}
	if filters.CategoryId != 0 {
		query = query.Where(article.CategoryIDEQ(filters.CategoryId))
	}

	page := int(filters.Page)
	pageSize := int(filters.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	ents, err := query.
		Order(entity.Desc(article.FieldCreateTime)).
		Offset(offset).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list articles: %w", err)
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count articles: %w", err)
	}

	articles := make([]*types.Article, len(ents))
	for i, ent := range ents {
		articles[i] = convpb.ConvertArticleToArticlePB(ent)
	}

	totalPages := int32(0)
	if total > 0 && pageSize > 0 {
		totalPages = int32((int64(total) + int64(pageSize) - 1) / int64(pageSize))
	}

	return &types.ListArticlesResponse{
		Articles:   articles,
		Total:      int32(total),
		Page:       int32(page),
		PageSize:   int32(pageSize),
		TotalPages: totalPages,
	}, nil
}

func (r *articleRepo) UpdateState(ctx context.Context, id string, state string) error {
	return r.data.db.Article.UpdateOneID(id).
		SetState(state).
		Exec(ctx)
}
