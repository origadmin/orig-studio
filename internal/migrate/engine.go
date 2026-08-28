package migrate

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/entity/media"
	"origadmin/application/origstudio/internal/pkg/idutil"

	"entgo.io/ent/dialect/sql"
	_ "github.com/sqlite3ent/sqlite3"
	_ "github.com/lib/pq"
)

type Engine struct {
	source   SourceAdapter
	target   *entity.Client
	mapper   *MemoryIDMapper
	reporter ProgressReporter
	config   *TargetConfig
	state    *MigrationState
	stateDir string
	logger   *log.Logger
}

func NewEngine(source SourceAdapter, reporter ProgressReporter, stateDir string, logger *log.Logger) *Engine {
	return &Engine{
		source:   source,
		mapper:   NewMemoryIDMapper(),
		reporter: reporter,
		stateDir: stateDir,
		logger:   logger,
	}
}

func (e *Engine) Source(adapter SourceAdapter) {
	e.source = adapter
}

func (e *Engine) ConnectTarget(ctx context.Context, cfg *TargetConfig) error {
	e.config = cfg
	drv, err := sql.Open(cfg.Dialect, cfg.DSN)
	if err != nil {
		return fmt.Errorf("open target db: %w", err)
	}
	e.target = entity.NewClient(entity.Driver(drv))

	// Create the target schema (idempotent) so a fresh database can be
	// migrated without a separate schema-init step.
	if err := e.target.Schema.Create(ctx); err != nil {
		return fmt.Errorf("create target schema: %w", err)
	}
	return nil
}

func (e *Engine) LoadState(migrationID string) error {
	statePath := filepath.Join(e.stateDir, migrationID, "state.json")
	state, err := LoadMigrationState(statePath)
	if err != nil {
		return err
	}
	if state != nil {
		e.state = state
		mapperPath := filepath.Join(e.stateDir, migrationID, "id_map.json")
		if err := e.mapper.LoadFromFile(mapperPath); err != nil {
			e.logger.Printf("warning: failed to load id map: %v", err)
		}
	}
	return nil
}

func (e *Engine) SaveState() error {
	if e.state == nil {
		return nil
	}
	statePath := filepath.Join(e.stateDir, e.state.ID, "state.json")
	if err := SaveMigrationState(statePath, e.state); err != nil {
		return err
	}
	mapperPath := filepath.Join(e.stateDir, e.state.ID, "id_map.json")
	return e.mapper.SaveToFile(mapperPath)
}

func (e *Engine) Run(ctx context.Context, sourceCfg *SourceConfig, targetCfg *TargetConfig) error {
	if err := e.source.Connect(ctx, sourceCfg); err != nil {
		return fmt.Errorf("connect source: %w", err)
	}
	defer e.source.Close()

	if err := e.ConnectTarget(ctx, targetCfg); err != nil {
		return fmt.Errorf("connect target: %w", err)
	}

	migrationID := fmt.Sprintf("migrate-%d", time.Now().Unix())
	e.state = &MigrationState{
		ID:         migrationID,
		Source:     *sourceCfg,
		Target:     *targetCfg,
		Progress:   Progress{Status: StatusRunning, StartedAt: time.Now(), UpdatedAt: time.Now()},
		PhaseState: make(map[string]string),
		CreatedAt:  time.Now().Format(time.RFC3339),
		UpdatedAt:  time.Now().Format(time.RFC3339),
	}

	phases := DefaultPhases()

	for _, phase := range phases {
		if e.state.Progress.Phase == phase.Name && e.state.Progress.Status == StatusCompleted {
			e.logger.Printf("Skipping completed phase: %s", phase.Name)
			continue
		}

		e.updatePhase(phase.Name)
		e.logger.Printf("Starting phase: %s", phase.Name)

		if err := phase.Fn(ctx, e); err != nil {
			e.updateStatus(StatusFailed, err.Error())
			_ = e.SaveState()
			return fmt.Errorf("phase %s failed: %w", phase.Name, err)
		}

		e.updatePhaseProgress(0, 0)
		e.logger.Printf("Completed phase: %s", phase.Name)
		_ = e.SaveState()
	}

	e.updateStatus(StatusCompleted, "")
	_ = e.SaveState()
	return nil
}

