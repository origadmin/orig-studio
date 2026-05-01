package dal

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/commentlike"
	"origadmin/application/origcms/internal/features/content/biz"
)

type commentLikeRepo struct {
	data *Data
	log  *log.Helper
}

func NewCommentLikeRepo(data *Data, logger log.Logger) biz.CommentLikeRepo {
	return &commentLikeRepo{data: data, log: log.NewHelper(log.With(logger, "module", "comment_like.data"))}
}

func (r *commentLikeRepo) Create(ctx context.Context, userID, commentID string, likeType string) (*biz.CommentLike, error) {
	ent, err := r.data.db.CommentLike.Create().
		SetCommentID(commentID).
		SetUserID(userID).
		SetLikeType(likeType).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return &biz.CommentLike{
		ID:        ent.ID,
		UserID:    userID,
		CommentID: commentID,
		LikeType:  ent.LikeType,
		CreateTime: ent.CreateTime,
	}, nil
}

func (r *commentLikeRepo) Delete(ctx context.Context, userID, commentID string) error {
	_, err := r.data.db.CommentLike.Delete().
		Where(
			commentlike.CommentIDEQ(commentID),
			commentlike.UserIDEQ(userID),
		).
		Exec(ctx)
	return err
}

func (r *commentLikeRepo) GetStatus(ctx context.Context, userID, commentID string) (string, error) {
	ent, err := r.data.db.CommentLike.Query().
		Where(
			commentlike.CommentIDEQ(commentID),
			commentlike.UserIDEQ(userID),
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

func (r *commentLikeRepo) CountByComment(ctx context.Context, commentID string, likeType string) (int64, error) {
	count, err := r.data.db.CommentLike.Query().
		Where(
			commentlike.CommentIDEQ(commentID),
			commentlike.LikeTypeEQ(likeType),
		).
		Count(ctx)
	return int64(count), err
}
