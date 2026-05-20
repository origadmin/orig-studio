package dal

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/history"
	"origadmin/application/origstudio/internal/dal/entity/media"
	"origadmin/application/origstudio/internal/dal/entity/predicate"
	"origadmin/application/origstudio/internal/dal/entity/user"
	"origadmin/application/origstudio/internal/features/content/biz"
)

type historyRepo struct {
	data *Data
	log  *log.Helper
}

func NewHistoryRepo(data *Data, logger log.Logger) biz.HistoryRepo {
	return &historyRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "history.data")),
	}
}

func toBizHistory(e *entity.History) *biz.History {
	return &biz.History{
		ID:              e.ID,
		UserID:          e.UserID,
		ContentID:       e.ContentID,
		ContentType:     string(e.ContentType),
		ProgressSeconds: e.ProgressSeconds,
		DurationSeconds: e.DurationSeconds,
		IsFinished:      e.IsFinished,
		Title:           e.Title,
		Thumbnail:       e.Thumbnail,
		ShortToken:      e.ShortToken,
		LastWatchedAt:   e.LastWatchedAt,
		CreateTime:      e.CreateTime,
		UpdateTime:      e.UpdateTime,
	}
}

func contentTypeToEnum(ct string) (history.ContentType, error) {
	switch ct {
	case "video":
		return history.ContentTypeVideo, nil
	case "article":
		return history.ContentTypeArticle, nil
	case "audio":
		return history.ContentTypeAudio, nil
	default:
		return "", fmt.Errorf("invalid content_type: %s", ct)
	}
}

func (r *historyRepo) Upsert(ctx context.Context, h *biz.History) (*biz.History, error) {
	ct, err := contentTypeToEnum(h.ContentType)
	if err != nil {
		return nil, err
	}

	// Try to find existing record by content_id (UUID)
	existing, err := r.data.db.History.Query().
		Where(
			history.HasUserWith(user.IDEQ(h.UserID)),
			history.ContentIDEQ(h.ContentID),
			history.ContentTypeEQ(ct),
		).
		Only(ctx)
	if err != nil && !entity.IsNotFound(err) {
		return nil, err
	}

	// If not found and we have a short_token, try finding by short_token
	// (compatibility: old records may have stored short_token as content_id)
	if entity.IsNotFound(err) && h.ShortToken != "" && h.ShortToken != h.ContentID {
		existingByToken, err2 := r.data.db.History.Query().
			Where(
				history.HasUserWith(user.IDEQ(h.UserID)),
				history.ContentIDEQ(h.ShortToken),
				history.ContentTypeEQ(ct),
			).
			Only(ctx)
		if err2 == nil {
			// Found old record with short_token as content_id - migrate it
			existing = existingByToken
			err = nil
			// Update content_id to UUID and fill denormalized fields
			update := r.data.db.History.UpdateOneID(existing.ID).
				SetContentID(h.ContentID).
				SetLastWatchedAt(time.Now())
			if h.Title != "" {
				update = update.SetTitle(h.Title)
			}
			if h.Thumbnail != "" {
				update = update.SetThumbnail(h.Thumbnail)
			}
			if h.ShortToken != "" {
				update = update.SetShortToken(h.ShortToken)
			}
			if h.DurationSeconds > 0 {
				update = update.SetDurationSeconds(h.DurationSeconds)
			}
			if h.ProgressSeconds > 0 {
				newProgress := h.ProgressSeconds
				if existing.ProgressSeconds > newProgress {
					newProgress = existing.ProgressSeconds
				}
				update = update.SetProgressSeconds(newProgress)
			}
			migrated, migrateErr := update.Save(ctx)
			if migrateErr != nil {
				return nil, migrateErr
			}
			return toBizHistory(migrated), nil
		}
	}

	if entity.IsNotFound(err) {
		created, err := r.data.db.History.Create().
			SetUserID(h.UserID).
			SetContentID(h.ContentID).
			SetContentType(ct).
			SetProgressSeconds(h.ProgressSeconds).
			SetDurationSeconds(h.DurationSeconds).
			SetIsFinished(h.IsFinished).
			SetTitle(h.Title).
			SetThumbnail(h.Thumbnail).
			SetShortToken(h.ShortToken).
			SetLastWatchedAt(time.Now()).
			Save(ctx)
		if err != nil {
			return nil, err
		}
		return toBizHistory(created), nil
	}

	newProgress := h.ProgressSeconds
	if existing.ProgressSeconds > newProgress {
		newProgress = existing.ProgressSeconds
	}

	isFinished := existing.IsFinished
	if h.DurationSeconds > 0 {
		threshold := 0.9
		if h.ContentType == "article" {
			threshold = 0.95
		}
		isFinished = float64(newProgress) >= float64(h.DurationSeconds)*threshold
	}

	update := r.data.db.History.UpdateOneID(existing.ID).
		SetProgressSeconds(newProgress).
		SetDurationSeconds(h.DurationSeconds).
		SetIsFinished(isFinished).
		SetLastWatchedAt(time.Now())

	if h.Title != "" {
		update = update.SetTitle(h.Title)
	}
	if h.Thumbnail != "" {
		update = update.SetThumbnail(h.Thumbnail)
	}
	if h.ShortToken != "" {
		update = update.SetShortToken(h.ShortToken)
	}

	updated, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}
	return toBizHistory(updated), nil
}

