package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/metadata"

	uploadv1 "origadmin/application/origstudio/api/gen/v1/upload"
	"origadmin/application/origstudio/internal/data/enums"
	"origadmin/application/origstudio/internal/features/media/biz"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/pkg/hashtag"
	"origadmin/application/origstudio/internal/domain/types"
)

// UploadServiceV1 adapts biz.UploadUseCase to the uploadv1.UploadServiceServer interface.
// This is the version registered on the media gRPC server so that the gateway's
// gRPC-to-HTTP bridge (uploadv1.NewUploadServiceGRPC2HTTP) can reach it.
type UploadServiceV1 struct {
	uploadv1.UnimplementedUploadServiceServer
	uc     *biz.UploadUseCase
	jwtMgr *auth.Manager
	log    *log.Helper
}

func NewUploadServiceV1(uc *biz.UploadUseCase, jwtMgr *auth.Manager, logger log.Logger) *UploadServiceV1 {
	return &UploadServiceV1{
		uc:     uc,
		jwtMgr: jwtMgr,
		log:    log.NewHelper(log.With(logger, "module", "service/upload_v1")),
	}
}

func (s *UploadServiceV1) extractUserID(ctx context.Context) *string {
	if id, ok := ctx.Value("user_id").(string); ok && id != "" {
		return &id
	}
	if id, ok := ctx.Value("user_id").(int64); ok && id != 0 {
		idStr := strconv.FormatInt(id, 10)
		return &idStr
	}
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
		claims, err := s.jwtMgr.Parse(token)
		if err != nil {
			s.log.Warnf("failed to parse JWT token from gRPC metadata: %v", err)
			continue
		}
		userID := claims.GetUserID()
		if userID != "" {
			return &userID
		}
	}
	return nil
}

func (s *UploadServiceV1) InitiateMultipartUpload(ctx context.Context, req *uploadv1.InitiateMultipartUploadRequest) (*uploadv1.InitiateMultipartUploadResponse, error) {
	userID := s.extractUserID(ctx)

	var categoryID *int64
	if req.CategoryId != 0 {
		categoryID = &req.CategoryId
	}

	var channelID *string
	if req.ChannelId != "" {
		channelID = &req.ChannelId
	}

	session, err := s.uc.InitiateMultipartUpload(
		ctx,
		req.Filename,
		req.FileSize,
		req.ContentType,
		req.Title,
		req.Description,
		categoryID,
		channelID,
		req.Tags,
		req.Thumbnail,
		userID,
	)
	if err != nil {
		s.log.Errorf("failed to initiate multipart upload: %v", err)
		return nil, err
	}

	return &uploadv1.InitiateMultipartUploadResponse{
		UploadId:   session.UploadID,
		TotalParts: int32(session.TotalParts),
		ChunkSize:  int32(session.ChunkSize),
	}, nil
}

func (s *UploadServiceV1) UploadPart(ctx context.Context, req *uploadv1.UploadPartRequest) (*uploadv1.UploadPartResponse, error) {
	etag, err := s.uc.UploadPart(ctx, req.UploadId, int(req.PartNumber), bytes.NewReader(req.Data), int64(len(req.Data)))
	if err != nil {
		s.log.Errorf("failed to upload part %d for upload %s: %v", req.PartNumber, req.UploadId, err)
		return nil, err
	}

	return &uploadv1.UploadPartResponse{
		Etag: etag,
		Size: int64(len(req.Data)),
	}, nil
}

func (s *UploadServiceV1) ListParts(ctx context.Context, req *uploadv1.ListPartsRequest) (*uploadv1.ListPartsResponse, error) {
	session, err := s.uc.GetSession(ctx, req.UploadId)
	if err != nil {
		s.log.Errorf("failed to get session for upload %s: %v", req.UploadId, err)
		return nil, err
	}

	parts := make([]*uploadv1.PartInfo, 0, len(session.Parts))
	for partNum, etag := range session.Parts {
		parts = append(parts, &uploadv1.PartInfo{
			PartNumber: int32(partNum),
			Etag:       etag,
			Size:       0,
		})
	}

	return &uploadv1.ListPartsResponse{
		Parts:        parts,
		TotalParts:   int32(session.TotalParts),
		UploadedSize: session.UploadedSize,
		TotalSize:    session.FileSize,
		Status:       string(session.Status),
	}, nil
}

func (s *UploadServiceV1) CompleteMultipartUpload(ctx context.Context, req *uploadv1.CompleteMultipartUploadRequest) (*uploadv1.CompleteMultipartUploadResponse, error) {
	session, err := s.uc.GetSession(ctx, req.UploadId)
	if err != nil {
		s.log.Errorf("failed to get session for upload %s: %v", req.UploadId, err)
		return nil, err
	}

	title := session.Title
	description := session.Description
	tags := session.Tags

	parsedHashtags := hashtag.ParseHashtags(title + " " + description)
	if len(parsedHashtags) > 0 {
		tags = mergeUploadTags(tags, parsedHashtags)
	}

	var categoryID *int64
	if session.CategoryID != nil {
		categoryID = session.CategoryID
	}

	media, err := s.uc.CompleteMultipartUpload(ctx, req.UploadId, req.Sha256,
		title, description, categoryID, session.ChannelID, tags, session.Thumbnail)
	if err != nil {
		s.log.Errorf("failed to complete multipart upload %s: %v", req.UploadId, err)
		return nil, err
	}

	return &uploadv1.CompleteMultipartUploadResponse{
		Media: media,
	}, nil
}