func (e *Engine) Resume(ctx context.Context) error {
	if e.state == nil {
		return fmt.Errorf("no state to resume")
	}

	e.logger.Printf("Resuming migration %s from phase %s", e.state.ID, e.state.Progress.Phase)

	if err := e.source.Connect(ctx, &e.state.Source); err != nil {
		return fmt.Errorf("reconnect source: %w", err)
	}
	defer e.source.Close()

	if err := e.ConnectTarget(ctx, &e.state.Target); err != nil {
		return fmt.Errorf("reconnect target: %w", err)
	}

	e.updateStatus(StatusRunning, "")

	phases := DefaultPhases()

	startFrom := e.state.Progress.Phase
	found := false
	for _, phase := range phases {
		if !found && phase.Name != startFrom {
			continue
		}
		found = true

		if e.state.PhaseState[string(phase.Name)] == string(StatusCompleted) && phase.Name != startFrom {
			continue
		}

		e.updatePhase(phase.Name)
		e.logger.Printf("Resuming phase: %s", phase.Name)

		if err := phase.Fn(ctx, e); err != nil {
			e.updateStatus(StatusFailed, err.Error())
			_ = e.SaveState()
			return fmt.Errorf("phase %s failed: %w", phase.Name, err)
		}

		e.state.PhaseState[string(phase.Name)] = string(StatusCompleted)
		_ = e.SaveState()
	}

	e.updateStatus(StatusCompleted, "")
	_ = e.SaveState()
	return nil
}

func (e *Engine) phaseDiscover(ctx context.Context) error {
	stats, err := e.source.Discover(ctx)
	if err != nil {
		return err
	}
	e.logger.Printf("Discovery: %d users, %d media, %d categories, %d tags, %d comments, %d channels",
		stats.Users, stats.Media, stats.Categories, stats.Tags, stats.Comments, stats.Channels)
	e.state.PhaseState[string(PhaseDiscover)] = string(StatusCompleted)
	return nil
}

func (e *Engine) phaseUsers(ctx context.Context) error {
	iter, err := e.source.Users(ctx)
	if err != nil {
		return err
	}
	defer iter.Close()

	var count, failed int64
	for iter.Next(ctx) {
		rec := iter.Item()
		count++
		e.updatePhaseProgress(count, 0)
		e.updateCurrentItem(rec.Username)

		if e.config.DryRun {
			e.mapper.Set("user", rec.ID, "dry-run-"+rec.ID)
			continue
		}

		targetID := idutil.GenUUIDv7()
		builder := e.target.User.Create().
			SetID(targetID).
			SetUsername(rec.Username).
			SetEmail(rec.Email).
			SetPassword(rec.Password).
			SetName(rec.DisplayName).
			SetSlug(rec.Username).
			SetAvatar(rec.Avatar).
			SetDescription(rec.Bio).
			SetStatus("ACTIVE").
			SetRole("user")

		if rec.Role == "admin" {
			builder.SetRole("admin").SetIsSuperuser(true)
		}
		if !rec.IsActive {
			builder.SetStatus("SUSPENDED")
		}

		if _, err := builder.Save(ctx); err != nil {
			failed++
			e.reporter.ReportError(PhaseUsers, rec.Username, err)
			continue
		}

		e.mapper.Set("user", rec.ID, targetID)
	}
	if iter.Err() != nil {
		return iter.Err()
	}

	e.updatePhaseProgress(count, failed)
	e.state.PhaseState[string(PhaseUsers)] = string(StatusCompleted)
	e.logger.Printf("Users: migrated %d, failed %d", count, failed)
	return nil
}

func (e *Engine) phaseCategories(ctx context.Context) error {
	iter, err := e.source.Categories(ctx)
	if err != nil {
		return err
	}
	defer iter.Close()

	var count, failed int64
	for iter.Next(ctx) {
		rec := iter.Item()
		count++
		e.updatePhaseProgress(count, 0)

		if e.config.DryRun {
			e.mapper.Set("category", rec.ID, "dry-run-"+rec.ID)
			continue
		}

		builder := e.target.Category.Create().
			SetName(rec.Name).
			SetSlug(rec.Slug).
			SetDescription(rec.Description).
			SetStatus("ACTIVE")

		if rec.ParentID != "" {
			if parentID, ok := e.mapper.Map("category", rec.ParentID); ok {
				pid, _ := strconv.ParseInt(parentID, 10, 64)
				builder.SetParentID(pid)
			}
		}

		saved, err := builder.Save(ctx)
		if err != nil {
			failed++
			e.reporter.ReportError(PhaseCategories, rec.Name, err)
			continue
		}

		e.mapper.Set("category", rec.ID, fmt.Sprintf("%d", saved.ID))
	}
	if iter.Err() != nil {
		return iter.Err()
	}

	e.updatePhaseProgress(count, failed)
	e.state.PhaseState[string(PhaseCategories)] = string(StatusCompleted)
	e.logger.Printf("Categories: migrated %d, failed %d", count, failed)
	return nil
}

