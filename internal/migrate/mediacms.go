package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/sqlite3ent/sqlite3"
	pq "github.com/lib/pq"

	"origadmin/application/origstudio/internal/conf"
)

// MediaCMSAdapter reads data from a MediaCMS Django database.
//
// Physical table names follow Django's {app_label}_{model} convention:
//   - files app  -> files_media, files_category, files_tag, files_comment,
//     files_playlist, files_playlistmedia, files_subtitle, files_language,
//     files_license, files_rating, files_ratingcategory, files_encoding,
//     files_encodeprofile, files_mediapermission
//   - users app  -> users_user, users_channel, users_channel_subscribers,
//     users_notification
//   - M2M pivots -> files_media_category, files_media_tags,
//     users_channel_subscribers
//
// Column timestamps are read via CAST(... AS TEXT) so the same query works on
// both PostgreSQL (timestamptz) and SQLite (datetime text).
type MediaCMSAdapter struct {
	db         *sql.DB
	dialect    string
	mediaDir   string
	paths      *conf.StoragePaths
	mediaTypes []string
}

func NewMediaCMSAdapter(paths *conf.StoragePaths) *MediaCMSAdapter {
	return &MediaCMSAdapter{paths: paths}
}

func init() {
	RegisterSource("mediacms", func(paths *conf.StoragePaths) SourceAdapter {
		return NewMediaCMSAdapter(paths)
	})
}

func (a *MediaCMSAdapter) Name() string {
	return "mediacms"
}

func (a *MediaCMSAdapter) Connect(ctx context.Context, cfg *SourceConfig) error {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return fmt.Errorf("connect mediacms db: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		db, err = sql.Open("sqlite3", cfg.DSN)
		if err != nil {
			return fmt.Errorf("connect mediacms db: %w", err)
		}
		if err := db.PingContext(ctx); err != nil {
			return fmt.Errorf("ping mediacms db: %w", err)
		}
		a.dialect = "sqlite3"
	} else {
		a.dialect = "postgres"
	}
	a.db = db
	a.mediaDir = cfg.MediaDir
	a.mediaTypes = cfg.MediaTypes
	return nil
}

// mediaTypeFilter returns a SQL fragment (no leading "WHERE") and its args
// that restrict media rows to the configured whitelist. An empty whitelist
// returns no filter (migrate all types). Column is left to the caller so the
// same filter works for aliased and non-aliased queries.
func (a *MediaCMSAdapter) mediaTypeFilter(col string) (string, []any) {
	if len(a.mediaTypes) == 0 {
		return "", nil
	}
	if a.dialect == "postgres" {
		return col + " = ANY($1)", []any{pq.Array(a.mediaTypes)}
	}
	placeholders := make([]string, len(a.mediaTypes))
	args := make([]any, len(a.mediaTypes))
	for i, t := range a.mediaTypes {
		placeholders[i] = "?"
		args[i] = t
	}
	return col + " IN (" + strings.Join(placeholders, ",") + ")", args
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
		"SELECT COUNT(*) FROM users_user":    &stats.Users,
		"SELECT COUNT(*) FROM files_category": &stats.Categories,
		"SELECT COUNT(*) FROM files_tag":      &stats.Tags,
		"SELECT COUNT(*) FROM files_comment":  &stats.Comments,
		"SELECT COUNT(*) FROM users_channel":  &stats.Channels,
		"SELECT COUNT(*) FROM files_playlist": &stats.Playlists,
	}

	for q, ptr := range queries {
		if err := a.db.QueryRowContext(ctx, q).Scan(ptr); err != nil {
			return nil, fmt.Errorf("discover query %q: %w", q, err)
		}
	}

	// Media count honors the type whitelist so the reported total matches
	// what will actually be migrated.
	mediaQ := "SELECT COUNT(*) FROM files_media"
	if clause, args := a.mediaTypeFilter("media_type"); clause != "" {
		mediaQ += " WHERE " + clause
		if err := a.db.QueryRowContext(ctx, mediaQ, args...).Scan(&stats.Media); err != nil {
			return nil, fmt.Errorf("discover media count: %w", err)
		}
	} else {
		if err := a.db.QueryRowContext(ctx, mediaQ).Scan(&stats.Media); err != nil {
			return nil, fmt.Errorf("discover media count: %w", err)
		}
	}
	return stats, nil
}

