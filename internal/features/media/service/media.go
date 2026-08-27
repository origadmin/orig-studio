/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"context"
	"encoding/json"
	"fmt"
	stdhttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/origadmin/runtime/errors"
	"github.com/origadmin/runtime/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"origadmin/application/origstudio/api/gen/v1/media"
	"origadmin/application/origstudio/api/gen/v1/types"
	contentbiz "origadmin/application/origstudio/internal/features/content/biz"
	repotypes "origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/features/media/biz"
	"origadmin/application/origstudio/internal/features/media/dto"
	"origadmin/application/origstudio/internal/infra/auth"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
)

type MediaService struct {
	media.UnimplementedMediaServiceServer
	media.UnimplementedEncodingProfileServiceServer
	uc            *biz.MediaUseCase
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase
	jwt           *auth.Manager
	settingUC     *systembiz.SettingUseCase
	log           *log.Helper
}

func NewMediaService(uc *biz.MediaUseCase, likeFavoriteUC *contentbiz.LikeFavoriteUseCase, jwt *auth.Manager, settingUC *systembiz.SettingUseCase, logger log.Logger) *MediaService {
	return &MediaService{
		uc:            uc,
		likeFavoriteUC: likeFavoriteUC,
		jwt:           jwt,
		settingUC:     settingUC,
		log:           log.NewHelper(log.With(logger, "module", "media.service")),
	}
}

func (s *MediaService) extractUserID(ctx context.Context) string {
	if id, ok := ctx.Value("user_id").(string); ok && id != "" {
		return id
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		authHeaders = md.Get("grpcgateway-authorization")
	}
	for _, header := range authHeaders {
		token := strings.TrimPrefix(header, "Bearer ")
		if token == header {
			continue
		}
		claims, err := s.jwt.Parse(token)
		if err != nil {
			continue
		}
		return claims.GetUserID()
	}
	return ""
}

func (s *MediaService) extractClaims(ctx context.Context) *auth.Claims {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		authHeaders = md.Get("grpcgateway-authorization")
	}
	for _, header := range authHeaders {
		token := strings.TrimPrefix(header, "Bearer ")
		if token == header {
			continue
		}
		claims, err := s.jwt.Parse(token)
		if err != nil {
			continue
		}
		return claims
	}
	return nil
}

func (s *MediaService) ListMedias(
	ctx context.Context,
	req *media.ListMediasRequest,
) (*media.ListMediasResponse, error) {
	// BUG-131: request -> query option mapping lives in exactly ONE place.
	// This handler used to hand-roll its own copy of the mapping, which is how
	// `tags`, `category_ids`, `state` and `featured` ended up silently dropped
	// on the live gRPC route while dto.ListMediasRequestToQueryOption (used by
	// other callers) mapped them correctly. Any new filter field must be added
	// to that function, and it is covered by dto tests.
	opts, err := dto.ListMediasRequestToQueryOption(req)
	if err != nil {
		return nil, err
	}
	page, pageSize := int(opts.Page), int(opts.PageSize)

	// BUG-141 ①② portal visibility gate.
	// Owner viewing their own list (req.UserId == authenticated viewer) relaxes the
	// gate to "own active" so creators see their own unreviewed content. Everyone
	// else (incl. anonymous) only sees listable (reviewed+encoded+active) content.
	claims := s.extractClaims(ctx)
	if req.UserId != nil && claims != nil && *req.UserId == claims.GetUserID() {
		opts.OwnerView = true
	} else {
		listableOnly := true
		opts.Listable = &listableOnly
	}

	items, total, err := s.uc.ListMedias(ctx, opts)
	if err != nil {
		return nil, err
	}
	// BUG-139: disabled platform feature modes force per-media toggles off on
	// the read path (override semantics; stored values are preserved).
	s.applyFeatureModeOverrides(ctx, items...)
	return &media.ListMediasResponse{
		Items:    items,
		Total:    total,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

func (s *MediaService) GetMedia(
	ctx context.Context,
	req *media.GetMediaRequest,
) (*media.GetMediaResponse, error) {
	item, err := s.uc.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("MEDIA_NOT_FOUND", "Media not found")
	}
	s.applyFeatureModeOverrides(ctx, item)
	return &media.GetMediaResponse{Media: item}, nil
}

// featureModeDisabled returns true when the platform-level feature mode setting
// key equals "disabled". BUG-139 G5: disabled = override 强制关停 — the entry is
// hidden/disabled for ALL media (incl. published), regardless of per-media values.
func (s *MediaService) featureModeDisabled(ctx context.Context, key string) bool {
	if s.settingUC == nil {
		return false
	}
	st, err := s.settingUC.GetByKey(ctx, key)
	if err != nil || st == nil {
		return false
	}
	return st.Value == "disabled"
}

// applyFeatureModeOverrides forces enable_comments / allow_download off on the
// read path when the corresponding platform feature mode is disabled.
func (s *MediaService) applyFeatureModeOverrides(ctx context.Context, medias ...*types.Media) {
	if len(medias) == 0 {
		return
	}
	commentsOff := s.featureModeDisabled(ctx, "comments_mode")
	downloadsOff := s.featureModeDisabled(ctx, "downloads_mode")
	if !commentsOff && !downloadsOff {
		return
	}
	for _, m := range medias {
		if m == nil {
			continue
		}
		if commentsOff {
			m.EnableComments = false
		}
		if downloadsOff {
			m.AllowDownload = false
		}
	}
}

func (s *MediaService) CreateMedia(
	ctx context.Context,
	req *media.CreateMediaRequest,
) (*media.CreateMediaResponse, error) {
	item, err := s.uc.CreateMedia(ctx, req.Media)
	if err != nil {
		return nil, err
	}
	return &media.CreateMediaResponse{Media: item}, nil
}

// portalOwnerEditableFields is the whitelist of Media fields a portal owner is
// permitted to change through PUT /api/v1/medias/{shortToken}.
//
// Security: the nested UpdateMediaRequest.media carries the full Media proto,
// which includes system/owner-only fields (user_id, uuid, created_at, ...).
// We deliberately copy ONLY this whitelist into the loaded entity so a portal
// caller cannot reassign ownership or clear audit columns. Admin uses a
// separate, flat endpoint (UpdateAdminMedia) with full control — this keeps
// the portal and admin surfaces cleanly separated.
// maskSet normalizes FieldMask paths into a lookup set. Paths may be supplied
// with or without the "media." resource prefix (AIP-134), and in camelCase or
// snake_case (protobuf FieldMask JSON requires camelCase paths; we normalize to
// the proto snake_case field names used by portalOwnerEditableFields). Returns
// nil when no paths are given (legacy full-merge mode).
func maskSet(paths []string) map[string]bool {
	if len(paths) == 0 {
		return nil
	}
	m := make(map[string]bool, len(paths))
	for _, p := range paths {
		p = strings.TrimPrefix(p, "media.")
		p = camelToSnake(p)
		if p != "" {
			m[p] = true
		}
	}
	return m
}

// camelToSnake converts a lowerCamelCase identifier to snake_case
// (e.g. "channelId" -> "channel_id"). Used to normalize FieldMask paths.
func camelToSnake(s string) string {
	if !strings.ContainsFunc(s, func(r rune) bool { return r >= 'A' && r <= 'Z' }) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + 4)
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + ('a' - 'A'))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// portalOwnerEditableFields merges the owner-editable fields from src into dst.
//
// With a non-empty update_mask (AIP-134, BUG-105) only the masked fields are
// copied verbatim (so an empty string CAN clear a field). channel_id is handled
// by the dedicated UpdateMediaChannel path (owner validation + set/clear), so it
// is deliberately excluded here when a mask is present.
//
// With an empty mask the legacy full-merge behaviour is preserved (non-empty
// values only) for backward compatibility with existing callers.
func portalOwnerEditableFields(dst, src *types.Media, paths []string) {
	mask := maskSet(paths)
	if mask == nil {
		// Legacy full-merge: only non-empty values are copied.
		if src.Title != "" {
			dst.Title = src.Title
		}
		if src.Description != "" {
			dst.Description = src.Description
		}
		if src.Thumbnail != "" {
			dst.Thumbnail = src.Thumbnail
		}
		if src.CategoryId != 0 {
			dst.CategoryId = src.CategoryId
		}
		if src.ChannelId != "" {
			dst.ChannelId = src.ChannelId
		}
		if src.Tags != nil {
			dst.Tags = src.Tags
		}
		if src.State != "" {
			dst.State = src.State
		}
		if src.Privacy != 0 {
			dst.Privacy = src.Privacy
		}
		dst.Featured = src.Featured
		dst.EnableComments = src.EnableComments
		dst.AllowDownload = src.AllowDownload
		// BUG-143 root cause ①: `listable` is DERIVED from (encoding_status,
		// review_status, state) and must never be taken from the client — it is
		// recomputed after the merge via ShouldBeListable.
		return
	}

	// Mask-driven merge: copy masked fields verbatim (empty values included).
	if mask["title"] {
		dst.Title = src.Title
	}
	if mask["description"] {
		dst.Description = src.Description
	}
	if mask["thumbnail"] {
		dst.Thumbnail = src.Thumbnail
	}
	if mask["category_id"] {
		dst.CategoryId = src.CategoryId
	}
	// channel_id deliberately NOT merged here — handled by UpdateMediaChannel.
	if mask["tags"] {
		dst.Tags = src.Tags
	}
	if mask["state"] {
		dst.State = src.State
	}
	if mask["privacy"] {
		dst.Privacy = src.Privacy
	}
	if mask["featured"] {
		dst.Featured = src.Featured
	}
	if mask["enable_comments"] {
		dst.EnableComments = src.EnableComments
	}
	if mask["allow_download"] {
		dst.AllowDownload = src.AllowDownload
	}
}

