/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/media"
	"origadmin/application/origstudio/internal/features/content/biz"
)

type feedRepo struct {
	data *Data
	log  *log.Helper
}

func NewFeedRepo(data *Data, logger log.Logger) biz.FeedRepo {
	return &feedRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "feed.data")),
	}
}

func (r *feedRepo) ListLatest(ctx context.Context, page, pageSize int) ([]*biz.MediaInfo, int, error) {
	query := r.data.db.Media.Query().Where(media.StateEQ("active"))
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	ents, err := query.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Order(entity.Desc(media.FieldCreateTime)).
		WithUser().
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	res := make([]*biz.MediaInfo, len(ents))
	for i, ent := range ents {
		res[i] = mapMediaInfo(ent)
	}
	return res, total, nil
}

func (r *feedRepo) ListTrending(ctx context.Context, page, pageSize int) ([]*biz.MediaInfo, int, error) {
	query := r.data.db.Media.Query().Where(media.StateEQ("active"))
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	ents, err := query.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Order(entity.Desc(media.FieldViewCount)).
		WithUser().
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	res := make([]*biz.MediaInfo, len(ents))
	for i, ent := range ents {
		res[i] = mapMediaInfo(ent)
	}
	return res, total, nil
}

func (r *feedRepo) ListFeatured(ctx context.Context, page, pageSize int) ([]*biz.MediaInfo, int, error) {
	query := r.data.db.Media.Query().Where(media.StateEQ("active"), media.FeaturedEQ(true))
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	ents, err := query.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Order(entity.Desc(media.FieldCreateTime)).
		WithUser().
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	res := make([]*biz.MediaInfo, len(ents))
	for i, ent := range ents {
		res[i] = mapMediaInfo(ent)
	}
	return res, total, nil
}

func mapMediaInfo(ent *entity.Media) *biz.MediaInfo {
	username := ""
	if ent.Edges.User != nil {
		username = ent.Edges.User.Username
	}
	return &biz.MediaInfo{
		ID:          int64(len(ent.ID)),
		Title:       ent.Title,
		Description: ent.Description,
		Thumbnail:   ent.Thumbnail,
		Duration:    ent.Duration,
		ViewCount:   ent.ViewCount,
		UserID:      int(len(ent.UserID)),
		Username:    username,
		Type:        ent.Type,
		URL:         fmt.Sprintf("/v/%d", len(ent.ID)),
	}
}
