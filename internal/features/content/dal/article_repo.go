/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/article"
	"origadmin/application/origcms/internal/features/content/biz"
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

func (r *articleRepo) Create(ctx context.Context, a *biz.Article) (*biz.Article, error) {
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
	if a.UserID != "" {
		builder = builder.SetUserID(a.UserID)
	}
	if a.CategoryID != 0 {
		builder = builder.SetCategoryID(a.CategoryID)
	}
	if a.MediaID != "" {
		builder = builder.SetMediaID(a.MediaID)
	}
	if a.Thumbnail != "" {
		builder = builder.SetThumbnail(a.Thumbnail)
	}
	if !a.PublishedAt.IsZero() {
		builder = builder.SetPublishedAt(a.PublishedAt)
	}

	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create article: %w", err)
	}
	return mapArticleEntity(ent), nil
}

func (r *articleRepo) Update(ctx context.Context, a *biz.Article) (*biz.Article, error) {
	builder := r.data.db.Article.UpdateOneID(a.ID).
		SetTitle(a.Title).
		SetContent(a.Content).
		SetState(a.State)

	if a.Summary != "" {
		builder = builder.SetSummary(a.Summary)
	}
	if a.Slug != "" {
		builder = builder.SetSlug(a.Slug)
	}
	builder = builder.SetFeatured(a.Featured)
	if a.Tags != nil {
		builder = builder.SetTags(a.Tags)
	}
	if a.CategoryID != 0 {
		builder = builder.SetCategoryID(a.CategoryID)
	}
	if a.MediaID != "" {
		builder = builder.SetMediaID(a.MediaID)
	}
	builder = builder.SetThumbnail(a.Thumbnail) // Allow empty string to clear
	if !a.PublishedAt.IsZero() {
		builder = builder.SetPublishedAt(a.PublishedAt)
	}

	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update article: %w", err)
	}
	return mapArticleEntity(ent), nil
}

func (r *articleRepo) Delete(ctx context.Context, id string) error {
	return r.data.db.Article.DeleteOneID(id).Exec(ctx)
}

func (r *articleRepo) Get(ctx context.Context, id string) (*biz.Article, error) {
	// Try short_token first (same pattern as Media repo)
	ent, err := r.data.db.Article.Query().
		Where(article.ShortTokenEQ(id)).
		WithMedia().
		WithCategory().
		Only(ctx)
	if err == nil {
		return mapArticleEntity(ent), nil
	}
	// Fall back to ID lookup
	ent, err = r.data.db.Article.Query().
		Where(article.IDEQ(id)).
		WithMedia().
		WithCategory().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get article: %w", err)
	}
	return mapArticleEntity(ent), nil
}

func (r *articleRepo) GetBySlug(ctx context.Context, slug string) (*biz.Article, error) {
	ent, err := r.data.db.Article.Query().
		Where(article.SlugEQ(slug)).
		WithMedia().
		WithCategory().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get article by slug: %w", err)
	}
	return mapArticleEntity(ent), nil
}

func (r *articleRepo) GetByShortToken(ctx context.Context, shortToken string) (*biz.Article, error) {
	ent, err := r.data.db.Article.Query().
		Where(article.ShortTokenEQ(shortToken)).
		WithMedia().
		WithCategory().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get article by short_token: %w", err)
	}
	return mapArticleEntity(ent), nil
}

func (r *articleRepo) List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*biz.Article, int64, error) {
	query := r.data.db.Article.Query()

	// Apply filters
	if state, ok := filters["state"]; ok {
		if s, ok := state.(string); ok && s != "" {
			query = query.Where(article.StateEQ(s))
		}
	}
	if featured, ok := filters["featured"]; ok {
		if f, ok := featured.(bool); ok && f {
			query = query.Where(article.FeaturedEQ(f))
		}
	}
	if userID, ok := filters["user_id"]; ok {
		if u, ok := userID.(string); ok && u != "" {
			query = query.Where(article.UserIDEQ(u))
		}
	}
	if categoryID, ok := filters["category_id"]; ok {
		switch v := categoryID.(type) {
		case int64:
			if v != 0 {
				query = query.Where(article.CategoryIDEQ(v))
			}
		case int:
			if v != 0 {
				query = query.Where(article.CategoryIDEQ(int64(v)))
			}
		}
	}

	// Count total
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count articles: %w", err)
	}

	// Apply pagination and ordering
	offset := (page - 1) * pageSize
	ents, err := query.
		Order(entity.Desc(article.FieldCreateTime)).
		Offset(offset).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list articles: %w", err)
	}

	result := make([]*biz.Article, len(ents))
	for i, ent := range ents {
		result[i] = mapArticleEntity(ent)
	}
	return result, int64(total), nil
}

func (r *articleRepo) UpdateState(ctx context.Context, id string, state string) error {
	return r.data.db.Article.UpdateOneID(id).
		SetState(state).
		Exec(ctx)
}

// mapArticleEntity maps an entity.Article to a biz.Article.
func mapArticleEntity(ent *entity.Article) *biz.Article {
	a := &biz.Article{
		ID:           ent.ID,
		Title:        ent.Title,
		Content:      ent.Content,
		Summary:      ent.Summary,
		Slug:         ent.Slug,
		ShortToken:   ent.ShortToken,
		State:        ent.State,
		ViewCount:    ent.ViewCount,
		CommentCount: ent.CommentCount,
		Featured:     ent.Featured,
		Tags:         ent.Tags,
		UserID:       ent.UserID,
		CategoryID:   ent.CategoryID,
		MediaID:      ent.MediaID,
		Thumbnail:    ent.Thumbnail,
		PublishedAt:  ent.PublishedAt,
		CreateTime:   ent.CreateTime,
		UpdateTime:   ent.UpdateTime,
	}
	if a.Tags == nil {
		a.Tags = []string{}
	}
	// Map eager-loaded media edge
	if media := ent.Edges.Media; media != nil {
		a.Media = &biz.MediaBrief{
			ID:         media.ID,
			Title:      media.Title,
			Thumbnail:  media.Thumbnail,
			Duration:   media.Duration,
			Type:       media.Type,
			ShortToken: media.ShortToken,
		}
		// Auto-populate thumbnail from media if not set
		if a.Thumbnail == "" && media.Thumbnail != "" {
			a.Thumbnail = media.Thumbnail
		}
	}
	return a
}