func (a *MediaCMSAdapter) Users(ctx context.Context) (Iterator[*UserRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT id, username, email, password, name, title, description, location,
		       logo, is_active, is_superuser, is_featured, "advancedUser",
		       is_editor, is_manager, media_count, notification_on_comments,
		       allow_contact, COALESCE(CAST(date_added AS TEXT), ''),
		       COALESCE(CAST(last_login AS TEXT), '')
		FROM users_user ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*UserRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*UserRecord, error) {
			var (
				r                                UserRecord
				name, title, bio, loc, avatar    sql.NullString
				dateAdded, lastLogin             sql.NullString
			)
			err := rows.Scan(&r.ID, &r.Username, &r.Email, &r.Password,
				&name, &title, &bio, &loc, &avatar,
				&r.IsActive, &r.IsSuperuser, &r.IsFeatured, &r.AdvancedUser,
				&r.IsEditor, &r.IsManager, &r.MediaCount, &r.Notifications,
				&r.AllowContact, &dateAdded, &lastLogin)
			if err != nil {
				return nil, err
			}
			if name.Valid {
				r.Name = name.String
			}
			if title.Valid {
				r.Title = title.String
			}
			if bio.Valid {
				r.Bio = bio.String
			}
			if loc.Valid {
				r.Location = loc.String
			}
			if avatar.Valid {
				r.Avatar = avatar.String
			}
			if dateAdded.Valid {
				r.DateAdded = dateAdded.String
			}
			if lastLogin.Valid {
				r.LastLogin = lastLogin.String
			}
			// DisplayName falls back to username when name is empty.
			r.DisplayName = r.Name
			if r.DisplayName == "" {
				r.DisplayName = r.Username
			}
			r.CreatedAt = dateAdded.String
			r.Role = "user"
			if r.IsSuperuser {
				r.Role = "admin"
			}
			return &r, nil
		},
	}, nil
}

func (a *MediaCMSAdapter) Categories(ctx context.Context) (Iterator[*CategoryRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT id, uid, title, description, is_global, media_count, thumbnail,
		       listings_thumbnail, COALESCE(CAST(user_id AS TEXT), ''),
		       COALESCE(CAST(identity_provider_id AS TEXT), ''),
		       is_rbac_category, COALESCE(CAST(add_date AS TEXT), '')
		FROM files_category ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*CategoryRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*CategoryRecord, error) {
			var (
				r                                     CategoryRecord
				desc, thumb, listingsThumb, uid        sql.NullString
				userID, identityProvider, addDate      sql.NullString
			)
			err := rows.Scan(&r.ID, &uid, &r.Name, &desc, &r.IsGlobal,
				&r.MediaCount, &thumb, &listingsThumb, &userID,
				&identityProvider, &r.IsRBACCategory, &addDate)
			if err != nil {
				return nil, err
			}
			if desc.Valid {
				r.Description = desc.String
			}
			if thumb.Valid {
				r.Thumbnail = thumb.String
			}
			if listingsThumb.Valid {
				r.ListingsThumbnail = listingsThumb.String
			}
			if uid.Valid {
				r.UID = uid.String
			}
			if userID.Valid {
				r.UserID = userID.String
			}
			if identityProvider.Valid {
				r.IdentityProvider = identityProvider.String
			}
			if addDate.Valid {
				r.AddDate = addDate.String
			}
			r.Slug = slugify(r.Name, "category-"+r.ID)
			return &r, nil
		},
	}, nil
}

func (a *MediaCMSAdapter) Tags(ctx context.Context) (Iterator[*TagRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT id, title, media_count, listings_thumbnail,
		       COALESCE(CAST(user_id AS TEXT), '')
		FROM files_tag ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*TagRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*TagRecord, error) {
			var (
				r              TagRecord
				listingsThumb  sql.NullString
				userID         sql.NullString
			)
			err := rows.Scan(&r.ID, &r.Name, &r.MediaCount, &listingsThumb, &userID)
			if err != nil {
				return nil, err
			}
			if listingsThumb.Valid {
				r.ListingsThumbnail = listingsThumb.String
			}
			if userID.Valid {
				r.UserID = userID.String
			}
			r.Slug = slugify(r.Name, "tag-"+r.ID)
			return &r, nil
		},
	}, nil
}

