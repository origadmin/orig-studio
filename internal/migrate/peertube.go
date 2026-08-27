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
	"time"

	"origadmin/application/origstudio/internal/conf"
)

// PeerTubeAdapter reads data from a PeerTube PostgreSQL database. It is the
// reference implementation for adding a second source platform: it
// self-registers in init() and implements the same SourceAdapter contract as
// MediaCMSAdapter, so neither the engine nor the CLI needs per-platform
// branches. Any new platform follows this same recipe (new file + RegisterSource).
//
// Schema notes (PeerTube stores camelCase identifiers, so every such column
// must be double-quoted in SQL):
//   - "user" is a reserved word -> always quoted.
//   - "videoFile"."videoStreamingPlaylistId" is NULL for web-video files
//     (resolution -1 = the original upload; resolution >= 0 = transcodes) and
//     NOT NULL for HLS segment files. Only the original file (or, if the
//     instance deleted it, the highest transcode) is migrated; HLS/transcodes
//     are left to the target's own transcoding pipeline.
//   - File storage under <mediaDir>: originals in videos/, transcodes in
//     web-videos/, thumbnails in thumbnails/, previews in previews/, captions
//     in captions/. findFile tries both video dirs so either convention works.
//   - Timestamps are BIGINT epoch milliseconds; converted to RFC3339 for the
//     engine's parseTime.
type PeerTubeAdapter struct {
	db         *sql.DB
	mediaDir   string
	paths      *conf.StoragePaths
	mediaTypes []string
}

func NewPeerTubeAdapter(paths *conf.StoragePaths) *PeerTubeAdapter {
	return &PeerTubeAdapter{paths: paths}
}

func init() {
	RegisterSource("peertube", func(paths *conf.StoragePaths) SourceAdapter {
		return NewPeerTubeAdapter(paths)
	})
}

func (a *PeerTubeAdapter) Name() string { return "peertube" }

func (a *PeerTubeAdapter) Connect(ctx context.Context, cfg *SourceConfig) error {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return fmt.Errorf("connect peertube db: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("ping peertube db: %w", err)
	}
	a.db = db
	a.mediaDir = cfg.MediaDir
	a.mediaTypes = cfg.MediaTypes
	return nil
}

func (a *PeerTubeAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// excludesVideo reports whether the configured media-type whitelist rejects
// PeerTube's video rows. PeerTube is video-centric (audio uploads are still
// stored as videos), so a whitelist that omits "video" means "migrate nothing"
// rather than a half-understood partial migration.
func (a *PeerTubeAdapter) excludesVideo() bool {
	if len(a.mediaTypes) == 0 {
		return false
	}
	for _, t := range a.mediaTypes {
		if strings.EqualFold(t, "video") {
			return false
		}
	}
	return true
}

func (a *PeerTubeAdapter) Discover(ctx context.Context) (*SourceStats, error) {
	stats := &SourceStats{}
	queries := map[string]*int{
		`SELECT COUNT(*) FROM "user"`:         &stats.Users,
		`SELECT COUNT(*) FROM tag`:            &stats.Tags,
		`SELECT COUNT(*) FROM "videoComment"`: &stats.Comments,
		`SELECT COUNT(*) FROM "videoChannel"`: &stats.Channels,
		`SELECT COUNT(*) FROM "videoPlaylist"`: &stats.Playlists,
	}
	for q, ptr := range queries {
		if err := a.db.QueryRowContext(ctx, q).Scan(ptr); err != nil {
			return nil, fmt.Errorf("discover query %q: %w", q, err)
		}
	}
	// Categories are a fixed PeerTube enum, synthesized rather than counted.
	stats.Categories = len(peerTubeCategoryNames)

	mediaQ := `SELECT COUNT(*) FROM video`
	if a.excludesVideo() {
		mediaQ += ` WHERE 1 = 0`
	}
	if err := a.db.QueryRowContext(ctx, mediaQ).Scan(&stats.Media); err != nil {
		return nil, fmt.Errorf("discover media count: %w", err)
	}
	return stats, nil
}

func (a *PeerTubeAdapter) Users(ctx context.Context) (Iterator[*UserRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT u.id, u.username, COALESCE(u.email, ''), COALESCE(u.password, ''),
		       COALESCE(acc."displayName", u.username), u.role,
		       COALESCE(u."createdAt", 0)
		FROM "user" u
		LEFT JOIN account acc ON acc."userId" = u.id
		ORDER BY u.id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*UserRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*UserRecord, error) {
			var (
				r              UserRecord
				displayName    string
				role           int
				createdAt      sql.NullInt64
			)
			if err := rows.Scan(&r.ID, &r.Username, &r.Email, &r.Password,
				&displayName, &role, &createdAt); err != nil {
				return nil, err
			}
			r.DisplayName = displayName
			r.Role = "user"
			// PeerTube UserRole: 0 USER, 1 MODERATOR, 2 ADMIN.
			if role == 2 {
				r.Role = "admin"
			}
			r.IsActive = true
			r.CreatedAt = peertubeTime(createdAt)
			return &r, nil
		},
	}, nil
}