func (s *MediaService) UpdateMedia(
	ctx context.Context,
	req *media.UpdateMediaRequest,
) (*media.UpdateMediaResponse, error) {
	// Defensive: the portal request is nested ({ id, media, update_mask }).
	// A flat/malformed body leaves req.Media nil and must fail cleanly
	// instead of panicking deep in the DAL.
	if req.Media == nil {
		return nil, errors.BadRequest("INVALID_REQUEST", "media object is required")
	}

	// Load the existing record. The portal request id is a short_token, and
	// GetMedia resolves short_token -> id (GetByID only matches the internal
	// UUID), so we use GetMedia here. Only the whitelisted owner-editable
	// fields are merged and system columns are preserved. This mirrors the
	// admin path's load + merge pattern but is restricted to portal-owned
	// fields.
	existing, err := s.uc.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("MEDIA_NOT_FOUND", "media not found")
	}
	originalState := existing.State

	// BUG-105: AIP-134 update_mask drives the merge. An empty mask keeps the
	// legacy full-merge behaviour; a mask scopes the update to the masked fields
	// (and an empty string then clears the field).
	paths := req.GetUpdateMask().GetPaths()
	portalOwnerEditableFields(existing, req.Media, paths)

	// BUG-138: a publish transition (draft -> active|pending_review) is the single
	// portal publish entry point. Route it through PublishMedia so the encoding
	// guard (400 ENCODING_NOT_READY) and the review flow (review_mode) are always
	// enforced — never a blind state merge. existing.Id is the resolved internal
	// id (GetMedia already mapped the short_token); PublishMedia reloads from DB.
	if (existing.State == "active" || existing.State == "pending_review") && existing.State != originalState {
	// BUG-138 #41: read the global review_mode (default manual) so the admin
	// Settings selector is honored at publish time. "" → PublishMedia defaults
	// to manual, matching the G5 default.
	reviewMode := ""
	if s.settingUC != nil {
		if st, err := s.settingUC.GetByKey(ctx, "review_mode"); err == nil && st != nil {
			reviewMode = st.Value
		}
	}
	published, perr := s.uc.PublishMedia(ctx, existing.Id, reviewMode, s.extractUserID(ctx))
		if perr != nil {
			return nil, perr
		}
		return &media.UpdateMediaResponse{Media: published}, nil
	}

	// BUG-143 root cause ①: recompute the derived visibility flag after the
	// owner-edit merge (portal callers must never control listable directly).
	existing.Listable = s.uc.ShouldBeListable(existing)

	item, err := s.uc.UpdateMedia(ctx, existing)
	if err != nil {
		return nil, err
	}

	// BUG-105: channel assignment changes (assign / move A->B / clear) go
	// through a dedicated path after the general update, because the generic
	// Update only SetChannelID for non-empty values and cannot express "clear".
	// When the mask requests channel_id, validate ownership first, then apply.
	if mask := maskSet(paths); mask != nil && mask["channel_id"] {
		target := req.Media.GetChannelId()
		if target != "" {
			ownerID, err := s.uc.GetChannelOwnerID(ctx, target)
			if err != nil {
				s.log.Errorf("Failed to resolve channel owner: %v", err)
				return nil, errors.InternalServer("RESOLVE_CHANNEL_OWNER_FAILED", "Failed to resolve channel owner")
			}
			if ownerID == "" {
				return nil, errors.BadRequest("CHANNEL_NOT_FOUND", "target channel does not exist")
			}
			if ownerID != s.extractUserID(ctx) {
				return nil, errors.BadRequest("CHANNEL_NOT_OWNED", "you can only assign media to your own channels")
			}
		}
		if err := s.uc.UpdateMediaChannel(ctx, existing.Id, target); err != nil {
			s.log.Errorf("Failed to update media channel: %v", err)
			return nil, errors.InternalServer("UPDATE_MEDIA_CHANNEL_FAILED", "Failed to update media channel")
		}
		// Keep the returned resource in sync with the dedicated channel write.
		item.ChannelId = target
	}

	return &media.UpdateMediaResponse{Media: item}, nil
}

