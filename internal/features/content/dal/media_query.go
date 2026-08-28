package dal

import (
	"context"

	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/entity/media"
)

// MediaByToken resolves a media row by short_token (subtitle owner/id checks, BUG-186).
func (d *Data) MediaByToken(ctx context.Context, token string) (*entity.Media, error) {
	return d.db.Media.Query().Where(media.ShortTokenEQ(token)).First(ctx)
}

// MediaByID resolves a media row by UUID id (subtitle delete owner check, BUG-186).
func (d *Data) MediaByID(ctx context.Context, id string) (*entity.Media, error) {
	return d.db.Media.Get(ctx, id)
}