// Categories synthesizes PeerTube's fixed category enum (no category table).
func (a *PeerTubeAdapter) Categories(ctx context.Context) (Iterator[*CategoryRecord], error) {
	recs := make([]*CategoryRecord, 0, len(peerTubeCategoryNames))
	for id, name := range peerTubeCategoryNames {
		idStr := strconv.Itoa(id)
		recs = append(recs, &CategoryRecord{
			ID:   idStr,
			Name: name,
			Slug: slugify(name, "category-"+idStr),
		})
	}
	return &sliceIterator[*CategoryRecord]{items: recs}, nil
}

func (a *PeerTubeAdapter) Tags(ctx context.Context) (Iterator[*TagRecord], error) {
	rows, err := a.db.QueryContext(ctx, `SELECT id, name FROM tag ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*TagRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*TagRecord, error) {
			var r TagRecord
			if err := rows.Scan(&r.ID, &r.Name); err != nil {
				return nil, err
			}
			r.Slug = slugify(r.Name, "tag-"+r.ID)
			return &r, nil
		},
	}, nil
}

// Media maps the video row plus the original media file (videoFile with
// videoStreamingPlaylistId NULL; original = resolution -1, else the highest
// transcode). One LATERAL subquery avoids per-row queries in the scan closure.
func (a *PeerTubeAdapter) Media(ctx context.Context) (Iterator[*MediaRecord], error) {
	query := `
		SELECT v.id, v.uuid, v.name, COALESCE(v.description, ''),
		       COALESCE(v.category, 0), COALESCE(v.licence, 0), COALESCE(v.language, ''),
		       v.privacy, v.duration, v.views, v.likes, v.dislikes, v.state,
		       COALESCE(v."channelId", 0), COALESCE(v."thumbnailPath", ''),
		       COALESCE(v."previewPath", ''), COALESCE(v."commentsEnabled", true),
		       COALESCE(v."downloadEnabled", true), COALESCE(v."publishedAt", 0),
		       COALESCE(v."createdAt", 0), COALESCE(acc."userId", ''),
		       COALESCE(f.filename, ''), COALESCE(f.size, 0), COALESCE(f.extname, ''),
		       COALESCE(f.resolution, 0)
		FROM video v
		LEFT JOIN "videoChannel" vc ON vc.id = v."channelId"
		LEFT JOIN account acc ON acc.id = vc."accountId"
		LEFT JOIN LATERAL (
			SELECT filename, size, extname, resolution
			FROM "videoFile" vf
			WHERE vf."videoId" = v.id AND vf."videoStreamingPlaylistId" IS NULL
			ORDER BY (vf.resolution = -1) DESC, vf.resolution DESC
			LIMIT 1
		) f ON true`
	if a.excludesVideo() {
		query += ` WHERE 1 = 0`
	}
	query += ` ORDER BY v.id`

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query media: %w", err)
	}
	return &sqlIterator[*MediaRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*MediaRecord, error) {
			var (
				r                        MediaRecord
				uuid, desc, lang         sql.NullString
				thumb, preview           sql.NullString
				fileName, ext            sql.NullString
				category, licence        sql.NullInt64
				privacy, state           int
				duration                 int64
				commentsOn, downloadOn   bool
				published, added         sql.NullInt64
				fileSize, channelID      int64
				resolution               int
			)
			err := rows.Scan(&r.ID, &uuid, &r.Title, &desc,
				&category, &licence, &lang,
				&privacy, &duration, &r.Views, &r.Likes, &r.Dislikes, &state,
				&channelID, &thumb, &preview,
				&commentsOn, &downloadOn,
				&published, &added,
				&r.UserID,
				&fileName, &fileSize, &ext, &resolution)
			if err != nil {
				return nil, err
			}
			r.Type = "video"
			r.Duration = float64(duration)
			r.Privacy = mapPeerTubePrivacy(privacy)
			r.State = strconv.Itoa(state)
			r.EncodingStatus = mapPeerTubeState(state)
			r.FileSize = fileSize
			r.AllowDownload = downloadOn
			r.EnableComments = commentsOn
			r.Listable = privacy != 3 && privacy != 4 // not private/internal
			r.CreatedAt = peertubeTime(added)
			if channelID > 0 {
				r.ChannelID = strconv.FormatInt(channelID, 10)
			}
			if category.Valid && category.Int64 > 0 {
				r.CategoryID = strconv.FormatInt(category.Int64, 10)
			}
			if licence.Valid && licence.Int64 > 0 {
				r.LicenseID = strconv.FormatInt(licence.Int64, 10)
			}
			if uuid.Valid {
				r.UID = uuid.String
			}
			if desc.Valid {
				r.Description = desc.String
			}
			if thumb.Valid {
				r.Thumbnail = thumb.String
			}
			if preview.Valid {
				r.Poster = preview.String
			}
			if fileName.Valid {
				r.FileName = fileName.String
			}
			if ext.Valid && ext.String != "" {
				r.MimeType = mimeByExt(ext.String)
			}
			if resolution > 0 {
				r.Height = resolution
			}
			r.Metadata = map[string]string{
				"source_language":     lang.String,
				"source_published_at": peertubeTime(published),
			}
			r.Tags = []string{}
			return &r, nil
		},
	}, nil
}

func (a *PeerTubeAdapter) Comments(ctx context.Context) (Iterator[*CommentRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT c.id, c.text, c."videoId", COALESCE(acc."userId", ''),
		       COALESCE(c."inReplyToCommentId", 0), COALESCE(c."createdAt", 0)
		FROM "videoComment" c
		LEFT JOIN account acc ON acc.id = c."accountId"
		ORDER BY c.id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*CommentRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*CommentRecord, error) {
			var (
				r             CommentRecord
				parentID      int64
				createdAt     sql.NullInt64
			)
			if err := rows.Scan(&r.ID, &r.Text, &r.MediaID, &r.UserID,
				&parentID, &createdAt); err != nil {
				return nil, err
			}
			if parentID > 0 {
				r.ParentID = strconv.FormatInt(parentID, 10)
			}
			r.CreatedAt = peertubeTime(createdAt)
			return &r, nil
		},
	}, nil
}

func (a *PeerTubeAdapter) Channels(ctx context.Context) (Iterator[*ChannelRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT vc.id, COALESCE(acc."userId", ''), vc.name,
		       COALESCE(vc.description, ''), COALESCE(vc."createdAt", 0)
		FROM "videoChannel" vc
		LEFT JOIN account acc ON acc.id = vc."accountId"
		ORDER BY vc.id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*ChannelRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*ChannelRecord, error) {
			var (
				r             ChannelRecord
				desc          sql.NullString
				createdAt     sql.NullInt64
			)
			if err := rows.Scan(&r.ID, &r.UserID, &r.Name, &desc, &createdAt); err != nil {
				return nil, err
			}
			if desc.Valid {
				r.Description = desc.String
			}
			r.Title = r.Name
			r.Slug = slugify(r.Name, "channel-"+r.ID)
			r.CreatedAt = peertubeTime(createdAt)
			return &r, nil
		},
	}, nil
}

func (a *PeerTubeAdapter) Playlists(ctx context.Context) (Iterator[*PlaylistRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT p.id, p.uuid, p.name, COALESCE(p.description, ''),
		       COALESCE(acc."userId", ''), COALESCE(p."createdAt", 0)
		FROM "videoPlaylist" p
		LEFT JOIN account acc ON acc.id = p."ownerAccountId"
		ORDER BY p.id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*PlaylistRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*PlaylistRecord, error) {
			var (
				r            PlaylistRecord
				uuid, desc   sql.NullString
				createdAt    sql.NullInt64
			)
			if err := rows.Scan(&r.ID, &uuid, &r.Name, &desc, &r.UserID, &createdAt); err != nil {
				return nil, err
			}
			if uuid.Valid {
				r.UID = uuid.String
			}
			if desc.Valid {
				r.Description = desc.String
			}
			r.Privacy = "public"
			r.CreatedAt = peertubeTime(createdAt)
			r.MediaIDs = []string{}
			return &r, nil
		},
	}, nil
}

// MediaTags reads the video↔tag pivot.
func (a *PeerTubeAdapter) MediaTags(ctx context.Context) (Iterator[*MediaTagRecord], error) {
	rows, err := a.db.QueryContext(ctx,
		`SELECT "videoId", "tagId" FROM "videoTag" ORDER BY id`)
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

// MediaCategories exposes PeerTube's single category per video as one pivot
// row per video, so the engine's primary-category logic works unchanged.
func (a *PeerTubeAdapter) MediaCategories(ctx context.Context) (Iterator[*MediaCategoryRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT v.id, v.category FROM video v
		WHERE v.category IS NOT NULL ORDER BY v.id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*MediaCategoryRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*MediaCategoryRecord, error) {
			var (
				r          MediaCategoryRecord
				categoryID int
			)
			if err := rows.Scan(&r.MediaID, &categoryID); err != nil {
				return nil, err
			}
			r.CategoryID = strconv.Itoa(categoryID)
			return &r, nil
		},
	}, nil
}

// PlaylistMedia reads the playlist element pivot (ordered by position).
func (a *PeerTubeAdapter) PlaylistMedia(ctx context.Context) (Iterator[*PlaylistMediaRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT e."videoPlaylistId", e."videoId", e.position,
		       COALESCE(e."createdAt", 0)
		FROM "videoPlaylistElement" e
		ORDER BY e."videoPlaylistId", e.position`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*PlaylistMediaRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*PlaylistMediaRecord, error) {
			var (
				r          PlaylistMediaRecord
				createdAt  sql.NullInt64
			)
			if err := rows.Scan(&r.PlaylistID, &r.MediaID, &r.Ordering, &createdAt); err != nil {
				return nil, err
			}
			if createdAt.Valid {
				r.ActionDate = peertubeTime(createdAt)
			}
			return &r, nil
		},
	}, nil
}