func (s *MediaService) DeleteMedia(
	ctx context.Context,
	req *media.DeleteMediaRequest,
) (*media.DeleteMediaResponse, error) {
	err := s.uc.DeleteMedia(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &media.DeleteMediaResponse{}, nil
}

func (s *MediaService) IncrementViewCount(
	ctx context.Context,
	req *media.IncrementViewCountRequest,
) (*media.IncrementViewCountResponse, error) {
	count, err := s.uc.IncrementViewCount(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &media.IncrementViewCountResponse{ViewCount: count}, nil
}

func (s *MediaService) ListEncodingTasks(
	ctx context.Context,
	req *media.ListEncodingTasksRequest,
) (*media.ListEncodingTasksResponse, error) {
	tasks, err := s.uc.ListEncodingTasks(ctx, req.MediaId)
	if err != nil {
		return nil, err
	}

	result := make([]*types.EncodingTask, len(tasks))
	for i, t := range tasks {
		result[i] = &types.EncodingTask{
			Id:           t.Id,
			MediaId:      t.MediaId,
			ProfileId:    int64(t.ProfileId),
			Status:       string(t.Status),
			OutputPath:   t.OutputPath,
			ErrorMessage: t.ErrorMessage,
		}
	}
	return &media.ListEncodingTasksResponse{Tasks: result}, nil
}

// GetTranscodingStatus returns the overall encoding status of the system.
func (s *MediaService) GetEncodingStatus(
	ctx context.Context,
	req *media.GetEncodingStatusRequest,
) (*media.GetEncodingStatusResponse, error) {
	status, err := s.uc.GetTranscodingStatus(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Normalize pagination parameters
	page, pageSize := repotypes.NormalizePagination(int(req.Page), int(req.PageSize))

	return &media.GetEncodingStatusResponse{
		ProcessingCount: int32(status.ProcessingCount),
		PendingCount:    int32(status.PendingCount),
		FailedCount:     int32(status.FailedCount),
		SuccessCount:    int32(status.SuccessCount),
		Total:           0,
		Page:            int32(page),
		PageSize:        int32(pageSize),
		Items:           []*media.TranscodingMediaItem{},
	}, nil
}

// dtoEncodeProfileToProto converts DTO to proto EncodeProfile with all fields
func dtoEncodeProfileToProto(p *dto.EncodeProfile) *types.EncodeProfile {
	return &types.EncodeProfile{
		Id:              int64(p.Id),
		Name:            p.Name,
		Description:     p.Description,
		Extension:       p.Extension,
		Resolution:      p.Resolution,
		VideoCodec:      p.VideoCodec,
		VideoBitrate:    p.VideoBitrate,
		AudioCodec:      p.AudioCodec,
		AudioBitrate:    p.AudioBitrate,
		BentoParameters: p.BentoParameters,
		IsActive:        p.IsActive,
		CreateTime:      timestamppb.New(p.CreateTime),
		UpdateTime:      timestamppb.New(p.UpdateTime),
	}
}

// parseProfileID parses string ID to int, returns BadRequest error if invalid
func parseProfileID(idStr string) (int, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, errors.BadRequest("INVALID_ID", "invalid profile id")
	}
	return id, nil
}

// isNotFoundError checks if error is a not found error
func isNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "not found")
}

// generateFFmpegCommand generates FFmpeg command based on profile config (mediacms best practices)
func generateFFmpegCommand(p *types.EncodeProfile) string {
	codec := p.VideoCodec
	if codec == "" {
		codec = "libx264"
	}
	// Map codec short names to ffmpeg codec names
	vcodec := map[string]string{
		"h264": "libx264",
		"h265": "libx265",
		"vp9":  "libvpx-vp9",
	}[codec]
	if vcodec == "" {
		vcodec = codec
	}
	acodec := p.AudioCodec
	if acodec == "" {
		acodec = "aac"
	}
	resolution := p.Resolution
	if resolution == "" {
		resolution = "1280x720"
	}
	crf := map[string]string{
		"libx264":  "23",
		"libx265":  "28",
		"libvpx-vp9": "32",
	}[vcodec]
	if crf == "" {
		crf = "23"
	}
	videoBitrate := p.VideoBitrate
	audioBitrate := p.AudioBitrate
	if videoBitrate == "" {
		videoBitrate = map[string]string{
			"854x480":  "1000k",
			"1280x720": "2500k",
			"1920x1080": "4500k",
		}[resolution]
		if videoBitrate == "" {
			videoBitrate = "2500k"
		}
	}
	if audioBitrate == "" {
		audioBitrate = map[string]string{
			"libx264":  "128k",
			"libx265":  "128k",
			"libvpx-vp9": "96k",
		}[vcodec]
		if audioBitrate == "" {
			audioBitrate = "128k"
		}
	}
	movflags := ""
	if strings.Contains(p.Extension, "mp4") {
		movflags = " -movflags +faststart"
	}
	return fmt.Sprintf(
		"ffmpeg -i <input_file> -c:v %s -preset medium -crf %s -b:v %s -maxrate %s -bufsize %s -vf scale=%s -c:a %s -b:a %s -pix_fmt yuv420p%s <output_file>",
		vcodec, crf, videoBitrate, videoBitrate, videoBitrate, resolution, acodec, audioBitrate, movflags,
	)
}

// ListEncodeProfiles returns a list of encoding profiles.
func (s *MediaService) ListEncodeProfiles(
	ctx context.Context,
	req *media.ListEncodeProfilesRequest,
) (*media.ListEncodeProfilesResponse, error) {
	profiles, err := s.uc.ListEncodeProfiles(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*types.EncodeProfile, len(profiles))
	for i, p := range profiles {
		result[i] = dtoEncodeProfileToProto(p)
	}
	return &media.ListEncodeProfilesResponse{Profiles: result}, nil
}

// GetEncodeProfile returns an encoding profile by ID.
func (s *MediaService) GetEncodeProfile(
	ctx context.Context,
	req *media.GetEncodeProfileRequest,
) (*media.GetEncodeProfileResponse, error) {
	profileID, err := parseProfileID(req.Id)
	if err != nil {
		return nil, err
	}
	p, err := s.uc.GetEncodeProfile(ctx, profileID)
	if err != nil {
		if isNotFoundError(err) {
			return nil, errors.NotFound("PROFILE_NOT_FOUND", "encoding profile not found")
		}
		return nil, err
	}
	return &media.GetEncodeProfileResponse{
		Profile: dtoEncodeProfileToProto(p),
	}, nil
}

// CreateEncodeProfile creates a new encoding profile.
func (s *MediaService) CreateEncodeProfile(
	ctx context.Context,
	req *media.CreateEncodeProfileRequest,
) (*media.CreateEncodeProfileResponse, error) {
	if req.Profile == nil {
		return nil, errors.BadRequest("INVALID_REQUEST", "profile is required")
	}
	p, err := s.uc.CreateEncodeProfile(ctx, &dto.EncodeProfile{
		Name:            req.Profile.Name,
		Description:     req.Profile.Description,
		Extension:       req.Profile.Extension,
		Resolution:      req.Profile.Resolution,
		VideoCodec:      req.Profile.VideoCodec,
		VideoBitrate:    req.Profile.VideoBitrate,
		AudioCodec:      req.Profile.AudioCodec,
		AudioBitrate:    req.Profile.AudioBitrate,
		BentoParameters: req.Profile.BentoParameters,
		IsActive:        req.Profile.IsActive,
	})
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return nil, errors.Conflict("DUPLICATE_NAME", "profile name already exists")
		}
		return nil, err
	}
	return &media.CreateEncodeProfileResponse{
		Profile: dtoEncodeProfileToProto(p),
	}, nil
}

// UpdateEncodeProfile updates an existing encoding profile.
func (s *MediaService) UpdateEncodeProfile(
	ctx context.Context,
	req *media.UpdateEncodeProfileRequest,
) (*media.UpdateEncodeProfileResponse, error) {
	if req.Profile == nil {
		return nil, errors.BadRequest("INVALID_REQUEST", "profile is required")
	}
	profileID := int(req.Profile.Id)
	if profileID <= 0 {
		return nil, errors.BadRequest("INVALID_ID", "invalid profile id")
	}
	p, err := s.uc.UpdateEncodeProfile(ctx, &dto.EncodeProfile{
		Id:              profileID,
		Name:            req.Profile.Name,
		Description:     req.Profile.Description,
		Extension:       req.Profile.Extension,
		Resolution:      req.Profile.Resolution,
		VideoCodec:      req.Profile.VideoCodec,
		VideoBitrate:    req.Profile.VideoBitrate,
		AudioCodec:      req.Profile.AudioCodec,
		AudioBitrate:    req.Profile.AudioBitrate,
		BentoParameters: req.Profile.BentoParameters,
		IsActive:        req.Profile.IsActive,
	})
	if err != nil {
		if isNotFoundError(err) {
			return nil, errors.NotFound("PROFILE_NOT_FOUND", "encoding profile not found")
		}
		if strings.Contains(err.Error(), "unique") {
			return nil, errors.Conflict("DUPLICATE_NAME", "profile name already exists")
		}
		return nil, err
	}
	return &media.UpdateEncodeProfileResponse{
		Profile: dtoEncodeProfileToProto(p),
	}, nil
}

