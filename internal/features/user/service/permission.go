package service

import (
	"context"

	mediav1 "origadmin/application/origstudio/api/gen/v1/media"
	typesv1 "origadmin/application/origstudio/api/gen/v1/types"

	"github.com/origadmin/runtime/log"
)

type PermissionService struct {
	mediav1.UnimplementedPermissionServiceServer
	log *log.Helper
}

func NewPermissionService(logger log.Logger) *PermissionService {
	return &PermissionService{
		log: log.NewHelper(log.With(logger, "module", "user.service.permission")),
	}
}

func (s *PermissionService) ListPermissions(
	ctx context.Context,
	req *mediav1.ListPermissionsRequest,
) (*mediav1.ListPermissionsResponse, error) {
	return &mediav1.ListPermissionsResponse{
		Permissions: []string{},
	}, nil
}

func (s *PermissionService) ListPermissionGroups(
	ctx context.Context,
	req *mediav1.ListPermissionGroupsRequest,
) (*mediav1.ListPermissionGroupsResponse, error) {
	return &mediav1.ListPermissionGroupsResponse{
		Groups: []*mediav1.PermissionGroupInfo{},
		Total:  0,
	}, nil
}

func (s *PermissionService) GetPermissionGroup(
	ctx context.Context,
	req *mediav1.GetPermissionGroupRequest,
) (*mediav1.GetPermissionGroupResponse, error) {
	return &mediav1.GetPermissionGroupResponse{Group: &mediav1.PermissionGroupInfo{}}, nil
}

func (s *PermissionService) CreatePermissionGroup(
	ctx context.Context,
	req *mediav1.CreatePermissionGroupRequest,
) (*mediav1.CreatePermissionGroupResponse, error) {
	return &mediav1.CreatePermissionGroupResponse{Group: &mediav1.PermissionGroupInfo{}}, nil
}

func (s *PermissionService) UpdatePermissionGroup(
	ctx context.Context,
	req *mediav1.UpdatePermissionGroupRequest,
) (*mediav1.UpdatePermissionGroupResponse, error) {
	return &mediav1.UpdatePermissionGroupResponse{Group: &mediav1.PermissionGroupInfo{}}, nil
}

func (s *PermissionService) DeletePermissionGroup(
	ctx context.Context,
	req *mediav1.DeletePermissionGroupRequest,
) (*mediav1.DeletePermissionGroupResponse, error) {
	return &mediav1.DeletePermissionGroupResponse{}, nil
}

func (s *PermissionService) TogglePermissionGroup(
	ctx context.Context,
	req *mediav1.TogglePermissionGroupRequest,
) (*mediav1.TogglePermissionGroupResponse, error) {
	return &mediav1.TogglePermissionGroupResponse{Success: true}, nil
}

func (s *PermissionService) ListPermissionGroupMembers(
	ctx context.Context,
	req *mediav1.ListPermissionGroupMembersRequest,
) (*mediav1.ListPermissionGroupMembersResponse, error) {
	return &mediav1.ListPermissionGroupMembersResponse{
		Members: []*typesv1.User{},
		Total:   0,
	}, nil
}

func (s *PermissionService) AddPermissionGroupMember(
	ctx context.Context,
	req *mediav1.AddPermissionGroupMemberRequest,
) (*mediav1.AddPermissionGroupMemberResponse, error) {
	return &mediav1.AddPermissionGroupMemberResponse{}, nil
}

func (s *PermissionService) RemovePermissionGroupMember(
	ctx context.Context,
	req *mediav1.RemovePermissionGroupMemberRequest,
) (*mediav1.RemovePermissionGroupMemberResponse, error) {
	return &mediav1.RemovePermissionGroupMemberResponse{}, nil
}