// Subtitles reads the videoCaption table (files under storage/captions/).
func (a *PeerTubeAdapter) Subtitles(ctx context.Context) (Iterator[*SubtitleRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT c."videoId", c.language, c.filename, COALESCE(c."createdAt", 0)
		FROM "videoCaption" c
		ORDER BY c.id`)
	if err != nil {
		return nil, err
	}
	return &sqlIterator[*SubtitleRecord]{
		rows: rows,
		scan: func(rows *sql.Rows) (*SubtitleRecord, error) {
			var (
				r         SubtitleRecord
				createdAt sql.NullInt64
			)
			if err := rows.Scan(&r.MediaID, &r.Language, &r.FileURL, &createdAt); err != nil {
				return nil, err
			}
			r.Label = r.Language
			return &r, nil
		},
	}, nil
}

// Subscriptions derives user→channel follows from ActivityPub actorFollow
// edges that point at a local channel. Remote followers are skipped because
// they have no local user row to map onto.
func (a *PeerTubeAdapter) Subscriptions(ctx context.Context) (Iterator[*SubscriptionRecord], error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT vc.id, acc."userId"
		FROM "actorFollow" f
		JOIN "videoChannel" vc ON vc."actorId" = f."targetActorId"
		JOIN actor act ON act.id = f."actorId"
		JOIN account acc ON acc."actorId" = act.id
		WHERE acc."userId" IS NOT NULL
		ORDER BY vc.id`)
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

// Licenses synthesizes PeerTube's fixed license enum (no license table).
func (a *PeerTubeAdapter) Licenses(ctx context.Context) (map[string]*LicenseRecord, error) {
	licenses := make(map[string]*LicenseRecord, len(peerTubeLicenseNames))
	for id, name := range peerTubeLicenseNames {
		idStr := strconv.Itoa(id)
		licenses[idStr] = &LicenseRecord{ID: idStr, Title: name}
	}
	return licenses, nil
}

// RatingsByMedia is not enumerable in PeerTube: likes/dislikes are stored as
// per-video aggregates (video.likes / video.dislikes), already migrated.
func (a *PeerTubeAdapter) RatingsByMedia(ctx context.Context, mediaID string) ([]*RatingRecord, error) {
	return nil, nil
}

// EncodingsByMedia: PeerTube transcode artifacts (videoFile rows beyond the
// original) are not migrated; the target re-transcodes. Nothing to fold in.
func (a *PeerTubeAdapter) EncodingsByMedia(ctx context.Context, mediaID string) ([]*EncodingRecord, error) {
	return nil, nil
}

// MediaPermissionsByMedia has no equivalent table in PeerTube.
func (a *PeerTubeAdapter) MediaPermissionsByMedia(ctx context.Context, mediaID string) ([]*MediaPermissionRecord, error) {
	return nil, nil
}

func (a *PeerTubeAdapter) FileRefs(ctx context.Context, media *MediaRecord) ([]*FileRef, error) {
	if a.mediaDir == "" {
		return nil, nil
	}

	var refs []*FileRef

	// Original media file: prefer the original upload, fall back to the
	// highest transcode. Try videos/ then web-videos/ for either convention.
	fileName, fileSize, err := a.originalFile(ctx, media.ID)
	if err != nil {
		return nil, err
	}
	if fileName != "" {
		if src := a.findFile(fileName); src != "" {
			refs = append(refs, &FileRef{
				SourcePath: src,
				TargetPath: a.paths.Relative("originals", media.UserID, filepath.Base(fileName)),
				Size:       fileSize,
				Kind:       "original",
			})
		}
	}

	thumb, preview, err := a.videoThumbnails(ctx, media.ID)
	if err != nil {
		return nil, err
	}
	if thumb != "" {
		if src := filepath.Join(a.mediaDir, "thumbnails", thumb); fileExists(src) {
			ext := filepath.Ext(thumb)
			if ext == "" {
				ext = ".jpg"
			}
			refs = append(refs, &FileRef{
				SourcePath: src,
				TargetPath: a.paths.Relative("thumbnails", media.ID+ext),
				Kind:       "thumbnail",
			})
		}
	}
	if preview != "" && preview != thumb {
		if src := filepath.Join(a.mediaDir, "previews", preview); fileExists(src) {
			ext := filepath.Ext(preview)
			if ext == "" {
				ext = ".jpg"
			}
			refs = append(refs, &FileRef{
				SourcePath: src,
				TargetPath: a.paths.Relative("thumbnails", media.ID+"_poster"+ext),
				Kind:       "poster",
			})
		}
	}

	captionRefs, err := a.captionRefs(ctx, media)
	if err != nil {
		return nil, err
	}
	refs = append(refs, captionRefs...)

	return refs, nil
}

// originalFile returns the filename and size of the video's original media
// file (videoStreamingPlaylistId NULL, resolution -1 preferred). A missing
// original falls back to the highest transcode; an empty filename means no
// file row exists at all.
func (a *PeerTubeAdapter) originalFile(ctx context.Context, videoID string) (string, int64, error) {
	var (
		filename string
		size     int64
	)
	err := a.db.QueryRowContext(ctx, `
		SELECT filename, size FROM "videoFile"
		WHERE "videoId" = $1 AND "videoStreamingPlaylistId" IS NULL
		ORDER BY (resolution = -1) DESC, resolution DESC
		LIMIT 1`, videoID).Scan(&filename, &size)
	if err == sql.ErrNoRows {
		return "", 0, nil
	}
	return filename, size, err
}

// videoThumbnails returns the thumbnail and preview filenames for a video.
func (a *PeerTubeAdapter) videoThumbnails(ctx context.Context, videoID string) (string, string, error) {
	var thumb, preview string
	err := a.db.QueryRowContext(ctx, `
		SELECT COALESCE("thumbnailPath", ''), COALESCE("previewPath", '')
		FROM video WHERE id = $1`, videoID).Scan(&thumb, &preview)
	if err == sql.ErrNoRows {
		return "", "", nil
	}
	return thumb, preview, err
}

// findFile locates a video file under either the original or transcode dir.
func (a *PeerTubeAdapter) findFile(filename string) string {
	for _, dir := range []string{"videos", "web-videos"} {
		candidate := filepath.Join(a.mediaDir, dir, filename)
		if fileExists(candidate) {
			return candidate
		}
	}
	return ""
}

// captionRefs resolves the on-disk subtitle tracks for one video into file
// refs under originals/subtitles/{srcMediaID}/{lang}{ext}, matching the DB
// path phaseSubtitles writes.
func (a *PeerTubeAdapter) captionRefs(ctx context.Context, media *MediaRecord) ([]*FileRef, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT filename, language FROM "videoCaption" WHERE "videoId" = $1`, media.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refs []*FileRef
	for rows.Next() {
		var filename, lang string
		if err := rows.Scan(&filename, &lang); err != nil {
			return nil, err
		}
		src := filepath.Join(a.mediaDir, "captions", filename)
		if !fileExists(src) {
			continue
		}
		ext := filepath.Ext(filename)
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

func (a *PeerTubeAdapter) OpenFile(ctx context.Context, ref *FileRef) (io.ReadCloser, error) {
	return os.Open(ref.SourcePath)
}

// fileExists reports whether path exists and is a regular file.
func fileExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

// peertubeTime converts PeerTube's BIGINT epoch-millisecond timestamp into an
// RFC3339 string (or "" for 0/NULL), which the engine's parseTime accepts.
func peertubeTime(ms sql.NullInt64) string {
	if !ms.Valid || ms.Int64 <= 0 {
		return ""
	}
	return time.UnixMilli(ms.Int64).UTC().Format(time.RFC3339)
}

// mapPeerTubePrivacy maps PeerTube privacy (1 Public, 2 Unlisted,
// 3 Private, 4 Internal) onto the target privacy enum.
func mapPeerTubePrivacy(p int) string {
	switch p {
	case 2:
		return "UNLISTED"
	case 3, 4:
		return "PRIVATE"
	default:
		return "PUBLIC"
	}
}

// mapPeerTubeState maps PeerTube video.state onto the target encoding-status
// enum the engine consumes (success / processing / failed / pending).
// 1 published, 4 waiting for transcoder, 5 transcoding, 6 transcoding failed.
func mapPeerTubeState(s int) string {
	switch s {
	case 1:
		return "success"
	case 4, 5:
		return "processing"
	case 6:
		return "failed"
	default:
		return "pending"
	}
}

// sliceIterator adapts an in-memory slice to the Iterator interface.
type sliceIterator[T any] struct {
	items []T
	idx   int
	err   error
}

func (it *sliceIterator[T]) Next(ctx context.Context) bool {
	if it.idx >= len(it.items) {
		return false
	}
	it.idx++
	return true
}

func (it *sliceIterator[T]) Item() T      { return it.items[it.idx-1] }
func (it *sliceIterator[T]) Err() error   { return it.err }
func (it *sliceIterator[T]) Close() error { return nil }

// peerTubeCategoryNames is PeerTube's fixed category enum (no DB table).
var peerTubeCategoryNames = map[int]string{
	1:  "Music",
	2:  "Films",
	3:  "Vehicles",
	4:  "Art",
	5:  "Sports",
	6:  "Travel",
	7:  "Gaming",
	8:  "People",
	9:  "Comedy",
	10: "Entertainment",
	11: "News & Politics",
	12: "Howto & Style",
	13: "Education",
	14: "Science & Technology",
	15: "Animals",
}

// peerTubeLicenseNames is PeerTube's fixed license enum (no DB table).
var peerTubeLicenseNames = map[int]string{
	1: "Attribution",
	2: "Attribution-ShareAlike",
	3: "Attribution-NonCommercial",
	4: "Attribution-NonCommercial-ShareAlike",
	5: "Attribution-NoDerivs",
	6: "Attribution-NonCommercial-NoDerivs",
	7: "Public Domain Dedication",
}