func (a *MediaCMSAdapter) Media(ctx context.Context) (Iterator[*MediaRecord], error) {
	query := `
		SELECT m.id, m.uid, m.user_id, m.title, m.description, m.media_type,
		       m.media_file, m.size, m.duration, m.encoding_status, m.thumbnail,
		       m.state, m.views, m.likes, m.dislikes, m.hls_file, m.sprites,
		       m.poster, m.preview_file_path, m.uploaded_thumbnail,
		       m.uploaded_poster, m.media_info, m.md5sum, m.friendly_token,
		       m.thumbnail_time, m.allow_download, m.enable_comments, m.featured,
		       m.listable, m.is_reviewed, m.reported_times, m.video_height,
		       COALESCE(CAST(m.channel_id AS TEXT), ''),
		       COALESCE(CAST(m.license_id AS TEXT), ''),
		       COALESCE(CAST(m.add_date AS TEXT), ''),
		       COALESCE(CAST(m.edit_date AS TEXT), ''),
		       COALESCE(u.username, '')
		FROM files_media m
		LEFT JOIN users_user u ON m.user_id = u.id`
	var args []any
	if clause, filterArgs := a.mediaTypeFilter("m.media_type"); clause != "" {
		query += " WHERE " + clause
		args = append(args, filterArgs...)
	}
	query += " ORDER BY m.id"

	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query media: %w", err)
	}
	return &sqlIterator[*MediaRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*MediaRecord, error) {
			var (
				r                                     MediaRecord
				desc, size, thumb, state              sql.NullString
				hls, sprites, poster, preview         sql.NullString
				upThumb, upPoster, mediaInfo, md5     sql.NullString
				friendlyToken, channelID, licenseID   sql.NullString
				addDate, editDate, sourceUser         sql.NullString
				uid                                   sql.NullString
				encStatus                             string
				duration                              int
				thumbnailTime                         sql.NullFloat64
			)
			err := rows.Scan(&r.ID, &uid, &r.UserID, &r.Title, &desc, &r.Type,
				&r.FilePath, &size, &duration, &encStatus, &thumb, &state,
				&r.Views, &r.Likes, &r.Dislikes, &hls, &sprites, &poster,
				&preview, &upThumb, &upPoster, &mediaInfo, &md5, &friendlyToken,
				&thumbnailTime, &r.AllowDownload, &r.EnableComments, &r.Featured,
				&r.Listable, &r.IsReviewed, &r.ReportedTimes, &r.Height,
				&channelID, &licenseID, &addDate, &editDate, &sourceUser)
			if err != nil {
				return nil, err
			}
			r.EncodingStatus = mapEncodingStatus(encStatus)
			if uid.Valid {
				r.UID = uid.String
			}
			if desc.Valid {
				r.Description = desc.String
			}
			if size.Valid {
				r.FileSize = parseFileSize(size.String)
			}
			if thumb.Valid {
				r.Thumbnail = thumb.String
			}
			if state.Valid {
				r.State = state.String
			}
			r.Privacy = mapPrivacy(r.State)
			if hls.Valid {
				r.HLSFile = hls.String
			}
			if sprites.Valid {
				r.Sprites = sprites.String
			}
			if poster.Valid {
				r.Poster = poster.String
			}
			if preview.Valid {
				r.PreviewFile = preview.String
			}
			if upThumb.Valid {
				r.UploadedThumb = upThumb.String
			}
			if upPoster.Valid {
				r.UploadedPoster = upPoster.String
			}
			if mediaInfo.Valid {
				r.MediaInfo = mediaInfo.String
			}
			if md5.Valid {
				r.Md5sum = md5.String
			}
			if friendlyToken.Valid {
				r.FriendlyToken = friendlyToken.String
			}
			if channelID.Valid {
				r.ChannelID = channelID.String
			}
			if licenseID.Valid {
				r.LicenseID = licenseID.String
			}
			if addDate.Valid {
				r.CreatedAt = addDate.String
			}
			if editDate.Valid {
				r.EditDate = editDate.String
			}
			if thumbnailTime.Valid {
				r.ThumbnailTime = thumbnailTime.Float64
			}
			if sourceUser.Valid {
				r.Metadata = map[string]string{"source_user": sourceUser.String}
			}
			r.Duration = float64(duration)
			r.FileName = filepath.Base(r.FilePath)
			if r.MimeType == "" || r.MimeType == "application/octet-stream" {
				r.MimeType = mimeByExt(r.FileName)
			}
			r.Tags = []string{}
			if r.Metadata == nil {
				r.Metadata = make(map[string]string)
			}
			return &r, nil
		},
	}, nil
}