func (s *UploadServiceV1) AbortMultipartUpload(ctx context.Context, req *uploadv1.AbortMultipartUploadRequest) (*uploadv1.AbortMultipartUploadResponse, error) {
	err := s.uc.AbortMultipartUpload(ctx, req.UploadId)
	if err != nil {
		s.log.Errorf("failed to abort multipart upload %s: %v", req.UploadId, err)
		return nil, err
	}
	return &uploadv1.AbortMultipartUploadResponse{}, nil
}

func (s *UploadServiceV1) UpdateMetadata(ctx context.Context, req *uploadv1.UpdateMetadataRequest) (*uploadv1.UpdateMetadataResponse, error) {
	var categoryID *int64
	if req.CategoryId != nil {
		categoryID = req.CategoryId
	}
	var channelID *string
	if req.ChannelId != nil && *req.ChannelId != "" {
		channelID = req.ChannelId
	}
	var thumbnail string
	if req.Thumbnail != nil {
		thumbnail = *req.Thumbnail
	}
	var title, description string
	if req.Title != nil {
		title = *req.Title
	}
	if req.Description != nil {
		description = *req.Description
	}
	err := s.uc.UpdateUploadMetadata(ctx, req.UploadId, title, description, categoryID, channelID, req.Tags, thumbnail)
	if err != nil {
		s.log.Errorf("failed to update metadata for upload %s: %v", req.UploadId, err)
		return nil, err
	}
	return &uploadv1.UpdateMetadataResponse{
		UploadId: req.UploadId,
		Status:   "updated",
	}, nil
}

func (s *UploadServiceV1) GetUploadSession(ctx context.Context, req *uploadv1.GetUploadSessionRequest) (*uploadv1.GetUploadSessionResponse, error) {
	session, err := s.uc.GetSession(ctx, req.UploadId)
	if err != nil {
		return nil, err
	}

	parts := make([]*uploadv1.PartInfo, 0, len(session.Parts))
	for partNum, etag := range session.Parts {
		parts = append(parts, &uploadv1.PartInfo{
			PartNumber: int32(partNum),
			Etag:       etag,
		})
	}

	return &uploadv1.GetUploadSessionResponse{
		UploadId:     session.UploadID,
		Filename:     session.Filename,
		FileSize:     session.FileSize,
		ContentType:  session.ContentType,
		TotalParts:   int32(session.TotalParts),
		ChunkSize:    int32(session.ChunkSize),
		UploadedSize: session.UploadedSize,
		Status:       string(session.Status),
		Parts:        parts,
		CreateTime:   session.CreateTime.Format(time.RFC3339),
		ExpiresAt:    session.ExpiresAt.Format(time.RFC3339),
	}, nil
}

func (s *UploadServiceV1) ListUploadSessions(ctx context.Context, req *uploadv1.ListUploadSessionsRequest) (*uploadv1.ListUploadSessionsResponse, error) {
	var userID string
	if uid := s.extractUserID(ctx); uid != nil {
		userID = *uid
	}

	page, pageSize := types.NormalizePagination(int(req.GetPage()), int(req.GetPageSize()))

	sessions, total, err := s.uc.ListSessions(ctx, userID, enums.UploadStatus(req.GetStatus()), page, pageSize)
	if err != nil {
		return nil, err
	}

	pbSessions := make([]*uploadv1.GetUploadSessionResponse, len(sessions))
	for i, session := range sessions {
		pbSessions[i] = &uploadv1.GetUploadSessionResponse{
			UploadId:     session.UploadID,
			Filename:     session.Filename,
			FileSize:     session.FileSize,
			ContentType:  session.ContentType,
			TotalParts:   int32(session.TotalParts),
			ChunkSize:    int32(session.ChunkSize),
			UploadedSize: session.UploadedSize,
			Status:       string(session.Status),
			CreateTime:   session.CreateTime.Format(time.RFC3339),
			ExpiresAt:    session.ExpiresAt.Format(time.RFC3339),
		}
	}

	return &uploadv1.ListUploadSessionsResponse{
		Sessions: pbSessions,
		Total:    int32(total),
	}, nil
}

func (s *UploadServiceV1) SimpleUpload(ctx context.Context, req *uploadv1.SimpleUploadRequest) (*uploadv1.SimpleUploadResponse, error) {
	userID := s.extractUserID(ctx)

	var categoryID *int64
	if req.CategoryId != 0 {
		categoryID = &req.CategoryId
	}
	var channelID *string
	if req.ChannelId != "" {
		channelID = &req.ChannelId
	}

	// Calculate SHA256
	hash := sha256.Sum256(req.Data)
	fileSha256 := hex.EncodeToString(hash[:])

	session, err := s.uc.InitiateMultipartUpload(
		ctx,
		req.Filename,
		int64(len(req.Data)),
		req.ContentType,
		req.Title,
		req.Description,
		categoryID,
		channelID,
		req.Tags,
		req.Thumbnail,
		userID,
	)
	if err != nil {
		return nil, err
	}

	_, err = s.uc.UploadPart(ctx, session.UploadID, 1, bytes.NewReader(req.Data), int64(len(req.Data)))
	if err != nil {
		_ = s.uc.AbortMultipartUpload(ctx, session.UploadID)
		return nil, err
	}

	media, err := s.uc.CompleteMultipartUpload(ctx, session.UploadID, fileSha256,
		req.Title, req.Description, categoryID, channelID, req.Tags, req.Thumbnail)
	if err != nil {
		_ = s.uc.AbortMultipartUpload(ctx, session.UploadID)
		return nil, err
	}

	return &uploadv1.SimpleUploadResponse{
		Media:     media,
		UploadUrl: "",
	}, nil
}