func (e *Engine) phaseTags(ctx context.Context) error {
	iter, err := e.source.Tags(ctx)
	if err != nil {
		return err
	}
	defer iter.Close()

	var count, failed int64
	for iter.Next(ctx) {
		rec := iter.Item()
		count++
		e.updatePhaseProgress(count, 0)

		if e.config.DryRun {
			e.mapper.Set("tag", rec.ID, "dry-run-"+rec.ID)
			continue
		}

		saved, err := e.target.Tag.Create().
			SetTitle(rec.Name).
			SetSlug(rec.Slug).
			SetStatus("ACTIVE").
			Save(ctx)
		if err != nil {
			failed++
			e.reporter.ReportError(PhaseTags, rec.Name, err)
			continue
		}

		e.mapper.Set("tag", rec.ID, fmt.Sprintf("%d", saved.ID))
	}
	if iter.Err() != nil {
		return iter.Err()
	}

	e.updatePhaseProgress(count, failed)
	e.state.PhaseState[string(PhaseTags)] = string(StatusCompleted)
	return nil
}

func (e *Engine) phaseChannels(ctx context.Context) error {
	iter, err := e.source.Channels(ctx)
	if err != nil {
		return err
	}
	defer iter.Close()

	var count, failed int64
	for iter.Next(ctx) {
		rec := iter.Item()
		count++
		e.updatePhaseProgress(count, 0)

		if e.config.DryRun {
			e.mapper.Set("channel", rec.ID, "dry-run-"+rec.ID)
			continue
		}

		userID, ok := e.mapper.Map("user", rec.UserID)
		if !ok {
			failed++
			e.reporter.ReportError(PhaseChannels, rec.Name, fmt.Errorf("user %s not found", rec.UserID))
			continue
		}

		targetID := idutil.GenUUIDv7()
		builder := e.target.Channel.Create().
			SetID(targetID).
			SetUserID(userID).
			SetName(rec.Name).
			SetTitle(firstNonEmpty(rec.Title, rec.Name)).
			SetSlug(rec.Slug).
			SetHandle(rec.Slug).
			SetDescription(rec.Description).
			SetStatus("ACTIVE")

		if rec.Avatar != "" {
			builder.SetAvatar(rec.Avatar)
		}
		if rec.Banner != "" {
			builder.SetBanner(rec.Banner)
		}

		if _, err := builder.Save(ctx); err != nil {
			failed++
			e.reporter.ReportError(PhaseChannels, rec.Name, err)
			continue
		}

		e.mapper.Set("channel", rec.ID, targetID)
	}
	if iter.Err() != nil {
		return iter.Err()
	}

	e.updatePhaseProgress(count, failed)
	e.state.PhaseState[string(PhaseChannels)] = string(StatusCompleted)
	return nil
}

