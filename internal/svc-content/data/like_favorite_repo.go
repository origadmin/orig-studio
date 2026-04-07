/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/favorite"
	"origadmin/application/origcms/internal/data/entity/like"
	"origadmin/application/origcms/internal/data/entity/media"
	"origadmin/application/origcms/internal/data/entity/user"
	"origadmin/application/origcms/internal/svc-content/biz"
)

type likeRepo struct {
	data *Data
	log  *log.Helper
}

type favoriteRepo struct {
	data *Data
	log  *log.Helper
}

func NewLikeRepo(data *Data, logger log.Logger) biz.LikeRepo {
	return &likeRepo{data: data, log: log.NewHelper(log.With(logger, "module", "like.data"))}
}

func NewFavoriteRepo(data *Data, logger log.Logger) biz.FavoriteRepo {
	return &favoriteRepo{data: data, log: log.NewHelper(log.With(logger, "module", "favorite.data"))}
}

// ─── Like repo ───────────────────────────────────────────────────────────────
//
// NOTE: The Like entity schema has no UserID/MediaID/LikeType fields.
// Relationships are stored as edges (media M2O, user O2M).
// LikeType is not persisted — all likes are treated as "like".

func (r *likeRepo) Create(ctx context.Context, userID, mediaID int, likeType string) (*biz.Like, error) {
	ent, err := r.data.db.Like.Create().
		SetMediaID(mediaID).
		SetUserID(userID).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return &biz.Like{
		ID:        ent.ID,
		UserID:    userID,
		MediaID:   mediaID,
		LikeType:  likeType, // returned as-is; not stored in schema
		CreatedAt: ent.CreatedAt,
	}, nil
}

func (r *likeRepo) Delete(ctx context.Context, userID, mediaID int) error {
	_, err := r.data.db.Like.Delete().
		Where(
			like.HasMediaWith(media.IDEQ(mediaID)),
			like.HasUserWith(user.IDEQ(userID)),
		).
		Exec(ctx)
	return err
}

func (r *likeRepo) GetStatus(ctx context.Context, userID, mediaID int) (string, error) {
	exists, err := r.data.db.Like.Query().
		Where(
			like.HasMediaWith(media.IDEQ(mediaID)),
			like.HasUserWith(user.IDEQ(userID)),
		).
		Exist(ctx)
	if err != nil {
		return "none", err
	}
	if exists {
		return "like", nil
	}
	return "none", nil
}

func (r *likeRepo) CountByMedia(ctx context.Context, mediaID int, likeType string) (int64, error) {
	// LikeType is not stored in schema; count all likes for the media.
	count, err := r.data.db.Like.Query().
		Where(like.HasMediaWith(media.IDEQ(mediaID))).
		Count(ctx)
	return int64(count), err
}

// ─── Favorite repo ────────────────────────────────────────────────────────────

func (r *favoriteRepo) Create(ctx context.Context, userID, mediaID int) (*biz.Favorite, error) {
	ent, err := r.data.db.Favorite.Create().
		SetMediaID(mediaID).
		SetUserID(userID).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return &biz.Favorite{
		ID:        ent.ID,
		UserID:    userID,
		MediaID:   mediaID,
		CreatedAt: ent.CreatedAt,
	}, nil
}

func (r *favoriteRepo) Delete(ctx context.Context, userID, mediaID int) error {
	_, err := r.data.db.Favorite.Delete().
		Where(
			favorite.HasMediaWith(media.IDEQ(mediaID)),
			favorite.HasUserWith(user.IDEQ(userID)),
		).
		Exec(ctx)
	return err
}

func (r *favoriteRepo) IsFavorited(ctx context.Context, userID, mediaID int) (bool, error) {
	return r.data.db.Favorite.Query().
		Where(
			favorite.HasMediaWith(media.IDEQ(mediaID)),
			favorite.HasUserWith(user.IDEQ(userID)),
		).
		Exist(ctx)
}

func (r *favoriteRepo) CountByMedia(ctx context.Context, mediaID int) (int64, error) {
	count, err := r.data.db.Favorite.Query().
		Where(favorite.HasMediaWith(media.IDEQ(mediaID))).
		Count(ctx)
	return int64(count), err
}

func (r *favoriteRepo) ListByUser(ctx context.Context, userID int) ([]*biz.Favorite, error) {
	ents, err := r.data.db.Favorite.Query().
		Where(favorite.HasUserWith(user.IDEQ(userID))).
		Order(entity.Desc(favorite.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*biz.Favorite, len(ents))
	for i, ent := range ents {
		res[i] = &biz.Favorite{
			ID:        ent.ID,
			UserID:    userID,
			MediaID:   0, // mediaID not directly on entity; load via edge if needed
			CreatedAt: ent.CreatedAt,
		}
	}
	return res, nil
}