func (a *MediaCMSAdapter) Comments(ctx context.Context) (Iterator[*CommentRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT id, uid, text, media_id, user_id,
		       COALESCE(CAST(parent_id AS TEXT), ''), level, tree_id,
		       COALESCE(CAST(add_date AS TEXT), '')
		FROM files_comment ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*CommentRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*CommentRecord, error) {
			var (
				r                  CommentRecord
				uid, parent, date  sql.NullString
			)
			err := rows.Scan(&r.ID, &uid, &r.Text, &r.MediaID, &r.UserID,
				&parent, &r.Level, &r.TreeID, &date)
			if err != nil {
				return nil, err
			}
			if uid.Valid {
				r.UID = uid.String
			}
			if parent.Valid {
				r.ParentID = parent.String
			}
			if date.Valid {
				r.CreatedAt = date.String
			}
			return &r, nil
		},
	}, nil
}

func (a *MediaCMSAdapter) Channels(ctx context.Context) (Iterator[*ChannelRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT id, user_id, title, description, friendly_token, banner_logo,
		       COALESCE(CAST(add_date AS TEXT), '')
		FROM users_channel ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*ChannelRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*ChannelRecord, error) {
			var (
				r                           ChannelRecord
				desc, token, banner, date   sql.NullString
			)
			err := rows.Scan(&r.ID, &r.UserID, &r.Title, &desc, &token, &banner, &date)
			if err != nil {
				return nil, err
			}
			if desc.Valid {
				r.Description = desc.String
			}
			if token.Valid {
				r.FriendlyToken = token.String
			}
			if banner.Valid {
				r.BannerLogo = banner.String
				r.Banner = banner.String
			}
			if date.Valid {
				r.AddDate = date.String
				r.CreatedAt = date.String
			}
			// users_channel has no separate name column; title is the display name.
			r.Name = r.Title
			r.Slug = slugify(r.Title, "channel-"+r.ID)
			return &r, nil
		},
	}, nil
}

func (a *MediaCMSAdapter) Playlists(ctx context.Context) (Iterator[*PlaylistRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT id, uid, user_id, title, description, friendly_token,
		       COALESCE(CAST(add_date AS TEXT), '')
		FROM files_playlist ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*PlaylistRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*PlaylistRecord, error) {
			var (
				r                         PlaylistRecord
				uid, desc, token, date    sql.NullString
			)
			err := rows.Scan(&r.ID, &uid, &r.UserID, &r.Name, &desc, &token, &date)
			if err != nil {
				return nil, err
			}
			if uid.Valid {
				r.UID = uid.String
			}
			if desc.Valid {
				r.Description = desc.String
			}
			if token.Valid {
				r.FriendlyToken = token.String
			}
			if date.Valid {
				r.AddDate = date.String
				r.CreatedAt = date.String
			}
			r.Privacy = "public"
			r.MediaIDs = []string{}
			return &r, nil
		},
	}, nil
}

func (a *MediaCMSAdapter) MediaTags(ctx context.Context) (Iterator[*MediaTagRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT media_id, tag_id FROM files_media_tags ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*MediaTagRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*MediaTagRecord, error) {
			var r MediaTagRecord
			err := rows.Scan(&r.MediaID, &r.TagID)
			return &r, err
		},
	}, nil
}

func (a *MediaCMSAdapter) MediaCategories(ctx context.Context) (Iterator[*MediaCategoryRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT media_id, category_id FROM files_media_category ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*MediaCategoryRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*MediaCategoryRecord, error) {
			var r MediaCategoryRecord
			err := rows.Scan(&r.MediaID, &r.CategoryID)
			return &r, err
		},
	}, nil
}