func (e *Engine) phaseMedia(ctx context.Context) error {
	iter, err := e.source.Media(ctx)
	if err != nil {
		return err
	}
	defer iter.Close()

	// Batch 2: preload M2M relations and name lookups once, so metadata
	// assembly inside the loop needs no extra per-row scans beyond the
	// per-media rating/encoding/permission queries.
	mediaCatMap, err := e.loadMediaCategoryMap(ctx)
	if err != nil {
		return err
	}
	mediaTagMap, err := e.loadMediaTagMap(ctx)
	if err != nil {
		return err
	}
	catNames, err := loadNameMap(ctx, e.source.Categories,
		func(r *CategoryRecord) string { return r.ID },
		func(r *CategoryRecord) string { return r.Name })
	if err != nil {
		return err
	}
	tagNames, err := loadNameMap(ctx, e.source.Tags,
		func(r *TagRecord) string { return r.ID },
		func(r *TagRecord) string { return r.Name })
	if err != nil {
		return err
	}
	licenses, err := e.source.Licenses(ctx)
	if err != nil {
		return err
	}

	var count, failed int64
	for iter.Next(ctx) {
		rec := iter.Item()
		count++
		e.updatePhaseProgress(count, 0)
		e.updateCurrentItem(rec.Title)

		if e.config.DryRun {
			e.mapper.Set("media", rec.ID, "dry-run-"+rec.ID)
			continue
		}

		userID, ok := e.mapper.Map("user", rec.UserID)
		if !ok {
			failed++
			e.reporter.ReportError(PhaseMedia, rec.Title, fmt.Errorf("user %s not found", rec.UserID))
			continue
		}

		targetID := idutil.GenUUIDv7()
		mediaType := "video"
		if rec.Type == "image" {
			mediaType = "image"
		} else if rec.Type == "audio" {
			mediaType = "audio"
		}

		// M2M: primary category -> category_id; the rest + all tags -> metadata.
		var (
			primaryCatID     string
			additionalCats   []string
			tagNamesForMedia []string
		)
		if catIDs := mediaCatMap[rec.ID]; len(catIDs) > 0 {
			primaryCatID = catIDs[0]
			for _, cid := range catIDs[1:] {
				if name, ok := catNames[cid]; ok {
					additionalCats = append(additionalCats, name)
				}
			}
		}
		for _, tid := range mediaTagMap[rec.ID] {
			if name, ok := tagNames[tid]; ok {
				tagNamesForMedia = append(tagNamesForMedia, name)
			}
		}

		// Metadata: all A-side data that has no B-side entity, stored in full.
		metadata := map[string]interface{}{
			"source_uid":   rec.UID,
			"source_state": rec.State,
			"media_info":   rec.MediaInfo,
			"md5sum":       rec.Md5sum,
		}
		if rec.LicenseID != "" {
			if lic, ok := licenses[rec.LicenseID]; ok {
				metadata["license"] = lic
			}
		}
		if ratings, err := e.source.RatingsByMedia(ctx, rec.ID); err == nil && len(ratings) > 0 {
			metadata["ratings"] = ratings
		}
		if encodings, err := e.source.EncodingsByMedia(ctx, rec.ID); err == nil && len(encodings) > 0 {
			metadata["encodings"] = encodings
		}
		if perms, err := e.source.MediaPermissionsByMedia(ctx, rec.ID); err == nil && len(perms) > 0 {
			metadata["media_permissions"] = perms
		}
		if len(additionalCats) > 0 {
			metadata["additional_categories"] = additionalCats
		}
		if len(tagNamesForMedia) > 0 {
			metadata["additional_tags"] = tagNamesForMedia
		}
		if len(rec.Metadata) > 0 {
			for k, v := range rec.Metadata {
				metadata["source_"+k] = v
			}
		}

		// Target storage keys: align every DB path field with where phaseFiles
		// will copy the file, so /files/{key} resolves on the target system.
		// When no source media dir is configured FileRefs returns nil and the
		// source-relative paths are kept as-is (metadata-only migration).
		targetURL, targetThumbnail, targetPoster, targetHLS :=
			rec.FilePath, rec.Thumbnail, rec.Poster, rec.HLSFile
		targetVtt, targetSpriteImg := rec.Sprites, ""
		if rec.Sprites != "" {
			targetSpriteImg = spriteImagePath(rec.Sprites)
		}
		// targetSize prefers the real on-disk source file size (authoritative);
		// falls back to the parsed DB size field when the file is unavailable.
		targetSize := rec.FileSize
		if refs, err := e.source.FileRefs(ctx, rec); err == nil {
			for _, ref := range refs {
				switch ref.Kind {
				case "original":
					targetURL = ref.TargetPath
					if fi, err := os.Stat(ref.SourcePath); err == nil && fi.Size() > 0 {
						targetSize = fi.Size()
					}
				case "thumbnail":
					targetThumbnail = ref.TargetPath
				case "poster":
					targetPoster = ref.TargetPath
				case "sprite_vtt":
					targetVtt = ref.TargetPath
				case "sprite_jpg":
					targetSpriteImg = ref.TargetPath
				}
			}
		}

		builder := e.target.Media.Create().
			SetID(targetID).
			SetUserID(userID).
			SetTitle(rec.Title).
			SetDescription(rec.Description).
			SetType(mediaType).
			SetURL(targetURL).
			SetSize(fmt.Sprintf("%d", targetSize)).
			SetDuration(int(rec.Duration)).
			SetWidth(rec.Width).
			SetHeight(rec.Height).
			SetMimeType(rec.MimeType).
			SetSha256(rec.Checksum).
			SetState("active").
			SetEncodingStatus(rec.EncodingStatus).
			SetShortToken(idutil.GenShortID()).
			SetPrivacy(media.Privacy(rec.Privacy)).
			SetViewCount(int64(rec.Views)).
			SetLikeCount(int64(rec.Likes)).
			SetDislikeCount(int64(rec.Dislikes)).
			SetAllowDownload(rec.AllowDownload).
			SetEnableComments(rec.EnableComments).
			SetFeatured(rec.Featured).
			SetListable(rec.Listable).
			SetReportedTimes(rec.ReportedTimes).
			SetReviewStatus(mapReviewStatus(rec.IsReviewed)).
			SetTags(tagNamesForMedia).
			SetMetadata(metadata)

		if ext := filepath.Ext(rec.FileName); ext != "" {
			builder.SetExtension(strings.TrimPrefix(ext, "."))
		}
		if rec.Thumbnail != "" {
			builder.SetThumbnail(targetThumbnail)
		}
		if rec.ChannelID != "" {
			if chID, ok := e.mapper.Map("channel", rec.ChannelID); ok {
				builder.SetChannelID(chID)
			}
		}
		if primaryCatID != "" {
			if catID, ok := e.mapper.Map("category", primaryCatID); ok {
				catIDInt, _ := strconv.ParseInt(catID, 10, 64)
				builder.SetCategoryID(catIDInt)
			}
		}
		if rec.HLSFile != "" {
			builder.SetHlsFile(targetHLS)
		}
		if rec.Poster != "" {
			builder.SetPoster(targetPoster)
		}
		if rec.PreviewFile != "" {
			builder.SetPreviewFilePath(rec.PreviewFile)
		}
		if rec.ThumbnailTime > 0 {
			builder.SetThumbnailTime(rec.ThumbnailTime)
		}
		// Sprites field holds the VTT file; the JPG sprite sheet is derived.
		if rec.Sprites != "" {
			builder.SetSpriteStatus("done").
				SetVttPath(targetVtt).
				SetSpritePath(targetSpriteImg)
		}
		if rec.CreatedAt != "" {
			if t, err := parseTime(rec.CreatedAt); err == nil {
				builder.SetPublishedAt(t)
			}
		}

		if _, err := builder.Save(ctx); err != nil {
			failed++
			e.reporter.ReportError(PhaseMedia, rec.Title, err)
			continue
		}

		e.mapper.Set("media", rec.ID, targetID)
	}
	if iter.Err() != nil {
		return iter.Err()
	}

	e.updatePhaseProgress(count, failed)
	e.state.PhaseState[string(PhaseMedia)] = string(StatusCompleted)
	e.logger.Printf("Media: migrated %d, failed %d", count, failed)
	return nil
}

