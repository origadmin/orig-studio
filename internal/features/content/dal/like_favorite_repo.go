/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/favorite"
	"origadmin/application/origcms/internal/data/entity/like"
	"origadmin/application/origcms/internal/data/entity/media"
	"origadmin/application/origcms/internal/data/entity/user"
	"origadmin/application/origcms/internal/features/content/biz"
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
	return &favoriteRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "favorite.data")),
	}
}

// ─── Like repo ───────────────────────────────────────────────────────────────

func (r *likeRepo) Create(
	ctx context.Context,
	userID, mediaID string,
	likeType string,
) (*biz.Like, error) {
	ent, err := r.data.db.Like.Create().
		SetMediaID(mediaID).
		SetUserID(userID).
		SetLikeType(likeType).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return &biz.Like{
		ID:        ent.ID,
		UserID:    userID,
		MediaID:   mediaID,
		LikeType:  ent.LikeType,
		CreateTime: ent.CreateTime,
	}, nil
}

func (r *likeRepo) Delete(ctx context.Context, userID, mediaID string) error {
	_, err := r.data.db.Like.Delete().
		Where(
			like.HasMediaWith(media.IDEQ(mediaID)),
			like.HasUserWith(user.IDEQ(userID)),
		).
		Exec(ctx)
	return err
}

func (r *likeRepo) GetStatus(ctx context.Context, userID, mediaID string) (string, error) {
	ent, err := r.data.db.Like.Query().
		Where(
			like.HasMediaWith(media.IDEQ(mediaID)),
			like.HasUserWith(user.IDEQ(userID)),
		).
		Only(ctx)
	if err != nil {
		if entity.IsNotFound(err) {
			return "none", nil
		}
		return "none", err
	}
	return ent.LikeType, nil
}

func (r *likeRepo) CountByMedia(ctx context.Context, mediaID string, likeType string) (int64, error) {
	count, err := r.data.db.Like.Query().
		Where(
			like.HasMediaWith(media.IDEQ(mediaID)),
			like.LikeTypeEQ(likeType),
		).
		Count(ctx)
	return int64(count), err
}

func (r *likeRepo) ListByUser(ctx context.Context, userID string) ([]*biz.Like, error) {
	ents, err := r.data.db.Like.Query().
		Where(like.HasUserWith(user.IDEQ(userID))).
		Order(entity.Desc(like.FieldCreateTime)).
		WithMedia().
		All(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*biz.Like, len(ents))
	for i, ent := range ents {
		mediaID := ""
		if ent.Edges.Media != nil {
			mediaID = ent.Edges.Media.ID
		}
		res[i] = &biz.Like{
			ID:        ent.ID,
			UserID:    userID,
			MediaID:   mediaID,
			LikeType:  ent.LikeType,
			CreateTime: ent.CreateTime,
		}
	}
	return res, nil
}

// ─── Favorite repo ────────────────────────────────────────────────────────────

func (r *favoriteRepo) Create(ctx context.Context, userID, mediaID string) (*biz.Favorite, error) {
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
		CreateTime: ent.CreateTime,
	}, nil
}

func (r *favoriteRepo) Delete(ctx context.Context, userID, mediaID string) error {
	_, err := r.data.db.Favorite.Delete().
		Where(
			favorite.HasMediaWith(media.IDEQ(mediaID)),
			favorite.HasUserWith(user.IDEQ(userID)),
		).
		Exec(ctx)
	return err
}

func (r *favoriteRepo) IsFavorited(ctx context.Context, userID, mediaID string) (bool, error) {
	return r.data.db.Favorite.Query().
		Where(
			favorite.HasMediaWith(media.IDEQ(mediaID)),
			favorite.HasUserWith(user.IDEQ(userID)),
		).
		Exist(ctx)
}

func (r *favoriteRepo) CountByMedia(ctx context.Context, mediaID string) (int64, error) {
	count, err := r.data.db.Favorite.Query().
		Where(favorite.HasMediaWith(media.IDEQ(mediaID))).
		Count(ctx)
	return int64(count), err
}

func (r *favoriteRepo) ListByUser(ctx context.Context, userID string) ([]*biz.Favorite, error) {
	ents, err := r.data.db.Favorite.Query().
		Where(favorite.HasUserWith(user.IDEQ(userID))).
		Order(entity.Desc(favorite.FieldCreateTime)).
		WithMedia().
		All(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*biz.Favorite, len(ents))
	for i, ent := range ents {
		mediaID := ""
		if ent.Edges.Media != nil {
			mediaID = ent.Edges.Media.ID
		}
		res[i] = &biz.Favorite{
			ID:        ent.ID,
			UserID:    userID,
			MediaID:   mediaID,
			CreateTime: ent.CreateTime,
		}
	}
	return res, nil
}
