/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "origadmin/application/origcms/api/gen/v1/upload"
	"origadmin/application/origcms/internal/svc-media/biz"
)

// UploadService is a service for handling file uploads.
type UploadService struct {
	pb.UnimplementedUploadServiceServer
	uc  *biz.UploadUseCase
	log *log.Helper
}

// NewUploadService creates a new UploadService.
func NewUploadService(uc *biz.UploadUseCase, logger log.Logger) *UploadService {
	return &UploadService{
		uc:  uc,
		log: log.NewHelper(log.With(logger, "module", "service/upload")),
	}
}

// InitiateMultipartUpload implements pb.UploadServiceServer.
func (s *UploadService) InitiateMultipartUpload(ctx context.Context, req *pb.InitiateMultipartUploadRequest) (*pb.InitiateMultipartUploadResponse, error) {
	// Extract user ID from context (assuming it's set by middleware)
	var userID *int64
	if id, ok := ctx.Value("user_id").(int64); ok {
		userID = &id
	}

	session, err := s.uc.InitiateMultipartUpload(
		ctx,
		req.Filename,
		req.FileSize,
		req.ContentType,
		req.Title,
		req.Description,
		req.CategoryId,
		req.Tags,
		userID,
	)
	if err != nil {
		s.log.Errorf("failed to initiate multipart upload: %v", err)
		return nil, err
	}

	return &pb.InitiateMultipartUploadResponse{
		UploadId:   session.UploadID,
		TotalParts: int32(session.TotalParts),
		ChunkSize:  int32(session.ChunkSize),
	}, nil
}

// UploadPart implements pb.UploadServiceServer.
func (s *UploadService) UploadPart(ctx context.Context, req *pb.UploadPartRequest) (*pb.UploadPartResponse, error) {
	etag, err := s.uc.UploadPart(ctx, req.UploadId, int(req.PartNumber), req.Data)
	if err != nil {
		s.log.Errorf("failed to upload part %d for upload %s: %v", req.PartNumber, req.UploadId, err)
		return nil, err
	}

	return &pb.UploadPartResponse{
		Etag: etag,
		Size: int64(len(req.Data)),
	}, nil
}

// ListParts implements pb.UploadServiceServer.
func (s *UploadService) ListParts(ctx context.Context, req *pb.ListPartsRequest) (*pb.ListPartsResponse, error) {
	session, err := s.uc.GetSession(ctx, req.UploadId)
	if err != nil {
		s.log.Errorf("failed to get session for upload %s: %v", req.UploadId, err)
		return nil, err
	}

	parts := make([]*pb.PartInfo, 0, len(session.Parts))
	for partNum, etag := range session.Parts {
		parts = append(parts, &pb.PartInfo{
			PartNumber: int32(partNum),
			Etag:       etag,
			Size:       0, // Placeholder
		})
	}

	return &pb.ListPartsResponse{
		Parts:        parts,
		TotalParts:   int32(session.TotalParts),
		UploadedSize: session.UploadedSize,
		TotalSize:    session.FileSize,
		Status:       session.Status,
	}, nil
}

// CompleteMultipartUpload implements pb.UploadServiceServer.
func (s *UploadService) CompleteMultipartUpload(ctx context.Context, req *pb.CompleteMultipartUploadRequest) (*pb.CompleteMultipartUploadResponse, error) {
	media, err := s.uc.CompleteMultipartUpload(ctx, req.UploadId, req.Sha256)
	if err != nil {
		s.log.Errorf("failed to complete multipart upload %s: %v", req.UploadId, err)
		return nil, err
	}

	return &pb.CompleteMultipartUploadResponse{
		Media: media,
	}, nil
}

// AbortMultipartUpload implements pb.UploadServiceServer.
func (s *UploadService) AbortMultipartUpload(ctx context.Context, req *pb.AbortMultipartUploadRequest) (*pb.AbortMultipartUploadResponse, error) {
	err := s.uc.AbortMultipartUpload(ctx, req.UploadId)
	if err != nil {
		s.log.Errorf("failed to abort multipart upload %s: %v", req.UploadId, err)
		return nil, err
	}
	return &pb.AbortMultipartUploadResponse{}, nil
}

// UploadFile implements pb.UploadServiceServer.
func (s *UploadService) UploadFile(ctx context.Context, req *pb.UploadFileRequest) (*pb.UploadFileResponse, error) {
	var userID *int64
	if id, ok := ctx.Value("user_id").(int64); ok {
		userID = &id
	}

	session, err := s.uc.InitiateMultipartUpload(
		ctx,
		req.Filename,
		int64(len(req.Data)),
		req.ContentType,
		req.Title,
		req.Description,
		req.CategoryId,
		req.Tags,
		userID,
	)
	if err != nil {
		return nil, err
	}

	_, err = s.uc.UploadPart(ctx, session.UploadID, 1, req.Data)
	if err != nil {
		_ = s.uc.AbortMultipartUpload(ctx, session.UploadID)
		return nil, err
	}

	media, err := s.uc.CompleteMultipartUpload(ctx, session.UploadID, "")
	if err != nil {
		_ = s.uc.AbortMultipartUpload(ctx, session.UploadID)
		return nil, err
	}

	return &pb.UploadFileResponse{
		Media: media,
	}, nil
}

// GetUploadSession implements pb.UploadServiceServer.
func (s *UploadService) GetUploadSession(ctx context.Context, req *pb.GetUploadSessionRequest) (*pb.GetUploadSessionResponse, error) {
	session, err := s.uc.GetSession(ctx, req.UploadId)
	if err != nil {
		return nil, err
	}

	parts := make([]*pb.PartInfo, 0, len(session.Parts))
	for partNum, etag := range session.Parts {
		parts = append(parts, &pb.PartInfo{
			PartNumber: int32(partNum),
			Etag:       etag,
		})
	}

	return &pb.GetUploadSessionResponse{
		UploadId:     session.UploadID,
		Filename:     session.Filename,
		FileSize:     session.FileSize,
		ContentType:  session.ContentType,
		TotalParts:   int32(session.TotalParts),
		ChunkSize:    int32(session.ChunkSize),
		UploadedSize: session.UploadedSize,
		Status:       session.Status,
		Parts:        parts,
		CreatedAt:    session.CreatedAt.Format(time.RFC3339),
		ExpiresAt:    session.ExpiresAt.Format(time.RFC3339),
	}, nil
}

// ListUploadSessions implements pb.UploadServiceServer.
func (s *UploadService) ListUploadSessions(ctx context.Context, req *pb.ListUploadSessionsRequest) (*pb.ListUploadSessionsResponse, error) {
	var userID int64
	if id, ok := ctx.Value("user_id").(int64); ok {
		userID = id
	}

	sessions, total, err := s.uc.ListSessions(ctx, userID, req.GetStatus(), int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}

	pbSessions := make([]*pb.GetUploadSessionResponse, len(sessions))
	for i, session := range sessions {
		pbSessions[i] = &pb.GetUploadSessionResponse{
			UploadId:     session.UploadID,
			Filename:     session.Filename,
			FileSize:     session.FileSize,
			ContentType:  session.ContentType,
			TotalParts:   int32(session.TotalParts),
			ChunkSize:    int32(session.ChunkSize),
			UploadedSize: session.UploadedSize,
			Status:       session.Status,
			CreatedAt:    session.CreatedAt.Format(time.RFC3339),
			ExpiresAt:    session.ExpiresAt.Format(time.RFC3339),
		}
	}

	return &pb.ListUploadSessionsResponse{
		Sessions: pbSessions,
		Total:    int32(total),
	}, nil
}