// loadMediaCategoryMap loads media_id -> ordered category ids from the M2M pivot.
func (e *Engine) loadMediaCategoryMap(ctx context.Context) (map[string][]string, error) {
	iter, err := e.source.MediaCategories(ctx)
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	m := make(map[string][]string)
	for iter.Next(ctx) {
		rec := iter.Item()
		m[rec.MediaID] = append(m[rec.MediaID], rec.CategoryID)
	}
	return m, iter.Err()
}

// loadMediaTagMap loads media_id -> ordered tag ids from the M2M pivot.
func (e *Engine) loadMediaTagMap(ctx context.Context) (map[string][]string, error) {
	iter, err := e.source.MediaTags(ctx)
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	m := make(map[string][]string)
	for iter.Next(ctx) {
		rec := iter.Item()
		m[rec.MediaID] = append(m[rec.MediaID], rec.TagID)
	}
	return m, iter.Err()
}

// loadNameMap builds id -> name lookups from any name-carrying iterator.
func loadNameMap[T any](ctx context.Context, next func(context.Context) (Iterator[*T], error), id func(*T) string, name func(*T) string) (map[string]string, error) {
	iter, err := next(ctx)
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	m := make(map[string]string)
	for iter.Next(ctx) {
		r := iter.Item()
		m[id(r)] = name(r)
	}
	return m, iter.Err()
}

// mapReviewStatus translates the A-side is_reviewed flag into B-side review_status.
func mapReviewStatus(reviewed bool) string {
	if reviewed {
		return "approved"
	}
	return "pending_review"
}

// spriteImagePath derives the sprite JPG path from a sprite VTT path.
func spriteImagePath(vttPath string) string {
	ext := filepath.Ext(vttPath)
	if ext == "" {
		return vttPath + ".jpg"
	}
	return strings.TrimSuffix(vttPath, ext) + ".jpg"
}

// firstNonEmpty returns the first non-empty string, falling back to the last.
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// parseTime parses PostgreSQL/SQLite timestamp text in the common layouts.
func parseTime(s string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05.999999-07",
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05-07",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unparsable time %q", s)
}

func (e *Engine) phaseComments(ctx context.Context) error {
	// Pass 1: insert every comment without its parent edge so parent/child
	// order in the source iterator never matters, and record the source→target
	// id mapping for pass 2.
	if err := e.pass1Comments(ctx); err != nil {
		return err
	}
	// Pass 2: wire up the parent edges using the mapping from pass 1.
	if err := e.pass2CommentParents(ctx); err != nil {
		return err
	}
	e.state.PhaseState[string(PhaseComments)] = string(StatusCompleted)
	return nil
}

func (e *Engine) pass1Comments(ctx context.Context) error {
	iter, err := e.source.Comments(ctx)
	if err != nil {
		return err
	}
	defer iter.Close()

	var count, failed int64
	for iter.Next(ctx) {
		rec := iter.Item()
		count++
		e.updatePhaseProgress(count, 0)

		if e.config.DryRun {
			e.mapper.Set("comment", rec.ID, "dry-run-"+rec.ID)
			continue
		}

		mediaID, ok := e.mapper.Map("media", rec.MediaID)
		if !ok {
			failed++
			continue
		}
		userID, ok := e.mapper.Map("user", rec.UserID)
		if !ok {
			failed++
			continue
		}

		created, err := e.target.Comment.Create().
			SetID(idutil.GenUUIDv7()).
			SetMediaID(mediaID).
			SetUserID(userID).
			SetText(rec.Text).
			SetStatus("APPROVED").
			Save(ctx)
		if err != nil {
			failed++
			e.reporter.ReportError(PhaseComments, rec.ID, err)
			continue
		}
		e.mapper.Set("comment", rec.ID, created.ID)
	}
	if iter.Err() != nil {
		return iter.Err()
	}
	e.updatePhaseProgress(count, failed)
	return nil
}

