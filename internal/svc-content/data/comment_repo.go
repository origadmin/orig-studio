/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/comment"
	"origadmin/application/origcms/internal/data/entity/media"
	"origadmin/application/origcms/internal/svc-content/biz"
)

type commentRepo struct {
	data *Data
	log  *log.Helper
}

// Data is a wrapper for ent client, similar to other services.
type Data struct {
	db *entity.Client
}

func NewData(db *entity.Client) *Data {
	return &Data{db: db}
}

func NewCommentRepo(data *Data, logger log.Logger) biz.CommentRepo {
	return &commentRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "comment.data")),
	}
}

func (r *commentRepo) Create(ctx context.Context, c *biz.Comment) (*biz.Comment, error) {
	builder := r.data.db.Comment.Create().
		SetText(c.Text).
		SetMediaID(c.MediaID).
		SetUserID(c.UserID)

	if c.ParentID != nil {
		builder.SetParentID(*c.ParentID)
	}

	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapComment(ent), nil
}

func (r *commentRepo) Get(ctx context.Context, id int) (*biz.Comment, error) {
	ent, err := r.data.db.Comment.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapComment(ent), nil
}

func (r *commentRepo) Update(ctx context.Context, c *biz.Comment) (*biz.Comment, error) {
	ent, err := r.data.db.Comment.UpdateOneID(c.ID).
		SetText(c.Text).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapComment(ent), nil
}

func (r *commentRepo) Delete(ctx context.Context, id int) error {
	return r.data.db.Comment.DeleteOneID(id).Exec(ctx)
}

func (r *commentRepo) ListByMedia(ctx context.Context, mediaID int, page, pageSize int) ([]*biz.Comment, int, error) {
	query := r.data.db.Comment.Query().Where(comment.HasMediaWith(media.IDEQ(mediaID)))
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	ents, err := query.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Order(entity.Desc(comment.FieldAddDate)).
		WithUser().
		WithReplies(func(rq *entity.CommentQuery) {
			rq.WithUser()
		}).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	res := make([]*biz.Comment, len(ents))
	for i, ent := range ents {
		res[i] = mapComment(ent)
	}
	return res, total, nil
}

func (r *commentRepo) ListAll(ctx context.Context, page, pageSize int) ([]*biz.Comment, int, error) {
	query := r.data.db.Comment.Query()
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	ents, err := query.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Order(entity.Desc(comment.FieldAddDate)).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	res := make([]*biz.Comment, len(ents))
	for i, ent := range ents {
		res[i] = mapComment(ent)
	}
	return res, total, nil
}

func mapComment(ent *entity.Comment) *biz.Comment {
	c := &biz.Comment{
		ID:        ent.ID,
		UID:       ent.UID,
		Text:      ent.Text,
		AddDate:   ent.AddDate,
		UpdatedAt: ent.AddDate, // Ent comment schema doesn't have updated_at yet?
	}

	if ent.ParentID != 0 {
		pid := ent.ParentID
		c.ParentID = &pid
	}
	if ent.Edges.Media != nil {
		c.MediaID = ent.Edges.Media.ID
	}
	if ent.Edges.User != nil {
		c.UserID = ent.Edges.User.ID
		u := ent.Edges.User
		c.User = &biz.User{
			ID:       u.ID,
			Username: u.Username,
			Name:     u.Name,
			Logo:     u.Logo,
		}
	}
	if len(ent.Edges.Replies) > 0 {
		c.Replies = make([]*biz.Comment, len(ent.Edges.Replies))
		for i, r := range ent.Edges.Replies {
			c.Replies[i] = mapComment(r)
		}
	}
	return c
}