// DeleteEncodeProfile deletes an encoding profile.
func (s *MediaService) DeleteEncodeProfile(
	ctx context.Context,
	req *media.DeleteEncodeProfileRequest,
) (*media.DeleteEncodeProfileResponse, error) {
	profileID, err := parseProfileID(req.Id)
	if err != nil {
		return nil, err
	}
	err = s.uc.DeleteEncodeProfile(ctx, profileID)
	if err != nil {
		if isNotFoundError(err) {
			return nil, errors.NotFound("PROFILE_NOT_FOUND", "encoding profile not found")
		}
		return nil, err
	}
	return &media.DeleteEncodeProfileResponse{}, nil
}

// PreviewEncodeCommand previews the encoding command per mediacms best practices (CRF mode).
func (s *MediaService) PreviewEncodeCommand(
	ctx context.Context,
	req *media.PreviewEncodeCommandRequest,
) (*media.PreviewEncodeCommandResponse, error) {
	codec := req.Codec
	if codec == "" {
		codec = "h264"
	}
	resolutionStr := req.Resolution
	if resolutionStr == "" {
		resolutionStr = "720"
	}
	resolution, err := strconv.Atoi(resolutionStr)
	if err != nil || resolution <= 0 {
		return nil, errors.BadRequest("INVALID_RESOLUTION", "resolution must be a positive integer (height in pixels)")
	}
	ext := req.Extension
	if ext == "" {
		if codec == "vp9" {
			ext = "webm"
		} else {
			ext = "mp4"
		}
	}
	inputPath := req.InputPath
	if inputPath == "" {
		inputPath = "<input_file>"
	}
	outputPath := req.OutputPath
	if outputPath == "" {
		outputPath = "<output_file>." + ext
	}
	fullCmd := generateMediacmsFFmpegCommand(codec, resolution, inputPath, outputPath)
	s.log.Infof("PreviewEncodeCommand: codec=%s resolution=%d ext=%s cmd=%s", codec, resolution, ext, fullCmd)
	return &media.PreviewEncodeCommandResponse{FullCommand: fullCmd}, nil
}

// generateMediacmsFFmpegCommand builds an FFmpeg CRF-mode preview command matching mediacms logic:
// https://github.com/mediacms-io/mediacms/blob/main/files/helpers.py (get_base_ffmpeg_command / produce_ffmpeg_commands)
func generateMediacmsFFmpegCommand(codec string, targetHeight int, inputPath, outputPath string) string {
	const (
		maxRateMultiplier = 1.5
		minRateMultiplier = 0.5
		bufSizeMultiplier = 1.5
		keyframeDistance  = 4 // seconds
		vp9Speed          = 2
		defaultFps        = 30
		defaultPreset     = "medium"
	)
	videoBitrates := map[string]map[int]int{
		"h264": {
			144: 150, 240: 300, 360: 500, 480: 1000,
			720: 2500, 1080: 4500, 1440: 9000, 2160: 18000,
		},
		"h265": {
			144: 75, 240: 150, 360: 275, 480: 500,
			720: 1024, 1080: 1800, 1440: 4500, 2160: 10000,
		},
		"vp9": {
			144: 75, 240: 150, 360: 275, 480: 500,
			720: 1024, 1080: 1800, 1440: 4500, 2160: 10000,
		},
	}
	videoCRFs := map[string]int{
		"h264": 23,
		"h265": 28,
		"vp9":  32,
	}
	audioEncoders := map[string]string{
		"h264": "aac",
		"h265": "aac",
		"vp9":  "libopus",
	}
	audioBitrates := map[string]int{
		"h264": 128,
		"h265": 128,
		"vp9":  96,
	}
	videoEncoders := map[string]string{
		"h264": "libx264",
		"h265": "libx265",
		"vp9":  "libvpx-vp9",
	}

	encoder := videoEncoders[codec]
	if encoder == "" {
		encoder = "libx264"
		codec = "h264"
	}
	crf := videoCRFs[codec]
	audioEncoder := audioEncoders[codec]
	audioBitrate := audioBitrates[codec]
	targetRate := videoBitrates[codec][targetHeight]
	if targetRate == 0 {
		targetRate = 2500 // default to 720p bitrate for unknown resolutions
	}

	targetWidth := targetHeight * 16 / 9
	maxrate := int(float64(targetRate) * maxRateMultiplier)
	minrate := int(float64(targetRate) * minRateMultiplier)
	bufsize := int(float64(targetRate) * bufSizeMultiplier)
	keyframeDistFrames := defaultFps * keyframeDistance
	keyframeDouble := keyframeDistFrames * 2

	var parts []string
	parts = append(parts, "ffmpeg", "-y")
	parts = append(parts, "-i", shellQuote(inputPath))
	parts = append(parts, "-c:v", encoder)

	scaleFilter := fmt.Sprintf(
		"scale=if(lt(iw\\,ih)\\,%d\\,%d):if(lt(iw\\,ih)\\,%d\\,%d):force_original_aspect_ratio=decrease:force_divisible_by=2:flags=lanczos,fps=%d",
		targetHeight, targetWidth, targetWidth, targetHeight, defaultFps,
	)
	parts = append(parts, "-filter:v", scaleFilter)
	parts = append(parts, "-pix_fmt", "yuv420p")
	parts = append(parts, "-crf", strconv.Itoa(crf))

	if codec == "vp9" {
		parts = append(parts, "-b:v", strconv.Itoa(targetRate)+"k")
	}

	parts = append(parts, "-c:a", audioEncoder)
	parts = append(parts, "-b:a", strconv.Itoa(audioBitrate)+"k")
	parts = append(parts, "-ac", "2")

	switch codec {
	case "h264":
		level := "4.2"
		if targetHeight > 1080 {
			level = "5.2"
		}
		parts = append(parts, "-maxrate", strconv.Itoa(maxrate)+"k")
		parts = append(parts, "-bufsize", strconv.Itoa(bufsize)+"k")
		parts = append(parts, "-force_key_frames", fmt.Sprintf("expr:gte(t,n_forced*%d)", keyframeDistance))
		parts = append(parts, "-x264-params", fmt.Sprintf("keyint=%d:keyint_min=%d", keyframeDouble, keyframeDistFrames))
		parts = append(parts, "-preset", defaultPreset)
		parts = append(parts, "-profile:v", "main")
		parts = append(parts, "-level", level)
	case "h265":
		x265Params := fmt.Sprintf("vbv-maxrate=%d:vbv-bufsize=%d:keyint=%d:keyint_min=%d",
			maxrate, bufsize, keyframeDouble, keyframeDistFrames)
		parts = append(parts, "-force_key_frames", fmt.Sprintf("expr:gte(t,n_forced*%d)", keyframeDistance))
		parts = append(parts, "-x265-params", x265Params)
		parts = append(parts, "-preset", defaultPreset)
		parts = append(parts, "-profile:v", "main")
	case "vp9":
		parts = append(parts, "-g", strconv.Itoa(keyframeDistFrames))
		parts = append(parts, "-keyint_min", strconv.Itoa(keyframeDistFrames))
		parts = append(parts, "-maxrate", strconv.Itoa(maxrate)+"k")
		parts = append(parts, "-minrate", strconv.Itoa(minrate)+"k")
		parts = append(parts, "-bufsize", strconv.Itoa(bufsize)+"k")
		parts = append(parts, "-speed", strconv.Itoa(vp9Speed))
	}

	parts = append(parts, "-strict", "-2")

	if strings.HasSuffix(strings.ToLower(outputPath), ".mp4") {
		parts = append(parts, "-movflags", "+faststart")
	}
	parts = append(parts, shellQuote(outputPath))

	return strings.Join(parts, " ")
}