func (e *Engine) pass2CommentParents(ctx context.Context) error {
	iter, err := e.source.Comments(ctx)
	if err != nil {
		return err
	}
	defer iter.Close()

	var count, failed int64
	for iter.Next(ctx) {
		rec := iter.Item()
		count++
		e.updatePhaseProgress(count, 0)

		if rec.ParentID == "" || e.config.DryRun {
			continue
		}
		targetID, ok := e.mapper.Map("comment", rec.ID)
		if !ok {
			failed++
			continue
		}
		parentID, ok := e.mapper.Map("comment", rec.ParentID)
		if !ok {
			failed++
			continue
		}
		if _, err := e.target.Comment.UpdateOneID(targetID).
			SetParentID(parentID).
			Save(ctx); err != nil {
			failed++
			e.reporter.ReportError(PhaseComments, rec.ID, err)
			continue
		}
	}
	if iter.Err() != nil {
		return iter.Err()
	}
	e.updatePhaseProgress(count, failed)
	return nil
}

func (e *Engine) phaseFiles(ctx context.Context) error {
	if e.config.MediaDir == "" {
		e.logger.Printf("No media dir configured, skipping file copy")
		e.state.PhaseState[string(PhaseFiles)] = string(StatusCompleted)
		return nil
	}

	iter, err := e.source.Media(ctx)
	if err != nil {
		return err
	}
	defer iter.Close()

	var count, failed int64
	for iter.Next(ctx) {
		rec := iter.Item()
		refs, err := e.source.FileRefs(ctx, rec)
		if err != nil {
			e.reporter.ReportWarning(PhaseFiles, rec.ID, fmt.Sprintf("get file refs: %v", err))
			continue
		}

		for _, ref := range refs {
			count++
			e.updatePhaseProgress(count, 0)

			if e.config.DryRun {
				continue
			}

			targetPath := filepath.Join(e.config.MediaDir, ref.TargetPath)
			targetDir := filepath.Dir(targetPath)
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				failed++
				e.reporter.ReportError(PhaseFiles, ref.SourcePath, err)
				continue
			}

			if _, err := os.Stat(targetPath); err == nil && !e.config.Overwrite {
				continue
			}

			reader, err := e.source.OpenFile(ctx, ref)
			if err != nil {
				failed++
				e.reporter.ReportError(PhaseFiles, ref.SourcePath, err)
				continue
			}

			writer, err := os.Create(targetPath)
			if err != nil {
				reader.Close()
				failed++
				e.reporter.ReportError(PhaseFiles, ref.TargetPath, err)
				continue
			}

			hasher := sha256.New()
			_, err = io.Copy(writer, io.TeeReader(reader, hasher))
			reader.Close()
			writer.Close()

			if err != nil {
				os.Remove(targetPath)
				failed++
				e.reporter.ReportError(PhaseFiles, ref.SourcePath, err)
				continue
			}

			if ref.Checksum != "" {
				actual := fmt.Sprintf("%x", hasher.Sum(nil))
				if actual != ref.Checksum {
					os.Remove(targetPath)
					failed++
					e.reporter.ReportError(PhaseFiles, ref.SourcePath,
						fmt.Errorf("checksum mismatch: expected %s, got %s", ref.Checksum, actual))
					continue
				}
			}
		}
	}
	if iter.Err() != nil {
		return iter.Err()
	}

	e.updatePhaseProgress(count, failed)
	e.state.PhaseState[string(PhaseFiles)] = string(StatusCompleted)
	e.logger.Printf("Files: copied %d, failed %d", count, failed)
	return nil
}

func (e *Engine) phaseVerify(ctx context.Context) error {
	e.logger.Printf("Verification phase - checking data integrity...")

	type entityCount struct {
		name   string
		target int64
	}
	counts := []entityCount{
		{"users", mustCount(ctx, e.target.User.Query().Count)},
		{"channels", mustCount(ctx, e.target.Channel.Query().Count)},
		{"categories", mustCount(ctx, e.target.Category.Query().Count)},
		{"tags", mustCount(ctx, e.target.Tag.Query().Count)},
		{"media", mustCount(ctx, e.target.Media.Query().Count)},
		{"comments", mustCount(ctx, e.target.Comment.Query().Count)},
		{"playlists", mustCount(ctx, e.target.Playlist.Query().Count)},
		{"playlist_media", mustCount(ctx, e.target.MediaPlaylist.Query().Count)},
		{"subtitles", mustCount(ctx, e.target.Subtitle.Query().Count)},
		{"subscriptions", mustCount(ctx, e.target.Subscription.Query().Count)},
	}

	// Compare against the source discovery only when the run actually wrote
	// to the target (dry-run leaves the target empty on purpose).
	if !e.config.DryRun {
		stats, err := e.source.Discover(ctx)
		if err != nil {
			return fmt.Errorf("verify re-discover source: %w", err)
		}
		expect := map[string]int{
			"users":      stats.Users,
			"channels":   stats.Channels,
			"categories": stats.Categories,
			"tags":       stats.Tags,
			"media":      stats.Media,
			"comments":   stats.Comments,
			"playlists":  stats.Playlists,
		}
		for _, c := range counts {
			if want, ok := expect[c.name]; ok && int(c.target) != want {
				e.reporter.ReportError(PhaseVerify, c.name,
					fmt.Errorf("count mismatch: source=%d target=%d", want, c.target))
			}
		}
	}

	var summary strings.Builder
	summary.WriteString("Verification: ")
	for i, c := range counts {
		if i > 0 {
			summary.WriteString(", ")
		}
		fmt.Fprintf(&summary, "%s=%d", c.name, c.target)
	}
	e.logger.Printf("%s", summary.String())
	e.state.PhaseState[string(PhaseVerify)] = string(StatusCompleted)
	return nil
}

