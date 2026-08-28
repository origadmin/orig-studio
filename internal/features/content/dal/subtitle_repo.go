package dal

import (
	"context"

	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/entity/subtitle"
)

// SubtitleItem is the DAL-level subtitle record (mirrors the ent entity).
type SubtitleItem struct {
	ID           string `json:"id"`
	MediaID      string `json:"media_id"`
	Language     string `json:"language"`
	Label        string `json:"label"`
	FileURL      string `json:"file_url"`
	Status       string `json:"status"` // processing / active / failed
	ErrorMessage string `json:"error_message"`
}

// SubtitleRepo abstracts subtitle persistence (BUG-186).
type SubtitleRepo interface {
	ListByMediaID(ctx context.Context, mediaID string) ([]*SubtitleItem, error)
	Create(ctx context.Context, item *SubtitleItem) (*SubtitleItem, error)
	GetByID(ctx context.Context, id string) (*SubtitleItem, error)
	Delete(ctx context.Context, id string) error
}

type subtitleRepo struct {
	data *Data
}

func NewSubtitleRepo(data *Data) SubtitleRepo {
	return &subtitleRepo{data: data}
}

func (r *subtitleRepo) ListByMediaID(ctx context.Context, mediaID string) ([]*SubtitleItem, error) {
	rows, err := r.data.db.Subtitle.Query().
		Where(subtitle.MediaIDEQ(mediaID)).
		Order(subtitle.ByCreateTime()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*SubtitleItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toSubtitleItem(row))
	}
	return items, nil
}

func (r *subtitleRepo) Create(ctx context.Context, item *SubtitleItem) (*SubtitleItem, error) {
	builder := r.data.db.Subtitle.Create().
		SetMediaID(item.MediaID).
		SetLanguage(item.Language).
		SetStatus(item.Status)
	if item.Label != "" {
		builder.SetLabel(item.Label)
	}
	if item.FileURL != "" {
		builder.SetFileURL(item.FileURL)
	}
	if item.ErrorMessage != "" {
		builder.SetErrorMessage(item.ErrorMessage)
	}
	row, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return toSubtitleItem(row), nil
}

func (r *subtitleRepo) GetByID(ctx context.Context, id string) (*SubtitleItem, error) {
	row, err := r.data.db.Subtitle.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return toSubtitleItem(row), nil
}

func (r *subtitleRepo) Delete(ctx context.Context, id string) error {
	return r.data.db.Subtitle.DeleteOneID(id).Exec(ctx)
}

func toSubtitleItem(row *entity.Subtitle) *SubtitleItem {
	return &SubtitleItem{
		ID:           row.ID,
		MediaID:      row.MediaID,
		Language:     row.Language,
		Label:        row.Label,
		FileURL:      row.FileURL,
		Status:       row.Status,
		ErrorMessage: row.ErrorMessage,
	}
}