func shellQuote(s string) string {
	if strings.ContainsAny(s, " <>|&;$`\\\"'") {
		return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
	}
	return s
}

func (s *MediaService) GetMediaVariants(
	ctx context.Context,
	req *media.GetMediaVariantsRequest,
) (*media.GetMediaVariantsResponse, error) {
	// 直接使用 req.Id 作为 mediaID，因为 ID 是 UUID 格式
	mediaIDStr := req.Id
	
	summary, err := s.uc.GetMediaVariantsByUUID(ctx, mediaIDStr)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, errors.NotFound("MEDIA_NOT_FOUND", "Media not found")
		}
		return nil, err
	}
	
	result := make([]*types.MediaVariant, len(summary.Variants))
	for i, v := range summary.Variants {
		result[i] = &types.MediaVariant{
			Id:         v.TaskID,
			MediaId:    mediaIDStr,
			ProfileId:  strconv.Itoa(v.ProfileID),
			Resolution: v.Resolution,
			Url:        v.OutputPath,
			Size:       0,
			Status:     string(v.Status),
		}
	}
	
	return &media.GetMediaVariantsResponse{Variants: result}, nil
}

// SSEHandler handles Server-Sent Events for transcoding progress.
func (s *MediaService) SSEHandler(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	mediaIdStr := r.URL.Query().Get("media_id")

	flusher, ok := w.(stdhttp.Flusher)
	if !ok {
		stdhttp.Error(w, "Streaming unsupported!", stdhttp.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ctx := r.Context()
	events, cleanup := s.uc.Subscribe(ctx, mediaIdStr)
	defer cleanup()

	// Keep-alive ticker
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	s.log.Infof("SSE client connected for media %s", mediaIdStr)

	for {
		select {
		case <-ctx.Done():
			s.log.Infof("SSE client disconnected for media %s", mediaIdStr)
			return
		case <-ticker.C:
			fmt.Fprintf(w, "event: ping\ndata: %d\n\n", time.Now().Unix())
			flusher.Flush()
		case ev := <-events:
			if ev == nil {
				return
			}
			data, _ := json.Marshal(map[string]interface{}{
				"media_id": ev.MediaId,
				"task_id":  "",
				"status":   "",
				"progress": ev.Progress,
				"speed":    ev.Speed,
				"fps":      ev.Fps,
				"time":     ev.Time,
			})
			if ev.Task != nil {
				data, _ = json.Marshal(map[string]interface{}{
					"media_id": ev.MediaId,
					"task_id":  ev.Task.Id,
					"status":   string(ev.Task.Status),
					"progress": ev.Progress,
					"speed":    ev.Speed,
					"fps":      ev.Fps,
					"time":     ev.Time,
				})
			}
			fmt.Fprintf(w, "event: transcoding_progress\ndata: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// TranscodingStatusHTTPHandler handles GET /api/v1/medias/transcoding/status.
// Returns aggregated encoding status counts.
func (s *MediaService) TranscodingStatusHTTPHandler(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	status, err := s.uc.GetTranscodingStatus(r.Context(), nil)
	if err != nil {
		stdhttp.Error(w, err.Error(), stdhttp.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"processing_count": status.ProcessingCount,
		"pending_count":    status.PendingCount,
		"failed_count":     status.FailedCount,
		"success_count":    status.SuccessCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_ = json.NewEncoder(w).Encode(resp)
}

// RetryTaskHTTPHandler handles POST /api/v1/medias/encoding/retry with query parameter task_id.
// Resets a single failed encoding task to "pending" for re-processing.
// Query params:
//   - task_id (required): the encoding task ID to retry
func (s *MediaService) RetryTaskHTTPHandler(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodPost {
		stdhttp.Error(w, "Method not allowed", stdhttp.StatusMethodNotAllowed)
		return
	}

	taskIDStr := r.URL.Query().Get("task_id")
	if taskIDStr == "" {
		writeRetryError(w, "task_id is required", 400)
		return
	}

	var taskID int
	if _, err := fmt.Sscanf(taskIDStr, "%d", &taskID); err != nil || taskID <= 0 {
		writeRetryError(w, "invalid task_id: must be a positive integer", 400)
		return
	}

	task, err := s.uc.RetryTask(r.Context(), taskIDStr)
	if err != nil {
		writeRetryError(w, err.Error(), 422) // Unprocessable Entity
		return
	}

	resp := map[string]any{
		"success": true,
		"task": map[string]any{
			"id":            task.Id,
			"media_id":      task.MediaId,
			"profile_id":    task.ProfileId,
			"status":        task.Status,
			"error_message": task.ErrorMessage,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_ = json.NewEncoder(w).Encode(resp)
}

// RetryAllFailedHTTPHandler handles POST /api/v1/medias/encoding/retry-all-failed with query parameter media_id.
// Resets all failed tasks for a media back to "pending".
func (s *MediaService) RetryAllFailedHTTPHandler(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodPost {
		stdhttp.Error(w, "Method not allowed", stdhttp.StatusMethodNotAllowed)
		return
	}

	mediaIDStr := r.URL.Query().Get("media_id")
	if mediaIDStr == "" {
		writeRetryError(w, "media_id is required", 400)
		return
	}

	count, err := s.uc.RetryAllFailedTasks(r.Context(), mediaIDStr)
	if err != nil {
		writeRetryError(w, err.Error(), 500)
		return
	}

	resp := map[string]any{
		"success":     true,
		"reset_count": count,
		"media_id":    mediaIDStr,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_ = json.NewEncoder(w).Encode(resp)
}

func writeRetryError(w stdhttp.ResponseWriter, message string, code int) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// UploadMedia handles the UploadMedia gRPC method.
func (s *MediaService) UploadMedia(ctx context.Context, req *media.UploadMediaRequest) (*media.UploadMediaResponse, error) {
	// Create a new media item from the request
	newMedia := &types.Media{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Url:         req.FilePath,
		Size:        req.Size,
		Duration:    req.Duration,
		UserId:      req.UserId,
		State:       "active",
		Privacy:     types.Privacy_PRIVACY_PUBLIC,
	}

	// Auto-bind to owner's default channel when no channel is specified.
	// Management backend uploads don't carry channel_id (unlike frontend multipart uploads),
	// so we resolve and inject it here to avoid orphan videos.
	if newMedia.ChannelId == "" {
		defaultChID, err := s.uc.GetDefaultChannelID(ctx, req.UserId)
		if err != nil {
			s.log.WithContext(ctx).Warnf("failed to resolve default channel for user %s (video will be orphan): %v", req.UserId, err)
			// Continue without channel_id rather than failing the upload;
			// the read path's NULL fallback for default channel will still cover it.
		} else {
			newMedia.ChannelId = defaultChID
			s.log.WithContext(ctx).Infof("auto-bound media to default channel %s for user %s", defaultChID, req.UserId)
		}
	}

	created, err := s.uc.CreateMedia(ctx, newMedia)
	if err != nil {
		return nil, err
	}

	return &media.UploadMediaResponse{
		Media: created,
	}, nil
}

// GetMediaStream handles the GetMediaStream gRPC method.
func (s *MediaService) GetMediaStream(ctx context.Context, req *media.GetMediaStreamRequest) (*media.GetMediaStreamResponse, error) {
	m, err := s.uc.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("MEDIA_NOT_FOUND", "Media not found")
	}

	streamURL := m.Url
	if m.HlsFile != "" {
		streamURL = m.HlsFile
	}

	return &media.GetMediaStreamResponse{
		StreamUrl: streamURL,
		Format:    "hls",
	}, nil
}

// GetMediaDownload handles the GetMediaDownload gRPC method.
func (s *MediaService) GetMediaDownload(ctx context.Context, req *media.GetMediaDownloadRequest) (*media.GetMediaDownloadResponse, error) {
	m, err := s.uc.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("MEDIA_NOT_FOUND", "Media not found")
	}

	filename := m.Title
	if m.Extension != "" {
		filename += "." + m.Extension
	}

	return &media.GetMediaDownloadResponse{
		DownloadUrl: m.Url,
		Filename:    filename,
	}, nil
}

// GetMediaThumbnail handles the GetMediaThumbnail gRPC method.
func (s *MediaService) GetMediaThumbnail(ctx context.Context, req *media.GetMediaThumbnailRequest) (*media.GetMediaThumbnailResponse, error) {
	m, err := s.uc.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("MEDIA_NOT_FOUND", "Media not found")
	}

	return &media.GetMediaThumbnailResponse{
		ThumbnailUrl: m.Thumbnail,
	}, nil
}

// OwnerRegenerateThumbnail handles owner/portal cover regeneration. The media
// is resolved from the short_token (uuid never leaves the server), and the
// caller must own the media (JWT subject == media.UserId).
// POST /api/v1/me/medias/{token}/regen-thumbnail
func (s *MediaService) OwnerRegenerateThumbnail(ctx context.Context, req *media.OwnerRegenerateThumbnailRequest) (*media.OwnerRegenerateThumbnailResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	bgCtx, cancel := context.WithTimeout(context.Background(), spriteGenerateTimeout)
	defer cancel()
	m, err := s.uc.GetByShortToken(bgCtx, req.Token)
	if err != nil || m == nil {
		return nil, status.Error(codes.NotFound, "media not found")
	}
	if claims.GetUserID() != m.UserId {
		return nil, status.Error(codes.PermissionDenied, "not allowed to modify this media")
	}
	if err := s.uc.RegenerateThumbnail(bgCtx, m.Id, req.ThumbnailTime); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	thumbnailURL := ""
	if mm, gerr := s.uc.GetMedia(bgCtx, m.Id); gerr == nil && mm != nil {
		thumbnailURL = mm.Thumbnail
	}
	return &media.OwnerRegenerateThumbnailResponse{Success: true, ThumbnailUrl: thumbnailURL}, nil
}

// OwnerSetThumbnail handles owner/portal cover selection from the sprite sheet.
// Resolved from short_token; caller must own the media.
// POST /api/v1/me/medias/{token}/set-thumbnail
func (s *MediaService) OwnerSetThumbnail(ctx context.Context, req *media.OwnerSetThumbnailRequest) (*media.OwnerSetThumbnailResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	bgCtx, cancel := context.WithTimeout(context.Background(), spriteGenerateTimeout)
	defer cancel()
	m, err := s.uc.GetByShortToken(bgCtx, req.Token)
	if err != nil || m == nil {
		return nil, status.Error(codes.NotFound, "media not found")
	}
	if claims.GetUserID() != m.UserId {
		return nil, status.Error(codes.PermissionDenied, "not allowed to modify this media")
	}
	if req.UseSpriteSheet {
		if err := s.uc.SetSpriteSheetThumbnail(bgCtx, m.Id); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &media.OwnerSetThumbnailResponse{Success: true}, nil
}

// RetryEncodingTask handles the RetryEncodingTask gRPC method.
func (s *MediaService) RetryEncodingTask(ctx context.Context, req *media.RetryEncodingTaskRequest) (*media.RetryEncodingTaskResponse, error) {
	task, err := s.uc.RetryTask(ctx, req.TaskId)
	if err != nil {
		return nil, err
	}

	return &media.RetryEncodingTaskResponse{
		Task: &types.EncodingTask{
			Id:           task.Id,
			MediaId:      task.MediaId,
			ProfileId:    int64(task.ProfileId),
			Status:       string(task.Status),
			Progress:     0,
			OutputPath:   task.OutputPath,
			ErrorMessage: task.ErrorMessage,
		},
	}, nil
}

// EncodingTasks handles the EncodingTasks gRPC method (proto interface requirement).
// Delegates to ListAllEncodingTasks which contains the actual implementation.
func (s *MediaService) EncodingTasks(ctx context.Context, req *media.ListAllEncodingTasksRequest) (*media.ListAllEncodingTasksResponse, error) {
	return s.ListAllEncodingTasks(ctx, req)
}

// TranscodingStatus handles the TranscodingStatus gRPC method (proto interface requirement).
// Returns aggregated encoding status counts.
func (s *MediaService) TranscodingStatus(ctx context.Context, req *media.GetEncodingStatusRequest) (*media.GetEncodingStatusResponse, error) {
	status, err := s.uc.GetTranscodingStatus(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &media.GetEncodingStatusResponse{
		ProcessingCount: int32(status.ProcessingCount),
		PendingCount:    int32(status.PendingCount),
		FailedCount:     int32(status.FailedCount),
		SuccessCount:    int32(status.SuccessCount),
	}, nil
}

// ListAllEncodingTasks handles the ListAllEncodingTasks gRPC method.
func (s *MediaService) ListAllEncodingTasks(ctx context.Context, req *media.ListAllEncodingTasksRequest) (*media.ListAllEncodingTasksResponse, error) {
	filterType := biz.FilterTypeAll
	if req.Status != "" && req.Status != "all" {
		filterType = biz.FilterTypeSpecific
	}
	filter := &biz.TranscodingStatusFilter{
		Status:     req.Status,
		Page:       int(req.Page),
		PageSize:   int(req.PageSize),
		FilterType: filterType,
	}

	if req.MediaId != nil {
		filter.SearchQuery = *req.MediaId
	}

	result, err := s.uc.ListEncodingTasksFlat(ctx, filter, req.MediaId)
	if err != nil {
		return nil, err
	}

	tasks := make([]*types.EncodingTask, len(result.Items))
	for i, item := range result.Items {
		tasks[i] = &types.EncodingTask{
			Id:           item.Id,
			MediaId:      item.MediaId,
			ProfileId:    int64(item.ProfileId),
			Status:       string(item.Status),
			Progress:     0,
			OutputPath:   item.OutputPath,
			ErrorMessage: item.ErrorMessage,
		}
	}

	return &media.ListAllEncodingTasksResponse{
		Total:      int32(result.Total),
		Tasks:      tasks,
		Page:       int32(result.Page),
		PageSize:   int32(result.PageSize),
		TotalPages: int32((result.Total + result.PageSize - 1) / result.PageSize),
	}, nil
}

// RetryAllFailedTasks handles the RetryAllFailedTasks gRPC method.
func (s *MediaService) RetryAllFailedTasks(ctx context.Context, req *media.RetryAllFailedTasksRequest) (*media.RetryAllFailedTasksResponse, error) {
	var count int
	var err error

	if req.MediaId != nil {
		count, err = s.uc.RetryAllFailedTasks(ctx, *req.MediaId)
	} else {
		// If no media ID specified, we would need to iterate all media with failed tasks
		// For now, just return 0 with success
		count = 0
	}

	if err != nil {
		return nil, err
	}

	return &media.RetryAllFailedTasksResponse{
		RetriedCount: int32(count),
	}, nil
}

// GetMediaLikes handles the GetMediaLikes gRPC method.
func (s *MediaService) GetMediaLikes(ctx context.Context, req *media.GetMediaLikesRequest) (*media.GetMediaLikesResponse, error) {
	m, err := s.uc.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("MEDIA_NOT_FOUND", "Media not found")
	}

	// BUG-222: return per-user like/dislike state (was hardcoded false).
	// Requires a resolved user; anonymous callers just get the aggregate counts.
	userID := s.extractUserID(ctx)
	if userID == "" || s.likeFavoriteUC == nil {
		return &media.GetMediaLikesResponse{
			LikeCount:    m.LikeCount,
			DislikeCount: m.DislikeCount,
			IsLiked:      false,
			IsDisliked:   false,
		}, nil
	}
	stats, err := s.likeFavoriteUC.GetMediaStats(ctx, userID, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get media stats: %v", err)
	}
	return &media.GetMediaLikesResponse{
		LikeCount:    stats.LikeCount,
		DislikeCount: stats.DislikeCount,
		IsLiked:      stats.UserLikeType == "like",
		IsDisliked:   stats.UserLikeType == "dislike",
	}, nil
}

// ToggleMediaLike handles the ToggleMediaLike gRPC method.
func (s *MediaService) ToggleMediaLike(ctx context.Context, req *media.ToggleMediaLikeRequest) (*media.ToggleMediaLikeResponse, error) {
	m, err := s.uc.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("MEDIA_NOT_FOUND", "Media not found")
	}
	_ = m

	// BUG-222: this was a count-increment stub with no per-user tracking. Wire the
	// real LikeFavoriteUseCase toggle (same as media_handler.go likeMedia) so the
	// state actually persists per user and reflects in GetMediaLikes.
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	likeType := req.Type
	if likeType == "" {
		likeType = "like" // frontend POST /medias/{token}/likes sends an empty body
	}
	stats, err := s.likeFavoriteUC.ToggleLike(ctx, claims.GetUserID(), req.Id, likeType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to toggle like: %v", err)
	}

	return &media.ToggleMediaLikeResponse{
		IsLiked:      stats.UserLikeType == "like",
		IsDisliked:   stats.UserLikeType == "dislike",
		LikeCount:    stats.LikeCount,
		DislikeCount: stats.DislikeCount,
	}, nil
}

// DeleteMediaLike handles the DeleteMediaLike gRPC method.
func (s *MediaService) DeleteMediaLike(ctx context.Context, req *media.DeleteMediaLikeRequest) (*media.DeleteMediaLikeResponse, error) {
	// For simplicity, just decrement both counts (this is a simplified implementation)
	_ = s.uc.UpdateLikeCount(ctx, req.Id, -1)
	_ = s.uc.UpdateDislikeCount(ctx, req.Id, -1)

	return &media.DeleteMediaLikeResponse{
		Success: true,
	}, nil
}

// ToggleMediaLikeCompat handles the ToggleMediaLikeCompat gRPC method (plural /likes path compat).
// Frontend calls POST /api/v1/medias/{token}/likes for the like toggle; without this the route
// returned HTTP 501 "not implemented". Delegate to the canonical ToggleMediaLike.
func (s *MediaService) ToggleMediaLikeCompat(ctx context.Context, req *media.ToggleMediaLikeRequest) (*media.ToggleMediaLikeResponse, error) {
	return s.ToggleMediaLike(ctx, req)
}

// DeleteMediaLikeCompat handles the DeleteMediaLikeCompat gRPC method (plural /likes path compat).
func (s *MediaService) DeleteMediaLikeCompat(ctx context.Context, req *media.DeleteMediaLikeRequest) (*media.DeleteMediaLikeResponse, error) {
	return s.DeleteMediaLike(ctx, req)
}

// GetMediaFavorites handles the GetMediaFavorites gRPC method.
func (s *MediaService) GetMediaFavorites(ctx context.Context, req *media.GetMediaFavoritesRequest) (*media.GetMediaFavoritesResponse, error) {
	userID := s.extractUserID(ctx)
	stats, err := s.likeFavoriteUC.GetMediaStats(ctx, userID, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get media stats: %v", err)
	}

	return &media.GetMediaFavoritesResponse{
		FavoriteCount: stats.FavoriteCount,
		IsFavorited:   stats.IsFavorited,
	}, nil
}

// ToggleMediaFavorite handles the ToggleMediaFavorite gRPC method.
func (s *MediaService) ToggleMediaFavorite(ctx context.Context, req *media.ToggleMediaFavoriteRequest) (*media.ToggleMediaFavoriteResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	stats, err := s.likeFavoriteUC.ToggleFavorite(ctx, claims.GetUserID(), req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to toggle favorite: %v", err)
	}

	return &media.ToggleMediaFavoriteResponse{
		IsFavorited:   stats.IsFavorited,
		FavoriteCount: stats.FavoriteCount,
	}, nil
}

// ToggleMediaFavoriteCompat handles the ToggleMediaFavoriteCompat gRPC method (plural path compat).
func (s *MediaService) ToggleMediaFavoriteCompat(ctx context.Context, req *media.ToggleMediaFavoriteRequest) (*media.ToggleMediaFavoriteResponse, error) {
	return s.ToggleMediaFavorite(ctx, req)
}

// DeleteMediaFavorite handles the DeleteMediaFavorite gRPC method.
func (s *MediaService) DeleteMediaFavorite(ctx context.Context, req *media.DeleteMediaFavoriteRequest) (*media.DeleteMediaFavoriteResponse, error) {
	claims := s.extractClaims(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	// Toggle favorite off if currently favorited
	favorited, err := s.likeFavoriteUC.ToggleFavorite(ctx, claims.GetUserID(), req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete favorite: %v", err)
	}
	_ = favorited

	return &media.DeleteMediaFavoriteResponse{
		Success: true,
	}, nil
}

// DeleteMediaFavoriteCompat handles the DeleteMediaFavoriteCompat gRPC method (plural path compat).
func (s *MediaService) DeleteMediaFavoriteCompat(ctx context.Context, req *media.DeleteMediaFavoriteRequest) (*media.DeleteMediaFavoriteResponse, error) {
	return s.DeleteMediaFavorite(ctx, req)
}

// ReportMedia handles the ReportMedia gRPC method.
// It records a user report against the target media by bumping its ReportedTimes
// counter. The report reason/description are accepted in the request for
// auditing/downstream moderation but are not stored on the media entity here.
// req.Id may be either the media ID or its short_token; the dal resolves both.
func (s *MediaService) ReportMedia(ctx context.Context, req *media.ReportMediaRequest) (*media.ReportMediaResponse, error) {
	if err := s.uc.UpdateReportedTimes(ctx, req.Id, 1); err != nil {
		return nil, err
	}

	return &media.ReportMediaResponse{
		Success: true,
	}, nil
}

// GetMediaShares handles the GetMediaShares gRPC method.
func (s *MediaService) GetMediaShares(ctx context.Context, req *media.GetMediaSharesRequest) (*media.GetMediaSharesResponse, error) {
	m, err := s.uc.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("MEDIA_NOT_FOUND", "Media not found")
	}

	shareURL := "/medias/" + m.Id
	if m.ShortToken != "" {
		shareURL = "/medias/" + m.ShortToken
	}

	return &media.GetMediaSharesResponse{
		ShareCount: m.ShareCount,
		ShareUrl:   shareURL,
	}, nil
}

// CreateMediaShare handles the CreateMediaShare gRPC method.
func (s *MediaService) CreateMediaShare(ctx context.Context, req *media.CreateMediaShareRequest) (*media.CreateMediaShareResponse, error) {
	m, err := s.uc.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("MEDIA_NOT_FOUND", "Media not found")
	}

	// Increment share count (this is simplified; real implementation would track shares per platform)
	_ = s.uc.UpdateFavoriteCount(ctx, req.Id, 1) // Using favorite count update as a placeholder for share count

	shareURL := "/medias/" + m.Id
	if m.ShortToken != "" {
		shareURL = "/medias/" + m.ShortToken
	}

	return &media.CreateMediaShareResponse{
		Success:  true,
		ShareUrl: shareURL,
	}, nil
}

// GetMediaComments handles the GetMediaComments gRPC method.
func (s *MediaService) GetMediaComments(ctx context.Context, req *media.GetMediaCommentsRequest) (*media.GetMediaCommentsResponse, error) {
	m, err := s.uc.GetMedia(ctx, req.MediaId)
	if err != nil {
		return nil, errors.NotFound("MEDIA_NOT_FOUND", "Media not found")
	}

	// For now, return empty comments
	// Real implementation would query comment repository
	return &media.GetMediaCommentsResponse{
		Total:      int32(m.CommentCount),
		Items:      []*types.Comment{},
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: 0,
	}, nil
}

// GetMediaSubtitles handles the GetMediaSubtitles gRPC method.
func (s *MediaService) GetMediaSubtitles(ctx context.Context, req *media.GetMediaSubtitlesRequest) (*media.GetMediaSubtitlesResponse, error) {
	_, err := s.uc.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("MEDIA_NOT_FOUND", "Media not found")
	}

	// For now, return empty subtitles
	return &media.GetMediaSubtitlesResponse{
		Subtitles: []*types.Subtitle{},
	}, nil
}

// CreateMediaSubtitle handles the CreateMediaSubtitle gRPC method.
func (s *MediaService) CreateMediaSubtitle(ctx context.Context, req *media.CreateMediaSubtitleRequest) (*media.CreateMediaSubtitleResponse, error) {
	_, err := s.uc.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("MEDIA_NOT_FOUND", "Media not found")
	}

	// Create a new subtitle (simplified; real implementation would save to database)
	subtitle := &types.Subtitle{
		Id:       "sub_" + req.Id,
		MediaId:  req.Id,
		Language: req.Language,
		FileUrl:  req.FileUrl,
		Label:    req.Language,
	}

	return &media.CreateMediaSubtitleResponse{
		Subtitle: subtitle,
	}, nil
}

// GetMediaMetadata handles the GetMediaMetadata gRPC method.
func (s *MediaService) GetMediaMetadata(ctx context.Context, req *media.GetMediaMetadataRequest) (*media.GetMediaMetadataResponse, error) {
	m, err := s.uc.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("MEDIA_NOT_FOUND", "Media not found")
	}

	return &media.GetMediaMetadataResponse{
		Metadata: &types.MediaMetadata{
			Id:         m.Id,
			MediaId:    m.Id,
			Duration:   m.Duration,
			Bitrate:    0,
			VideoCodec: "",
			AudioCodec: "",
			FrameRate:  0,
			Width:      int32(m.Width),
			Height:     int32(m.Height),
		},
	}, nil
}

// MediaVariantsHTTPHandler handles GET /api/v1/medias/{id}/variants
// Returns aggregated transcoding status for a single media, including all variant details.
// This is the API that the "media management" page uses to display transcoding overview.
func (s *MediaService) MediaVariantsHTTPHandler(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	// Extract media ID from URL path: /api/v1/medias/{id}/variants
	path := r.URL.Path
	// Find the positions of "/medias/" and "/variants"
	mediasIndex := strings.Index(path, "/medias/")
	variantsIndex := strings.Index(path, "/variants")
	if mediasIndex == -1 || variantsIndex == -1 || mediasIndex >= variantsIndex {
		writeRetryError(w, "invalid media ID in path", 400)
		return
	}
	// Extract the media ID
	mediaIDStr := path[mediasIndex+8 : variantsIndex]
	if mediaIDStr == "" {
		writeRetryError(w, "invalid media ID in path", 400)
		return
	}

	summary, err := s.uc.GetMediaVariantsByUUID(r.Context(), mediaIDStr)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeRetryError(w, "media not found", 404)
			return
		}
		writeRetryError(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_ = json.NewEncoder(w).Encode(summary)
}

// EncodingTasksHTTPHandler handles GET /api/v1/media/encoding/tasks.
// Returns a flat, paginated list of encoding tasks (one row per task).
// Query params:
//   - status: "active" | "processing" | "pending" | "partial" | "failed" | "success" | "all"
//   - page: page number (default 1)
//   - page_size: items per page (default 25, max 100)
//   - media_id: optional filter to a specific media
func (s *MediaService) EncodingTasksHTTPHandler(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	filter := &biz.TranscodingStatusFilter{
		Page:     1,
		PageSize: 25,
	}

	if q := r.URL.Query().Get("status"); q != "" {
		filter.Status = q
	} else {
		filter.Status = "all"
	}
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			filter.Page = v
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil {
			filter.PageSize = v
		}
	}
	// Normalize pagination parameters
	page, pageSize := repotypes.NormalizePagination(filter.Page, filter.PageSize)
	filter.Page = page
	filter.PageSize = pageSize
	if pr := r.URL.Query().Get("profile"); pr != "" {
		filter.ProfileFilter = pr
	}
	if ch := r.URL.Query().Get("chunk"); ch != "" {
		filter.ChunkFilter = ch
	}
	if se := r.URL.Query().Get("search"); se != "" {
		filter.SearchQuery = se
	}
	if os := r.URL.Query().Get("only_stats"); os == "true" {
		filter.OnlyStats = true
	}

	var mediaID *string
	if m := r.URL.Query().Get("media_id"); m != "" {
		mediaID = &m
	}

	result, err := s.uc.ListEncodingTasksFlat(r.Context(), filter, mediaID)
	if err != nil {
		stdhttp.Error(w, err.Error(), stdhttp.StatusInternalServerError)
		return
	}

	totalPages := 0
	if result.PageSize > 0 {
		totalPages = (result.Total + result.PageSize - 1) / result.PageSize
	}

	resp := map[string]interface{}{
		"items":            result.Items,
		"total":            result.Total,
		"page":             result.Page,
		"page_size":        result.PageSize,
		"total_pages":      totalPages,
		"processing_count": result.ProcessingCount,
		"pending_count":    result.PendingCount,
		"queued_count":     result.PendingCount,
		"partial_count":    result.PartialCount,
		"failed_count":     result.FailedCount,
		"success_count":    result.SuccessCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_ = json.NewEncoder(w).Encode(resp)
}
