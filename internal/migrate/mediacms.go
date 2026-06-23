package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"

	_ "github.com/sqlite3ent/sqlite3"
	_ "github.com/lib/pq"
)

type MediaCMSAdapter struct {
	db       *sql.DB
	mediaDir string
}

func NewMediaCMSAdapter() *MediaCMSAdapter {
	return &MediaCMSAdapter{}
}

func (a *MediaCMSAdapter) Name() string {
	return "mediacms"
}

func (a *MediaCMSAdapter) Connect(ctx context.Context, cfg *SourceConfig) error {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		db, err = sql.Open("sqlite3", cfg.DSN)
		if err != nil {
			return fmt.Errorf("connect mediacms db: %w", err)
		}
	}
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping mediacms db: %w", err)
	}
	a.db = db
	a.mediaDir = cfg.MediaDir
	return nil
}

func (a *MediaCMSAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

func (a *MediaCMSAdapter) Discover(ctx context.Context) (*SourceStats, error) {
	stats := &SourceStats{}
	queries := map[string]*int{
		"SELECT COUNT(*) FROM auth_user":         &stats.Users,
		"SELECT COUNT(*) FROM content_content":   &stats.Media,
		"SELECT COUNT(*) FROM content_category":  &stats.Categories,
		"SELECT COUNT(*) FROM content_tag":       &stats.Tags,
		"SELECT COUNT(*) FROM content_comment":   &stats.Comments,
		"SELECT COUNT(*) FROM content_channel":   &stats.Channels,
		"SELECT COUNT(*) FROM content_playlist":  &stats.Playlists,
	}

	for q, ptr := range queries {
		if err := a.db.QueryRowContext(ctx, q).Scan(ptr); err != nil {
			return nil, fmt.Errorf("discover query %q: %w", q, err)
		}
	}
	return stats, nil
}

func (a *MediaCMSAdapter) Users(ctx context.Context) (Iterator[*UserRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT id, username, email, password, name, avatar_url, bio, is_active, date_joined
		FROM auth_user ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*UserRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*UserRecord, error) {
			var r UserRecord
			var dateJoined sql.NullString
			err := rows.Scan(&r.ID, &r.Username, &r.Email, &r.Password, &r.DisplayName,
				&r.Avatar, &r.Bio, &r.IsActive, &dateJoined)
			if dateJoined.Valid {
				r.CreatedAt = dateJoined.String
			}
			r.Role = "user"
			return &r, err
		},
	}, nil
}

func (a *MediaCMSAdapter) Categories(ctx context.Context) (Iterator[*CategoryRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT id, title, slug, description, parent_id
		FROM content_category ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*CategoryRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*CategoryRecord, error) {
			var r CategoryRecord
			var parentID sql.NullString
			err := rows.Scan(&r.ID, &r.Name, &r.Slug, &r.Description, &parentID)
			if parentID.Valid {
				r.ParentID = parentID.String
			}
			return &r, err
		},
	}, nil
}

