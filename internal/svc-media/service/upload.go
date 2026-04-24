package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/metadata"

	pb "origadmin/application/origcms/api/gen/v1/upload"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/data/enums"
	"origadmin/application/origcms/internal/svc-media/biz"
)

type UploadService struct {
	pb.UnimplementedUploadServiceServer
	uc     *biz.UploadUseCase
	jwtMgr *auth.Manager
	log    *log.Helper
}

func NewUploadService(uc *biz.UploadUseCase, jwtMgr *auth.Manager, logger log.Logger) *UploadService {
	return &UploadService{
		uc:     uc,
		jwtMgr: jwtMgr,
		log:    log.NewHelper(log.With(logger, "module", "service/upload")),
	}
}

func (s *UploadService) extractUserID(ctx context.Context) *string {
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

func (s *UploadService) InitiateMultipartUpload(ctx context.Context, req *pb.InitiateMultipartUploadRequest) (*pb.InitiateMultipartUploadResponse, error) {
	userID := s.extractUserID(ctx)

	var categoryID *string
	if req.CategoryId != 0 {
		idStr := strconv.FormatInt(req.CategoryId, 10)
		categoryID = &idStr
	}

	session, err := s.uc.InitiateMultipartUpload(
		ctx,
		req.Filename,
		req.FileSize,
		req.ContentType,
		req.Title,
		req.Description,
		categoryID,
		req.Tags,
		"",
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
			Size:       0,
		})
	}

	return &pb.ListPartsResponse{
		Parts:        parts,
		TotalParts:   int32(session.TotalParts),
		UploadedSize: session.UploadedSize,
		TotalSize:    session.FileSize,
		Status:       string(session.Status),
	}, nil
}

func (s *UploadService) CompleteMultipartUpload(ctx context.Context, req *pb.CompleteMultipartUploadRequest) (*pb.CompleteMultipartUploadResponse, error) {
	media, err := s.uc.CompleteMultipartUpload(ctx, req.UploadId, req.Sha256,
		"", "", nil, nil, "")
	if err != nil {
		s.log.Errorf("failed to complete multipart upload %s: %v", req.UploadId, err)
		return nil, err
	}

	return &pb.CompleteMultipartUploadResponse{
		Media: media,
	}, nil
}

func (s *UploadService) AbortMultipartUpload(ctx context.Context, req *pb.AbortMultipartUploadRequest) (*pb.AbortMultipartUploadResponse, error) {
	err := s.uc.AbortMultipartUpload(ctx, req.UploadId)
	if err != nil {
		s.log.Errorf("failed to abort multipart upload %s: %v", req.UploadId, err)
		return nil, err
	}
	return &pb.AbortMultipartUploadResponse{}, nil
}

func (s *UploadService) UploadFile(ctx context.Context, req *pb.UploadFileRequest) (*pb.UploadFileResponse, error) {
	userID := s.extractUserID(ctx)

	var categoryID *string
	if req.CategoryId != 0 {
		idStr := strconv.FormatInt(req.CategoryId, 10)
		categoryID = &idStr
	}

	session, err := s.uc.InitiateMultipartUpload(
		ctx,
		req.Filename,
		int64(len(req.Data)),
		req.ContentType,
		req.Title,
		req.Description,
		categoryID,
		req.Tags,
		"",
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

	media, err := s.uc.CompleteMultipartUpload(ctx, session.UploadID, "",
		req.Title, req.Description, categoryID, req.Tags, "")
	if err != nil {
		_ = s.uc.AbortMultipartUpload(ctx, session.UploadID)
		return nil, err
	}

	return &pb.UploadFileResponse{
		Media: media,
	}, nil
}

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
		Status:       string(session.Status),
		Parts:        parts,
		CreatedAt:    session.CreatedAt.Format(time.RFC3339),
		ExpiresAt:    session.ExpiresAt.Format(time.RFC3339),
	}, nil
}

func (s *UploadService) ListUploadSessions(ctx context.Context, req *pb.ListUploadSessionsRequest) (*pb.ListUploadSessionsResponse, error) {
	var userID string
	if uid := s.extractUserID(ctx); uid != nil {
		userID = *uid
	}

	sessions, total, err := s.uc.ListSessions(ctx, userID, enums.UploadStatus(req.GetStatus()), int(req.Page), int(req.PageSize))
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
			Status:       string(session.Status),
			CreatedAt:    session.CreatedAt.Format(time.RFC3339),
			ExpiresAt:    session.ExpiresAt.Format(time.RFC3339),
		}
	}

	return &pb.ListUploadSessionsResponse{
		Sessions: pbSessions,
		Total:    int32(total),
	}, nil
}