func (r *historyRepo) List(ctx context.Context, userID string, contentType string, page, pageSize int) ([]*biz.History, int, error) {
	query := r.data.db.History.Query().
		Where(history.HasUserWith(user.IDEQ(userID)))

	if contentType != "" {
		ct, err := contentTypeToEnum(contentType)
		if err != nil {
			return nil, 0, err
		}
		query = query.Where(history.ContentTypeEQ(ct))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	ents, err := query.
		Order(entity.Desc(history.FieldLastWatchedAt)).
		Offset(offset).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*biz.History, len(ents))
	var needBackfill []string
	for i, e := range ents {
		result[i] = toBizHistory(e)
		if e.Title == "" && e.ContentType == history.ContentTypeVideo {
			needBackfill = append(needBackfill, e.ContentID)
		}
	}

	// Backfill title/thumbnail/short_token for legacy records that lack them
	if len(needBackfill) > 0 {
		medias, err := r.data.db.Media.Query().
			Where(media.IDIn(needBackfill...)).
			Select(media.FieldID, media.FieldShortToken, media.FieldTitle, media.FieldThumbnail).
			All(ctx)
		if err == nil {
			mediaMap := make(map[string]*entity.Media, len(medias))
			for _, m := range medias {
				mediaMap[m.ID] = m
			}
			for _, h := range result {
				if h.Title == "" && h.ContentType == "video" {
					if m, ok := mediaMap[h.ContentID]; ok {
						h.Title = m.Title
						h.Thumbnail = m.Thumbnail
						h.ShortToken = m.ShortToken
					}
				}
			}
		}
	}

	return result, total, nil
}

func (r *historyRepo) GetByUserContent(ctx context.Context, userID, contentID, contentType string) (*biz.History, error) {
	ct, err := contentTypeToEnum(contentType)
	if err != nil {
		return nil, err
	}

	e, err := r.data.db.History.Query().
		Where(
			history.HasUserWith(user.IDEQ(userID)),
			history.ContentIDEQ(contentID),
			history.ContentTypeEQ(ct),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return toBizHistory(e), nil
}

func (r *historyRepo) Delete(ctx context.Context, id string) error {
	return r.data.db.History.DeleteOneID(id).Exec(ctx)
}

func (r *historyRepo) DeleteAll(ctx context.Context, userID string) (int, error) {
	n, err := r.data.db.History.Delete().
		Where(history.HasUserWith(user.IDEQ(userID))).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (r *historyRepo) Sync(ctx context.Context, userID string, items []*biz.History) ([]*biz.History, int, error) {
	mergedCount := 0
	for _, item := range items {
		item.UserID = userID
		_, err := r.Upsert(ctx, item)
		if err != nil {
			r.log.Warnf("sync upsert failed for content_id=%s: %v", item.ContentID, err)
			continue
		}
		mergedCount++
	}

	allHistory, _, err := r.List(ctx, userID, "", 1, 500)
	if err != nil {
		return nil, mergedCount, err
	}
	return allHistory, mergedCount, nil
}

func (r *historyRepo) CountByUser(ctx context.Context, userID string) (int, error) {
	return r.data.db.History.Query().
		Where(history.HasUserWith(user.IDEQ(userID))).
		Count(ctx)
}

func (r *historyRepo) DeleteOldest(ctx context.Context, userID string, n int) error {
	oldest, err := r.data.db.History.Query().
		Where(history.HasUserWith(user.IDEQ(userID))).
		Order(entity.Asc(history.FieldLastWatchedAt)).
		Limit(n).
		All(ctx)
	if err != nil {
		return err
	}

	ids := make([]string, len(oldest))
	for i, o := range oldest {
		ids[i] = o.ID
	}

	if len(ids) > 0 {
		predicates := make([]predicate.History, len(ids))
		for i, id := range ids {
			predicates[i] = history.IDEQ(id)
		}
		_, err = r.data.db.History.Delete().
			Where(history.Or(predicates...)).
			Exec(ctx)
	}
	return err
}