func (a *MediaCMSAdapter) PlaylistMedia(ctx context.Context) (Iterator[*PlaylistMediaRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT playlist_id, media_id, ordering,
		       COALESCE(CAST(action_date AS TEXT), '')
		FROM files_playlistmedia ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*PlaylistMediaRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*PlaylistMediaRecord, error) {
			var (
				r               PlaylistMediaRecord
				actionDate      sql.NullString
			)
			err := rows.Scan(&r.PlaylistID, &r.MediaID, &r.Ordering, &actionDate)
			if err != nil {
				return nil, err
			}
			if actionDate.Valid {
				r.ActionDate = actionDate.String
			}
			return &r, nil
		},
	}, nil
}

func (a *MediaCMSAdapter) Subtitles(ctx context.Context) (Iterator[*SubtitleRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT s.media_id, COALESCE(l.code, ''), COALESCE(l.title, ''),
		       s.subtitle_file, COALESCE(CAST(s.user_id AS TEXT), '')
		FROM files_subtitle s
		LEFT JOIN files_language l ON s.language_id = l.id
		ORDER BY s.id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*SubtitleRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*SubtitleRecord, error) {
			var r SubtitleRecord
			err := rows.Scan(&r.MediaID, &r.Language, &r.Label, &r.FileURL, &r.UserID)
			return &r, err
		},
	}, nil
}

func (a *MediaCMSAdapter) Subscriptions(ctx context.Context) (Iterator[*SubscriptionRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT channel_id, user_id FROM users_channel_subscribers ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*SubscriptionRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*SubscriptionRecord, error) {
			var r SubscriptionRecord
			err := rows.Scan(&r.ChannelID, &r.SubscriberID)
			return &r, err
		},
	}, nil
}