func mustCount(ctx context.Context, fn func(context.Context) (int, error)) int64 {
	n, err := fn(ctx)
	if err != nil {
		return -1
	}
	return int64(n)
}

func (e *Engine) updatePhase(phase Phase) {
	if e.state != nil {
		e.state.Progress.Phase = phase
		e.state.Progress.UpdatedAt = time.Now()
	}
	e.reporter.UpdateProgress(&e.state.Progress)
}

func (e *Engine) updatePhaseProgress(done, failed int64) {
	if e.state != nil {
		e.state.Progress.DoneItems = done
		e.state.Progress.FailedItems = failed
		e.state.Progress.UpdatedAt = time.Now()
	}
	e.reporter.UpdateProgress(&e.state.Progress)
}

func (e *Engine) updateCurrentItem(item string) {
	if e.state != nil {
		e.state.Progress.CurrentItem = item
	}
}

func (e *Engine) updateStatus(status Status, errMsg string) {
	if e.state != nil {
		e.state.Progress.Status = status
		e.state.Progress.Error = errMsg
		e.state.Progress.UpdatedAt = time.Now()
		e.state.UpdatedAt = time.Now().Format(time.RFC3339)
	}
	e.reporter.UpdateProgress(&e.state.Progress)
}

func (e *Engine) phasePlaylists(ctx context.Context) error {
	iter, err := e.source.Playlists(ctx)
	if err != nil {
		return err
	}
	defer iter.Close()

	var count, failed int64
	for iter.Next(ctx) {
		rec := iter.Item()
		count++
		e.updatePhaseProgress(count, 0)

		if e.config.DryRun {
			e.mapper.Set("playlist", rec.ID, "dry-run-"+rec.ID)
			continue
		}

		userID, ok := e.mapper.Map("user", rec.UserID)
		if !ok {
			failed++
			e.reporter.ReportError(PhasePlaylists, rec.Name, fmt.Errorf("user %s not found", rec.UserID))
			continue
		}

		targetID := idutil.GenUUIDv7()
		_, err := e.target.Playlist.Create().
			SetID(targetID).
			SetUserID(userID).
			AddUserIDs(userID). // required edge "user" is a M2M join in this schema
			SetTitle(rec.Name).
			SetDescription(rec.Description).
			SetShortToken(idutil.GenShortID()).
			SetPrivacy("PUBLIC").
			SetStatus("ACTIVE").
			Save(ctx)
		if err != nil {
			failed++
			e.reporter.ReportError(PhasePlaylists, rec.Name, err)
			continue
		}

		e.mapper.Set("playlist", rec.ID, targetID)
	}
	if iter.Err() != nil {
		return iter.Err()
	}

	e.updatePhaseProgress(count, failed)
	e.state.PhaseState[string(PhasePlaylists)] = string(StatusCompleted)
	e.logger.Printf("Playlists: migrated %d, failed %d", count, failed)
	return nil
}