func (a *MediaCMSAdapter) Tags(ctx context.Context) (Iterator[*TagRecord], error) {
	rows, err := a.db.QueryContext(ctx, `SELECT id, title, slug FROM content_tag ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*TagRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*TagRecord, error) {
			var r TagRecord
			err := rows.Scan(&r.ID, &r.Name, &r.Slug)
			return &r, err
		},
	}, nil
}

func (a *MediaCMSAdapter) Media(ctx context.Context) (Iterator[*MediaRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT c.id, c.user_id, c.title, c.description, c.media_type, c.media_file,
		       c.file_size, c.duration, c.encoding_status, c.thumbnail, c.created_at,
		       COALESCE(u.username, '')
		FROM content_content c
		LEFT JOIN auth_user u ON c.user_id = u.id
		ORDER BY c.id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*MediaRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*MediaRecord, error) {
			var r MediaRecord
			var duration sql.NullFloat64
			var fileSize sql.NullInt64
			var desc, thumb, createdAt, sourceUser sql.NullString
			r.Metadata = make(map[string]string)
			err := rows.Scan(&r.ID, &r.UserID, &r.Title, &desc, &r.Type, &r.FilePath,
				&fileSize, &duration, &r.MimeType, &thumb, &createdAt, &sourceUser)
			if desc.Valid {
				r.Description = desc.String
			}
			if fileSize.Valid {
				r.FileSize = fileSize.Int64
			}
			if duration.Valid {
				r.Duration = duration.Float64
			}
			if thumb.Valid {
				r.Thumbnail = thumb.String
			}
			if createdAt.Valid {
				r.CreatedAt = createdAt.String
			}
			if sourceUser.Valid {
				r.Metadata["source_user"] = sourceUser.String
			}
			r.FilePath = filepath.Base(r.FilePath)
			r.Privacy = "public"
			return &r, err
		},
	}, nil
}

func (a *MediaCMSAdapter) Comments(ctx context.Context) (Iterator[*CommentRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT id, content_id, user_id, text, created_at
		FROM content_comment ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*CommentRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*CommentRecord, error) {
			var r CommentRecord
			var createdAt sql.NullString
			err := rows.Scan(&r.ID, &r.MediaID, &r.UserID, &r.Text, &createdAt)
			if createdAt.Valid {
				r.CreatedAt = createdAt.String
			}
			return &r, err
		},
	}, nil
}

func (a *MediaCMSAdapter) Channels(ctx context.Context) (Iterator[*ChannelRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT id, user_id, name, slug, description, avatar_url, banner_url, created_at
		FROM content_channel ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*ChannelRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*ChannelRecord, error) {
			var r ChannelRecord
			var desc, avatar, banner, createdAt sql.NullString
			err := rows.Scan(&r.ID, &r.UserID, &r.Name, &r.Slug, &desc, &avatar, &banner, &createdAt)
			if desc.Valid {
				r.Description = desc.String
			}
			if avatar.Valid {
				r.Avatar = avatar.String
			}
			if banner.Valid {
				r.Banner = banner.String
			}
			if createdAt.Valid {
				r.CreatedAt = createdAt.String
			}
			return &r, err
		},
	}, nil
}

func (a *MediaCMSAdapter) Playlists(ctx context.Context) (Iterator[*PlaylistRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT id, user_id, title, description, privacy, created_at
		FROM content_playlist ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*PlaylistRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*PlaylistRecord, error) {
			var r PlaylistRecord
			var desc, createdAt sql.NullString
			err := rows.Scan(&r.ID, &r.UserID, &r.Name, &desc, &r.Privacy, &createdAt)
			if desc.Valid {
				r.Description = desc.String
			}
			if createdAt.Valid {
				r.CreatedAt = createdAt.String
			}
			return &r, err
		},
	}, nil
}

func (a *MediaCMSAdapter) FileRefs(ctx context.Context, media *MediaRecord) ([]*FileRef, error) {
	if a.mediaDir == "" {
		return nil, nil
	}

	var refs []*FileRef

	sourcePath := filepath.Join(a.mediaDir, "media", media.FilePath)
	if _, err := os.Stat(sourcePath); err == nil {
		refs = append(refs, &FileRef{
			SourcePath: sourcePath,
			TargetPath: filepath.Join("originals", media.UserID, media.FilePath),
			Size:       media.FileSize,
			Checksum:   media.Checksum,
		})
	}

	if media.Thumbnail != "" {
		thumbSource := filepath.Join(a.mediaDir, "media", "thumbnails", media.Thumbnail)
		if _, err := os.Stat(thumbSource); err == nil {
			refs = append(refs, &FileRef{
				SourcePath: thumbSource,
				TargetPath: filepath.Join("thumbnails", media.ID+filepath.Ext(media.Thumbnail)),
				Size:       0,
			})
		}
	}

	return refs, nil
}

func (a *MediaCMSAdapter) OpenFile(ctx context.Context, ref *FileRef) (io.ReadCloser, error) {
	return os.Open(ref.SourcePath)
}

type sqlIterator[T any] struct {
	rows *sql.Rows
	scan func(*sql.Rows) (T, error)
	item T
	err  error
}

func (it *sqlIterator[T]) Next(ctx context.Context) bool {
	if !it.rows.Next() {
		return false
	}
	it.item, it.err = it.scan(it.rows)
	return it.err == nil
}

func (it *sqlIterator[T]) Item() T       { return it.item }
func (it *sqlIterator[T]) Err() error    { return it.err }
func (it *sqlIterator[T]) Close() error  { return it.rows.Close() }