// Licenses loads the full files_license table keyed by id. Used by phaseMedia
// to fold license data into Media.metadata (no B-side entity).
func (a *MediaCMSAdapter) Licenses(ctx context.Context) (map[string]*LicenseRecord, error) {
	rows, err := a.db.QueryContext(ctx, `SELECT id, title, description FROM files_license ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	licenses := make(map[string]*LicenseRecord)
	for rows.Next() {
		var r LicenseRecord
		var desc sql.NullString
		if err := rows.Scan(&r.ID, &r.Title, &desc); err != nil {
			return nil, err
		}
		if desc.Valid {
			r.Description = desc.String
		}
		licenses[r.ID] = &r
	}
	return licenses, rows.Err()
}

// RatingsByMedia returns all files_rating rows for one media id.
func (a *MediaCMSAdapter) RatingsByMedia(ctx context.Context, mediaID string) ([]*RatingRecord, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT id, media_id, user_id, rating_category_id, score,
		       COALESCE(CAST(add_date AS TEXT), '')
		FROM files_rating WHERE media_id = $1 ORDER BY id`, mediaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []*RatingRecord
	for rows.Next() {
		var r RatingRecord
		var addDate sql.NullString
		if err := rows.Scan(&r.ID, &r.MediaID, &r.UserID, &r.RatingCategoryID, &r.Score, &addDate); err != nil {
			return nil, err
		}
		if addDate.Valid {
			r.AddDate = addDate.String
		}
		ratings = append(ratings, &r)
	}
	return ratings, rows.Err()
}

// EncodingsByMedia returns all files_encoding rows for one media id.
func (a *MediaCMSAdapter) EncodingsByMedia(ctx context.Context, mediaID string) ([]*EncodingRecord, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT id, media_id, profile_id, status, progress, media_file, size,
		       COALESCE(md5sum, ''), chunk, chunk_file_path, chunks_info, logs,
		       commands, temp_file, task_id, worker, total_run_time, retries,
		       COALESCE(CAST(add_date AS TEXT), ''),
		       COALESCE(CAST(update_date AS TEXT), '')
		FROM files_encoding WHERE media_id = $1 ORDER BY id`, mediaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var encodings []*EncodingRecord
	for rows.Next() {
		var r EncodingRecord
		var addDate, updateDate sql.NullString
		if err := rows.Scan(&r.ID, &r.MediaID, &r.ProfileID, &r.Status, &r.Progress,
			&r.MediaFile, &r.Size, &r.MD5Sum, &r.Chunk, &r.ChunkFilePath,
			&r.ChunksInfo, &r.Logs, &r.Commands, &r.TempFile, &r.TaskID,
			&r.Worker, &r.TotalRunTime, &r.Retries, &addDate, &updateDate); err != nil {
			return nil, err
		}
		if addDate.Valid {
			r.AddDate = addDate.String
		}
		if updateDate.Valid {
			r.UpdateDate = updateDate.String
		}
		encodings = append(encodings, &r)
	}
	return encodings, rows.Err()
}

// MediaPermissionsByMedia returns all files_mediapermission rows for one media id.
func (a *MediaCMSAdapter) MediaPermissionsByMedia(ctx context.Context, mediaID string) ([]*MediaPermissionRecord, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT id, media_id, user_id, owner_user_id, permission,
		       COALESCE(CAST(created_at AS TEXT), '')
		FROM files_mediapermission WHERE media_id = $1 ORDER BY id`, mediaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []*MediaPermissionRecord
	for rows.Next() {
		var r MediaPermissionRecord
		var createdAt sql.NullString
		if err := rows.Scan(&r.ID, &r.MediaID, &r.UserID, &r.OwnerUserID, &r.Permission, &createdAt); err != nil {
			return nil, err
		}
		if createdAt.Valid {
			r.CreatedAt = createdAt.String
		}
		perms = append(perms, &r)
	}
	return perms, rows.Err()
}

func (a *MediaCMSAdapter) FileRefs(ctx context.Context, media *MediaRecord) ([]*FileRef, error) {
	if a.mediaDir == "" {
		return nil, nil
	}

	var refs []*FileRef

	// Primary media file: DB stores paths relative to MEDIA_ROOT
	// (e.g. original/user/{username}/{uid}.{ext}).
	if media.FilePath != "" {
		sourcePath := filepath.Join(a.mediaDir, media.FilePath)
		if _, err := os.Stat(sourcePath); err == nil {
			refs = append(refs, &FileRef{
				SourcePath: sourcePath,
				TargetPath: a.paths.Relative("originals", media.UserID, media.FileName),
				Size:       media.FileSize,
				Kind:       "original",
			})
		}
	}

	// Thumbnail: prefer uploaded thumbnail, else generated thumbnail.
	thumb := media.UploadedThumb
	if thumb == "" {
		thumb = media.Thumbnail
	}
	if thumb != "" {
		thumbSource := filepath.Join(a.mediaDir, thumb)
		if _, err := os.Stat(thumbSource); err == nil {
			ext := filepath.Ext(thumb)
			if ext == "" {
				ext = ".jpg"
			}
			refs = append(refs, &FileRef{
				SourcePath: thumbSource,
				TargetPath: a.paths.Relative("thumbnails", media.ID+ext),
				Size:       0,
				Kind:       "thumbnail",
			})
		}
	}

	// Poster: distinct from the thumbnail, keep a separate copy.
	if media.Poster != "" && media.Poster != thumb {
		posterSource := filepath.Join(a.mediaDir, media.Poster)
		if _, err := os.Stat(posterSource); err == nil {
			ext := filepath.Ext(media.Poster)
			if ext == "" {
				ext = ".jpg"
			}
			refs = append(refs, &FileRef{
				SourcePath: posterSource,
				TargetPath: a.paths.Relative("thumbnails", media.ID+"_poster"+ext),
				Size:       0,
				Kind:       "poster",
			})
		}
	}

	// Sprite sheet: vtt plus the derived jpg.
	if media.Sprites != "" {
		spriteVTT := filepath.Join(a.mediaDir, media.Sprites)
		if _, err := os.Stat(spriteVTT); err == nil {
			refs = append(refs, &FileRef{
				SourcePath: spriteVTT,
				TargetPath: a.paths.Relative("sprites", media.ID, "sprite.vtt"),
				Kind:       "sprite_vtt",
			})
		}
		spriteJPG := filepath.Join(a.mediaDir, spriteImagePath(media.Sprites))
		if _, err := os.Stat(spriteJPG); err == nil {
			refs = append(refs, &FileRef{
				SourcePath: spriteJPG,
				TargetPath: a.paths.Relative("sprites", media.ID, "sprite.jpg"),
				Kind:       "sprite_jpg",
			})
		}
	}

	// Subtitle files (files_subtitle -> subtitle_file relative to MEDIA_ROOT).
	subRefs, err := a.subtitleRefs(ctx, media)
	if err != nil {
		return nil, err
	}
	refs = append(refs, subRefs...)

	return refs, nil
}

// subtitleRefs resolves the on-disk subtitle tracks for one media item into
// file refs, targeting originals/subtitles/{srcMediaID}/{lang}{ext} so the
// DB path written by phaseSubtitles stays aligned with where phaseFiles lands.
func (a *MediaCMSAdapter) subtitleRefs(ctx context.Context, media *MediaRecord) ([]*FileRef, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT s.subtitle_file, COALESCE(l.code, 'und')
		FROM files_subtitle s
		LEFT JOIN files_language l ON s.language_id = l.id
		WHERE s.media_id = $1`, media.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refs []*FileRef
	for rows.Next() {
		var subFile, lang string
		if err := rows.Scan(&subFile, &lang); err != nil {
			return nil, err
		}
		src := filepath.Join(a.mediaDir, subFile)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		ext := filepath.Ext(subFile)
		if ext == "" {
			ext = ".vtt"
		}
		refs = append(refs, &FileRef{
			SourcePath: src,
			TargetPath: a.paths.Relative("originals", "subtitles", media.ID, lang+ext),
			Kind:       "subtitle",
		})
	}
	return refs, rows.Err()
}

func (a *MediaCMSAdapter) OpenFile(ctx context.Context, ref *FileRef) (io.ReadCloser, error) {
	return os.Open(ref.SourcePath)
}

// slugify converts a title into a URL-safe slug. Falls back to fallback when
// the result would be empty.
func slugify(s, fallback string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	out := strings.TrimSuffix(b.String(), "-")
	if out == "" {
		return fallback
	}
	return out
}

// parseFileSize parses a Django CharField size into bytes. MediaCMS stores
// sizes both as plain byte counts ("123456") and as human-readable strings
// ("0.1MB", "1.2 GB", "512KB"). Empty or unparseable values yield 0.
func parseFileSize(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return n
	}
	// Human-readable forms: "0.1MB", "1.2 GB", "512KB", "3G".
	fields := strings.Fields(s)
	numStr := fields[0]
	unit := ""
	if len(fields) > 1 {
		unit = strings.ToUpper(fields[1])
	} else {
		// Unit attached to the number, e.g. "0.1MB" -> split trailing letters.
		i := len(numStr)
		for i > 0 {
			c := numStr[i-1]
			if (c >= '0' && c <= '9') || c == '.' {
				break
			}
			i--
		}
		unit = strings.ToUpper(numStr[i:])
		numStr = numStr[:i]
	}
	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}
	var mult float64
	switch strings.TrimSuffix(unit, "B") {
	case "K":
		mult = 1 << 10
	case "M":
		mult = 1 << 20
	case "G":
		mult = 1 << 30
	case "T":
		mult = 1 << 40
	case "":
		mult = 1
	default:
		return 0
	}
	return int64(val * mult)
}

// mapPrivacy translates MediaCMS state (private/public/unlisted) into B-side
// privacy enum values.
func mapPrivacy(state string) string {
	switch strings.ToLower(state) {
	case "private":
		return "PRIVATE"
	case "unlisted":
		return "UNLISTED"
	default:
		return "PUBLIC"
	}
}

// mapEncodingStatus translates MediaCMS encoding status into B-side values.
// The mapping is MediaCMS-specific, so it lives on the adapter instead of the
// engine: the engine consumes EncodingStatus already normalized to the target
// enum (success / processing / failed / pending).
func mapEncodingStatus(s string) string {
	switch strings.ToLower(s) {
	case "success":
		return "success"
	case "running":
		return "processing"
	case "fail", "failed":
		return "failed"
	default:
		return "pending"
	}
}

// mimeByExt guesses a MIME type from a file name extension.
func mimeByExt(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mov":
		return "video/quicktime"
	case ".mkv":
		return "video/x-matroska"
	case ".m3u8":
		return "application/x-mpegURL"
	case ".ts":
		return "video/mp2t"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".ogg", ".oga":
		return "audio/ogg"
	case ".pdf":
		return "application/pdf"
	case ".vtt":
		return "text/vtt"
	case ".srt":
		return "application/x-subrip"
	default:
		return "application/octet-stream"
	}
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

func (it *sqlIterator[T]) Item() T      { return it.item }
func (it *sqlIterator[T]) Err() error   { return it.err }
func (it *sqlIterator[T]) Close() error { return it.rows.Close() }
