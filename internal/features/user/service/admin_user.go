package service

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mediav1 "origadmin/application/origstudio/api/gen/v1/media"
	typesv1 "origadmin/application/origstudio/api/gen/v1/types"
	userbiz "origadmin/application/origstudio/internal/features/user/biz"
	userdto "origadmin/application/origstudio/internal/features/user/dto"
	repotypes "origadmin/application/origstudio/internal/domain/types"

	"github.com/origadmin/runtime/log"
)

type AdminUserService struct {
	mediav1.UnimplementedAdminUserServiceServer
	uc  *userbiz.UserUseCase
	log *log.Helper
}

func NewAdminUserService(uc *userbiz.UserUseCase, logger log.Logger) *AdminUserService {
	return &AdminUserService{
		uc:  uc,
		log: log.NewHelper(log.With(logger, "module", "user.service.admin")),
	}
}

func (s *AdminUserService) ListAdminUsers(
	ctx context.Context,
	req *mediav1.ListAdminUsersRequest,
) (*mediav1.ListAdminUsersResponse, error) {
	page := req.GetPage()
	pageSize := req.GetPageSize()
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	opts := &userdto.UserQueryOption{
		QueryOption: repotypes.QueryOption{
			Page:     page,
			PageSize: pageSize,
			Keyword:  req.GetKeyword(),
		},
	}

	if role := req.GetRole(); role != "" && role != "all" {
		opts.Role = role
	}

	if statusStr := req.GetStatus(); statusStr != "" && statusStr != "all" {
		statusMap := map[string]int32{
			"pending":   1,
			"active":    2,
			"inactive":  3,
			"suspended": 4,
			"rejected":  5,
		}
		if s, ok := statusMap[statusStr]; ok {
			opts.Status = &s
		}
	}

	users, total, err := s.uc.ListUsers(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &mediav1.ListAdminUsersResponse{
		Items:    users,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *AdminUserService) GetAdminUser(
	ctx context.Context,
	req *mediav1.GetAdminUserRequest,
) (*mediav1.GetAdminUserResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	userInfo, err := s.uc.GetUser(ctx, req.GetId())
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, err
	}

	return &mediav1.GetAdminUserResponse{User: userInfo}, nil
}

func (s *AdminUserService) CreateAdminUser(
	ctx context.Context,
	req *mediav1.CreateAdminUserRequest,
) (*mediav1.CreateAdminUserResponse, error) {
	if req.GetUsername() == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	password := req.GetPassword()
	if password == "" {
		password = "admin123456"
	}

	if len(password) < 6 {
		return nil, status.Error(codes.InvalidArgument, "password must be at least 6 characters")
	}

	existing, _ := s.uc.GetUserByUsername(ctx, req.GetUsername())
	if existing != nil {
		return nil, status.Error(codes.AlreadyExists, "username already exists")
	}

	existingEmail, _ := s.uc.GetUserByEmail(ctx, req.GetEmail())
	if existingEmail != nil {
		return nil, status.Error(codes.AlreadyExists, "email already exists")
	}

	hashedPassword, err := s.uc.HashPassword(password)
	if err != nil {
		s.log.Errorf("Failed to hash password: %v", err)
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	userInfo := &typesv1.User{
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Nickname: req.GetNickname(),
		Status:   typesv1.UserStatus_USER_STATUS_ACTIVE,
	}

	roleIds := req.GetRoleIds()
	var role string
	if len(roleIds) > 0 {
		role = roleIds[0]
	} else {
		role = "user"
	}

	created, err := s.uc.CreateUser(ctx, userInfo, hashedPassword)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		s.log.Errorf("Failed to create user: %v", err)
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	if role != "" && role != "user" {
		_ = s.uc.SetUserRole(ctx, created.Id, role)
	}

	return &mediav1.CreateAdminUserResponse{User: created}, nil
}

func (s *AdminUserService) UpdateAdminUser(
	ctx context.Context,
	req *mediav1.UpdateAdminUserRequest,
) (*mediav1.UpdateAdminUserResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	existing, err := s.uc.GetUser(ctx, req.GetId())
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, err
	}

	if req.GetNickname() != "" {
		existing.Nickname = req.GetNickname()
	}
	if req.GetEmail() != "" {
		existing.Email = req.GetEmail()
	}
	if req.GetPhone() != "" {
		existing.Phone = req.GetPhone()
	}

	updated, err := s.uc.UpdateUser(ctx, existing)
	if err != nil {
		return nil, err
	}

	roleIds := req.GetRoleIds()
	if len(roleIds) > 0 {
		_ = s.uc.SetUserRole(ctx, req.GetId(), roleIds[0])
	}

	return &mediav1.UpdateAdminUserResponse{User: updated}, nil
}

func (s *AdminUserService) DeleteAdminUser(
	ctx context.Context,
	req *mediav1.DeleteAdminUserRequest,
) (*mediav1.DeleteAdminUserResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	err := s.uc.DeleteUser(ctx, req.GetId())
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, err
	}

	return &mediav1.DeleteAdminUserResponse{}, nil
}

func (s *AdminUserService) UpdateAdminUserStatus(
	ctx context.Context,
	req *mediav1.UpdateAdminUserStatusRequest,
) (*mediav1.UpdateAdminUserStatusResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	err := s.uc.UpdateUserStatus(ctx, req.GetId(), int8(req.GetStatus()))
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, err
	}

	return &mediav1.UpdateAdminUserStatusResponse{Success: true}, nil
}

func (s *AdminUserService) UpdateAdminUserRole(
	ctx context.Context,
	req *mediav1.UpdateAdminUserRoleRequest,
) (*mediav1.UpdateAdminUserRoleResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	roleIds := req.GetRoleIds()
	if len(roleIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "role ids are required")
	}

	err := s.uc.SetUserRole(ctx, req.GetId(), roleIds[0])
	if err != nil {
		return nil, err
	}

	return &mediav1.UpdateAdminUserRoleResponse{Success: true}, nil
}

func (s *AdminUserService) GetAdminUserPermissions(
	ctx context.Context,
	req *mediav1.GetAdminUserPermissionsRequest,
) (*mediav1.GetAdminUserPermissionsResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	userInfo, err := s.uc.GetUser(ctx, req.GetId())
	if err != nil {
		if repotypes.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, err
	}

	permissions := []string{}
	roles := []string{}
	if userInfo.Role != "" {
		roles = append(roles, userInfo.Role)
	}

	return &mediav1.GetAdminUserPermissionsResponse{
		Permissions: permissions,
		Roles:       roles,
	}, nil
}