// phasePlaylistMedia migrates the playlist-media pivot (files_playlistmedia ->
// content_playlist_media). Batch 3.
func (e *Engine) phasePlaylistMedia(ctx context.Context) error {
	iter, err := e.source.PlaylistMedia(ctx)
	if err != nil {
		return err
	}
	defer iter.Close()

	var count, failed int64
	for iter.Next(ctx) {
		rec := iter.Item()
		count++
		e.updatePhaseProgress(count, 0)

		if e.config.DryRun {
			continue
		}

		playlistID, ok := e.mapper.Map("playlist", rec.PlaylistID)
		if !ok {
			failed++
			e.reporter.ReportError(PhasePlaylistMedia, rec.PlaylistID, fmt.Errorf("playlist %s not found", rec.PlaylistID))
			continue
		}
		mediaID, ok := e.mapper.Map("media", rec.MediaID)
		if !ok {
			failed++
			e.reporter.ReportError(PhasePlaylistMedia, rec.MediaID, fmt.Errorf("media %s not found", rec.MediaID))
			continue
		}

		builder := e.target.MediaPlaylist.Create().
			SetID(idutil.GenUUIDv7()).
			SetPlaylistID(playlistID).
			SetMediaID(mediaID).
			SetOrdering(rec.Ordering)
		if rec.ActionDate != "" {
			if t, err := parseTime(rec.ActionDate); err == nil {
				builder.SetActionDate(t)
			}
		}
		if _, err := builder.Save(ctx); err != nil {
			failed++
			e.reporter.ReportError(PhasePlaylistMedia, rec.PlaylistID, err)
			continue
		}
	}
	if iter.Err() != nil {
		return iter.Err()
	}

	e.updatePhaseProgress(count, failed)
	e.state.PhaseState[string(PhasePlaylistMedia)] = string(StatusCompleted)
	e.logger.Printf("PlaylistMedia: migrated %d, failed %d", count, failed)
	return nil
}

// phaseSubtitles migrates subtitle tracks (files_subtitle -> subtitles).
// Language code is flattened from files_language; B-side stores .vtt URLs.
// Batch 3.
func (e *Engine) phaseSubtitles(ctx context.Context) error {
	iter, err := e.source.Subtitles(ctx)
	if err != nil {
		return err
	}
	defer iter.Close()

	var count, failed int64
	for iter.Next(ctx) {
		rec := iter.Item()
		count++
		e.updatePhaseProgress(count, 0)

		if e.config.DryRun {
			continue
		}

		mediaID, ok := e.mapper.Map("media", rec.MediaID)
		if !ok {
			failed++
			e.reporter.ReportError(PhaseSubtitles, rec.MediaID, fmt.Errorf("media %s not found", rec.MediaID))
			continue
		}

		lang := rec.Language
		if lang == "" {
			lang = "und"
		}
		builder := e.target.Subtitle.Create().
			SetID(idutil.GenUUIDv7()).
			SetMediaID(mediaID).
			SetLanguage(lang).
			SetStatus("active")
		if rec.Label != "" {
			builder.SetLabel(rec.Label)
		}
		// file_url stores the target storage key (originals/subtitles/...),
		// matching the path phaseFiles copies the subtitle file to.
		if rec.FileURL != "" {
			ext := filepath.Ext(rec.FileURL)
			if ext == "" {
				ext = ".vtt"
			}
			builder.SetFileURL("originals/subtitles/" + rec.MediaID + "/" + lang + ext)
		}
		if _, err := builder.Save(ctx); err != nil {
			failed++
			e.reporter.ReportError(PhaseSubtitles, rec.MediaID, err)
			continue
		}
	}
	if iter.Err() != nil {
		return iter.Err()
	}

	e.updatePhaseProgress(count, failed)
	e.state.PhaseState[string(PhaseSubtitles)] = string(StatusCompleted)
	e.logger.Printf("Subtitles: migrated %d, failed %d", count, failed)
	return nil
}

// phaseSubscriptions migrates channel subscriptions
// (users_channel_subscribers -> user_subscriptions). Batch 3.
func (e *Engine) phaseSubscriptions(ctx context.Context) error {
	iter, err := e.source.Subscriptions(ctx)
	if err != nil {
		return err
	}
	defer iter.Close()

	var count, failed int64
	for iter.Next(ctx) {
		rec := iter.Item()
		count++
		e.updatePhaseProgress(count, 0)

		if e.config.DryRun {
			continue
		}

		channelID, ok := e.mapper.Map("channel", rec.ChannelID)
		if !ok {
			failed++
			e.reporter.ReportError(PhaseSubscriptions, rec.ChannelID, fmt.Errorf("channel %s not found", rec.ChannelID))
			continue
		}
		subscriberID, ok := e.mapper.Map("user", rec.SubscriberID)
		if !ok {
			failed++
			e.reporter.ReportError(PhaseSubscriptions, rec.SubscriberID, fmt.Errorf("subscriber %s not found", rec.SubscriberID))
			continue
		}

		if _, err := e.target.Subscription.Create().
			SetID(idutil.GenUUIDv7()).
			SetChannelID(channelID).
			SetSubscriberID(subscriberID).
			Save(ctx); err != nil {
			failed++
			e.reporter.ReportError(PhaseSubscriptions, rec.SubscriberID, err)
			continue
		}
	}
	if iter.Err() != nil {
		return iter.Err()
	}

	e.updatePhaseProgress(count, failed)
	e.state.PhaseState[string(PhaseSubscriptions)] = string(StatusCompleted)
	e.logger.Printf("Subscriptions: migrated %d, failed %d", count, failed)
	return nil
}

func init() {
	_ = entity.IsNotFound
}
