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
	"time"

	"origadmin/application/origstudio/internal/data/entity"
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

		builder := e.target.Media.Create().
			SetID(targetID).
			SetUserID(userID).
			SetTitle(rec.Title).
			SetDescription(rec.Description).
			SetType(mediaType).
			SetURL(rec.FilePath).
			SetSize(fmt.Sprintf("%d", rec.FileSize)).
			SetDuration(int(rec.Duration)).
			SetWidth(rec.Width).
			SetHeight(rec.Height).
			SetMimeType(rec.MimeType).
			SetMd5sum(rec.Checksum).
			SetState("active").
			SetEncodingStatus("done").
			SetShortToken(idutil.GenShortID()).
			SetPrivacy("PUBLIC")

		if rec.Thumbnail != "" {
			builder.SetThumbnail(rec.Thumbnail)
		}
		if rec.CategoryID != "" {
			if catID, ok := e.mapper.Map("category", rec.CategoryID); ok {
				catIDInt, _ := strconv.ParseInt(catID, 10, 64)
				builder.SetCategoryID(catIDInt)
			}
		}
		if rec.ChannelID != "" {
			if chID, ok := e.mapper.Map("channel", rec.ChannelID); ok {
				builder.SetChannelID(chID)
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

func (e *Engine) phaseComments(ctx context.Context) error {
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

		_, err := e.target.Comment.Create().
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
	}
	if iter.Err() != nil {
		return iter.Err()
	}

	e.updatePhaseProgress(count, failed)
	e.state.PhaseState[string(PhaseComments)] = string(StatusCompleted)
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

	userCount, err := e.target.User.Query().Count(ctx)
	if err != nil {
		return fmt.Errorf("verify users: %w", err)
	}
	mediaCount, err := e.target.Media.Query().Count(ctx)
	if err != nil {
		return fmt.Errorf("verify media: %w", err)
	}
	categoryCount, err := e.target.Category.Query().Count(ctx)
	if err != nil {
		return fmt.Errorf("verify categories: %w", err)
	}

	e.logger.Printf("Verification: %d users, %d media, %d categories in target", userCount, mediaCount, categoryCount)
	e.state.PhaseState[string(PhaseVerify)] = string(StatusCompleted)
	return nil
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

func init() {
	_ = entity.IsNotFound
}
